#!/bin/bash

arc_info_output=$(arc info 2>&1)
if echo "$arc_info_output" | grep -q "Not a mounted arc repository"; then
    diff=$(git diff --merge-base origin/master --unified=0 --no-color --diff-filter=d)
else
    diff=$(arc diff $(arc merge-base HEAD cloudia/trunk) . -U0)
fi

changed_lines=$(echo "$diff" | awk '
    /^\+\+\+ / {
        file = $2
        sub(/^b\//, "", file)
        next
    }
    /^@@ / {
        range = $3
        sub(/^\+/, "", range)
        split(range, parts, ",")
        start = parts[1] - 3
        if (start < 1) start = 1
        count = parts[2] == "" ? 1 : parts[2]
        end = parts[1] + count + 2
        for (line = start; line <= end; line++) print file ":" line ":"
    }
')

if [[ -z "$changed_lines" ]]; then
  echo "No issues found."
  exit 0
fi

output=$(go run lint/cmd/provider-linter/main.go ./yandex-framework/... ./yandex/... 2>&1 | grep -F -f <(echo "$changed_lines"))

if [[ -n "$output" ]]; then
  echo "Found issues:"
  echo "$output"
  exit 1
else
  echo "No issues found."
fi
