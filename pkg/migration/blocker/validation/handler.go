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

package validation

import (
	"context"
	"github.com/kubernetes-sigs/service-catalog/pkg/webhookutil"

	admissionTypes "k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Handler handles Service Catalog CRDs
type Handler struct {
}

var _ admission.Handler = &Handler{}

// Handle handles admission requests.

func (h *Handler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start handling validation operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)

	switch req.Operation {
	case admissionTypes.Create, admissionTypes.Delete, admissionTypes.Update:
		return admission.Denied("Operation denied - migration is in progress")
	default:
		traced.Infof("Validation wehbook does not support action %q", req.Operation)
		return admission.Allowed("action not taken")
	}
}
