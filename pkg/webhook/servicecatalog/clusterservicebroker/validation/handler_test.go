package validation_test

import (
	"context"
	"errors"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/clusterservicebroker/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)

const (
	ALLOWED_SECRET_NAME = "csb-secret-name"
	DENIED_SECRET_NAME  = "denied-csb-secret-name"
)

// Reactors are not implemented in 'sigs.k8s.io/controller-runtime/pkg/client/fake' package
// https://github.com/kubernetes-sigs/controller-runtime/issues/72
// instead it is used custom client with override Create method
type fakedClient struct {
	client.Client
}

func (m *fakedClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOptionFunc) error {
	if _, ok := obj.(*v1.SubjectAccessReview); !ok {
		return errors.New("Input object is not SubjectAccessReview type")
	}

	if obj.(*v1.SubjectAccessReview).Spec.ResourceAttributes.Name == ALLOWED_SECRET_NAME {
		obj.(*v1.SubjectAccessReview).Status.Allowed = true
	}

	return nil
}

func TestAdmissionHandler_HandleAllowed(t *testing.T) {
	//Given
	err := sc.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UID:       "3333-cccc",
			Name:      "test-broker",
			Namespace: "test-handler",
			Kind: metav1.GroupVersionKind{
				Kind:    "ClusterServiceBroker",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{},
		},
	}

	decoder, err := admission.NewDecoder(scheme.Scheme)
	require.NoError(t, err)

	tests := map[string]struct {
		operation admissionv1beta1.Operation
		object    []byte
	}{
		"Request for Create ClusterServiceBroker should be allowed": {
			admissionv1beta1.Create,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local"
  				}
			}`),
		},
		"Request for Update ClusterServiceBroker should be allowed": {
			admissionv1beta1.Update,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local"
  				}
			}`),
		},
		"Request for Create ClusterServiceBroker with AuthInfo should be allowed": {
			admissionv1beta1.Create,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local",
				  "authInfo": {
    			    "basic": {
      				  "secretRef": {
        			    "namespace": "test-handler",
						"name": "` + ALLOWED_SECRET_NAME + `"
					  }
					}
				  }
  				}
			}`),
		},
		"Request for Update ClusterServiceBroker with AuthInfo should be allowed": {
			admissionv1beta1.Update,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local",
				  "authInfo": {
    			    "bearer": {
      				  "secretRef": {
        			    "namespace": "test-handler",
						"name": "` + ALLOWED_SECRET_NAME + `"
					  }
					}
				  }
				}
			}`),
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// Given
			handler := &validation.AdmissionHandler{}

			err := handler.InjectDecoder(decoder)
			require.NoError(t, err)
			err = handler.InjectClient(&fakedClient{})
			require.NoError(t, err)

			request.AdmissionRequest.Operation = test.operation
			request.AdmissionRequest.Object.Raw = test.object

			// When
			response := handler.Handle(context.TODO(), request)

			// Then
			assert.True(t, response.AdmissionResponse.Allowed)
		})
	}
}

func TestAdmissionHandler_HandleDenied(t *testing.T) {
	//Given
	err := sc.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UID:       "4444-dddd",
			Name:      "test-broker",
			Namespace: "test-handler",
			Kind: metav1.GroupVersionKind{
				Kind:    "ClusterServiceBroker",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{},
		},
	}

	sch := scheme.Scheme

	decoder, err := admission.NewDecoder(sch)
	require.NoError(t, err)

	tests := map[string]struct {
		operation admissionv1beta1.Operation
		object    []byte
	}{
		"Request for Create ClusterServiceBroker should be denied": {
			admissionv1beta1.Create,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local",
				  "authInfo": {
    			    "bearer": {
      				  "secretRef": {
        			    "namespace": "test-handler",
						"name": "` + DENIED_SECRET_NAME + `"
					  }
					}
				  }
  				}
			}`),
		},
		"Request for Update ClusterServiceBroker should be denied": {
			admissionv1beta1.Update,
			[]byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ClusterServiceBroker",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-broker"
  				},
  				"spec": {
				  "url": "http://test-broker.local",
				  "authInfo": {
    			    "basic": {
      				  "secretRef": {
        			    "namespace": "test-handler",
						"name": "` + DENIED_SECRET_NAME + `"
					  }
					}
				  }
  				}
			}`),
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// Given
			handler := &validation.AdmissionHandler{}

			fakeClient := fake.NewFakeClientWithScheme(sch, &sc.ClusterServiceBroker{})

			err := handler.InjectDecoder(decoder)
			require.NoError(t, err)
			err = handler.InjectClient(fakeClient)
			require.NoError(t, err)

			request.AdmissionRequest.Operation = test.operation
			request.AdmissionRequest.Object.Raw = test.object

			// When
			response := handler.Handle(context.TODO(), request)

			// Then
			assert.False(t, response.AdmissionResponse.Allowed)
		})
	}
}
