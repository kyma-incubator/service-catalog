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

// CreateUpdateHandler handles ClusterServicePlann
type CreateUpdateHandler struct {
	decoder *admission.Decoder
}

var _ admission.Handler = &CreateUpdateHandler{}


// Handle handles admission requests.
func (h *CreateUpdateHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	sb := &sc.ClusterServicePlan{}

	if err := h.decoder.Decode(req, sb); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	mutated := sb.DeepCopy()
	switch req.Operation {
	case admissionTypes.Create:
		h.mutateOnCreate(ctx, req, mutated)
	case admissionTypes.Update:
		h.mutateOnUpdate(ctx, mutated)
	default:
		klog.Warning("ClusterServicePlan mutation wehbook does not support action %q", req.Operation)
	}
	h.syncLabels(mutated)
	rawMutated, err := json.Marshal(mutated)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.AdmissionRequest.Object.Raw, rawMutated)
}

var _ admission.DecoderInjector = &CreateUpdateHandler{}

// InjectDecoder injects the decoder
func (h *CreateUpdateHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *CreateUpdateHandler) mutateOnCreate(ctx context.Context, req admission.Request, binding *sc.ClusterServicePlan) {

}

func (h *CreateUpdateHandler) mutateOnUpdate(ctx context.Context, obj *sc.ClusterServicePlan) {
	// TODO: implement logic from pkg/registry/servicecatalog/binding/strategy.go
}

func (h *CreateUpdateHandler) syncLabels(obj *sc.ClusterServicePlan) {
	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}

	obj.Labels[sc.GroupName+"/spec.externalID"] = obj.Spec.ExternalID
	obj.Labels[sc.GroupName+"/spec.externalName"] = obj.Spec.ExternalName
	obj.Labels[sc.GroupName+"/spec.clusterServiceClassRef.name"] = obj.Spec.ClusterServiceClassRef.Name
	obj.Labels[sc.GroupName+"/spec.clusterServiceBrokerName"] = obj.Spec.ClusterServiceBrokerName
}
