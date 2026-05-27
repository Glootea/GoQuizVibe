#!/bin/bash
set -e

echo "=== Fix all import issues ==="

# Fix middleware - services should be auth service
sed -i '' 's|"github.com/goquizvibe/services"|"github.com/goquizvibe/backend/feature/auth/services"|g' \
    backend/shared/middleware/require_auth.go \
    backend/shared/middleware/require_role.go

# Fix di/wire_test.go
sed -i '' \
    -e 's|"github.com/goquizvibe/handlers"|"github.com/goquizvibe/backend/feature/admin/handlers"|g' \
    -e 's|"github.com/goquizvibe/services"|"github.com/goquizvibe/backend/feature/auth/services"|g' \
    backend/shared/di/wire_test.go

# Fix di/providers.go
sed -i '' \
    -e 's|"github.com/goquizvibe/handlers"|"github.com/goquizvibe/backend/feature/admin/handlers"|g' \
    -e 's|"github.com/goquizvibe/services"|"github.com/goquizvibe/backend/feature/auth/services"|g' \
    backend/shared/di/providers.go

# Fix pages imports - feature ui
sed -i '' \
    -e 's|"github.com/goquizvibe/pages"|"github.com/goquizvibe/backend/feature/auth/ui"|g' \
    backend/feature/auth/handlers/auth.go

sed -i '' \
    -e 's|"github.com/goquizvibe/pages"|"github.com/goquizvibe/backend/feature/admin/ui/admin"|g' \
    backend/feature/admin/handlers/admin.go \
    backend/feature/admin/handlers/middleware.go

sed -i '' \
    -e 's|"github.com/goquizvibe/pages"|"github.com/goquizvibe/backend/feature/quiz/ui"|g' \
    backend/feature/quiz/handlers/quiz.go

sed -i '' \
    -e 's|"github.com/goquizvibe/pages"|"github.com/goquizvibe/backend/feature/dashboard/ui"|g' \
    backend/feature/dashboard/handlers/dashboard.go

sed -i '' \
    -e 's|"github.com/goquizvibe/pages/admin"|"github.com/goquizvibe/backend/feature/admin/ui/admin"|g' \
    backend/feature/learning_materials/handlers/learning_materials.go

sed -i '' \
    -e 's|"github.com/goquizvibe/pages/editor"|"github.com/goquizvibe/backend/feature/editor/ui"|g' \
    backend/feature/editor/handlers/editor.go

# Fix templ generated files
sed -i '' \
    -e 's|"github.com/goquizvibe/pages/components"|"github.com/goquizvibe/backend/shared/ui/components"|g' \
    -e 's|"github.com/goquizvibe/pages"|"github.com/goquizvibe/backend/shared/ui"|g' \
    backend/feature/admin/ui/admin/*.go

echo "Done fixing imports"