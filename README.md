![Skuttle](skuttle.png)

# `skuttle`

When there's no hope left for your Kubernetes nodes, Skuttle will prune them from the Kubernetes API.

## What it does

Skuttle watches for nodes entering a `NotReady` state.
If a node has been `NotReady` for some time, Skuttle will use the node's `ProviderID` to query the cloud provider and check if it's still available.
Skuttle will only delete a node if the cloud provider reports it as terminated or missing.

## Usage

```
Usage of skuttle:
  -dry-run
      dry run mode to only log instead of scheduling deletion
  -kubeconfig string
      path to kubeconfig file if not running in-cluster
  -log-level string
      log level (debug, info, warn, error) (default "info")
  -node-selector string
      selector used to filter nodes skuttle should manage (default "node.kubernetes.io/node")
  -not-ready-duration duration
      time duration to tolerate NotReady nodes (default 10m0s)
  -providers string
      comma-separated list of enabled providers
  -refresh-duration duration
      refresh duration (default 10s)
```

## Supported cloud providers

Skuttle supports multiple cloud providers at a time, specified with the `-providers` flag.
The node's `ProviderID` is expected to be in the format `<prefix>://...`.
`<prefix>` is used to determine which of the specified cloud providers to query.

### `aws`: AWS EC2

The `aws` provider will handle nodes with a provider ID `aws://<region>/<instance ID>`.
IAM credentials with permissions to query the existence and state of EC2 instances will need to be available.
