package clusterlogging

import (
	"fmt"
	lokiv1 "github.com/grafana/loki/operator/apis/loki/v1"
	"github.com/openshift-kni/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

var (
	defaultLokiStackName      = "lokistack-test"
	defaultLokiStackNamespace = "lokistack-space"
)

func TestPullLokiStack(t *testing.T) {
	generateLokiStack := func(name, namespace string) *lokiv1.LokiStack {
		return &lokiv1.LokiStack{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
	}

	testCases := []struct {
		name                string
		namespace           string
		addToRuntimeObjects bool
		expectedError       error
		client              bool
	}{
		{
			name:                defaultLokiStackName,
			namespace:           defaultLokiStackNamespace,
			addToRuntimeObjects: true,
			expectedError:       nil,
			client:              true,
		},
		{
			name:                "",
			namespace:           defaultLokiStackNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("lokiStack 'name' cannot be empty"),
			client:              true,
		},
		{
			name:                defaultLokiStackName,
			namespace:           "",
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("lokiStack 'nsname' cannot be empty"),
			client:              true,
		},
		{
			name:                "lokitest",
			namespace:           defaultLokiStackNamespace,
			addToRuntimeObjects: false,
			expectedError: fmt.Errorf("lokiStack object lokitest does not exist " +
				"in namespace lokistack-space"),
			client: true,
		},
		{
			name:                "triggerauthtest",
			namespace:           defaultLokiStackNamespace,
			addToRuntimeObjects: true,
			expectedError:       fmt.Errorf("lokiStack 'apiClient' cannot be empty"),
			client:              false,
		},
	}

	for _, testCase := range testCases {
		// Pre-populate the runtime objects
		var runtimeObjects []runtime.Object

		var testSettings *clients.Settings

		testLokiStack := generateLokiStack(testCase.name, testCase.namespace)

		if testCase.addToRuntimeObjects {
			runtimeObjects = append(runtimeObjects, testLokiStack)
		}

		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects: runtimeObjects,
			})
		}

		builderResult, err := PullLokiStack(testSettings, testCase.name, testCase.namespace)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError != nil {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		} else {
			assert.Equal(t, testLokiStack.Name, builderResult.Object.Name)
			assert.Equal(t, testLokiStack.Namespace, builderResult.Object.Namespace)
			assert.Nil(t, err)
		}
	}
}

func TestNewLokiStackBuilder(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		expectedError string
		client        bool
	}{
		{
			name:          defaultLokiStackName,
			namespace:     defaultLokiStackNamespace,
			expectedError: "",
			client:        true,
		},
		{
			name:          "",
			namespace:     defaultLokiStackNamespace,
			expectedError: "lokiStack 'name' cannot be empty",
			client:        true,
		},
		{
			name:          defaultLokiStackName,
			namespace:     "",
			expectedError: "lokiStack 'nsname' cannot be empty",
			client:        true,
		},
		{
			name:          defaultLokiStackName,
			namespace:     defaultLokiStackNamespace,
			expectedError: "",
			client:        false,
		},
	}

	for _, testCase := range testCases {
		var testSettings *clients.Settings
		if testCase.client {
			testSettings = clients.GetTestClients(clients.TestClientParams{})
		}

		testLokiStackBuilder := NewLokiStackBuilder(testSettings, testCase.name, testCase.namespace)

		if testCase.expectedError == "" {
			if testCase.client {
				assert.Equal(t, testCase.name, testLokiStackBuilder.Definition.Name)
				assert.Equal(t, testCase.namespace, testLokiStackBuilder.Definition.Namespace)
			} else {
				assert.Nil(t, testLokiStackBuilder)
			}
		} else {
			assert.Equal(t, testCase.expectedError, testLokiStackBuilder.errorMsg)
			assert.NotNil(t, testLokiStackBuilder.Definition)
		}
	}
}

