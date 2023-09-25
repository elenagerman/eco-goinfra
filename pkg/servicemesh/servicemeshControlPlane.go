package servicemesh

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	istioV1 "istio.io/api/mesh/v1alpha1"
	istioV2 "istio.io/api/networking/v1alpha3"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ControlPlaneBuilder provides a struct for serviceMeshControlPlane object from the cluster and
// a serviceMeshControlPlane definition.
type ControlPlaneBuilder struct {
	// serviceMeshControlPlane definition, used to create the serviceMeshControlPlane object.
	Definition *istioV2.ServiceMeshControlPlane
	// Created serviceMeshControlPlane object.
	Object *istioV2.ServiceMeshControlPlane
	// Used in functions that define or mutate serviceMeshControlPlane definition. errorMsg is processed
	// before the serviceMeshControlPlane object is created
	errorMsg string
	// api client to interact with the cluster.
	apiClient *clients.Settings
}

// NewServiceMeshControlPlane method creates new instance of builder.
func NewServiceMeshControlPlane(apiClient *clients.Settings, name, nsname string) *ControlPlaneBuilder {
	glog.V(100).Infof("Initializing new serviceMeshControlPlane ControlPlaneBuilder structure with the following "+
		"params: name: %s, namespace: %s", name, nsname)

	builder := &ControlPlaneBuilder{
		apiClient: apiClient,
		Definition: &istioV2.ServiceMeshControlPlane{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the serviceMeshControlPlane is empty")

		builder.errorMsg = "The serviceMeshControlPlane 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the serviceMeshControlPlane is empty")

		builder.errorMsg = "The serviceMeshControlPlane 'namespace' cannot be empty"
	}

	return builder
}

// WithAllAddonsDisabled disable all addons to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithAllAddonsDisabled() *ControlPlaneBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	enablement := false

	glog.V(100).Infof(
		"Creating serviceMeshControlPlane %s with the all addons disabled", builder.Definition.Name)

	if builder.Definition.Spec.Addons == nil {
		builder.Definition.Spec.Addons = &istioV2.AddonsConfig{}
	}

	if builder.Definition.Spec.Addons.Grafana == nil {
		builder.Definition.Spec.Addons.Grafana = &istioV2.GrafanaAddonConfig{}
	}

	if builder.Definition.Spec.Addons.Kiali == nil {
		builder.Definition.Spec.Addons.Kiali = &istioV2.KialiAddonConfig{}
	}

	if builder.Definition.Spec.Addons.Jaeger == nil {
		builder.Definition.Spec.Addons.Jaeger = &istioV2.JaegerAddonConfig{}
	}

	if builder.Definition.Spec.Addons.Prometheus == nil {
		builder.Definition.Spec.Addons.Prometheus = &istioV2.PrometheusAddonConfig{}
	}

	if builder.Definition.Spec.Addons.Stackdriver == nil {
		builder.Definition.Spec.Addons.Stackdriver = &istioV2.StackdriverAddonConfig{}
	}

	if builder.Definition.Spec.Addons.ThreeScale == nil {
		builder.Definition.Spec.Addons.ThreeScale = &istioV2.ThreeScaleAddonConfig{}
	}

	disableAllAddons := &istioV2.AddonsConfig{
		Prometheus: &istioV2.PrometheusAddonConfig{
			Enablement: istioV2.Enablement{
				Enabled: &enablement,
			},
		},
		Jaeger: &istioV2.JaegerAddonConfig{
			Install: &istioV2.JaegerInstallConfig{
				Ingress: &istioV2.JaegerIngressConfig{
					Enablement: istioV2.Enablement{
						Enabled: &enablement,
					},
				},
			},
		},
		Grafana: &istioV2.GrafanaAddonConfig{
			Enablement: istioV2.Enablement{
				Enabled: &enablement,
			},
		},
		Kiali: &istioV2.KialiAddonConfig{
			Enablement: istioV2.Enablement{
				Enabled: &enablement,
			},
		},
	}

	builder.Definition.Spec.Addons = disableAllAddons

	return builder
}

