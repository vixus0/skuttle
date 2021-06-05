package provider

type Provider interface {
	InstanceExists(providerID string) (bool, error)
}
