package servicemesh

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	istioV1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/openshift-kni/eco-goinfra/pkg/msg"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	goclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// MemberRollBuilder provides a struct for serviceMeshMemberRoll object from the cluster and
// a serviceMeshMemberRoll definition.
type MemberRollBuilder struct {
	// serviceMeshMemberRoll definition, used to create the serviceMeshMemberRoll object.
	Definition *istioV1.ServiceMeshMemberRoll
	// Created serviceMeshMemberRoll object.
	Object *istioV1.ServiceMeshMemberRoll
	// Used in functions that define or mutate serviceMeshMemberRoll definition. errorMsg is processed
	// before the serviceMeshMemberRoll object is created
	errorMsg string
	// api client to interact with the cluster.
	apiClient *clients.Settings
}

// NewServiceMeshMemberRoll method creates new instance of builder.
func NewServiceMeshMemberRoll(apiClient *clients.Settings, name, nsname string) *MemberRollBuilder {
	glog.V(100).Infof("Initializing new serviceMeshMemberRoll ControlPlaneBuilder structure with the following "+
		"params: name: %s, namespace: %s", name, nsname)

	builder := &MemberRollBuilder{
		apiClient: apiClient,
		Definition: &istioV1.ServiceMeshMemberRoll{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: nsname,
			},
		},
	}

	if name == "" {
		glog.V(100).Infof("The name of the serviceMeshMemberRoll is empty")

		builder.errorMsg = "The serviceMeshMemberRoll 'name' cannot be empty"
	}

	if nsname == "" {
		glog.V(100).Infof("The namespace of the serviceMeshMemberRoll is empty")

		builder.errorMsg = "The serviceMeshMemberRoll 'namespace' cannot be empty"
	}

	return builder
}

// Exists checks whether the given serviceMeshMemberRoll exists.
func (builder *MemberRollBuilder) Exists() bool {
	if valid, _ := builder.validate(); !valid {
		return false
	}

	glog.V(100).Infof("Checking if serviceMeshMemberRoll %s exists in namespace %s",
		builder.Definition.Name, builder.Definition.Namespace)

	err := builder.Discover()

	return err == nil || !k8serrors.IsNotFound(err)
}

// Discover fetches existing serviceMeshMemberRoll from cluster.
func (builder *MemberRollBuilder) Discover() error {
	if valid, err := builder.validate(); !valid {
		return err
	}

	glog.V(100).Infof("Pulling existing serviceMeshMemberRoll with name %s under namespace %s from cluster",
		builder.Definition.Name, builder.Definition.Namespace)

	smomr := &istioV1.ServiceMeshMemberRoll{}
	err := builder.apiClient.Get(context.TODO(), goclient.ObjectKey{
		Name:      builder.Definition.Name,
		Namespace: builder.Definition.Namespace,
	}, smomr)

	builder.Object = smomr

	return err
}

// validate will check that the builder and builder definition are properly initialized before
// accessing any member fields.
func (builder *MemberRollBuilder) validate() (bool, error) {
	resourceCRD := "ServiceMeshMemberRoll"

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
