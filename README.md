![Skuttle](skuttle.png)

# `skuttle`

When there's no hope left for your Kubernetes nodes, Skuttle will prune them from the Kubernetes API.

## How it works

Skuttle watches for nodes entering a `NotReady` state.
If a node has been `NotReady` for some time, Skuttle will use the node's `ProviderID` to query the cloud provider and check if it's still available.
Skuttle will only delete a node if the cloud provider reports it as terminated or missing.

## Supported cloud providers

- AWS: EC2
