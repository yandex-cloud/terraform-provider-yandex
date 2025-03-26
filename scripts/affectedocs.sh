#!/bin/bash

arc_info_output=$(arc info 2>&1)
if echo "$arc_info_output" | grep -q "Not a mounted arc repository"; then
    files=$(git diff --merge-base origin/master --name-only --no-color --diff-filter=d)
else
    files=$(arc diff $(arc merge-base HEAD cloudia/trunk) . --name-only)
fi

file_pattern=$(echo "$files" | sed 's/[./]/\\&/g' | sed 's/^/(/' | sed 's/$/)/' | sed 's/$/|/' | tr -d '\n' | sed 's/\\|$//' | sed 's/.$//')

output=$(go run lint/cmd/provider-linter/main.go ./yandex-framework/... ./yandex/... 2>&1 | grep -E "$file_pattern")

if [[ -n "$output" ]]; then
  echo "Found issues:"
  echo "$output"
  exit 1
else
  echo "No issues found."
fi
