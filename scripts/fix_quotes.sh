#!/bin/bash
set -e

echo "=== Fix missing opening quotes ==="

# The issue: imports like "github.com/goquizvibe/backend/shared/db"
# got transformed to github.com/goquizvibe/backend/shared/db" (missing opening quote)
# We need to add " at the start of each import

# Fix imports in all Go files that have the pattern:
# \tgithub.com/goquizvibe/... should be \t"github.com/goquizvibe/...

# Fix shared imports
find backend -name "*.go" -exec sed -i '' \
    -e 's|\tgithub\.com/goquizvibe/|\t"github.com/goquizvibe/|g' \
    {} \;

# Fix lines that start with github.com (should start with ")
find backend -name "*.go" -exec sed -i '' \
    -e 's|^github\.com/goquizvibe/|"github.com/goquizvibe/|g' \
    {} \;

echo "Done fixing quotes"