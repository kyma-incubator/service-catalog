package main

import (
	"flag"
	"fmt"

	"github.com/kubernetes-sigs/service-catalog/upgrade/examiner/internal"
	"github.com/kubernetes-sigs/service-catalog/upgrade/examiner/internal/runner"
	"github.com/kubernetes-sigs/service-catalog/upgrade/examiner/pkg/tests/cluster_service_broker"
	"github.com/kubernetes-sigs/service-catalog/upgrade/examiner/pkg/tests/service_broker"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
)

const (
	prepareDataActionName  = "prepareData"
	executeTestsActionName = "executeTests"
)

func registeredTests(cs *internal.ClientStorage) map[string]runner.UpgradeTest {
	return map[string]runner.UpgradeTest{
		"test-broker":         service_broker.NewTestBroker(cs),
		"test-cluster-broker": cluster_service_broker.NewTestBroker(cs),
	}
}

// Config collects all parameters from env variables
type Config struct {
	Local          bool         `envconfig:"default=false"`
	KubeconfigPath string       `envconfig:"optional"`
	KubeConfig     *rest.Config `envconfig:"-"`
	internal.ServiceCatalogConfig
}

// ConfigFlag collects all parameters from flags
type ConfigFlag struct {
	Action string
}

func main() {
	// setup all configurations: envs, flags, stop channel
	flg := readFlags()
	cfg, err := readConfig()
	fatalOnError(err, "while create config")
	stop := internal.SetupChannel()

	// create client storage - struct with all required clients
	cs, err := internal.NewClientStorage(cfg.KubeConfig)
	fatalOnError(err, "while create kubernetes client storage")

	// get tests
	upgradeTests := registeredTests(cs)

	// get runner
	testRunner, err := runner.NewTestRunner(cs.KubernetesClient().CoreV1().Namespaces(), upgradeTests)
	fatalOnError(err, "while creating test runner")

	// launch runner
	switch flg.Action {
	case prepareDataActionName:
		// make sure ServiceCatalog and TestBroker are ready
		ready := internal.NewReadiness(cs, cfg.ServiceCatalogConfig)
		err = ready.TestEnvironmentIsReady()
		fatalOnError(err, "while check ServiceCatalog/TestBroker readiness")

		// prepare data for tests
		err := testRunner.PrepareData(stop)
		fatalOnError(err, "while executing prepare data for all registered tests")
	case executeTestsActionName:
		err := testRunner.ExecuteTests(stop)
		fatalOnError(err, "while executing tests for all registered tests")
	default:
		klog.Fatalf("Unrecognized runner action. Allowed actions: %s or %s.", prepareDataActionName, executeTestsActionName)
	}
}

func fatalOnError(err error, context string) {
	if err != nil {
		klog.Fatalf("%s: %v", context, err)
	}
}

func readFlags() ConfigFlag {
	klog.InitFlags(nil)
	var action string

	flag.StringVar(&action, "action", "", fmt.Sprintf("Define what kind of action runner should execute. Possible values: %s or %s", prepareDataActionName, executeTestsActionName))

	err := flag.Set("logtostderr", "true")
	fatalOnError(err, "while set flag logtostderr")

	err = flag.Set("alsologtostderr", "true")
	fatalOnError(err, "while set flag alsologtostderr")

	flag.Parse()

	return ConfigFlag{
		Action: action,
	}
}

func readConfig() (Config, error) {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(err, "while reading configuration from environment variables")

	if cfg.Local && cfg.KubeconfigPath == "" {
		return cfg, errors.New("KubeconfigPath is required for local mode")
	}

	if cfg.Local {
		cfg.KubeConfig, err = clientcmd.BuildConfigFromFlags("", cfg.KubeconfigPath)
	} else {
		cfg.KubeConfig, err = rest.InClusterConfig()
	}
	if err != nil {
		return cfg, errors.Wrap(err, "while get kubernetes client config")
	}

	return cfg, nil
}
