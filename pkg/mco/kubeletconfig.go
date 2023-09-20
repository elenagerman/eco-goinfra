package mco

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	mcv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

// KCBuilder provides struct for KubeletConfig Object which contains connection to cluster
// and KubeletConfig definitions.
type KCBuilder struct {
	// KubeletConfig definition. Used to create KubeletConfig object with minimum set of required elements.
	Definition *mcv1.KubeletConfig
	// Created KubeletConfig object on the cluster.
	Object *mcv1.KubeletConfig
	// api client to interact with the cluster.
	apiClient *clients.Settings
	// errorMsg is processed before KubeletConfig object is created.
	errorMsg string
}

// KCAdditionalOptions for kubeletconfig object.
type KCAdditionalOptions func(builder *KCBuilder) (*KCBuilder, error)

// NewKCBuilder provides struct for KubeletConfig object which contains connection to cluster
// and KubeletConfig definition.
func NewKCBuilder(apiClient *clients.Settings, name string) *KCBuilder {
	glog.V(100).Infof("Initializing new KCBuilder structure with following params: %s", name)

	builder := KCBuilder{
		apiClient: apiClient,
		Definition: &mcv1.KubeletConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the KubeletConfig is empty")

		builder.errorMsg = "KubeletConfig 'name' cannot be empty"
	}

	return &builder
}

// PullKubeletConfig fetches existing kubeletconfig from cluster.
func PullKubeletConfig(apiClient *clients.Settings, name string) (*KCBuilder, error) {
	glog.V(100).Infof("Pulling existing kubeletconfig name %s from cluster", name)

	builder := KCBuilder{
		apiClient: apiClient,
		Definition: &mcv1.KubeletConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the kubeletconfig is empty")

		builder.errorMsg = "kubeletconfig 'name' cannot be empty"
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("kubeletconfig object %s doesn't exist", name)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Create generates a kubeletconfig in the cluster and stores the created object in struct.
func (builder *KCBuilder) Create() (*KCBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating KubeletConfig %s", builder.Definition.Name)

	var err error
	if !builder.Exists() {
		builder.Object, err = builder.apiClient.KubeletConfigs().Create(
			context.TODO(), builder.Definition, metav1.CreateOptions{})
	}

	return builder, err
}

// Delete removes the kubeletconfig.
func (builder *KCBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the kubeletconfig object %s", builder.Definition.Name)

	if !builder.Exists() {
		return fmt.Errorf("kubeletconfig cannot be deleted because it does not exist")
	}

	err := builder.apiClient.KubeletConfigs().Delete(
		context.TODO(), builder.Object.Name, metav1.DeleteOptions{})

	if err != nil {
		return fmt.Errorf("cannot delete kubeletconfig: %w", err)
	}

	builder.Object = nil

	return err
}

// Exists checks whether the given kubeletconfig exists.
func (builder *KCBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if the kubeletconfig object %s exists", builder.Definition.Name)

	var err error
	builder.Object, err = builder.apiClient.KubeletConfigs().Get(
		context.Background(), builder.Definition.Name, metav1.GetOptions{})

	return err == nil || !k8serrors.IsNotFound(err)
}

// WithMCPoolSelector redefines kubeletconfig definition with the given machineConfigPoolSelector field.
func (builder *KCBuilder) WithMCPoolSelector(key, value string) *KCBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Labeling the kubeletconfig %s with %s=%s", builder.Definition.Name, key, value)

	if key == "" {
		glog.V(100).Infof("The key can't be empty")

		builder.errorMsg = "'key' cannot be empty"

		return builder
	}

	if builder.Definition.Spec.MachineConfigPoolSelector == nil {
		builder.Definition.Spec.MachineConfigPoolSelector = &metav1.LabelSelector{}
	}

	if builder.Definition.Spec.MachineConfigPoolSelector.MatchLabels == nil {
		builder.Definition.Spec.MachineConfigPoolSelector.MatchLabels = map[string]string{}
	}

	builder.Definition.Spec.MachineConfigPoolSelector.MatchLabels[key] = value

	return builder
}

// WithSystemReserved redefines kubeletconfig definition with the given systemreserved fields.
func (builder *KCBuilder) WithSystemReserved(cpu, memory string) *KCBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting up %s cpu and %s memory to the %s kubeletconfig",
		cpu, memory, builder.Definition.Name)

	if cpu == "" {
		glog.V(100).Infof("The cpu can't be empty")

		builder.errorMsg = "'cpu' cannot be empty"

		return builder
	}

	if memory == "" {
		glog.V(100).Infof("The memory can't be empty")

		builder.errorMsg = "'memory' cannot be empty"

		return builder
	}

	if builder.Definition.Spec.KubeletConfig == nil {
		builder.Definition.Spec.KubeletConfig = &runtime.RawExtension{}
	}

	if builder.Definition.Spec.KubeletConfig.Raw == nil {
		builder.Definition.Spec.KubeletConfig.Raw = []byte{}
	}

	systemReservedKubeletConfiguration := &kubeletconfigv1beta1.KubeletConfiguration{
		SystemReserved: map[string]string{
			cpu:    cpu,
			memory: memory,
		},
	}

	builder.Definition.Spec.KubeletConfig.Object = systemReservedKubeletConfiguration

	return builder
}

// WithOptions creates the kubeletconfig with generic mutation options.
func (builder *KCBuilder) WithOptions(options ...KCAdditionalOptions) *KCBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof("Setting kubeletconfig additional options")

	for _, option := range options {
		if option != nil {
			builder, err := option(builder)

			if err != nil {
				glog.V(100).Infof("Error occurred in mutation function")

				builder.errorMsg = err.Error()

				return builder
			}
		}
	}

	return builder
}

func (builder *KCBuilder) validate() (bool, error) {
	resourceCRD := "KubeletConfig"

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
