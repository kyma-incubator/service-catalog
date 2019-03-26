package validation

import (
	"context"
	"fmt"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type ValidationHandler struct {
	decoder *admission.Decoder
	client client.Client
}

var _ admission.Handler = &ValidationHandler{}

// Handle checks if instance reference for ServiceBinding is not marked for deletion
// fail Create/Update ServiceBinding operation if the ServiceInstance is marked for deletion
func (v *ValidationHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start validation handling operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)

	sb := &sc.ServiceBinding{}
	if err := webhookutil.MatchKinds(sb, req.Kind); err != nil {
		traced.Errorf("Error matching kinds: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := v.decoder.Decode(req, sb); err != nil {
		traced.Errorf("Could not decode request object: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	instanceRef := sb.Spec.InstanceRef
	instance := &sc.ServiceInstance{}

	err := v.client.Get(ctx, types.NamespacedName{Namespace: sb.Namespace, Name: instanceRef.Name}, instance)
	if err != nil {
		traced.Errorf("Could not get ServiceInstance by name %q: %v", instanceRef.Name, err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if instance.DeletionTimestamp != nil {
		traced.Infof(
			"Could not handle %s operation for %s because ServiceInstance %s is marked for deletion",
			req.Operation,
			req.Kind.Kind,
			instanceRef.Name)
		return admission.Denied(fmt.Sprintf("Could not %s ServiceBinding %q", req.Operation, sb.Name))
	}

	traced.Infof("Completed successfully validation operation: %s for %s: %q", req.Operation, req.Kind.Kind, sb.Name)
	return admission.Allowed("Validation successful")
}

var _ inject.Client = &ValidationHandler{}

// InjectClient injects the client into the ValidationHandler
func (v *ValidationHandler) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

var _ admission.DecoderInjector = &ValidationHandler{}

// InjectDecoder injects the decoder into the ValidationHandler
func (v *ValidationHandler) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
