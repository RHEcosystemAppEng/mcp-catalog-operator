/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mcpv1 "github.com/dmartinol/mcp-registry-operator/api/v1"
)

// McpRegistryReconciler reconciles a McpRegistry object
type McpRegistryReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpregistries,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpregistries/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpregistries/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the McpRegistry object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *McpRegistryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var mcpRegistry mcpv1.McpRegistry
	if err := r.Get(ctx, req.NamespacedName, &mcpRegistry); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	registryName := mcpRegistry.Name
	saName := mcpRegistry.Name

	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: mcpRegistry.Namespace,
		},
	}

	if err := r.Create(ctx, &sa); err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("ServiceAccount %q already exists\n", saName)
		} else {
			return ctrl.Result{}, fmt.Errorf("Error creating ServiceAccount: %w", err)
		}
	} else {
		fmt.Printf("ServiceAccount %q created in namespace %q\n", saName, mcpRegistry.Namespace)
	}
	if err := controllerutil.SetControllerReference(&mcpRegistry, &sa, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner reference on ServiceAccount: %w", err)
	}

	// TODO: we really need these? Use K8S API Job instead and remove this code
	for _, role := range []string{
		"mcpregistry-admin-role",
		"mcpserver-admin-role",
		"mcpcertifiedserver-admin-role",
		"mcpserverpool-admin-role",
		"pipeline-as-code-controller-clusterrole"} {
		crbName := fmt.Sprintf("%s-is-%s", saName, role)
		crb := rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      crbName,
				Namespace: mcpRegistry.Namespace,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      saName,
					Namespace: mcpRegistry.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.GroupName,
				Kind:     "ClusterRole",
				Name:     role,
			},
		}
		if err := r.Create(ctx, &crb); err != nil {
			if apierrors.IsAlreadyExists(err) {
				fmt.Printf("ClusterRoleBinding %q already exists\n", crbName)
			} else {
				return ctrl.Result{}, fmt.Errorf("failed creating ClusterRoleBinding: %w", err)
			}
		} else {
			fmt.Printf("ClusterRoleBinding %q created in namespace %q\n", crbName, mcpRegistry.Namespace)
		}
		if err := controllerutil.SetControllerReference(&mcpRegistry, &crb, r.Scheme); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set owner reference on ClusterRoleBinding: %w", err)
		}
	}

	deployName := registryName
	labels := map[string]string{
		"app": registryName,
	}
	replicas := int32(1)

	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployName,
			Namespace: mcpRegistry.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: saName,
					Containers: []corev1.Container{
						{
							Name:            "mcp-registry",
							Image:           "quay.io/ecosystem-appeng/mcp-registry:amd64-0.1",
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8000,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "MCP_REGISTRY_NAME",
									Value: registryName,
								},
								{
									Name:  "MCP_SERVERPOOL_NAME",
									Value: "foo",
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(&mcpRegistry, &deploy, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner reference on deployment: %w", err)
	}
	if err := r.Create(ctx, &deploy); err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("deployment %q already exists\n", deployName)
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to create deployment: %w", err)
		}
	}

	serviceName := mcpRegistry.Name + "-svc"
	serviceNamespace := mcpRegistry.Namespace
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": mcpRegistry.Name},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   "TCP",
					Port:       8000,
					TargetPort: intstr.FromInt(8000),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	if err := controllerutil.SetControllerReference(&mcpRegistry, &service, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner reference on service: %w", err)
	}
	if err := r.Create(ctx, &service); err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("service %q already exists\n", serviceName)
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to create service: %w", err)
		}
	}

	// Continue with normal reconcile logic...
	return ctrl.Result{}, nil
}
func (r *McpRegistryReconciler) createClientServiceAccount(ctx context.Context, mcpRegistry mcpv1.McpRegistry) error {
	registryName := mcpRegistry.Name
	saName := fmt.Sprintf("%s-client", registryName)
	namespace := mcpRegistry.Namespace

	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: namespace,
		},
	}

	if err := r.Create(ctx, &sa); err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("ServiceAccount %q already exists\n", saName)
		} else {
			return fmt.Errorf("error creating ServiceAccount: %w", err)
		}
	} else {
		fmt.Printf("ServiceAccount %q created in namespace %q\n", saName, namespace)
	}
	if err := controllerutil.SetControllerReference(&mcpRegistry, &sa, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference on ServiceAccount: %w", err)
	}
	return nil
}

func (r *McpRegistryReconciler) createClientRole(ctx context.Context, mcpRegistry mcpv1.McpRegistry) error {
	registryName := mcpRegistry.Name
	namespace := mcpRegistry.Namespace
	roleName := fmt.Sprintf("%s-client", registryName)

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleName,
			Namespace: namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{""},
				Resources:     []string{"mcpregistries"},
				Verbs:         []string{"get"},
				ResourceNames: []string{registryName},
			},
		},
	}
	if err := r.Create(ctx, role); err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("Role %q already exists\n", roleName)
		} else {
			return fmt.Errorf("error creating Role: %w", err)
		}
	} else {
		fmt.Printf("Role %q created in namespace %q\n", roleName, namespace)
	}
	if err := controllerutil.SetControllerReference(&mcpRegistry, role, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference on Role: %w", err)
	}
	return nil
}

func (r *McpRegistryReconciler) createClientRoleBinding(ctx context.Context, mcpRegistry mcpv1.McpRegistry) error {
	registryName := mcpRegistry.Name
	saName := fmt.Sprintf("%s-client", registryName)
	roleName := fmt.Sprintf("%s-client", registryName)
	namespace := mcpRegistry.Namespace

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-client-can-use-access-token", registryName),
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      saName,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		},
	}
	if err := r.Create(ctx, roleBinding); err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("RoleBinding %q already exists\n", roleBinding.Name)
		} else {
			return fmt.Errorf("error creating RoleBinding: %w", err)
		}
	} else {
		fmt.Printf("RoleBinding %q created in namespace %q\n", roleBinding.Name, namespace)
	}
	if err := controllerutil.SetControllerReference(&mcpRegistry, roleBinding, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference on RoleBinding: %w", err)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *McpRegistryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1.McpRegistry{}).
		Named("mcpregistry").
		Complete(r)
}
