package validation_test

import (
	"context"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhook/servicecatalog/servicebinding/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
	"time"
)

const (
	UP_TO_DATE_INSTANCE  = "up-to-date-instance"
	OUT_OF_DATE_INSTANCE = "out-of-date-instance"
)

func TestAdmissionHandler_HandleAllowed(t *testing.T) {
	//Given
	namespace := "test-handler"
	err := sc.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UID:       "1111-aaaa",
			Name:      "test-binding",
			Namespace: namespace,
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceBinding",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{Raw: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBinding",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-binding"
  				},
  				"spec": {
				  "instanceRef": {
					"name": "` + UP_TO_DATE_INSTANCE + `"
				  },
				  "externalID": "123-abc",
				  "secretName": "test-binding"
  				}
			}`)},
		},
	}

	sch, err := sc.SchemeBuilderRuntime.Build()
	require.NoError(t, err)

	decoder, err := admission.NewDecoder(sch)
	require.NoError(t, err)

	tests := map[string]struct {
		operation       admissionv1beta1.Operation
		instanceRefName string
	}{
		"Request for Create ServiceBinding should be allowed": {
			admissionv1beta1.Create,
			UP_TO_DATE_INSTANCE,
		},
		"Request for Update ServiceBinding should be allowed": {
			admissionv1beta1.Update,
			UP_TO_DATE_INSTANCE,
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// Given
			handler := &validation.AdmissionHandler{}

			fakeClient := fake.NewFakeClientWithScheme(sch, &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      test.instanceRefName,
					Namespace: namespace,
				},
			})

			err := handler.InjectDecoder(decoder)
			require.NoError(t, err)
			err = handler.InjectClient(fakeClient)
			require.NoError(t, err)

			request.AdmissionRequest.Operation = test.operation

			// When
			response := handler.Handle(context.TODO(), request)

			// Then
			assert.True(t, response.AdmissionResponse.Allowed)
		})
	}
}

func TestAdmissionHandler_HandleDenied(t *testing.T) {
	//Given
	namespace := "test-handler"
	err := sc.AddToScheme(scheme.Scheme)
	require.NoError(t, err)

	request := admission.Request{
		AdmissionRequest: admissionv1beta1.AdmissionRequest{
			UID:       "2222-bbbb",
			Name:      "test-binding",
			Namespace: namespace,
			Kind: metav1.GroupVersionKind{
				Kind:    "ServiceBinding",
				Version: "v1beta1",
				Group:   "servicecatalog.k8s.io",
			},
			Object: runtime.RawExtension{Raw: []byte(`{
  				"apiVersion": "servicecatalog.k8s.io/v1beta1",
  				"kind": "ServiceBinding",
  				"metadata": {
				  "finalizers": ["kubernetes-incubator/service-catalog"],
  				  "creationTimestamp": null,
  				  "name": "test-binding"
  				},
  				"spec": {
				  "instanceRef": {
					"name": "` + OUT_OF_DATE_INSTANCE + `"
				  },
				  "externalID": "123-abc",
				  "secretName": "test-binding"
  				}
			}`)},
		},
	}

	sch, err := sc.SchemeBuilderRuntime.Build()
	require.NoError(t, err)

	decoder, err := admission.NewDecoder(sch)
	require.NoError(t, err)

	tests := map[string]struct {
		operation       admissionv1beta1.Operation
		instanceRefName string
	}{
		"Request for Create ServiceBinding should be denied": {
			admissionv1beta1.Create,
			OUT_OF_DATE_INSTANCE,
		},
		"Request for Update ServiceBinding should be denied": {
			admissionv1beta1.Update,
			OUT_OF_DATE_INSTANCE,
		},
	}

	for desc, test := range tests {
		t.Run(desc, func(t *testing.T) {
			// Given
			handler := &validation.AdmissionHandler{}

			fakeClient := fake.NewFakeClientWithScheme(sch, &sc.ServiceInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name:              test.instanceRefName,
					Namespace:         namespace,
					DeletionTimestamp: &metav1.Time{time.Now()},
				},
			})

			err := handler.InjectDecoder(decoder)
			require.NoError(t, err)
			err = handler.InjectClient(fakeClient)
			require.NoError(t, err)

			request.AdmissionRequest.Operation = test.operation

			// When
			response := handler.Handle(context.TODO(), request)

			// Then
			assert.False(t, response.AdmissionResponse.Allowed)
		})
	}
}
