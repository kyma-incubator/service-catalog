package validation

import (
	"context"
	"fmt"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"
	authenticationapi "k8s.io/api/authentication/v1"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type AdmissionHandler struct {
	decoder *admission.Decoder
	client  client.Client
}

var _ admission.Handler = &AdmissionHandler{}

func (h *AdmissionHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start validation handling operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)

	csb := &sc.ClusterServiceBroker{}
	if err := webhookutil.MatchKinds(csb, req.Kind); err != nil {
		traced.Errorf("Error matching kinds: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := h.decoder.Decode(req, csb); err != nil {
		traced.Errorf("Could not decode request object: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if csb.Spec.AuthInfo == nil {
		traced.Infof("%s %q has no AuthInfo. Operation completed", csb.Kind, csb.Name)
		return admission.Allowed("Validation successful")
	}

	var secretRef *sc.ObjectReference
	if csb.Spec.AuthInfo.Basic != nil {
		secretRef = csb.Spec.AuthInfo.Basic.SecretRef
	} else if csb.Spec.AuthInfo.Bearer != nil {
		secretRef = csb.Spec.AuthInfo.Bearer.SecretRef
	}

	if secretRef == nil {
		traced.Infof("%s %q has no SecretRef neither in Basic nor Bearer auth. Operation completed", csb.Kind, csb.Name)
		return admission.Allowed("Validation successful")
	}

	user := req.UserInfo
	sar := &authorizationapi.SubjectAccessReview{
		Spec: authorizationapi.SubjectAccessReviewSpec{
			ResourceAttributes: &authorizationapi.ResourceAttributes{
				Namespace: secretRef.Namespace,
				Verb:      "get",
				Group:     corev1.SchemeGroupVersion.Group,
				Version:   corev1.SchemeGroupVersion.Version,
				Resource:  corev1.ResourceSecrets.String(),
				Name:      secretRef.Name,
			},
			User:   user.Username,
			Groups: user.Groups,
			Extra:  convertToSARExtra(user.Extra),
			UID:    user.UID,
		},
	}

	err := h.client.Create(ctx, sar)
	if err != nil {
		traced.Errorf("Could not create SubjectAccessReview for %s %q: %v", csb.Kind, csb.Name, err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if !sar.Status.Allowed {
		traced.Infof(
			"Could not handle %s operation for %s %q because SubjectAccessReview has allowed status set to false",
			req.Operation,
			csb.Kind,
			csb.Name,
		)
		return admission.Denied(fmt.Sprintf("Could not %s %s %q", req.Operation, csb.Kind, csb.Name))
	}

	traced.Infof("Completed successfully validation operation: %s for %s: %q", req.Operation, csb.Kind, csb.Name)
	return admission.Allowed("Validation successful")
}

func convertToSARExtra(extra map[string]authenticationapi.ExtraValue) map[string]authorizationapi.ExtraValue {
	if extra == nil {
		return nil
	}

	ret := map[string]authorizationapi.ExtraValue{}
	for k, v := range extra {
		ret[k] = authorizationapi.ExtraValue(v)
	}

	return ret
}

var _ inject.Client = &AdmissionHandler{}

// InjectClient injects the client into the AdmissionHandler
func (h *AdmissionHandler) InjectClient(c client.Client) error {
	h.client = c
	return nil
}

var _ admission.DecoderInjector = &AdmissionHandler{}

// InjectDecoder injects the decoder into the AdmissionHandler
func (h *AdmissionHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}
