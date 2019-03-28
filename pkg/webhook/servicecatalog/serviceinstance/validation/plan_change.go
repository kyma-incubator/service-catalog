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


func (h *AdmissionHandler) ValidatePlanUpdate(ctx context.Context, req admission.Request, si *sc.ServiceInstance, traced *webhookutil.TracedLogger) error {
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
		h.decoder.DecodeRaw(req.OldObject, origInstance)
		if err := h.decoder.Decode(req, si); err != nil {
			traced.Errorf("Could not decode request oldObject: %v", err)
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