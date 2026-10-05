# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
#
# The runner-side view of `GO_LINKMODE`, for `make print-linkmode-for-runner`.
#
# This is a VERBATIM copy of Makefile lines 41-47 — the GO_LINKMODE default and the
# CGO_ENABLED=0 downgrade — and deliberately nothing else. It exists because the
# real Makefile cannot be asked that question from a Windows host: Makefile:52-54
# applies
#
#     ifeq ($(OS),Windows_NT)
#         GO_LINKMODE := internal
#     endif
#
# as an unconditional `:=` assignment AFTER the CGO_ENABLED rule, so on Windows it
# overrides the environment no matter what is exported. A guard that compares an
# action's GO_LINKMODE against `make print-linkmode` therefore sees `internal` on a
# Windows dev box, which equals every possible override, and reports OK on the very
# defect it exists to catch. That is not hypothetical: `scripts/ci/
# check-build-linkmode.mjs` did exactly that on its first run against this tree, and
# only its POSIX fixture suite caught the pre-fix `setup-macos` action.
#
# CI runners are ubuntu or macos, never Windows_NT, so omitting the Windows branch
# here reproduces what they derive. `OS` cannot simply be cleared for the whole
# Makefile because `make/init.mk:12-29` branches on it to decide whether to demand
# direnv, and clearing it turns a clean query into a hard error.
#
# DRIFT IS A FAILURE MODE, NOT A STYLE POINT: this file can fall behind Makefile. It
# is guarded, not trusted — `scripts/ci/check-build-linkmode.test.mjs` case 10
# re-derives both answers from the REAL Makefile on any non-Windows host and fails
# if they disagree, so a divergence is caught the moment CI runs it rather than
# being discovered as a mysteriously wrong verdict months later.

GO_LINKMODE            ?= external
CGO_ENABLED            ?= $(shell go env CGO_ENABLED)
ifeq ($(CGO_ENABLED),0)
	ifeq ($(GO_LINKMODE),external)
		GO_LINKMODE := internal
	endif
endif

.PHONY: print-linkmode
print-linkmode:
	@echo "$(GO_LINKMODE)"
