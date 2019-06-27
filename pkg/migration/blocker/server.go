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

package blocker

import (
	scTypes "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-sigs/service-catalog/pkg/migration/blocker/validation"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// RunServer runs the webhook server with configuration according to opts
func RunServer(opts *WebhookServerOptions, stopCh <-chan struct{}) error {
	if stopCh == nil {
		/* the caller of RunServer should generate the stop channel
		if there is a need to stop the Webhook server */
		stopCh = make(chan struct{})
	}

	return run(opts, stopCh)
}

func run(opts *WebhookServerOptions, stopCh <-chan struct{}) error {
	cfg := config.GetConfigOrDie()
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		return errors.Wrap(err, "while set up overall controller manager for webhook server")
	}

	err = scTypes.AddToScheme(mgr.GetScheme())
	if err != nil {
		return errors.Wrap(err, "while register Service Catalog scheme into manager")
	}

	// setup webhook server
	webhookSvr := &webhook.Server{
		Port:    opts.SecureServingOptions.BindPort,
		CertDir: opts.SecureServingOptions.ServerCert.CertDirectory,
	}

	webhookSvr.Register("/reject-changes",
		&webhook.Admission{
			Handler: &validation.Handler{},
		})

	// register server
	if err := mgr.Add(webhookSvr); err != nil {
		return errors.Wrap(err, "while registering webhook server with manager")
	}

	// starts the server blocks until the Stop channel is closed
	if err := mgr.Start(stopCh); err != nil {
		return errors.Wrap(err, "while running the webhook manager")

	}

	return nil
}
