# Simple usage with a mounted data directory:
# > docker build -t virtengine .
#
# Server:
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.virtengine:/root/.virtengine virtengine virtengined init test-chain
# TODO: need to set validator in genesis so start runs
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.virtengine:/root/.virtengine virtengine virtengined start
#
# Client: (Note the virtengine binary always looks at ~/.virtengine we can bind to different local storage)
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.virtenginecli:/root/.virtengine virtengine virtengined keys add foo
# > docker run -it -p 26657:26657 -p 26656:26656 -v ~/.virtenginecli:/root/.virtengine virtengine virtengined keys list
# TODO: demo connecting rest-server (or is this in server now?)
FROM golang:alpine AS build-env

# Install minimum necessary dependencies,
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3
RUN apk add --no-cache $PACKAGES

# Set working directory for the build
WORKDIR /go/src/github.com/virtengine/virtengine

# Add source files
COPY . .

# install virtengine, remove packages
RUN make virtengine


# Final image
FROM alpine:edge

# Install ca-certificates
RUN apk add --update ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/src/github.com/virtengine/virtengine/.cache/bin/virtengine /usr/bin/virtengined

EXPOSE 26656 26657 1317 9090

# Run virtengined by default, omit entrypoint to ease using container with simcli
CMD ["virtengined"]