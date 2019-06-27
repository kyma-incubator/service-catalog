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
	genericserveroptions "k8s.io/apiserver/pkg/server/options"
)

const (
	certDirectory            = "/var/run/service-catalog-blocker"
	defaultWebhookServerPort = 8443
	defaultHealthzServerPort = 8080
)

// WebhookServerOptions holds configuration for mutating/validating webhook server.
type WebhookServerOptions struct {
	SecureServingOptions  *genericserveroptions.SecureServingOptions
	ReleaseName           string
	HealthzServerBindPort int
}

// NewWebhookServerOptions creates a new WebhookServerOptions with a default settings.
func NewWebhookServerOptions() *WebhookServerOptions {
	opt := WebhookServerOptions{
		SecureServingOptions: genericserveroptions.NewSecureServingOptions(),
	}

	opt.SecureServingOptions.BindPort = defaultWebhookServerPort
	opt.HealthzServerBindPort = defaultHealthzServerPort
	opt.SecureServingOptions.ServerCert.CertDirectory = certDirectory

	return &opt
}
