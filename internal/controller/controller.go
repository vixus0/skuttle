package controller

import (
	"context"
	"strings"
	"time"

	"github.com/vixus0/skuttle/v2/internal/logging"
	"github.com/vixus0/skuttle/v2/internal/provider"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

var (
	log *logging.Logger = logging.NewLogger("ctrl")
)

type NodeDeleter interface {
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

type Controller struct {
	Config
	nodeDeleter NodeDeleter
	ctx         context.Context
}

type Config struct {
	DryRun           bool
	NotReadyDuration time.Duration
	Providers        provider.Store
}

func NewController(
	cfg *Config,
	ctx context.Context,
	nodeDeleter NodeDeleter,
	nodeInformer cache.SharedIndexInformer,
) *Controller {
	controller := &Controller{
		Config:      *cfg,
		ctx:         ctx,
		nodeDeleter: nodeDeleter,
	}

	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.Add,
		UpdateFunc: controller.Update,
		DeleteFunc: controller.Delete,
	})

	return controller
}

// When a new node gets created
func (c *Controller) Add(obj interface{}) {
	n := coerce(obj)
	log.Debug("add node %s", n.Name())
	if err := c.Handle(n); err != nil {
		log.Error(err.Error())
	}
}

// When a node gets updated
func (c *Controller) Update(_ interface{}, obj interface{}) {
	n := coerce(obj)
	log.Debug("update node %s", n.Name())
	if err := c.Handle(n); err != nil {
		log.Error(err.Error())
	}
}

// When a node gets deleted
func (c *Controller) Delete(obj interface{}) {
	n := coerce(obj)
	log.Debug("remove node %s", n.Name())
}

// Handle a node
func (c *Controller) Handle(n *node) error {
	cond, err := n.ReadyCondition()
	if err != nil {
		return err
	}

	// node is Ready, no need to handle
	if cond.Status == v1.ConditionTrue {
		return nil
	}

	// handle if transition to NotReady is greater than tolerance
	sinceTransition := time.Since(cond.LastTransitionTime.Time)
	threshold := c.NotReadyDuration

	if sinceTransition > threshold {
		log.Info(
			"node %s has been NotReady for %s (> threshold %s)",
			n.Name(),
			sinceTransition.String(),
			threshold.String(),
		)

		// Get Provider for Node
		prefixParts := strings.Split(n.ProviderID(), ":")
		prefix := prefixParts[0]
		provider, err := c.Providers.Get(prefix)

		// Check if instance exists
		exists, err := provider.InstanceExists(n.ProviderID())
		if err != nil {
			return err
		}

		// Delete node if not
		if exists {
			log.Warn("node %s exists at provider", n.Name())
		} else {
			log.Info("deleting node %s", n.Name())
			return c.deleteNode(n.Name())
		}
	}

	return nil
}

func (c *Controller) deleteNode(name string) error {
	if c.DryRun {
		log.Info("*** DRY RUN *** deleted node %s", name)
	} else {
		return c.nodeDeleter.Delete(c.ctx, name, metav1.DeleteOptions{})
	}
	return nil
}

func coerce(obj interface{}) *node {
	v1node := obj.(*v1.Node)
	return &node{v1node}
}
