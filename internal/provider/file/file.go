package file

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Provider struct {
	Nodes []string
}

func NewProvider() (*Provider, error) {
	path, ok := os.LookupEnv("NODE_LIST")
	if !ok {
		return nil, fmt.Errorf("Need to specify path to node list in NODE_LIST")
	}

	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	var nodes []string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		nodes = append(nodes, strings.TrimSpace(scanner.Text()))
	}

	return &Provider{
		Nodes: nodes,
	}, nil
}

func (provider *Provider) InstanceExists(providerID string) (bool, error) {
	noPrefixID := strings.TrimPrefix(providerID, "file://")

	for _, id := range provider.Nodes {
		if id == noPrefixID {
			return true, nil
		}
	}
	return false, nil
}
