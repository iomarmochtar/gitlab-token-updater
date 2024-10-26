#!/bin/bash

# we use file CHANGELOG.md as version reference, take the first section as the latest one
VERSION=$(grep -m 1 -E '^# ' CHANGELOG.md | awk '{print $2}')
TAG_VER="v${VERSION}"

echo "set with version ${VERSION} with git tag ${TAG_VER}"

git tag "v${VERSION}"