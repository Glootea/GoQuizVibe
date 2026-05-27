#!/bin/bash
set -e

echo "=== Fix admin.go comprehensively ==="

# Fix imports first
sed -i '' \
    -e 's|"github.com/goquizvibe/db"|"github.com/goquizvibe/backend/shared/db"|g' \
    -e 's|"github.com/goquizvibe/models"|"github.com/goquizvibe/backend/shared/models"|g' \
    -e 's|"github.com/goquizvibe/types"|"github.com/goquizvibe/backend/shared/types"|g' \
    -e 's|"github.com/goquizvibe/repositories"|"github.com/goquizvibe/backend/shared/repositories"|g' \
    backend/feature/admin/services/admin.go

# Add import aliases
sed -i '' \
    -e 's|r "github.com/goquizvibe/backend/shared/repositories"|r "github.com/goquizvibe/backend/shared/repositories"\n\tauthSvc "github.com/goquizvibe/backend/feature/auth/services"\n\tstorageSvc "github.com/goquizvibe/backend/shared/infrastructure/storage"\n\tcacheSvc "github.com/goquizvibe/backend/shared/infrastructure/cache"\n\timgHelpers "github.com/goquizvibe/backend/shared/infrastructure/cache"|g' \
    backend/feature/admin/services/admin.go

# Fix type references
sed -i '' \
    -e 's|\*AuthService|*authSvc.AuthService|g' \
    -e 's|\*StorageService|*storageSvc.StorageService|g' \
    -e 's|\*CacheService|*cacheSvc.CacheService|g' \
    -e 's|AttachImagesToQuestions|imgHelpers.AttachImagesToQuestions|g' \
    -e 's|\.Delete(|\.Delete(ctx, |g' \
    -e 's|\.SaveOrUpdate(|\.SaveOrUpdate(ctx, |g' \
    -e 's|GetOrFetch|cacheSvc.GetOrFetch|g' \
    -e 's|cache\.cache|cacheSvc|g' \
    backend/feature/admin/services/admin.go

echo "Done"