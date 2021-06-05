package controller

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

type node struct {
	*v1.Node
}

type Node interface {
	Name() string
	ProviderID() string
	ReadyCondition() (v1.NodeCondition, error)
}

func (n *node) Name() string {
	return n.ObjectMeta.Name
}

func (n *node) ProviderID() string {
	return n.Spec.ProviderID
}

func (n *node) ReadyCondition() (v1.NodeCondition, error) {
	for _, c := range n.Status.Conditions {
		if c.Type == v1.NodeReady {
			return c, nil
		}
	}
	return v1.NodeCondition{}, fmt.Errorf("node missing Ready condition")
}
