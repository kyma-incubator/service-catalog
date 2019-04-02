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
	"fmt"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// ReferenceDeletion handles ServiceBinding validation
type ReferenceDeletion struct {
	decoder *admission.Decoder
	client  client.Client
}

// InjectDecoder injects the decoder
func (h *ReferenceDeletion) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// InjectClient injects the client
func (h *ReferenceDeletion) InjectClient(c client.Client) error {
	h.client = c
	return nil
}

// Validate checks if instance reference for ServiceBinding is not marked for deletion
// fail ServiceBinding operation if the ServiceInstance is marked for deletion
func (h *ReferenceDeletion) Validate(ctx context.Context, req admission.Request, sb *sc.ServiceBinding, traced *webhookutil.TracedLogger) error {
	instanceRef := sb.Spec.InstanceRef
	instance := &sc.ServiceInstance{}

	err := h.client.Get(ctx, types.NamespacedName{Namespace: sb.Namespace, Name: instanceRef.Name}, instance)
	if err != nil {
		traced.Errorf("Could not get ServiceInstance by name %q: %v", instanceRef.Name, err)
		return err
	}

	if instance.DeletionTimestamp != nil {
		traced.Infof(
			"Could not handle %s operation for %s because ServiceInstance %s is marked for deletion",
			req.Operation,
			req.Kind.Kind,
			instanceRef.Name)
		return fmt.Errorf("could not %s %s %q", req.Operation, sb.Kind, sb.Name)
	}

	return nil
}
