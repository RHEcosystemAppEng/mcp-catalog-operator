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

	mcpv1 "github.com/dmartinol/mcp-catalog-operator/api/v1"
)

// McpCatalogReconciler reconciles a McpCatalog object
type McpCatalogReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpcatalogs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpcatalogs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpcatalogs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the McpCatalog object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *McpCatalogReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var mcpCatalog mcpv1.McpCatalog
	if err := r.Get(ctx, req.NamespacedName, &mcpCatalog); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	catalogName := mcpCatalog.Name
	saName := mcpCatalog.Name

	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: mcpCatalog.Namespace,
		},
	}

	if err := r.Create(ctx, &sa); err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("ServiceAccount %q already exists\n", saName)
		} else {
			return ctrl.Result{}, fmt.Errorf("Error creating ServiceAccount: %w", err)
		}
	} else {
		fmt.Printf("ServiceAccount %q created in namespace %q\n", saName, mcpCatalog.Namespace)
	}
	if err := controllerutil.SetControllerReference(&mcpCatalog, &sa, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner reference on ServiceAccount: %w", err)
	}

	for _, role := range []string{
		"mcpcatalog-admin-role",
		"mcpblueprint-admin-role",
		"mcpserverdefinition-admin-role",
		"mcpregistry-admin-role",
		"pipeline-as-code-controller-clusterrole"} {
		crbName := fmt.Sprintf("%s-is-%s", saName, role)
		crb := rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      crbName,
				Namespace: mcpCatalog.Namespace,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      saName,
					Namespace: mcpCatalog.Namespace,
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
			fmt.Printf("ClusterRoleBinding %q created in namespace %q\n", crbName, mcpCatalog.Namespace)
		}
		if err := controllerutil.SetControllerReference(&mcpCatalog, &crb, r.Scheme); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set owner reference on ClusterRoleBinding: %w", err)
		}
	}

	deployName := catalogName
	labels := map[string]string{
		"app": catalogName,
	}
	replicas := int32(1)

	deploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployName,
			Namespace: mcpCatalog.Namespace,
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
							Name:            "mcp-server",
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
									Name:  "MCP_CATALOG_NAME",
									Value: catalogName,
								},
								{
									Name:  "MCP_REGISTRY_NAME",
									Value: "foo",
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(&mcpCatalog, &deploy, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner reference on deployment: %w", err)
	}
	if err := r.Create(ctx, &deploy); err != nil {
		if apierrors.IsAlreadyExists(err) {
			fmt.Printf("deployment %q already exists\n", deployName)
		} else {
			return ctrl.Result{}, fmt.Errorf("failed to create deployment: %w", err)
		}
	}

	serviceName := mcpCatalog.Name + "-svc"
	serviceNamespace := mcpCatalog.Namespace
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": mcpCatalog.Name},
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

	if err := controllerutil.SetControllerReference(&mcpCatalog, &service, r.Scheme); err != nil {
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

// SetupWithManager sets up the controller with the Manager.
func (r *McpCatalogReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1.McpCatalog{}).
		Named("mcpcatalog").
		Complete(r)
}
