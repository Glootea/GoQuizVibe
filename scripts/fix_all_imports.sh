#!/bin/bash
set -e

echo "=== Fix malformed imports ==="

# Fix common patterns of malformed imports

# 1. Fix double module names like "goquizvibe/goquizvibe/backend"
find backend -name "*.go" -exec sed -i '' \
    -e 's|goquizvibe/goquizvibe/|goquizvibe/|g' \
    {} \;

# 2. Fix "ce " prefix issues (ce github.com -> "github.com)
find backend -name "*.go" -exec sed -i '' \
    -e 's|	ce |	|g' \
    {} \;

# 3. Fix "import github.com" (missing quote at start)
find backend -name "*.go" -exec sed -i '' \
    -e 's|import github\.com|"github.com|g' \
    {} \;

# 4. Fix imports that should have quote but got mangled
# Pattern: "github.com/foo"bar" -> should be "github.com/foo/bar"
find backend -name "*.go" -exec sed -i '' \
    -e 's|"github\.com/\([^"]*\)"\([^"]*\)"|"github.com/\1\2"|g' \
    {} \;

# 5. Fix missing opening quote for goquizvibe imports
find backend -name "*.go" -exec sed -i '' \
    -e 's|\tgithub\.com/goquizvibe/|\t"github.com/goquizvibe/|g' \
    {} \;

# 6. Fix missing closing quote for goquizvibe imports (line ends with " already)
# This would be: github.com/goquizvibe/backend/xxx" -> should have opening quote
find backend -name "*.go" -exec sed -i '' \
    -e 's|\tgithub\.com/goquizvibe/\([^\n"]*\)"$|\t"github.com/goquizvibe/\1"|g' \
    {} \;

echo "Done"