/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mutation

import (
	"context"
	"encoding/json"
	"net/http"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"

	admissionTypes "k8s.io/api/admission/v1beta1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// CreateUpdateHandler handles ClusterServiceBroker
type CreateUpdateHandler struct {
	// To use the client, you need to do the following:
	// - uncomment it
	// - "sigs.k8s.io/controller-runtime/pkg/client"
	// - "sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	// - uncomment the InjectClient method at the bottom of this file.
	//client client.Client

	// Decoder decodes objects
	decoder *admission.Decoder
}

var _ admission.Handler = &CreateUpdateHandler{}

// Handle handles admission requests.
func (h *CreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	cb := &sc.ClusterServiceBroker{}
	if err := h.decoder.Decode(req, cb); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	mutated := cb.DeepCopy()
	switch req.Operation {
	case admissionTypes.Create:
		h.mutateOnCreate(ctx, mutated)
	case admissionTypes.Update:
		h.mutateOnUpdate(ctx, mutated)
	default:
		klog.Warning("ClusterServiceBroker mutation wehbook does not support action %q", req.Operation)
	}

	rawMutated, err := json.Marshal(mutated)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, rawMutated)
}

//var _ inject.Client = &CreateUpdateHandler{}
//
//// InjectClient injects the client into the CreateUpdateHandler
//func (h *CreateUpdateHandler) InjectClient(c client.Client) error {
//	h.client = c
//	return nil
//}

var _ admission.DecoderInjector = &CreateUpdateHandler{}

// InjectDecoder injects the decoder into the CreateUpdateHandler
func (h *CreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *CreateUpdateHandler) mutateOnCreate(ctx context.Context, obj *sc.ClusterServiceBroker) {
	obj.Finalizers = []string{sc.FinalizerServiceCatalog}

	obj.Spec.RelistBehavior = sc.ServiceBrokerRelistBehaviorDuration
}

func (h *CreateUpdateHandler) mutateOnUpdate(ctx context.Context, obj *sc.ClusterServiceBroker) {
	// TODO: implement logic from pkg/registry/servicecatalog/clusterservicebroker/strategy.go
}
