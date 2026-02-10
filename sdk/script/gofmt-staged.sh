#!/usr/bin/env bash

for file in $(git diff --cached --name-only --diff-filter=ACMRTUXB | grep "\.go$")
do
	echo "(gofmt) $file"
	gofmt -w "$file"
	git add "$file"
done