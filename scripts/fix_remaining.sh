#!/bin/bash
set -e

echo "=== Fix remaining import issues ==="

# Fix repositories imports (now in shared)
find backend -name "*.go" -exec sed -i '' \
    -e 's|"github.com/goquizvibe/repositories"|"github.com/goquizvibe/backend/shared/repositories"|g' \
    {} \;

# Fix pkg/storage imports (now in shared/pkg)
find backend -name "*.go" -exec sed -i '' \
    -e 's|"github.com/goquizvibe/pkg/storage"|"github.com/goquizvibe/backend/shared/pkg/storage"|g' \
    {} \;

# Fix internal/grpc/proto imports (now in shared/internal)
find backend -name "*.go" -exec sed -i '' \
    -e 's|"github.com/goquizvibe/internal/grpc/proto"|"github.com/goquizvibe/backend/shared/internal/grpc/proto"|g' \
    {} \;

# Fix images_helpers.go (was in wrong location)
if [ -f backend/shared/images_helpers.go ]; then
    mv backend/shared/images_helpers.go backend/shared/infrastructure/cache/
fi

# Now fix images_helpers imports
find backend -name "*.go" -exec sed -i '' \
    -e 's|"github.com/goquizvibe/backend/shared/infrastructure/cache/images_helpers"|"github.com/goquizvibe/backend/shared/infrastructure/cache"|g' \
    {} \;

echo "Done fixing imports"