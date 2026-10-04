#!/usr/bin/env bash

# generated changelog depends on the tag type
# only SEMVER tags are accounted
# there are two type of tag distinguished by minor part of the semver:
#   - even number: mainnet
#   - odd number: testnet
# net detection is done in section s1

# there are two type release notes generated
#   - prerelease: changelog between current and nearest lower prerelease (or previous release)
#     for example current tag v0.1.1-rc.10 and previous was v0.1.1-rc.9, so changelog is generated between
#   - release: changelog between current and previous release tags
# mainnet status is taken care as well. if current tag is edgenet (e.g. v0.1.1-rc.10) it
# will be generated to edgenet changes only

PATH=$PATH:$(pwd)/.cache/bin
export PATH=$PATH

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

usage() {
	echo "usage: genchangelog.sh [--next-tag] <tag> <output-file>"
	echo ""
	echo "  <tag>          the release tag to generate notes for"
	echo "  <output-file>  where the markdown notes are written"
	echo "  --next-tag     preview mode: generate notes for a tag that does not exist yet"
	echo "                 (passes --next-tag to git-chglog; nothing is tagged or published)"
	exit 1
}

# s0
# Preview mode (release plan defect D5). Without --next-tag, git-chglog refuses to run
# for a tag that is not in the repository:
#   ERROR commits corresponding to "v0.3.0" was not found
# which made it impossible to review release notes before cutting the tag.
next_tag=false
if [[ "${1:-}" == "--next-tag" ]]; then
	next_tag=true
	shift
fi

if [[ $# -ne 2 ]]; then
	usage
fi

to_tag=$1

# s1
# shellcheck disable=SC1073
if ! "${SCRIPT_DIR}"/mainnet-from-tag.sh "$to_tag" ; then
	version_rel="^[v|V]?(0|[1-9][0-9]*)\\.(\\d*[13579])\\.(0|[1-9][0-9]*)$"
	version_prerel="^[v|V]?(0|[1-9][0-9]*)\\.(\\d*[13579])\\.(0|[1-9][0-9]*)(\\-[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"
else
	version_rel="^[v|V]?(0|[1-9][0-9]*)\.(\d*[02468])\.(0|[1-9][0-9]*)$"
	version_prerel="^[v|V]?(0|[1-9][0-9]*)\.(\d*[02468])\.(0|[1-9][0-9]*)(\-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$"
fi

# s2
if [[ -z $("${SCRIPT_DIR}"/semver.sh get prerel "$to_tag") ]]; then
	tag_regexp=$version_rel
else
	tag_regexp=$version_prerel
fi

# s3
# --config is mandatory: without it git-chglog looks for .chglog/config.yml, but this
# repository ships .chglog/config.yaml, so the command dies with
#   ERROR open .chglog/config.yml: The system cannot find the file specified
# --tag-filter-pattern keeps checkpoint/* automation tags out of the compare range.
if [[ "$next_tag" == true ]]; then
	git-chglog --config .chglog/config.yaml --tag-filter-pattern="$tag_regexp" --next-tag "$to_tag" --output "$2"
else
	git-chglog --config .chglog/config.yaml --tag-filter-pattern="$tag_regexp" --output "$2" "$to_tag"
fi
