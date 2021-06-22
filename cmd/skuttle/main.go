package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vixus0/skuttle/v2/internal/controller"
	"github.com/vixus0/skuttle/v2/internal/logging"
	"github.com/vixus0/skuttle/v2/internal/provider"
	"github.com/vixus0/skuttle/v2/internal/provider/aws"
	"github.com/vixus0/skuttle/v2/internal/provider/file"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	log = logging.NewLogger("main")
)

func main() {
	// Startup flags
	var (
		argDryRun           bool
		argLogLevel         string
		argKubeconfig       string
		argNodeSelector     string
		argNotReadyDuration time.Duration
		argRefreshDuration  time.Duration
		argProviders        string
	)

	flag.BoolVar(&argDryRun, "dry-run", BoolEnv("DRY_RUN", false),
		"dry run mode to only log instead of scheduling deletion",
	)

	flag.StringVar(&argLogLevel, "log-level", StringEnv("LOG_LEVEL", "info"),
		"log level (debug, info, warn, error)",
	)

	flag.StringVar(&argKubeconfig, "kubeconfig", StringEnv("KUBECONFIG", ""),
		"path to kubeconfig file if not running in-cluster",
	)

	flag.StringVar(&argNodeSelector, "node-selector", StringEnv("NODE_SELECTOR", "node.kubernetes.io/node"),
		"selector used to filter nodes skuttle should manage",
	)

	flag.DurationVar(&argNotReadyDuration, "not-ready-duration", DurationEnv("NOT_READY_DURATION", "10m"),
		"time duration to tolerate NotReady nodes",
	)

	flag.DurationVar(&argRefreshDuration, "refresh-duration", DurationEnv("REFRESH_DURATION", "10s"),
		"refresh duration",
	)

	flag.StringVar(&argProviders, "providers", StringEnv("PROVIDERS", ""),
		"comma-separated list of enabled providers",
	)

	flag.Parse()

	// Set log level
	switch argLogLevel {
	case "debug":
		logging.SetLevel(logging.DEBUG)
	case "info":
		logging.SetLevel(logging.INFO)
	case "warn":
		logging.SetLevel(logging.WARN)
	case "error":
		logging.SetLevel(logging.ERROR)
	default:
		log.Fatalf("unknown log level: %s", argLogLevel)
	}

	// Create Kubernetes client
	log.Info("init")

	if argKubeconfig != "" {
		log.Info("using config from: %s\n", argKubeconfig)
	} else {
		log.Info("assuming we're running in-cluster")
	}

	config, err := clientcmd.BuildConfigFromFlags("", argKubeconfig)
	if err != nil {
		log.Fatalf("could not build kubeconfig: %v", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("could not create kube client: %v", err.Error())
	}

	// Define cancellable context
	ctx := signals.SetupSignalHandler()

	// Handle Kube API crashes
	defer runtime.HandleCrash()

	// Populate store of cloud instance providers
	cleanProviders := strings.TrimSpace(argProviders)
	if cleanProviders == "" {
		log.Fatal("No providers specified!")
	}

	providerStore := &provider.DefaultStore{}
	providerPrefixes := strings.Split(cleanProviders, ",")

	for _, prefix := range providerPrefixes {
		var (
			err error
			p   provider.Provider
		)

		switch prefix {
		case "aws":
			p, err = aws.NewProvider(ctx)
		case "file":
			p, err = file.NewProvider()
		default:
			log.Fatalf("no provider available for %s", prefix)
		}

		if err != nil {
			log.Fatalf("error creating provider %v: %v", prefix, err)
		}

		providerStore.Add(prefix, p)
	}

	// Create node informer
	tweakListOptions := informers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = argNodeSelector
	})
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientset, argRefreshDuration, tweakListOptions)
	nodeInformer := informerFactory.Core().V1().Nodes().Informer()

	// Create controller
	cfg := &controller.Config{
		DryRun:           argDryRun,
		NotReadyDuration: argNotReadyDuration,
		Providers:        providerStore,
	}
	nodeClient := clientset.CoreV1().Nodes()
	controller.NewController(cfg, ctx, nodeClient, nodeInformer)

	// Start all informers created by factory
	informerFactory.Start(ctx.Done())

	// Wait for informer to sync
	log.Info("wait for sync")
	if !cache.WaitForCacheSync(ctx.Done(), nodeInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
	}

	log.Info("starting")
	<-ctx.Done()
	if err = ctx.Err(); err != nil {
		runtime.HandleError(err)
	}
}

func StringEnv(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func BoolEnv(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		ret, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatal(err)
		}
		return ret
	}
	return defaultVal
}

func DurationEnv(key string, defaultVal string) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok {
		val = defaultVal
	}

	ret, err := time.ParseDuration(val)
	if err != nil {
		log.Fatal(err)
	}

	return ret
}

func IntEnv(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		ret, err := strconv.Atoi(val)
		if err != nil {
			log.Fatal(err)
		}
		return ret
	}
	return defaultVal
}
