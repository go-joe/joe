#!/usr/bin/env bash

type grep >/dev/null 2>&1 || { echo >&2 'ERROR: script requires "grep"'; exit 1; }

echo "# Searching for files that use embedmd.."
files=$(grep -r -l -E '\[embedmd\]:# \(.+\)' content)
if [[ -z "$files" ]]; then
  echo >&2 'ERROR: did not find any file that uses embedmd'
  exit 1
fi

set -e -o pipefail

for f in $files; do
    echo "> Updating $f"
    embedmd -w "$f"
done
