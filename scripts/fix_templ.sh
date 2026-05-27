#!/bin/bash
set -e

echo "=== Fix all templ package issues ==="

# 1. Fix base.templ to use sharedui package
sed -i '' 's/^package pages$/package sharedui/' backend/shared/ui/base.templ
sed -i '' 's/import sharedui "github.com\/goquizvibe\/backend\/shared\/ui"//' backend/shared/ui/base.templ

# 2. Fix all feature ui packages to import sharedui properly
for file in $(find backend/feature -name "*.templ"); do
    # Add sharedui import if not present
    if ! grep -q 'sharedui "github.com\/goquizvibe\/backend\/shared\/ui"' "$file"; then
        # Check if file has multiple imports
        if grep -q '^import ($' "$file"; then
            sed -i '/^import ($/,/^)$/s/)$/sharedui "github.com\/goquizvibe\/backend\/shared\/ui"\n)/' "$file"
        fi
    fi

    # Fix @Base to @sharedui.Base
    sed -i '' 's/@Base(/@sharedui.Base(/g' "$file"
done

echo "Done fixing templ files"