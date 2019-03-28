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
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"net/http"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"

	admissionTypes "k8s.io/api/admission/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	strings "strings"
)

type CreateValidator interface {
	Validate(ctx context.Context, cli client.Client, si *sc.ServiceInstance) error
}

type UpdateValidator interface {
	Validate(ctx context.Context, cli client.Client, old, new *sc.ServiceInstance) error
}

// AdmissionHandler handles ServiceInstance
type AdmissionHandler struct {
	decoder *admission.Decoder
	client  client.Client

	CreateValidators []CreateValidator
	UpdateValidators []UpdateValidator
}

var _ admission.Handler = &AdmissionHandler{}
var _ admission.DecoderInjector = &AdmissionHandler{}
var _ inject.Client = &AdmissionHandler{}

// Handle handles admission requests.
func (h *AdmissionHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	traced := webhookutil.NewTracedLogger(req.UID)
	traced.Infof("Start handling AdmissionHandler operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)

	si := &sc.ServiceInstance{}
	if err := webhookutil.MatchKinds(si, req.Kind); err != nil {
		traced.Errorf("Error matching kinds: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := h.decoder.Decode(req, si); err != nil {
		traced.Errorf("Could not decode request object: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	var errs multiError
	switch req.Operation {
	case admissionTypes.Create:
		for _, v := range h.CreateValidators {
			errs = append(errs, v.Validate(ctx,h.client, si))
		}
	case admissionTypes.Update:
		old := &sc.ServiceInstance{}
		if err := h.decoder.DecodeRaw(req.OldObject, si); err != nil {
			traced.Errorf("Could not decode request object: %v", err)
			return admission.Errored(http.StatusBadRequest, err)
		}
		for _, v := range h.UpdateValidators {
			errs = append(errs, v.Validate(ctx,h.client, old, si))
		}
	default:
		traced.Infof("ServiceInstance AdmissionHandler wehbook does not support action %q", req.Operation)
		return admission.Allowed("action not taken")
	}
	if len(errs) > 0 {
		return admission.Denied(errs.Error())
	}

	traced.Infof("Completed successfully AdmissionHandler operation: %s for %s: %q", req.Operation, req.Kind.Kind, req.Name)
	return admission.Allowed("ServiceInstance AdmissionHandler successful")
}

// InjectDecoder injects the decoder
func (h *AdmissionHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// InjectClient injects the client into the CreateUpdateHandler
func (h *AdmissionHandler) InjectClient(c client.Client) error {
	h.client = c
	return nil
}

type multiError []error

func (m multiError) Error() string {
	msgs := make([]string, len(m))
	for _, e := range m {
		msgs = append(msgs, e.Error())
	}
	return strings.join(msgs, "; ")
}