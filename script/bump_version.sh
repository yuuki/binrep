#!/bin/bash

set -ex

if [ -z "$1" ]; then
    echo "required patch/minor/major" 1>&2
    exit 1
fi

MAIN='./'

# gobump
new_version=$(gobump "$1" -w -v "${MAIN}" | jq -r '.[]')
git add "${MAIN}/version.go"
git commit -m "Bump version $new_version"
git push origin master
git tag "v$new_version"
git push origin "v$new_version"
