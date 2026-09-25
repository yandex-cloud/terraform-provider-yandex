#!/usr/bin/env bash
set -euo pipefail

# Release preparation has already run go mod tidy with a temporary proxy.
# Its SDK and proto versions only exist in vendor after that proxy is removed.
# These versions remain in go.mod after release, so all later PRs use vendor too.
if grep -Eq '^[[:space:]]*(require[[:space:]]+)?github[.]com/yandex-cloud/(go-genproto|go-sdk)(/[^[:space:]]*)?[[:space:]]+v[0-9]+[.]0[.]0-tf[.][0-9a-f]{40}([[:space:]]|$)' go.mod; then
    go list -mod=vendor -deps -test ./... >/dev/null
else
    go mod tidy
fi

changes=$(arc diff --name-only)
if [ -n "$changes" ]; then
    echo "##teamcity[buildProblem description='The arc working directory is not empty after Go dependency check.' identity='arc-diff-not-empty-after-go-mod-check']"
    exit 1
fi
