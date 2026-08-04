# The immutable digest printed by buildx is the only image reference accepted
# by the Kubernetes workload. Signing and SBOM generation are explicit release
# steps so CI can use workload identity and never place signing keys in source.
IMAGE ?= ghcr.io/virtengine/veid-inference
TAG ?= dev

.PHONY: image sbom sign verify
image:
	docker buildx build --pull --provenance=mode=max --sbom=true --platform linux/amd64 --tag $(IMAGE):$(TAG) --file Dockerfile ../..

sbom:
	syft $(IMAGE):$(TAG) --output cyclonedx-json > veid-inference.sbom.cdx.json

sign:
	cosign sign --yes $(IMAGE)@$(DIGEST)
	cosign attest --yes --type cyclonedx --predicate veid-inference.sbom.cdx.json $(IMAGE)@$(DIGEST)

verify:
	@test -n "$(DIGEST)"
	cosign verify $(IMAGE)@$(DIGEST)
	cosign verify-attestation --type cyclonedx $(IMAGE)@$(DIGEST)
