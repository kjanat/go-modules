#!/usr/bin/env bash
set -eu
awk -v heading="[$1]" '
  /^## / {
    if (found) exit
    found = ($2 == heading)
    next
  }
  found { print; if (NF) content = 1 }
  END { if (!content) { print "No release notes for " heading > "/dev/stderr"; exit 1 } }
' CHANGELOG.md
