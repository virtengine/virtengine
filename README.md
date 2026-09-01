# VirtEngine - Decentralized Serverless Network

![tests](https://github.com/virtengine/virtengine/actions/workflows/quality-gate.yaml/badge.svg?branch=main&event=push)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

VirtEngine is a secure, transparent, and decentralized cloud computing marketplace that connects those who need computing resources (tenants) with those that have computing capacity to lease (providers).

# Roadmap and contributing

VirtEngine is written in Golang and is Apache 2.0 licensed. Contributions are welcome whether that means providing feedback, testing existing and new features, or hacking on the source.

## Intellectual Property Notice

This project is licensed under the [Apache License 2.0](LICENSE). You are free to use, modify, and distribute the source code under the terms of that license. However, certain methods, processes, and system architectures implemented in this codebase are protected by patent claims, including but not limited to [patent AU2024203136B2](https://patents.google.com/patent/AU2024203136B2/). The Apache 2.0 license grants a patent license only to the extent of the contributions made by the patent holder to this project. Reproducing or deploying the patented methods or processes outside the scope of this license may require separate authorization from the patent holder. For questions regarding patent licensing, please contact the project maintainers.

To become a contributor, please see the guide on [contributing](CONTRIBUTING.md).

# Repository Status

This repository currently ships active development from `main`.

- The release automation and version helper scripts remain in place for tagged releases.
- **TestNet is planned for the January 2027 launch window.** It is the public pre-production phase for multi-operator validation and may be reset.
- **MainNet is planned for the March 2027 launch window.** It is the production launch and requires the TestNet exit criteria plus a fresh go/no-go approval.
- The launches are separated so January TestNet evidence can be observed, remediated, and re-tested during February before production artifacts and approvals are finalized for March.
- Neither TestNet nor MainNet should be described as live until its launch is officially confirmed. The April 2026 MainNet windows did not proceed and remain historical evidence only.

See [docs/NETWORK_LAUNCH_SCHEDULE.md](docs/NETWORK_LAUNCH_SCHEDULE.md), [RELEASE.md](RELEASE.md), [VERIFICATION.md](VERIFICATION.md), and [_docs/operations/network-launch-schedule.md](_docs/operations/network-launch-schedule.md) for the current release and launch posture.

## VirtEngine Suite

VirtEngine Suite is the reference implementation of the VirtEngine Protocol detailed in [patent AU2024203136B2](https://patents.google.com/patent/AU2024203136B2/). It is an actively developed implementation of the marketplace, identity, provider, and operations stack, with launch-readiness work tracked in-repo.

The Suite is composed of one binary, `virtengine`, which contains a [CometBFT](https://github.com/cometbft/cometbft)-powered blockchain node that implements the decentralized exchange as well as client functionality to access the exchange and network data in general.

The basis of this repository includes some source code derived from the [Akash Network: Decentralized Cloud Infrastructure Marketplace](https://ipfs.io/ipfs/QmVwsi5kTrg7UcUEGi5UfdheVLBWoHjze2pHy4tLqYvLYv).

## Get Started with VirtEngine

The quickest way to explore the project is to build from source and use the deployment and operations guides in this repository. Installation artifacts may exist for tagged releases, but operators should verify the target tag, network posture, and launch status before treating any release as production-approved.

# Supported platforms

| Platform | Arch                | Status               |
| -------- | ------------------- | :------------------- |
| Darwin   | amd64               | Supported            |
| Darwin   | arm64               | Supported            |
| Linux    | amd64               | Supported            |
| Linux    | arm64 (aka aarch64) | Supported            |
| Linux    | armhf GOARM=5,6,7   | Not supported        |
| Windows  | amd64               | Experimental         |

# Installing

The repository includes GoReleaser, Homebrew, and install-script paths for tagged releases. Before using those artifacts in production, confirm:

1. The release tag you intend to install has actually been published.
2. The target network has an approved launch or upgrade decision.
3. The verification and support posture in [VERIFICATION.md](VERIFICATION.md) and [docs/COMPATIBILITY.md](docs/COMPATIBILITY.md) matches your intended deployment.

Example installation commands for an already-published tag:

```sh
brew tap virtengine/tap
brew install virtengine
```

```sh
curl -sSfL https://raw.githubusercontent.com/virtengine/virtengine/main/install.sh | sh
```

## Development environment

[This doc](_docs/development-environment.md) guides through setting up a local development environment.

VirtEngine is developed against [Go 1.25.5](https://go.dev/). Building requires a working Go installation, a properly set `GOPATH`, and `$GOPATH/bin` present in `$PATH`. It is also required to have a C/C++ compiler installed (`gcc` or `clang`) as there are C dependencies in use (`libusb`, `libhid`).

VirtEngine build processes and examples are heavily tied to the `Makefile`.

## Building from Source

The command below compiles the `virtengine` executable and writes it into `.cache/bin`.

```shell
make virtengine
```

Once the binary is compiled, it is used in preference to any system-wide `virtengine` installed inside the repository environment.

## Bosun - AI Agent Orchestrator

[Bosun](https://github.com/virtengine/bosun?tab=readme-ov-file#bosun) is VirtEngine's autonomous AI agent orchestrator. It supervises fleets of AI coding agents with failover, crash analysis, auto-restarts, and Telegram or WhatsApp notifications.

```bash
npm install -g bosun
bosun --daemon
```

See the [Bosun repo](https://github.com/virtengine/bosun?tab=readme-ov-file#bosun) for full documentation, or install from [npm](https://www.npmjs.com/package/bosun).

## Running

Deployment, operations, and environment setup are consolidated into:

- [Deployment and Operations Guide](docs/operations/DEPLOYMENT_GUIDE.md)

<!-- GitHub Analytics Pixel -->
<img src="https://cloud.umami.is/p/iR78WZdwe" alt="" width="1" height="1" style="display:none;" />
