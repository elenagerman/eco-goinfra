package storage

import (
	"context"
	"fmt"
	ocsoperatorv1 "github.com/red-hat-storage/ocs-operator/api/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang/glog"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StorageClusterBuilder provides struct for PullStorageCluster object containing connection
// to the cluster and the storageCluster definitions.
type StorageClusterBuilder struct {
	// StorageCluster definition. Used to create a storageCluster object
	Definition *ocsoperatorv1.StorageCluster
	// Created storageCluster object
	Object *ocsoperatorv1.StorageCluster
	// api client to interact with the cluster.
	apiClient goclient.Client
	// Used in functions that define or mutate storageCluster definition. errorMsg is processed before the
	// storageCluster object is created.
	errorMsg string
}

// StorageClusterNewBuilder creates a new instance of Builder.
func StorageClusterNewBuilder(apiClient *clients.Settings, name, nsname string) *StorageClusterBuilder {
	glog.V(100).Infof(
		"Initializing new storageCluster structure with the following params: %s, %s", name, nsname)

	builder := StorageClusterBuilder{
		apiClient: apiClient,
		Definition: &ocsoperatorv1.StorageCluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       StorageClusterKind,
				APIVersion: fmt.Sprintf("%s/%s", APIGroup, APIVersion),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the storageCluster is empty")

		builder.errorMsg = "storageCluster 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the storageCluster is empty")

		builder.errorMsg = "storageCluster 'nsname' cannot be empty"
	}

	return &builder
}

// PullStorageCluster gets an existing storageCluster from the cluster.
func PullStorageCluster(apiClient *clients.Settings, name, namespace string) (*StorageClusterBuilder, error) {
	if apiClient == nil {
		glog.V(100).Infof("The apiClient is empty")

		return nil, fmt.Errorf("storageCluster 'apiClient' cannot be empty")
	}

	glog.V(100).Infof("Pulling existing storageCluster object %s from namespace %s",
		name, namespace)

	builder := StorageClusterBuilder{
		apiClient: apiClient,
		Definition: &ocsoperatorv1.StorageCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the storageCluster is empty")

		return nil, fmt.Errorf("storageCluster 'name' cannot be empty")
	}

	if namespace == "" {
		glog.V(100).Infof("The namespace of the storageCluster is empty")

		return nil, fmt.Errorf("storageCluster 'namespace' cannot be empty")
	}

	if !builder.Exists() {
		return nil, fmt.Errorf("storageCluster object %s doesn't exist in namespace %s",
			name, namespace)
	}

	builder.Definition = builder.Object

	return &builder, nil
}

// Get fetches existing storageCluster from cluster.
func (builder *StorageClusterBuilder) Get() (*ocsoperatorv1.StorageCluster, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting existing storageCluster with name %s from the namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	storageClusterObj := &ocsoperatorv1.StorageCluster{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name: builder.Definition.Name,
	}, storageClusterObj)

	if err != nil {
		glog.V(100).Infof("storageCluster object %s doesn't exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)

		return nil, err
	}

	return storageClusterObj, nil
}

// Exists checks whether the given storageCluster exists.
func (builder *StorageClusterBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if storageCluster %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	var err error
	builder.Object, err = builder.Get()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Create makes a storageCluster in the cluster and stores the created object in struct.
func (builder *StorageClusterBuilder) Create() (*StorageClusterBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Creating the storageCluster %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace,
	)

	var err error
	if !builder.Exists() {
		err = builder.apiClient.Create(context.TODO(), builder.Definition)
		if err == nil {
			builder.Object = builder.Definition
		}
	}

	return builder, err
}

// Delete removes storageCluster object from a cluster.
func (builder *StorageClusterBuilder) Delete() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Deleting the storageCluster object %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	err := builder.apiClient.Delete(context.TODO(), builder.Definition)

	if err != nil {

		return fmt.Errorf("can not delete storageCluster: %w", err)
	}

	builder.Object = nil

	return nil
}

