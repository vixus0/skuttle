package provider

import (
	"fmt"
)

type Store interface {
	Get(string) (Provider, error)
}

type DefaultStore map[string]Provider

func (m *DefaultStore) Add(prefix string, p Provider) {
	(*m)[prefix] = p
}

func (m *DefaultStore) Get(prefix string) (Provider, error) {
	if provider, ok := (*m)[prefix]; ok {
		return provider, nil
	}
	return nil, fmt.Errorf("no provider for prefix %s", prefix)
}
