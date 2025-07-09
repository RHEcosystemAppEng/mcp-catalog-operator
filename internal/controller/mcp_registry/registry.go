package mcp_registry

import (
	"context"
	"fmt"
	"os"

	mcpv1alpha1 "github.com/RHEcosystemAppEng/mcp-registry-operator/api/v1alpha1"
	"github.com/RHEcosystemAppEng/mcp-registry-operator/internal/types"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func CreateMcpRegistryResources(ctx context.Context, c client.Client, scheme *runtime.Scheme, mcpRegistry *mcpv1alpha1.McpRegistry) error {
	// Create registry deployment
	registryLabels := map[string]string{"app": mcpRegistry.Name, "component": "mcp-registry"}

	if err := createRegistryDeployment(ctx, c, scheme, mcpRegistry, registryLabels); err != nil {
		return fmt.Errorf("failed to create registry deployment: %w", err)
	}

	// Create registry service
	if err := createRegistryService(ctx, c, scheme, mcpRegistry, registryLabels); err != nil {
		return fmt.Errorf("failed to create registry service: %w", err)
	}

	// Create registry UI deployment
	registryUiLabels := map[string]string{"app": mcpRegistry.Name, "component": "mcp-registry-ui"}

	if err := createRegistryUiDeployment(ctx, c, scheme, mcpRegistry, registryUiLabels); err != nil {
		return fmt.Errorf("failed to create registry UI deployment: %w", err)
	}

	// Create registry UI service and route
	if err := createRegistryUiServiceAndRoute(ctx, c, scheme, mcpRegistry, registryUiLabels); err != nil {
		return fmt.Errorf("failed to create registry UI service and route: %w", err)
	}
	return nil
}

func createRegistryDeployment(ctx context.Context, c client.Client, scheme *runtime.Scheme, mcpRegistry *mcpv1alpha1.McpRegistry, registryLabels map[string]string) error {
	registryName := mcpRegistry.Name
	namespace := mcpRegistry.Namespace
	mongoSvcName := GetMongoServiceName(registryName)

	registryDeploymentName := registryName + "-registry"

	registryDeployment := &appsv1.Deployment{}
	if err := c.Get(ctx, k8stypes.NamespacedName{Name: registryDeploymentName, Namespace: namespace}, registryDeployment); apierrors.IsNotFound(err) {
		registryDeployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registryDeploymentName,
				Namespace: namespace,
				Labels:    registryLabels,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{MatchLabels: registryLabels},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: registryLabels},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "mcp-registry",
							Image: types.RegistryImage,
							Env: []corev1.EnvVar{
								{Name: "MCP_REGISTRY_DATABASE_TYPE", Value: types.MongoDBDatabaseType},
								{Name: "MCP_REGISTRY_DATABASE_URL", Value: fmt.Sprintf("mongodb://%s:%s", mongoSvcName, types.MongoDBPort)},
								{Name: "MCP_REGISTRY_DATABASE_NAME", Value: types.MongoDBDatabaseName},
								{Name: "MCP_REGISTRY_COLLECTION_NAME", Value: types.MongoDBCollectionName},
								{Name: "MCP_REGISTRY_LOG_LEVEL", Value: types.MongoDBLogLevel},
								{Name: "MCP_REGISTRY_SEED_IMPORT", Value: types.MongoDBSeedImport},
							},
							Ports: []corev1.ContainerPort{{ContainerPort: types.RegistryPort}},
						}},
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(mcpRegistry, registryDeployment, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on registry deployment: %w", err)
		}
		if err := c.Create(ctx, registryDeployment); err != nil {
			return fmt.Errorf("failed to create registry deployment: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func createRegistryService(ctx context.Context, c client.Client, scheme *runtime.Scheme, mcpRegistry *mcpv1alpha1.McpRegistry, registryLabels map[string]string) error {
	registryName := mcpRegistry.Name
	namespace := mcpRegistry.Namespace

	registrySvcName := registryName + "-registry"

	regSvc := &corev1.Service{}
	if err := c.Get(ctx, k8stypes.NamespacedName{Name: registrySvcName, Namespace: namespace}, regSvc); apierrors.IsNotFound(err) {
		regSvc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registrySvcName,
				Namespace: namespace,
				Labels:    registryLabels,
			},
			Spec: corev1.ServiceSpec{
				Selector: registryLabels,
				Ports: []corev1.ServicePort{{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				}},
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		if err := controllerutil.SetControllerReference(mcpRegistry, regSvc, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on registry service: %w", err)
		}
		if err := c.Create(ctx, regSvc); err != nil {
			return fmt.Errorf("failed to create registry service: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func createRegistryUiDeployment(ctx context.Context, c client.Client, scheme *runtime.Scheme, mcpRegistry *mcpv1alpha1.McpRegistry, registryUiLabels map[string]string) error {
	registryName := mcpRegistry.Name
	namespace := mcpRegistry.Namespace

	registrySvcName := registryName + "-registry"
	registryUiDeploymentName := registryName + "-registry-ui"

	registryUiDeployment := &appsv1.Deployment{}
	if err := c.Get(ctx, k8stypes.NamespacedName{Name: registryUiDeploymentName, Namespace: namespace}, registryUiDeployment); apierrors.IsNotFound(err) {
		registryUiDeployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registryUiDeploymentName,
				Namespace: namespace,
				Labels:    registryUiLabels,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: int32Ptr(1),
				Selector: &metav1.LabelSelector{MatchLabels: registryUiLabels},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: registryUiLabels},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "mcp-registry-ui",
							Image: types.RegistryUiImage,
							Env: []corev1.EnvVar{
								{Name: "API_URL", Value: fmt.Sprintf("http://%s:%d", registrySvcName, types.RegistryUiPort)},
							},
							Ports: []corev1.ContainerPort{{ContainerPort: types.RegistryUiPort}},
						}},
					},
				},
			},
		}
		if err := controllerutil.SetControllerReference(mcpRegistry, registryUiDeployment, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on registry UI deployment: %w", err)
		}
		if err := c.Create(ctx, registryUiDeployment); err != nil {
			return fmt.Errorf("failed to create registry UI deployment: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func createRegistryUiServiceAndRoute(ctx context.Context, c client.Client, scheme *runtime.Scheme, mcpRegistry *mcpv1alpha1.McpRegistry, registryUiLabels map[string]string) error {
	registryName := mcpRegistry.Name
	namespace := mcpRegistry.Namespace

	registryUiSvcName := registryName + "-registry-ui"

	// Create service
	regSvc := &corev1.Service{}
	if err := c.Get(ctx, k8stypes.NamespacedName{Name: registryUiSvcName, Namespace: namespace}, regSvc); apierrors.IsNotFound(err) {
		regSvc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      registryUiSvcName,
				Namespace: namespace,
				Labels:    registryUiLabels,
			},
			Spec: corev1.ServiceSpec{
				Selector: registryUiLabels,
				Ports: []corev1.ServicePort{{
					Name:       "http",
					Port:       types.RegistryUiPort,
					TargetPort: intstr.FromInt(types.RegistryUiPort),
				}},
				Type: corev1.ServiceTypeClusterIP,
			},
		}
		if err := controllerutil.SetControllerReference(mcpRegistry, regSvc, scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on registry UI service: %w", err)
		}
		if err := c.Create(ctx, regSvc); err != nil {
			return fmt.Errorf("failed to create registry UI service: %w", err)
		}
	} else if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	if os.Getenv("DEPLOY_ROUTE") != "false" {
		// Create route
		registryUiRouteName := registryName + "-registry-ui"
		route := &routev1.Route{}

		if err := c.Get(ctx, k8stypes.NamespacedName{Name: registryUiRouteName, Namespace: namespace}, route); apierrors.IsNotFound(err) {
			route = &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:      registryUiRouteName,
					Namespace: namespace,
					Labels:    registryUiLabels,
				},
				Spec: routev1.RouteSpec{
					Port: &routev1.RoutePort{
						TargetPort: intstr.FromInt(8080),
					},
					TLS: &routev1.TLSConfig{
						Termination: routev1.TLSTerminationEdge,
					},
					To: routev1.RouteTargetReference{
						Kind: "Service",
						Name: registryUiSvcName,
					},
				},
			}
			if err := controllerutil.SetControllerReference(mcpRegistry, route, scheme); err != nil {
				return fmt.Errorf("failed to set owner reference on registry UI route: %w", err)
			}
			if err := c.Create(ctx, route); err != nil {
				return fmt.Errorf("failed to create registry UI route: %w", err)
			}
		} else if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func int32Ptr(i int32) *int32 { return &i }
