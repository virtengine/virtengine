#!/usr/bin/env bash

set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

if [ -L "${BASH_SOURCE[0]}" ]; then
	SCRIPT_DIR=$(cd "$(dirname "$(realpath "${BASH_SOURCE[0]}")")" >/dev/null 2>&1 && pwd)
fi

source "$SCRIPT_DIR/semver_funcs.sh"

GO_ROOT=$(realpath "$SCRIPT_DIR/../go")

gomod="$GO_ROOT/go.mod"

function get_gotoolchain() {
	local gotoolchain
	local goversion
	local local_goversion

	set +o pipefail
	if [[ "$GOWORK" != "off" ]] && [[ -f "$GOWORK" ]]; then
		gotoolchain=$(grep -E '^toolchain go[0-9]{1,}.[0-9]{1,}.[0-9]{1,}$' "$GOWORK" | cut -d ' ' -f 2 | tr -d '\n')
		goversion=$(grep -E '^go [0-9]{1,}.[0-9]{1,}(.[0-9]{1,})?$' "$GOWORK" | cut -d ' ' -f 2 | tr -d '\n')
	else
		gotoolchain=$(grep -E '^toolchain go[0-9]{1,}.[0-9]{1,}.[0-9]{1,}$' "$gomod" | cut -d ' ' -f 2 | tr -d '\n')
		goversion=$(grep -E '^go [0-9]{1,}.[0-9]{1,}(.[0-9]{1,})?$' "$gomod" | cut -d ' ' -f 2 | tr -d '\n')
	fi

	set -o pipefail

	if [[ ${gotoolchain} == "" ]]; then
		# determine go toolchain from go version in go.mod
		if which go >/dev/null 2>&1; then
			local_goversion=$(GOTOOLCHAIN=local go version | cut -d ' ' -f 3 | sed 's/go*//' | tr -d '\n')
			if [[ $(command_compare "v$local_goversion" v"$goversion") -ge 0 ]]; then
				goversion=$local_goversion
			else
				local_goversion=
			fi
		fi

		if [[ "$local_goversion" == "" ]]; then
			goversion=$(curl -s "https://go.dev/dl/?mode=json&include=all" \
				| jq -r --arg regexp "^go$goversion" '.[] | select(.stable == true) | select(.version | match($regexp)) | .version' \
				| head -n 1 \
				| sed -e s/^go//)
		fi

		if [[ $goversion != "" ]] && [[ $(command_compare "v$goversion" v1.21.0) -ge 0 ]]; then
			gotoolchain=go${goversion}
		else
			gotoolchain=go$(grep -E '^go [0-9]{1,}.[0-9]{1,}$' <"$gomod" | cut -d ' ' -f 2 | tr -d '\n').0
		fi
	fi

	echo -n "$gotoolchain"
}

replace_paths() {
	local file="${1}"
	local cimport="${2}"
	local nimport="${3}"
	local sedcmd=sed

	if [[ "$OSTYPE" == "darwin"* ]]; then
		sedcmd=gsed
	fi

	$sedcmd -ri "s~$cimport~$nimport~" "${file}"
}

function replace_import_path() {
	local next_major_version=$1
	local curr_module_name
	local curr_version
	local new_module_name

	curr_module_name=$(
		cd go
		go list -m
	)
	curr_version=$(echo "$curr_module_name" | sed -n 's/.*v\([0-9]*\).*/\1/p')
	new_module_name=${curr_module_name%/"v$curr_version"}/$next_major_version

	echo "current import paths are $curr_module_name, replacing with $new_module_name"

	declare -a modules_to_upgrade_manually

	modules_to_upgrade_manually+=("./go/go.mod")

	echo "preparing files to replace"

	declare -a files

	while IFS= read -r line; do
		files+=("$line")
	done < <(find . -type f -not \( \
		-path "./install.sh" \
		-or -path "./upgrades/*" \
		-or -path "./.cache/*" \
		-or -path "./dist/*" \
		-or -path "./.git*" \
		-or -name "*.md" \
		-or -path "./.idea/*" \))

	echo "updating all files"

	for file in "${files[@]}"; do
		if test -f "$file"; then
			# skip files that need manual upgrading
			for excluded_file in "${modules_to_upgrade_manually[@]}"; do
				if [[ "$file" == *"$excluded_file"* ]]; then
					continue 2
				fi
			done

			replace_paths "$file" "\"$curr_module_name" "\"$new_module_name"
		fi
	done

	echo "updating go.mod"
	for retract in $(
		cd go
		go mod edit --json | jq -cr '.Retract | if . != null then .[] else empty end'
	); do
		local low
		local high

		low=$(jq -r '.Low' <<<"$retract")
		high=$(jq -r '.High' <<<"$retract")
		echo "    dropping retract: [$low, $high]"
		go mod edit -dropretract=["$low","$high"]
	done

	replace_paths "./go/go.mod" "$curr_module_name" "$new_module_name"
}

function run_gotest() {
	declare -a modules

	modules=("$1")

	if [ -z "$AKASH_ROOT" ]; then
		echo "AKASH_ROOT environment variable is not set"
		exit 1
	fi

	if [ -z "$GO111MODULE" ]; then
		echo "GO111MODULE environment variable is not set"
		exit 1
	fi

	# shellcheck disable=SC2068
	for module in ${modules[@]}; do
		pushd "$(pwd)"
		echo "running tests in $module"
		cd "$module"
		# shellcheck disable=SC2086
		go test ${2} ${3}
		popd
	done
}

function run_golint() {
	declare -a modules

	modules=("$1")
	dirs="$2"

	# shellcheck disable=SC2068
	for module in ${modules[@]}; do
		pushd "$(pwd)"
		echo "running lint on $module"
		cd "$module"
		# shellcheck disable=SC2086
		golangci-lint run --issues-exit-code=0 --timeout=15m "$dirs"
		popd
	done
}

function run_gocoverage() {
	declare -a modules

	modules=("$1")

	if [ -z "$AKASH_ROOT" ]; then
		echo "AKASH_ROOT environment variable is not set"
		exit 1
	fi

	if [ -z "$GO111MODULE" ]; then
		echo "GO111MODULE environment variable is not set"
		exit 1
	fi

	# shellcheck disable=SC2068
	for module in ${modules[@]}; do
		pushd "$(pwd)"
		cd "$module"
		local coverpkgs

		# shellcheck disable=SC2086
		coverpkgs=$(go list ${3} | grep -v mock | paste -sd, -)

		local coverprofile
		coverprofile="$AKASH_ROOT/coverage-$(echo "$module" | tr '/' '-').txt"

		# shellcheck disable=SC2086
		go test ${2} -coverprofile="$coverprofile" \
			-covermode=count \
			-coverpkg="$coverpkgs" \
			${3}

		popd
	done
}

function run_bump_go_module() {
	local cmd
	local prefix
	local nversion
	local cversion

	if [[ $# -ne 2 ]]; then
		echo "Invalid arguments. expected 2, received $#"
		exit 1
	fi

	cmd="$1"
	prefix="$2"

	cversion=v$(git tag -l | grep -i "^$prefix/v[0-9]*" | sed -e "s/^${prefix//\//\\/}\///" | semver sort --filter=latest)
	nversion=v$(semver bump "$cmd" "$cversion")

	read -r -p "bumping module $prefix: $cversion -> $nversion? [Y/n]: " response
	response=${response,,} # tolower
	if [[ $response =~ ^(y| ) ]] || [[ -z $response ]]; then
		git tag -a "$prefix/$nversion" -m "$prefix/$nversion"

		read -r -p "would you like to push $prefix/$nversion? [Y/n]: " response
		response=${response,,} # tolower
		if [[ $response =~ ^(y| ) ]] || [[ -z $response ]]; then
			git push origin "$prefix/$nversion"
		fi
	fi
}

function run_go_mod_tidy() {
	declare -a modules

	modules=("$1")

	if [ -z "$AKASH_ROOT" ]; then
		echo "AKASH_ROOT environment variable is not set"
		exit 1
	fi

	# shellcheck disable=SC2068
	for module in ${modules[@]}; do
		pushd "$(pwd)"
		echo "running go mod tidy for $module"
		cd "$module"
		# shellcheck disable=SC2086
		go mod tidy
		popd
	done
}

case "$1" in
	gotoolchain)
		get_gotoolchain
		;;
	replace-import-path)
		shift
		replace_import_path "$@"
		;;
	gotest)
		shift
		run_gotest "$@"
		;;
	golint)
		shift
		run_golint "$@"
		;;
	gocoverage)
		shift
		run_gocoverage "$@"
		;;
	bump-go)
		shift
		run_bump_go_module "$@"
		;;
	go-mod-tidy)
		shift
		run_go_mod_tidy "$@"
		;;
esac
