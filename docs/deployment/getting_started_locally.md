# Deploying Gardener locally

This document will walk you through deploying Gardener on your local machine.
If you encounter difficulties, please open an issue so that we can make this process easier.

Gardener runs in any Kubernetes cluster.
In this guide, we will start a [KinD](https://kind.sigs.k8s.io/) cluster which is used as both garden and seed cluster (please refer to the [architecture overview](../concepts/architecture.md)) for simplicity.

Based on [Skaffold](https://skaffold.dev/), the container images for all required components will be built and deployed into the cluster (via their [Helm charts](https://helm.sh/)).

![Architecture Diagram](content/getting_started_locally.png)

## Prerequisites

- Make sure your Docker daemon is up-to-date, up and running and has enough resources (at least `8` CPUs and `8Gi` memory; see [here](https://docs.docker.com/desktop/mac/#resources) how to configure the resources for Docker for Mac).
  > Please note that 8 CPU / 8Gi memory might not be enough for more than two `Shoot` clusters, i.e., you might need to increase these values if you want to run additional `Shoot`s.

  Additionally, please configure at least `120Gi` of disk size for the Docker daemon.
  > Tip: With `docker system df` and `docker system prune -a` you can cleanup unused data.

## Setting up the KinD cluster (garden and seed)

```bash
make kind-up
```

This command sets up a new KinD cluster named `gardener-local` and stores the kubeconfig in the `./example/gardener-local/kind/kubeconfig` file.

> It might be helpful to copy this file to `$HOME/.kube/config` since you will need to target this KinD cluster multiple times.
Alternatively, make sure to set your `KUBECONFIG` environment variable to `./example/gardener-local/kind/kubeconfig` for all future steps via `export KUBECONFIG=example/gardener-local/kind/kubeconfig`.

All following steps assume that your are using this kubeconfig.

## Setting up Gardener

```bash
make gardener-up
```

This will first build the images based (which might take a bit if you do it for the first time).
Afterwards, the Gardener resources will be deployed into the cluster.

## Creating a `Shoot` cluster

You can wait for the `Seed` to be ready by running

```bash
kubectl wait --for=condition=gardenletready --for=condition=extensionsready --for=condition=bootstrapped seed local --timeout=5m
```

Alternatively, you can run `kubectl get seed local` and wait for the `STATUS` to indicate readiness:

```bash
NAME    STATUS   PROVIDER   REGION   AGE     VERSION       K8S VERSION
local   Ready    local      local    4m42s   vX.Y.Z-dev    v1.21.1
```

In order to create a first shoot cluster, just run

```bash
kubectl apply -f example/provider-local/shoot.yaml
```

You can wait for the `Shoot` to be ready by running

```bash
kubectl wait --for=condition=apiserveravailable --for=condition=controlplanehealthy --for=condition=everynodeready --for=condition=systemcomponentshealthy shoot local -n garden-local --timeout=10m
```

Alternatively, you can run `kubectl -n garden-local get shoot local` and wait for the `LAST OPERATION` to reach `100%`:

```bash
NAME    CLOUDPROFILE   PROVIDER   REGION   K8S VERSION   HIBERNATION   LAST OPERATION            STATUS    AGE
local   local          local      local    1.21.0        Awake         Create Processing (43%)   healthy   94s
```

(Optional): You could also execute a simple e2e test (creating and deleting a shoot) by running

```shell
make test-e2e-local-fast KUBECONFIG="$PWD/example/gardener-local/kind/kubeconfig"
```

⚠️ Please note that in this setup shoot clusters are not accessible by default when you download the kubeconfig and try to communicate with them.
The reason is that your host most probably cannot resolve the DNS names of the clusters since `provider-local` extension runs inside the KinD cluster (see [this](../extensions/provider-local.md#dnsrecord) for more details).
Hence, if you want to access the shoot cluster, you have to run the following command which will extend your `/etc/hosts` file with the required information to make the DNS names resolvable:

```bash
cat <<EOF | sudo tee -a /etc/hosts

# Manually created to access local Gardener shoot clusters with names 'local' or 'e2e-local' in the 'garden-local' namespace.
# TODO: Remove this again when the shoot cluster access is no longer required.
127.0.0.1 api.local.local.external.local.gardener.cloud
127.0.0.1 api.local.local.internal.local.gardener.cloud
127.0.0.1 api.e2e-local.local.external.local.gardener.cloud
127.0.0.1 api.e2e-local.local.internal.local.gardener.cloud
EOF
```

Now you can access it by running

```bash
kubectl -n garden-local get secret local.kubeconfig -o jsonpath={.data.kubeconfig} | base64 -d > /tmp/kubeconfig-shoot-local.yaml
kubectl --kubeconfig=/tmp/kubeconfig-shoot-local.yaml get nodes
```

## Deleting the `Shoot` cluster

```shell
./hack/usage/delete shoot local garden-local
```

## Tear down the Gardener environment

```shell
make kind-down
```

## Further reading

This setup makes use of the local provider extension. You can read more about it in [this document](../extensions/provider-local.md).