// Update renovates the storageCluster in the cluster and stores the created object in struct.
func (builder *StorageClusterBuilder) Update() (*StorageClusterBuilder, error) {
	if valid, err := builder.validate(); !valid {
		return builder, err
	}

	glog.V(100).Infof("Updating the storageCluster %s in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return nil, fmt.Errorf("storageCluster object %s doesn't exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	err := builder.apiClient.Update(context.TODO(), builder.Definition)
	if err == nil {
		builder.Object = builder.Definition
	}

	return builder, err
}

// GetManageNodes fetches storageCluster manageNodes value.
func (builder *StorageClusterBuilder) GetManageNodes() (bool, error) {
	if valid, err := builder.validate(); !valid {
		return false, err
	}

	glog.V(100).Infof("Getting storageCluster %s in namespace %s manageNodes configuration",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return false, fmt.Errorf("storageCluster object %s doesn't exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return builder.Object.Spec.ManageNodes, nil
}

// GetManagedResources fetches storageCluster managedResources value.
func (builder *StorageClusterBuilder) GetManagedResources() (*ocsoperatorv1.ManagedResourcesSpec, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting storageCluster %s in namespace %s managedResources configuration",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return nil, fmt.Errorf("storageCluster object %s doesn't exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return &builder.Object.Spec.ManagedResources, nil
}

// GetMonDataDirHostPath fetches storageCluster monDataDirHostPath value.
func (builder *StorageClusterBuilder) GetMonDataDirHostPath() (string, error) {
	if valid, err := builder.validate(); !valid {
		return "", err
	}

	glog.V(100).Infof("Getting storageCluster %s in namespace %s monDataDirHostPath configuration",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return "", fmt.Errorf("storageCluster object %s doesn't exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return builder.Object.Spec.MonDataDirHostPath, nil
}

// GetMultiCloudGateway fetches storageCluster multiCloudGateway value.
func (builder *StorageClusterBuilder) GetMultiCloudGateway() (*ocsoperatorv1.MultiCloudGatewaySpec, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting storageCluster %s in namespace %s multiCloudGateway configuration",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return nil, fmt.Errorf("storageCluster object %s doesn't exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return builder.Object.Spec.MultiCloudGateway, nil
}

// GetStorageDeviceSets fetches storageCluster storageDeviceSets value.
func (builder *StorageClusterBuilder) GetStorageDeviceSets() ([]ocsoperatorv1.StorageDeviceSet, error) {
	if valid, err := builder.validate(); !valid {
		return nil, err
	}

	glog.V(100).Infof("Getting storageCluster %s in namespace %s storageDeviceSets configuration",
		builder.Definition.Name, builder.Definition.Namespace)

	if !builder.Exists() {
		return nil, fmt.Errorf("storageCluster object %s doesn't exist in namespace %s",
			builder.Definition.Name, builder.Definition.Namespace)
	}

	return builder.Object.Spec.StorageDeviceSets, nil
}

// WithManagedNodes sets the storageCluster's managedNodes value.
func (builder *StorageClusterBuilder) WithManagedNodes(expectedManagedNodesValue bool) *StorageClusterBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with managedNodes value: %v",
		builder.Definition.Name, builder.Definition.Namespace, expectedManagedNodesValue)

	builder.Definition.Spec.ManageNodes = expectedManagedNodesValue

	return builder
}

// WithManagedResources sets the storageCluster's managedResources value.
func (builder *StorageClusterBuilder) WithManagedResources(
	expectedManagedResources ocsoperatorv1.ManagedResourcesSpec) *StorageClusterBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with managedResources value: %v",
		builder.Definition.Name, builder.Definition.Namespace, expectedManagedResources)

	if &expectedManagedResources == nil {
		glog.V(100).Infof("the expectedManagedResources can not be empty")

		builder.errorMsg = "the expectedManagedResources can not be empty"

		return builder
	}

	builder.Definition.Spec.ManagedResources = expectedManagedResources

	return builder
}

// WithMonDataDirHostPath sets the storageCluster's monDataDirHostPath value.
func (builder *StorageClusterBuilder) WithMonDataDirHostPath(
	expectedMonDataDirHostPath string) *StorageClusterBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with monDataDirHostPath value: %v",
		builder.Definition.Name, builder.Definition.Namespace, expectedMonDataDirHostPath)

	if expectedMonDataDirHostPath == "" {
		glog.V(100).Infof("the expectedMonDataDirHostPath can not be empty")

		builder.errorMsg = "the expectedMonDataDirHostPath can not be empty"

		return builder
	}

	builder.Definition.Spec.MonDataDirHostPath = expectedMonDataDirHostPath

	return builder
}

// WithMultiCloudGateway sets the storageCluster's multiCloudGateway value.
func (builder *StorageClusterBuilder) WithMultiCloudGateway(
	expectedMultiCloudGateway ocsoperatorv1.MultiCloudGatewaySpec) *StorageClusterBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with multiCloudGateway value: %v",
		builder.Definition.Name, builder.Definition.Namespace, expectedMultiCloudGateway)

	if &expectedMultiCloudGateway == nil {
		glog.V(100).Infof("the expectedMultiCloudGateway can not be empty")

		builder.errorMsg = "the expectedMultiCloudGateway can not be empty"

		return builder
	}

	builder.Definition.Spec.MultiCloudGateway = &expectedMultiCloudGateway

	return builder
}

// WithStorageDeviceSet sets the storageCluster's storageDeviceSets value.
func (builder *StorageClusterBuilder) WithStorageDeviceSet(
	expectedStorageDeviceSet ocsoperatorv1.StorageDeviceSet) *StorageClusterBuilder {
	if valid, _ := builder.validate(); !valid {
		return builder
	}

	glog.V(100).Infof(
		"Setting storageCluster %s in namespace %s with storageDeviceSets value: %v",
		builder.Definition.Name, builder.Definition.Namespace, expectedStorageDeviceSet)

	if &expectedStorageDeviceSet == nil {
		glog.V(100).Infof("the expectedStorageDeviceSet can not be empty")

		builder.errorMsg = "the expectedStorageDeviceSet can not be empty"

		return builder
	}

	if builder.Definition.Spec.StorageDeviceSets == nil {
		glog.V(100).Infof("Plugins are nil. Initializing one")

		builder.Definition.Spec.StorageDeviceSets = []ocsoperatorv1.StorageDeviceSet{}
	}

	builder.Definition.Spec.StorageDeviceSets =
		append(builder.Definition.Spec.StorageDeviceSets, expectedStorageDeviceSet)

	return builder
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *StorageClusterBuilder) validate() (bool, error) {
	resourceCRD := "StorageCluster"

	if builder == nil {
		glog.V(100).Infof("The %s builder is uninitialized", resourceCRD)

		return false, fmt.Errorf("error: received nil %s builder", resourceCRD)
	}

	if builder.Definition == nil {
		glog.V(100).Infof("The %s is undefined", resourceCRD)

		return false, fmt.Errorf(msg.UndefinedCrdObjectErrString(resourceCRD))
	}

	if builder.apiClient == nil {
		glog.V(100).Infof("The %s builder apiclient is nil", resourceCRD)

		return false, fmt.Errorf("%s builder cannot have nil apiClient", resourceCRD)
	}

	return true, nil
}