// WithGrafanaAddon add grafana addons to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithGrafanaAddon(
	enablement bool,
	install *istioV2.GrafanaInstallConfig,
	address *string) *ControlPlaneBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating serviceMeshControlPlane %s with the Grafana addons defined: enablement %s, install %s, "+
			"address %s", builder.Definition.Name, enablement, install, address)

	grafanaAddon := &istioV2.GrafanaAddonConfig{
		Enablement: istioV2.Enablement{
			Enabled: &enablement,
		},
		Install: install,
		Address: address,
	}

	if builder.Definition.Spec.Addons == nil {
		builder.Definition.Spec.Addons = &istioV2.AddonsConfig{}
	}

	if builder.Definition.Spec.Addons.Grafana == nil {
		builder.Definition.Spec.Addons.Grafana = &istioV2.GrafanaAddonConfig{}
	}

	builder.Definition.Spec.Addons.Grafana = grafanaAddon

	return builder
}

// WithJaegerAddon add joeger addons to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithJaegerAddon(
	name string,
	storageType istioV2.JaegerStorageType,
	memoryStorageMaxTraces int64,
	elasticSearchNodesCount int32,
	elasticsearchStorage istioV1.HelmValues,
	elasticsearchRedundancyPolicy string,
	elasticsearchIndexCleaner istioV1.HelmValues,
	ingressEnablement bool,
	ingressLabels map[string]string,
	ingressAnnotations map[string]string,
) *ControlPlaneBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating serviceMeshControlPlane %s with the Jaeger addons defined", builder.Definition.Name)

	var jaegerAddon = &istioV2.JaegerAddonConfig{
		Name: name,
		Install: &istioV2.JaegerInstallConfig{
			Storage: &istioV2.JaegerStorageConfig{
				Type: storageType,
				Memory: &istioV2.JaegerMemoryStorageConfig{
					MaxTraces: &memoryStorageMaxTraces,
				},
				Elasticsearch: &istioV2.JaegerElasticsearchStorageConfig{
					NodeCount:        &elasticSearchNodesCount,
					Storage:          &elasticsearchStorage,
					RedundancyPolicy: elasticsearchRedundancyPolicy,
					IndexCleaner:     &elasticsearchIndexCleaner,
				},
			},
			Ingress: &istioV2.JaegerIngressConfig{
				Enablement: istioV2.Enablement{
					Enabled: &ingressEnablement,
				},
				Metadata: &istioV2.MetadataConfig{
					Labels:      ingressLabels,
					Annotations: ingressAnnotations,
				},
			},
		},
	}
	if builder.Definition.Spec.Addons == nil {
		builder.Definition.Spec.Addons = &istioV2.AddonsConfig{}
	}

	if builder.Definition.Spec.Addons.Jaeger == nil {
		builder.Definition.Spec.Addons.Jaeger = &istioV2.JaegerAddonConfig{}
	}

	builder.Definition.Spec.Addons.Jaeger = jaegerAddon

	return builder
}

// WithKialiAddon add kiali addons to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithKialiAddon(
	enablement bool,
	name string,
	dashboardViewOnly bool,
	dashboardEnableGrafana bool,
	dashboardEnablePrometheus bool,
	dashboardEnableTracing bool,
	serviceLabels map[string]string,
	serviceAnnotations map[string]string,
	serviceNodePort int32,
	ingressEnablement bool,
	ingressLabels map[string]string,
	ingressAnnotations map[string]string,
	ingressHosts []string,
	ingressContextPath string,
	ingressTLS istioV1.HelmValues,
) *ControlPlaneBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating serviceMeshControlPlane %s with the Kiali addons defined", builder.Definition.Name)

	kialiAddon := &istioV2.KialiAddonConfig{
		Enablement: istioV2.Enablement{
			Enabled: &enablement,
		},
		Name: name,
		Install: &istioV2.KialiInstallConfig{
			Dashboard: &istioV2.KialiDashboardConfig{
				ViewOnly:         &dashboardViewOnly,
				EnableGrafana:    &dashboardEnableGrafana,
				EnablePrometheus: &dashboardEnablePrometheus,
				EnableTracing:    &dashboardEnableTracing,
			},
			Service: &istioV2.ComponentServiceConfig{
				Metadata: &istioV2.MetadataConfig{
					Labels:      serviceLabels,
					Annotations: serviceAnnotations,
				},
				NodePort: &serviceNodePort,
				Ingress: &istioV2.ComponentIngressConfig{
					Enablement: istioV2.Enablement{
						Enabled: &ingressEnablement,
					},
					Metadata: &istioV2.MetadataConfig{
						Labels:      ingressLabels,
						Annotations: ingressAnnotations,
					},
					Hosts:       ingressHosts,
					ContextPath: ingressContextPath,
					TLS:         &ingressTLS,
				},
			},
		},
	}

	if builder.Definition.Spec.Addons == nil {
		builder.Definition.Spec.Addons = &istioV2.AddonsConfig{}
	}

	if builder.Definition.Spec.Addons.Kiali == nil {
		builder.Definition.Spec.Addons.Kiali = &istioV2.KialiAddonConfig{}
	}

	builder.Definition.Spec.Addons.Kiali = kialiAddon

	return builder
}