func TestLokiStackExists(t *testing.T) {
	testCases := []struct {
		testLokiStack  *LokiStackBuilder
		expectedStatus bool
	}{
		{
			testLokiStack:  buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedStatus: true,
		},
		{
			testLokiStack:  buildInValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedStatus: false,
		},
		{
			testLokiStack:  buildValidLokiStackBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedStatus: false,
		},
	}

	for _, testCase := range testCases {
		exist := testCase.testLokiStack.Exists()
		assert.Equal(t, testCase.expectedStatus, exist)
	}
}

func TestLokiStackGet(t *testing.T) {
	testCases := []struct {
		testLokiStack *LokiStackBuilder
		expectedError error
	}{
		{
			testLokiStack: buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testLokiStack: buildInValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: fmt.Errorf("lokiStack 'name' cannot be empty"),
		},
		{
			testLokiStack: buildValidLokiStackBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: fmt.Errorf("lokistacks.loki.grafana.com \"lokistack-test\" not found"),
		},
	}

	for _, testCase := range testCases {
		lokiStackObj, err := testCase.testLokiStack.Get()

		if testCase.expectedError == nil {
			assert.Equal(t, lokiStackObj.Name, testCase.testLokiStack.Definition.Name)
			assert.Equal(t, lokiStackObj.Namespace, testCase.testLokiStack.Definition.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestLokiStackCreate(t *testing.T) {
	testCases := []struct {
		testLokiStack *LokiStackBuilder
		expectedError string
	}{
		{
			testLokiStack: buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: "",
		},
		{
			testLokiStack: buildInValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: "lokiStack 'name' cannot be empty",
		},
		{
			testLokiStack: buildValidLokiStackBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: "",
		},
	}

	for _, testCase := range testCases {
		testLokiStackBuilder, err := testCase.testLokiStack.Create()

		if testCase.expectedError == "" {
			assert.Equal(t, testLokiStackBuilder.Definition.Name, testLokiStackBuilder.Object.Name)
			assert.Equal(t, testLokiStackBuilder.Definition.Namespace, testLokiStackBuilder.Object.Namespace)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError, err.Error())
		}
	}
}

func TestLokiStackDelete(t *testing.T) {
	testCases := []struct {
		testLokiStack *LokiStackBuilder
		expectedError error
	}{
		{
			testLokiStack: buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: nil,
		},
		{
			testLokiStack: buildValidLokiStackBuilder(clients.GetTestClients(clients.TestClientParams{})),
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		_, err := testCase.testLokiStack.Delete()

		if testCase.expectedError == nil {
			assert.Nil(t, testCase.testLokiStack.Object)
			assert.Nil(t, err)
		} else {
			assert.Equal(t, testCase.expectedError.Error(), err.Error())
		}
	}
}

func TestLokiStackUpdate(t *testing.T) {
	testCases := []struct {
		testLokiStack *LokiStackBuilder
		expectedError string
		testSize      lokiv1.LokiStackSizeType
	}{
		{
			testLokiStack: buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject()),
			expectedError: "",
			testSize:      lokiv1.SizeOneXDemo,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, lokiv1.LokiStackSizeType(""), testCase.testLokiStack.Definition.Spec.Size)
		assert.Nil(t, nil, testCase.testLokiStack.Object)
		testCase.testLokiStack.WithSize(testCase.testSize)
		_, err := testCase.testLokiStack.Update()

		if testCase.expectedError != "" {
			assert.Equal(t, testCase.expectedError, err.Error())
		} else {
			assert.Equal(t, testCase.testSize, testCase.testLokiStack.Definition.Spec.Size)
		}
	}
}

func TestLokiStackWithSize(t *testing.T) {
	testCases := []struct {
		testSize          lokiv1.LokiStackSizeType
		expectedError     bool
		expectedErrorText string
	}{
		{
			testSize:          lokiv1.SizeOneXDemo,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testSize:          lokiv1.SizeOneXSmall,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testSize:          lokiv1.SizeOneXMedium,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testSize:          lokiv1.SizeOneXExtraSmall,
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testSize:          "",
			expectedError:     false,
			expectedErrorText: "'size' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithSize(testCase.testSize)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testSize, result.Definition.Spec.Size)
		}
	}
}

func TestLokiStackWithStorage(t *testing.T) {
	testCases := []struct {
		testStorage       lokiv1.ObjectStorageSpec
		expectedError     bool
		expectedErrorText string
	}{
		{
			testStorage: lokiv1.ObjectStorageSpec{
				Secret: lokiv1.ObjectStorageSecretSpec{
					Type: "s3",
					Name: "test",
				},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithStorage(testCase.testStorage)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testStorage, result.Definition.Spec.Storage)
		}
	}
}

func TestLokiStackWithStorageClassName(t *testing.T) {
	testCases := []struct {
		testStorageClassName string
		expectedError        bool
		expectedErrorText    string
	}{
		{
			testStorageClassName: "gp2",
			expectedError:        false,
			expectedErrorText:    "",
		},
		{
			testStorageClassName: "",
			expectedError:        false,
			expectedErrorText:    "'storageClassName' argument cannot be empty",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithStorageClassName(testCase.testStorageClassName)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testStorageClassName, result.Definition.Spec.StorageClassName)
		}
	}
}

func TestLokiStackWithTenants(t *testing.T) {
	testCases := []struct {
		testTenants       lokiv1.TenantsSpec
		expectedError     bool
		expectedErrorText string
	}{
		{
			testTenants: lokiv1.TenantsSpec{
				Mode: "openshift-logging",
			},
			expectedError:     false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithTenants(testCase.testTenants)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testTenants.Mode, result.Definition.Spec.Tenants.Mode)
		}
	}
}

func TestLokiStackWithRules(t *testing.T) {
	testCases := []struct {
		testRules         lokiv1.RulesSpec
		expectedError     bool
		expectedErrorText string
	}{
		{
			testRules: lokiv1.RulesSpec{
				Enabled: true,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"openshift.io/cluster-monitoring": "true"},
				},
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"openshift.io/cluster-monitoring": "true"},
				},
			},
			expectedError:     false,
			expectedErrorText: "",
		},
		{
			testRules: lokiv1.RulesSpec{
				Enabled: false,
			},
			expectedError:     false,
			expectedErrorText: "",
		},
	}

	for _, testCase := range testCases {
		testBuilder := buildValidLokiStackBuilder(buildLokiStackClientWithDummyObject())

		result := testBuilder.WithRules(testCase.testRules)

		if testCase.expectedError {
			if testCase.expectedErrorText != "" {
				assert.Equal(t, testCase.expectedErrorText, result.errorMsg)
			}
		} else {
			assert.NotNil(t, result)
			assert.Equal(t, testCase.testRules.Enabled, result.Definition.Spec.Rules.Enabled)
			if testCase.testRules.Enabled {
				assert.Equal(t, testCase.testRules.Selector, result.Definition.Spec.Rules.Selector)
				assert.Equal(t, testCase.testRules.NamespaceSelector, result.Definition.Spec.Rules.NamespaceSelector)
			}
		}
	}
}

func buildValidLokiStackBuilder(apiClient *clients.Settings) *LokiStackBuilder {
	triggerAuthBuilder := NewLokiStackBuilder(
		apiClient, defaultLokiStackName, defaultLokiStackNamespace)

	return triggerAuthBuilder
}

func buildInValidLokiStackBuilder(apiClient *clients.Settings) *LokiStackBuilder {
	triggerAuthBuilder := NewLokiStackBuilder(
		apiClient, "", defaultLokiStackNamespace)

	return triggerAuthBuilder
}

func buildLokiStackClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects: buildDummyTriggerAuthentication(),
	})
}

func buildDummyTriggerAuthentication() []runtime.Object {
	return append([]runtime.Object{}, &lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultLokiStackName,
			Namespace: defaultLokiStackNamespace,
		},
	})
}
