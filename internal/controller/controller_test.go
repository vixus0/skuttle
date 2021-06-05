package controller_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vixus0/skuttle/v2/internal/controller"
	"github.com/vixus0/skuttle/v2/internal/logging"
	"github.com/vixus0/skuttle/v2/internal/provider"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	//clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

var _ = Describe("Controller", func() {
	var (
		client       kubernetes.Interface
		deletedNodes []string
	)

	logging.SetLevel(logging.DEBUG)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the fake client.
	client = fake.NewSimpleClientset()

	// We will create an informer that writes added nodes to a channel.
	factory := informers.NewSharedInformerFactory(client, 0)

	// Create node informer that will also collect deleted nodes
	nodeInformer := factory.Core().V1().Nodes().Informer()
	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			node := obj.(*v1.Node)
			deletedNodes = append(deletedNodes, node.ObjectMeta.Name)
		},
	})

	// populate fake cloud provider
	fakeProvider := &FakeProvider{
		Nodes: map[string]bool{
			"node-ready":           true,
			"node-unready-below":   false,
			"node-unready-above":   false,
			"node-unready-exists":  true,
			"node-unready-missing": false,
		},
	}

	// Setup controller
	providerStore := &provider.DefaultStore{}
	providerStore.Add("fake", fakeProvider)
	cfg := &controller.Config{
		DryRun:           false,
		NotReadyDuration: 10 * time.Minute,
		Providers:        providerStore,
	}

	// add nodes to kube API
	AddNode(client, FakeNode{
		Name:           "node-ready",
		Ready:          true,
		TransitionTime: time.Now(),
	})
	AddNode(client, FakeNode{
		Name:           "node-unready-below",
		Ready:          false,
		TransitionTime: time.Now().Add(-5 * time.Minute),
	})
	AddNode(client, FakeNode{
		Name:           "node-unready-exists",
		Ready:          false,
		TransitionTime: time.Now().Add(-15 * time.Minute),
	})
	AddNode(client, FakeNode{
		Name:           "node-unready-missing",
		Ready:          false,
		TransitionTime: time.Now().Add(-15 * time.Minute),
	})
	AddNode(client, FakeNode{
		Name:           "node-error",
		Ready:          false,
		TransitionTime: time.Now().Add(-20 * time.Minute),
	})

	// run controller
	controller.NewController(cfg, ctx, client.CoreV1().Nodes(), nodeInformer)
	factory.Start(ctx.Done())
	cache.WaitForCacheSync(ctx.Done(), nodeInformer.HasSynced)

	Describe("Handle node", func() {
		Context("Node is ready", func() {
			It("Should do nothing", func() {
				Expect(deletedNodes).ToNot(ContainElement("node-ready"))
			})
		})

		Context("Node is not ready for shorter than threshold", func() {
			It("Should do nothing", func() {
				Expect(deletedNodes).ToNot(ContainElement("node-unready-below"))
			})
		})

		Context("Node is not ready for longer than threshold", func() {
			Context("Node exists at provider", func() {
				It("Should do nothing", func() {
					Expect(deletedNodes).ToNot(ContainElement("node-unready-exists"))
				})
			})

			Context("Node does not exist at provider", func() {
				It("Should delete the node", func() {
					Expect(deletedNodes).To(ContainElement("node-unready-missing"))
				})
			})
		})

		Context("Provider returns error", func() {
			It("Should return the error", func() {
				Expect(deletedNodes).ToNot(ContainElement("node-unready-exists"))
			})
		})
	})
})

type FakeProvider struct {
	Nodes map[string]bool
}

func (p *FakeProvider) InstanceExists(providerID string) (bool, error) {
	id := strings.TrimPrefix(providerID, "fake://")
	if exists, ok := p.Nodes[id]; ok {
		return exists, nil
	}
	return false, fmt.Errorf("unknown provider ID: %v", providerID)
}

type FakeNode struct {
	Name           string
	Ready          bool
	TransitionTime time.Time
}

func AddNode(client kubernetes.Interface, fn FakeNode) {
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: fn.Name},
		Spec:       v1.NodeSpec{ProviderID: fmt.Sprintf("fake://%s", fn.Name)},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: nodeCondition(fn.Ready), LastTransitionTime: metav1.Time{fn.TransitionTime}},
			},
		},
	}
	_, err := client.CoreV1().Nodes().Create(context.TODO(), node, metav1.CreateOptions{})
	if err != nil {
		Fail(fmt.Sprintf("error adding node: %v", err))
	}
}

func nodeCondition(ready bool) v1.ConditionStatus {
	if ready {
		return v1.ConditionTrue
	} else {
		return v1.ConditionFalse
	}
}