// WithPrometheusAddon add prometheus addons to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithPrometheusAddon(
	enablement bool,
	metricsExpiryDuration string,
	scrape bool,
	installRetention string,
	installScrapeInterval string,
	serviceLabels map[string]string,
	serviceAnnotations map[string]string,
	serviceNodePort int32,
	ingressEnablement bool,
	ingressLabels map[string]string,
	ingressAnnotations map[string]string,
	ingressHosts []string,
	ingressContextPath string,
	ingressTLS istioV1.HelmValues,
	useTLS bool,
	address string,
) *ControlPlaneBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating serviceMeshControlPlane %s with the Prometheus addons defined", builder.Definition.Name)

	prometheusAddon := &istioV2.PrometheusAddonConfig{
		Enablement: istioV2.Enablement{
			Enabled: &enablement,
		},
		MetricsExpiryDuration: metricsExpiryDuration,
		Scrape:                &scrape,
		Install: &istioV2.PrometheusInstallConfig{
			SelfManaged:    false,
			Retention:      installRetention,
			ScrapeInterval: installScrapeInterval,
			Service: &istioV2.ComponentServiceConfig{
				Metadata: &istioV2.MetadataConfig{
					Labels:      serviceLabels,
					Annotations: serviceAnnotations,
				},
				NodePort: &serviceNodePort,
				Ingress: &istioV2.ComponentIngressConfig{
					Enablement: istioV2.Enablement{
						Enabled: &ingressEnablement,
					},
					Metadata: &istioV2.MetadataConfig{
						Labels:      ingressLabels,
						Annotations: ingressAnnotations,
					},
					Hosts:       ingressHosts,
					ContextPath: ingressContextPath,
					TLS:         &ingressTLS,
				},
			},
			UseTLS: &useTLS,
		},
		Address: &address,
	}

	if builder.Definition.Spec.Addons == nil {
		builder.Definition.Spec.Addons = &istioV2.AddonsConfig{}
	}

	if builder.Definition.Spec.Addons.Prometheus == nil {
		builder.Definition.Spec.Addons.Prometheus = &istioV2.PrometheusAddonConfig{}
	}

	builder.Definition.Spec.Addons.Prometheus = prometheusAddon

	return builder
}

// WithGatewaysEnablement add gateway enablement to the serviceMeshControlPlane.
func (builder *ControlPlaneBuilder) WithGatewaysEnablement(enablement bool) *ControlPlaneBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Creating serviceMeshControlPlane %s with enabled Gateways", builder.Definition.Name)

	gatewaysConfig := istioV2.Enablement{
		Enabled: &enablement,
	}

	if builder.Definition.Spec.Gateways == nil {
		builder.Definition.Spec.Gateways = &istioV2.GatewaysConfig{}
	}

	builder.Definition.Spec.Gateways.Enablement = gatewaysConfig

	return builder
}

// Exists checks whether the given serviceMeshControlPlane exists.
func (builder *ControlPlaneBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if serviceMeshControlPlane %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	err := builder.Discover()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Discover fetches existing serviceMeshControlPlane from cluster.
func (builder *ControlPlaneBuilder) Discover() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Pulling existing serviceMeshControlPlane with name %s under namespace %s from cluster",
		builder.Definition.Name, builder.Definition.Namespace)

	smocp := &istioV2.ServiceMeshControlPlane{}
	err := builder.apiClient.Get(context.Background(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, smocp)

	builder.Object = smocp

	return err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *ControlPlaneBuilder) validate() (bool, error) {
	resourceCRD := "ServiceMeshControlPlane"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		builder.errorMsg = msg.UndefinedCrdObjectErrString(resourceCRD)
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		builder.errorMsg = fmt.Sprintf("%s builder cannot have nil apiClient", resourceCRD)
	}

	if builder.errorMsg != "" {
		glog.V(100).Infof("The %s builder has error message: %s", resourceCRD, builder.errorMsg)

		return false, fmt.Errorf(builder.errorMsg)
	}

	return true, nil
}
