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

package migration

import (
	"fmt"
	restclient "k8s.io/client-go/rest"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	k8sClientSet "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"github.com/kubernetes-incubator/service-catalog/pkg/migration"
	"k8s.io/klog"
)

func RunCommand(opt *MigrationOptions) error {
	if err := opt.Validate(); nil != err {
		return err
	}

	restConfig, err := newRestClientConfig()
	if err != nil {
		return fmt.Errorf("failed to get Kubernetes client config: %s", err)
	}

	scClient, err := sc.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create Service Catalog client: %s", err)
	}
	scInterface := scClient.ServicecatalogV1beta1()
	k8sCli, err := k8sClientSet.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %s", err)
	}

	svc := migration.NewMigrationService(scInterface, opt.StoragePath, k8sCli.CoreV1())
	scalingSvc := migration.NewScalingService(opt.ReleaseNamespace, opt.ControllerManagerName, k8sCli.AppsV1())

	switch opt.Action {
	case backupActionName:
		klog.Infoln("Executing backup action")
		err := scalingSvc.ScaleDown()
		if err != nil {
			return err
		}

		res, err := svc.BackupResources()
		if err != nil {
			return err
		}

		err = svc.RemoveOwnerReferenceFromSecrets()
		if err != nil {
			return err
		}

		return svc.Cleanup(res)
	case restoreActionName:
		klog.Infoln("Executing restore action")
		err := scalingSvc.ScaleDown()
		if err != nil {
			return err
		}

		res, err := svc.LoadResources()
		if err != nil {
			return err
		}

		err = svc.Restore(res)
		if err != nil {
			return err
		}

		err = scalingSvc.ScaleUp()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown action %s\n", opt.Action)
	}
	return nil
}

func newRestClientConfig() (*restclient.Config, error) {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	}

	klog.V(4).Info("KUBECONFIG not defined, creating in-cluster config")
	return restclient.InClusterConfig()
}
