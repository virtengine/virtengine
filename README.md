# VirtEngine - Decentralized Serverless Network

![tests](https://github.com/virtengine/node/workflows/tests/badge.svg)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

VirtEngine is a secure, transparent, and decentralized cloud computing marketplace that connects those who need computing resources (tenants) with those that have computing capacity to lease (providers).

# Roadmap and contributing

VirtEngine is written in Golang and is Apache 2.0 licensed - contributions are welcomed whether that means providing feedback, testing existing and new feature or hacking on the source.

To become a contributor, please see the guide on [contributing](CONTRIBUTING.md)

# Branching and Versioning

The `main` branch contains new features and is under active development; the `mainnet/main` branch contains the current, stable release.

* **stable** releases will have even minor numbers ( `v0.8.0` ) and be cut from the `mainnet/main` branch.
* **unstable** releases will have odd minor numbers ( `v0.9.0` ) and be cut from the `main` branch.

## VirtEngine Suite

VirtEngine Suite is the reference implementation of the VirtEngine Protocol detailed in [patent AU2024203136B2](https://patents.google.com/patent/AU2024203136B2/). [VirtEngine](https://virtengine.com) is an actively-developed prototype currently focused on the distributed marketplace functionality with Proof-of-Identity baked into the protocol.

The Suite is composed of one binary, `virtengine`, which contains a ([tendermint](https://github.com/cometbft/cometbft)-powered) blockchain node that
implements the decentralized exchange as well as client functionality to access the exchange and network data in general.

The basis of this repository includes some source code derived from the [Akash Protocol](https://akash.network/l/whitepaper)

## Get Started with VirtEngine

The easiest way to get started with VirtEngine is by following the Quick Start Guide.

# Supported platforms

Platform | Arch | Status
--- | --- | :---
Darwin | amd64 | ✅ **Supported**
Darwin | arm64 | ✅ **Supported**
Linux | amd64 | ✅ **Supported**
Linux | arm64 (aka aarch64) | ✅ **Supported**
Linux | armhf GOARM=5,6,7 | ⚠️ **Not supported**
Windows | amd64 | ⚠️ **Experimental**

# Installing

The [latest](https://github.com/virtengine/node/releases/latest) binary release can be installed with [Homebrew](https://brew.sh/):

```sh
$ brew tap virtengine/tap
$ brew install virtengine
```

Or [GoDownloader](https://github.com/goreleaser/godownloader):

```sh
$ curl -sSfL https://raw.githubusercontent.com/virtengine/node/main/install.sh | sh
```

## Development environment
[This doc](_docs/development-environment.md) guides through setting up local development environment

VirtEngine is developed and tested with [golang 1.21.0+](https://golang.org/). 
Building requires a working [golang](https://golang.org/) installation, a properly set `GOPATH`, and `$GOPATH/bin` present in `$PATH`.
It is also required to have C/C++ compiler installed (gcc/clang) as there are C dependencies in use (libusb/libhid)
VirtEngine build process and examples are heavily tied to Makefile.


## Building from Source
Command below will compile virtengine executable and put it into `.cache/bin`
```shell
make virtengine # virtengine is set as default target thus `make` is equal to `make virtengine`
```
once binary compiled it exempts system-wide installed virtengine within virtengine repo

## Running

We use thin integration testing environments to simplify
the development and testing process.  We currently have three environments:

* [Single node](_run/lite): simple (no workloads) single node running locally.
* [Single node with workloads](_run/single): single node and provider running locally, running workloads within a virtual machine.
* [full k8s](_run/kube): same as above but with node and provider running inside Kubernetes.
