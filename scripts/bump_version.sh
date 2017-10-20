#!/bin/bash

set -ex

if [ -z "$1" ]; then
    echo "required patch/minor/major" 1>&2
    exit 1
fi

next_version=$(gobump "$1" -w -v | jq -r '.[]')
ghch -w -N "v$next_version"

git commit -am "Bump version $next_version"
git tag "v$next_version"
git push && git push --tags
