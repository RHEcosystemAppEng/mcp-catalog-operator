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
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mcpv1alpha1 "github.com/dmartinol/mcp-registry-operator/api/v1alpha1"
)

// McpServerRunReconciler reconciles a McpServerRun object
type McpServerRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpserverruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpserverruns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=mcp.opendatahub.io,resources=mcpserverruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the McpServerRun object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *McpServerRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var mcpServerRun mcpv1alpha1.McpServerRun
	if err := r.Get(ctx, req.NamespacedName, &mcpServerRun); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var serverImage *string = nil
	var command string
	args := mcpServerRun.Spec.McpServer.Args

	registryRef := mcpServerRun.Spec.RegistryRef
	if registryRef.Name == "" {
		return ctrl.Result{}, fmt.Errorf("invalid registryRef: name missing")
	}

	authConfig := mcpServerRun.Spec.McpServer.Auth
	if authConfig == nil {
		authConfig = &mcpv1alpha1.AuthConfig{Enabled: true}
		mcpServerRun.Spec.McpServer.Auth = authConfig
		fmt.Printf("Initialized auth to %v", authConfig)
	}

	var registry mcpv1alpha1.McpRegistry
	registryNs := mcpServerRun.Namespace
	if registryRef.Namespace != nil {
		registryNs = *registryRef.Namespace
	}
	fmt.Printf("Looking for McpRegistry %s in %s", registryRef.Name, registryNs)
	if err := r.Get(ctx, types.NamespacedName{Name: registryRef.Name, Namespace: registryNs}, &registry); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get referenced McpRegistry: %w", err)
	}

	// Set McpRegistry as owner of McpServerRun
	if err := controllerutil.SetControllerReference(&registry, &mcpServerRun, r.Scheme); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to set owner reference: %w", err)
	}

	// Persist change if necessary
	if err := r.Update(ctx, &mcpServerRun); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update McpServerRun with owner ref: %w", err)
	}

	// Get mcpServerRef from spec
	if mcpServerRun.Spec.ServerMode == "blueprint" {
		fmt.Printf("mcpServer.Spec is %+v\n", mcpServerRun.Spec)
		ref := mcpServerRun.Spec.McpServer.McpServerRef
		if ref.Name == "" {
			return ctrl.Result{}, fmt.Errorf("invalid mcpServerRef: name missing")
		}

		var mcpCertServer mcpv1alpha1.McpCertifiedServer
		ns := mcpServerRun.Namespace
		if ref.Namespace != nil {
			ns = *ref.Namespace
		}
		fmt.Printf("Looking for McpCertifiedServer %s in %s", ref.Name, ns)
		if err := r.Get(ctx, types.NamespacedName{Name: ref.Name, Namespace: ns}, &mcpCertServer); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get referenced McpCertifiedServer: %w", err)
		}

		serverImage = &mcpCertServer.Spec.McpServer.Image
		command = mcpCertServer.Spec.McpServer.Command
		args = append(args, mcpCertServer.Spec.McpServer.Args...)
	}

	if mcpServerRun.Spec.ServerMode == "container" {
		serverImage = &mcpServerRun.Spec.McpServer.ServerImage
	}

	if serverImage != nil {
		deployName := mcpServerRun.Name + "-deployment"
		deployNamespace := mcpServerRun.Namespace

		var deploy appsv1.Deployment
		err := r.Get(ctx, types.NamespacedName{Name: deployName, Namespace: deployNamespace}, &deploy)
		if err == nil {
			return ctrl.Result{}, nil
		}
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}

		// 4. Create a new Deployment
		labels := map[string]string{
			"app": mcpServerRun.Name,
		}

		replicas := int32(1)
		if mcpServerRun.Spec.Replicas != nil {
			replicas = *mcpServerRun.Spec.Replicas
		}

		if mcpServerRun.Spec.McpServer.Proxy != nil && *mcpServerRun.Spec.McpServer.Proxy {
			// TODO Have supergateway already in the base image
			args = []string{"-y", "supergateway", "--stdio", fmt.Sprintf("%s %s", command, strings.Join(args, " "))}
			command = "npx"
		}

		containers := []corev1.Container{
			{
				Name:    "mcp-server",
				Image:   *serverImage,
				Args:    args,
				EnvFrom: mcpServerRun.Spec.EnvFrom,
				Command: []string{command},
			},
		}
		// if mcpServer.Spec.McpServer.Proxy != nil && *mcpServer.Spec.McpServer.Proxy {
		// 	containers = append(containers, corev1.Container{
		// 		Name:  "mcp-proxy",
		// 		Image: "quay.io/dmartino/mcp-proxy:amd64",
		// 		// Args:  []string{"--sse-port", "8000", "http://0.0.0.0:8080/sse"},
		// 		Args: []string{"--sse-port", "8000"},
		// 		Ports: []corev1.ContainerPort{
		// 			{ContainerPort: 8000},
		// 		},
		// 	})
		// }

		deploy = appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deployName,
				Namespace: deployNamespace,
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
						Containers: containers,
					},
				},
			},
		}

		if err := controllerutil.SetControllerReference(&mcpServerRun, &deploy, r.Scheme); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to set owner reference on deployment: %w", err)
		}
		if err := r.Create(ctx, &deploy); err != nil {
			if apierrors.IsAlreadyExists(err) {
				fmt.Printf("deployment %q already exists\n", deployName)
			} else {
				return ctrl.Result{}, fmt.Errorf("failed to create deployment: %w", err)
			}
		}
	}

	serviceName := mcpServerRun.Name + "-svc"
	serviceNamespace := mcpServerRun.Namespace
	labels := map[string]string{
		"app": mcpServerRun.Name,
	}
	service := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: serviceNamespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": mcpServerRun.Name},
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

	if err := controllerutil.SetControllerReference(&mcpServerRun, &service, r.Scheme); err != nil {
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
func (r *McpServerRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcpv1alpha1.McpServerRun{}).
		Named("mcpserverrun").
		Complete(r)
}
