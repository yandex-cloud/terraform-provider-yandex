#!/bin/bash

files=$(git diff --merge-base origin/master --name-only --no-color --diff-filter=d)

file_pattern=$(echo "$files" | sed 's/[./]/\\&/g' | sed 's/^/(/' | sed 's/$/)/' | sed 's/$/|/' | tr -d '\n' | sed 's/\\|$//' | sed 's/.$//')

output=$(go run lint/cmd/provider-linter/main.go ./yandex-framework/... ./yandex/... 2>&1 | grep -E "$file_pattern")

if [[ -n "$output" ]]; then
  echo "Found issues:"
  echo "$output"
  exit 1
else
  echo "No issues found."
fi