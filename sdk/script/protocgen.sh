#!/usr/bin/env bash

set -e
set -o pipefail

if [[ $# -lt 1 ]]; then
	echo "invalid number of parameters"
	exit 1
fi

PATH=$(pwd)/.cache/bin:$PATH
export PATH=$PATH

function gen() {
	mapfile -t proto_files < <( buf ls-files | grep ^proto)

	for file in "${proto_files[@]}"; do
		find ./"$(dirname "$file")" -name "*.pb.go" -o -name "*.pb.gw.go" -type f -delete
		buf generate --config buf.yaml --template "$1" "$file"
	done
}

function gen_go() {
	if [[ $# -ne 2 ]]; then
		echo "invalid number of parameters"
		exit 1
	fi

	IFS=/ read -r pkg_domain _ <<<"$1"

	function cleanup_go() {
		rm -rf "$pkg_domain"
	}

	trap cleanup_go EXIT ERR

	gen buf.gen.gogo.yaml

	# shellcheck disable=SC2086
	cp -r ./$1/* ./$2/
}

function gen_rust() {
	if [[ $# -ne 2 ]]; then
		echo "invalid number of parameters"
		exit 1
	fi

	IFS=/ read -r pkg_domain _ <<<"$1"

	function cleanup_go() {
		rm -rf "$pkg_domain"
	}

	trap cleanup_go EXIT ERR

	gen buf.gen.gogo.yaml

	# shellcheck disable=SC2086
	cp -r ./$1/* ./$2/
}


function gen_pulsar() {
	gen buf.gen.pulsar.yaml

	cp -r pulsar/akash/* go/node/
}

function gen_ts() {
	rm -rf ts/src/generated
	PROTO_SOURCE=node buf generate --template buf.gen.ts.yaml ./proto/node
	PROTO_SOURCE=cosmos buf generate --template buf.gen.ts.yaml ./go/vendor/github.com/cosmos/cosmos-sdk/proto
	PROTO_SOURCE=ibc-go buf generate --template buf.gen.ts.yaml ./go/vendor/github.com/cosmos/ibc-go/v10/proto
#	PROTO_SOURCE=ics23 buf generate --template buf.gen.ts.yaml ./go/vendor/github.com/cosmos/ics23/proto
	PROTO_SOURCE=provider buf generate --template buf.gen.ts.yaml ./proto/provider
	node --experimental-strip-types --no-warnings ts/script/fix-ts-proto-generated-types.ts
}

function gen_doc() {
	local workdir

	workdir=$AKASH_DEVCACHE_TMP/swagger-gen

	function cleanup {
		rm -rf "$workdir"
	}

	# clean swagger files
	trap cleanup EXIT ERR

	buf generate --template ./proto/node/buf.gen.doc.yaml ./proto/node 2>/dev/null
	buf generate --template ./proto/provider/buf.gen.doc.yaml ./proto/provider 2>/dev/null

	mapfile -t proto_files < <(buf ls-files | grep ^proto | grep -e "service.proto$" -e "query.proto$")

	for file in "${proto_files[@]}"; do
		buf generate --template buf.gen.swagger.yaml "$file"
	done

	mkdir -p ./docs/swagger-ui

	# combine swagger files
	# uses nodejs package `swagger-combine`.
	# all the individual swagger files need to be configured in `config.json` for merging
	swagger-combine \
		./docs/config.yaml \
		-o ./docs/swagger-ui/swagger.yaml \
		-f yaml \
		--continueOnConflictingPaths=true \
		--includeDefinitions=true
}

case $1 in
	go)
		shift
		gen_go "$@"
		;;
	rust)
		shift
		gen_rust "$@"
		;;
	pulsar)
		shift
		gen_pulsar "$@"
		;;
	ts)
		shift
		gen_ts
		;;
	doc)
		gen_doc
		;;
	*)
		echo "Invalid argument"
		exit 1
		;;
esac
