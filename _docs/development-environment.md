# Setting up development environment

## Install dependencies

### macOS

> **WARNING**: macOS uses ancient version of the `make`. VirtEngine's environment uses some tricks available in `make 4`.
> We recommend use homebrew to installs most up-to-date version of the `make`. Keep in mind `make` is keg-only, and you'll need manually add its location to the `PATH`.
> Make sure homebrew's make path takes precedence of `/usr/bin`

```shell
brew install curl wget jq direnv coreutils make npm

# Depending on your shell, it may go to .zshrc, .bashrc etc
export PATH="$(brew --prefix)/opt/make/libexec/gnubin:$PATH"
```

### Linux

#### Debian based

**TODO** validate

```shell
sudo apt update
sudo apt install -y jq curl wget build-essentials ca-certificates npm direnv gcc
```

### Node.js + pnpm (frontend tooling)

VirtEngine uses pnpm for frontend workspaces. Ensure pnpm is available:

```shell
corepack enable
corepack prepare pnpm@10.28.2 --activate
pnpm --version
```

Install workspace dependencies when you plan to touch frontend files:

```shell
# Portal app
pnpm -C portal install

# TypeScript SDK
pnpm -C sdk/ts install
```

## Direnv

Both [virtengine](https://github.com/virtengine/virtengine) [provider-services](https://github.com/virtengine/provider) are extensively using `direnv` to set up and seamlessly update environment
while traversing across various directories. It is especially handy for running `provider-services` examples.

> [!WARNING]
> Some distributions may provider outdated version of `direnv`. Currently minimum required version is `2.32.x`.
> Latest version can be installed using [binary builds](https://direnv.net/docs/installation.html#from-binary-builds)

You may enable auto allow by whitelisting specific directories in `direnv.toml`.
To do so use following template to edit `${XDG_CONFIG_HOME:-$HOME/.config}/direnv/direnv.toml`

```toml
[whitelist]
prefix = [
    "<path to virtengine sources>",
    "<path to provider-services sources>"
]
```

## Cache

Build environment will create `.cache` directory in the root of source-tree. We use it to install specific versions of temporary build tools. Refer to `make/setup-cache.mk` for exact list.
It is possible to set custom path to `.cache` with `VE_DEVCACHE` environment variable.

All tools are referred as `makefile targets` and set as dependencies thus installed (to `.cache/bin`) only upon necessity.
For example `protoc` installed only when `proto-gen` target called.

The structure of the dir:

```shell
./cache
    bin/ # build tools
    run/ # work directories for _run examples (provider-services
    versions/ # versions of installed build tools (make targets use them to detect change of version of build tool and install new version if changed)
```

### Add new tool

We will use `modevendor` as an example.
All variables must be capital case.

Following are added to `make/init.mk`

1. Add version variable as `<NAME>_VERSION ?= <version>` to the "# ==== Build tools versions ====" section
   ```makefile
   MODVENDOR_VERSION                  ?= v0.3.0
   ```
2. Add variable tracking version file `<NAME>_VERSION_FILE := $(VE_DEVCACHE_VERSIONS)/<tool>/$(<TOOL>)` to the `# ==== Build tools version tracking ====` section
   ```makefile
   MODVENDOR_VERSION_FILE             := $(VE_DEVCACHE_VERSIONS)/modvendor/$(MODVENDOR)
   ```
3. Add variable referencing executable to the `# ==== Build tools executables ====` section

   ```makefile
   MODVENDOR                          := $(VE_DEVCACHE_VERSIONS)/bin/modvendor
   ```

4. Add installation rules. Following template is used followed by the example

   ```makefile
   $(<TOOL>_VERSION_FILE): $(VE_DEVCACHE)
   	@echo "installing <tool> $(<TOOL>_VERSION) ..."
   	rm -f $(<TOOL>)      # remove current binary if exists
   	# installation procedure depends on distribution type. Check make/setup-cache.mk for various examples
   	rm -rf "$(dir $@)"   # remove current version file if exists
   	mkdir -p "$(dir $@)" # make new version directory
   	touch $@             # create new version file
   $(<TOOL>): $(<TOOL>_VERSION_FILE)
   ```

   Following are added to `make/setup-cache.mk`

   ```makefile
   $(MODVENDOR_VERSION_FILE): $(VE_DEVCACHE)
   	@echo "installing modvendor $(MODVENDOR_VERSION) ..."
   	rm -f $(MODVENDOR)
   	GOBIN=$(VE_DEVCACHE_BIN) $(GO) install github.com/goware/modvendor@$(MODVENDOR_VERSION)
   	rm -rf "$(dir $@)"
   	mkdir -p "$(dir $@)"
   	touch $@
   $(MODVENDOR): $(MODVENDOR_VERSION_FILE)
   ```

## Local Development Network (Localnet)

The localnet provides a complete local development environment with all VirtEngine services running in Docker containers.

### Prerequisites

- Docker and Docker Compose installed
- Go 1.25.5+ (matches go.mod)
- Bash shell (WSL2 on Windows, native on Linux/macOS)

### Quick Start

Start the complete localnet with a single command:

```bash
# Make the script executable (first time only)
chmod +x scripts/localnet.sh scripts/init-chain.sh

# Start the localnet
./scripts/localnet.sh start
```

This starts the following services:
| Service | Port | Description |
|---------|------|-------------|
| VirtEngine Chain (RPC) | 26657 | Tendermint RPC endpoint |
| VirtEngine Chain (REST) | 1317 | Cosmos REST API |
| VirtEngine Chain (gRPC) | 9090 | gRPC endpoint |
| Waldur API | 8080 | Resource management (placeholder) |
| Portal | 3000 | Web UI (placeholder) |
| Provider Daemon | 8443 | Provider API (placeholder) |
| API Gateway (Kong) | 8000 | Gateway proxy |
| Kong Admin API | 8001 | Gateway management API |
| Developer Portal | 3001 | Swagger UI docs |
| Prometheus | 9095 | Metrics UI |
| Grafana | 3002 | Dashboards |

### Localnet Commands

```bash
# Start localnet (runs in background)
./scripts/localnet.sh start

# Stop all services
./scripts/localnet.sh stop

# View service status
./scripts/localnet.sh status

# Tail logs from all services
./scripts/localnet.sh logs

# Tail logs from specific service
./scripts/localnet.sh logs virtengine-node

# Run integration tests
./scripts/localnet.sh test

# Reset localnet (deletes all data)
./scripts/localnet.sh reset

# Open shell in test-runner container
./scripts/localnet.sh shell
```

### Test Accounts

The localnet automatically creates the following test accounts with the `test` keyring:

| Account   | Purpose               |
| --------- | --------------------- |
| validator | Chain validator       |
| alice     | Test user             |
| bob       | Test user             |
| charlie   | Test user             |
| provider  | Test provider account |
| operator  | Test operator account |

Query account addresses:

```bash
# After starting localnet, query accounts from the chain
curl http://localhost:26657/status
```

### Using Docker Compose Directly

You can also use docker-compose directly for more control:

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Stop and remove volumes (full reset)
docker-compose down -v
```

### Running Integration Tests

Integration tests run against the localnet:

```bash
# Start localnet first
./scripts/localnet.sh start

# Run all integration tests
go test -v -tags="e2e.integration" ./tests/integration/...

# Or use the script
./scripts/localnet.sh test
```

### Environment Variables

Configure the localnet with environment variables:

| Variable  | Default               | Description       |
| --------- | --------------------- | ----------------- |
| CHAIN_ID  | virtengine-localnet-1 | Chain identifier  |
| LOG_LEVEL | info                  | Logging level     |
| DETACH    | true                  | Run in background |

Example:

```bash
CHAIN_ID=mytest-1 LOG_LEVEL=debug ./scripts/localnet.sh start
```

### Troubleshooting

**Chain not starting:**

```bash
# Check container status
docker-compose ps

# View chain logs
docker-compose logs virtengine-node

# Reset and try again
./scripts/localnet.sh reset
```

**Port conflicts:**
Edit `docker-compose.yaml` to change port mappings if you have conflicts with existing services.

**Windows (WSL2):**
Run all commands from within WSL2. Ensure Docker Desktop is configured to use WSL2 backend.

## Releasing

With following release instructions VirtEngine team attempted to unify build and release processes:

- reproducible builds
- correct Go toolchains and CGO environment (required for Ledger devices support)

Build is performed by [Goreleaser-cross](https://github.com/goreleaser/goreleaser-cross).
This project was created and is maintained by [virtengine Network](https://github.com/virtengine) core member @troian, and it solves cross-compilation for Golang project with CGO on various hosts for various platforms.

> [!CAUTION]
> The goreleaser-cross image is roughly 7GB in size. You've been warned!

1. To start build simply type

   ```shell
   make release
   ```

2. To release with custom docker image names prepend release command with `RELEASE_DOCKER_IMAGE` variable

   ```shell
   RELEASE_DOCKER_IMAGE=ghcr.io/virtengine/virtengine make release
   ```

3. To build just docker images one case use following command.
   ```shell
   make docker-image
   ```
   or one with custom registry
   ```shell
   RELEASE_DOCKER_IMAGE=ghcr.io/virtengine/virtengine make docker-image
   ```
