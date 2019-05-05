package server

import (
	"fmt"
	"github.com/spf13/pflag"
)

const (
	removeCRD                        = "remove-crd"
	releaseName                      = "release-name"
	serviceCatalogNamespaceParameter = "service-catalog-namespace"
	controllerManagerNameParameter   = "controller-manager-deployment"
)

// CleanerOptions holds configuration for cleaner jobs
type CleanerOptions struct {
	Command               string
	ReleaseName           string
	ReleaseNamespace      string
	ControllerManagerName string
}

// NewCleanerOptions creates and returns a new CleanerOptions
func NewCleanerOptions() *CleanerOptions {
	return &CleanerOptions{}
}

// AddFlags adds flags for a CleanerOptions to the specified FlagSet.
func (c *CleanerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.Command, "cleaner-command", removeCRD, "Command name to execute")
	fs.StringVar(&c.ReleaseName, releaseName, "", "Name of ServiceCatalog release used in helm")
	fs.StringVar(&c.ReleaseNamespace, serviceCatalogNamespaceParameter, "", "Name of namespace where Service Catalog is released")
	fs.StringVar(&c.ControllerManagerName, controllerManagerNameParameter, "", "Name of controller manager deployment")
}

// Validate checks flag has been set and has a proper value
func (c *CleanerOptions) Validate() error {
	if c.Command != removeCRD {
		return fmt.Errorf("Command %q is not supported", c.Command)
	}
	for name, value := range map[string]string{
		releaseName:                      c.ReleaseName,
		serviceCatalogNamespaceParameter: c.ReleaseNamespace,
		controllerManagerNameParameter:   c.ControllerManagerName,
	} {
		if value == "" {
			return fmt.Errorf("command %q requires three parameters, one of them (%q) is empty", removeCRD, name)
		}
	}

	return nil
}
