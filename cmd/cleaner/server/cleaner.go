/*
Copyright 2017 The Kubernetes Authors.

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

package server

import (
	"fmt"
	"github.com/kubernetes-incubator/service-catalog/pkg/cleaner"
)

// RunCommand executes one of the command from CleanerOptions
func RunCommand(opt *CleanerOptions) error {
	if err := opt.Validate(); nil != err {
		return err
	}

	clr, err := cleaner.NewCleaner()
	if err != nil {
		return fmt.Errorf("failed get new Cleaner: %s", err)
	}

	return clr.RemoveCRDs(opt.ReleaseName, opt.ReleaseNamespace, opt.ControllerManagerName)
}
