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

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/webhookutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// DenyPlanChangeIfNotUpdatable checks if Plan can be changed
func (h *AdmissionHandler) DenyPlanChangeIfNotUpdatable(ctx context.Context, req admission.Request, si *sc.ServiceInstance, traced *webhookutil.TracedLogger) error {

	if si.Spec.ClusterServiceClassRef == nil {
		traced.Infof("Service class does not exist")
		return nil // user chose a service class that doesn't exist
	}

	csc := &sc.ClusterServiceClass{}
	if err := h.client.Get(ctx, types.NamespacedName{
		Namespace: "",
		Name:      si.Spec.ClusterServiceClassRef.Name,
	}, csc); err != nil {
		traced.Infof("Could not locate service class '%v', can not determine if UpdateablePlan.", si.Spec.ClusterServiceClassRef.Name)
		return err
	}

	if csc.Spec.PlanUpdatable {
		traced.Info("AdmissionHandler passed - UpdateablePlan is set to true.")
		return nil
	}

	if si.Spec.GetSpecifiedClusterServicePlan() != "" {
		origInstance := &sc.ServiceInstance{}
		if err := h.decoder.DecodeRaw(req.OldObject, origInstance); err != nil {
			traced.Errorf("Could not decode oldObject: %v", err)
			return err
		}

		externalPlanNameUpdated := si.Spec.ClusterServicePlanExternalName != origInstance.Spec.ClusterServicePlanExternalName
		externalPlanIDUpdated := si.Spec.ClusterServicePlanExternalID != origInstance.Spec.ClusterServicePlanExternalID
		k8sPlanUpdated := si.Spec.ClusterServicePlanName != origInstance.Spec.ClusterServicePlanName
		if externalPlanNameUpdated || externalPlanIDUpdated || k8sPlanUpdated {
			var oldPlan, newPlan string
			if externalPlanNameUpdated {
				oldPlan = origInstance.Spec.ClusterServicePlanExternalName
				newPlan = si.Spec.ClusterServicePlanExternalName
			} else if externalPlanIDUpdated {
				oldPlan = origInstance.Spec.ClusterServicePlanExternalID
				newPlan = si.Spec.ClusterServicePlanExternalID
			} else {
				oldPlan = origInstance.Spec.ClusterServicePlanName
				newPlan = si.Spec.ClusterServicePlanName
			}
			traced.Infof("update Service Instance %v/%v request specified Plan %v while original instance had %v", si.Namespace, si.Name, newPlan, oldPlan)
			msg := fmt.Sprintf("The Service Class %v does not allow plan changes.", csc.Name)
			traced.Error(msg)
			return errors.New(msg)
		}
	}

	return nil
}
