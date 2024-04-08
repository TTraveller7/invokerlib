# invokerlib

invokerlib contains the implementation of **Fission Stream**, a toolkit that transforms user processor logic to a stream processing pipeline composed of Fission functions.

## Install

0. Verify your terminal has `kubectl` installed, and it can connect to a Kubernetes cluster.

```
kubectl get cs
```

1. Verify you have fission running on your Kubernetes cluster. If you have not installed fission, check the [official installation guide](https://fission.io/docs/installation/) from fission.

```
fission check
```

2. Install fctl with script

```
wget https://raw.githubusercontent.com/TTraveller7/invokerlib/main/scripts/install_fctl.sh
./install_fctl.sh
```

3. Initialize fctl

```
fctl init
```
