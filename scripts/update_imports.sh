#!/bin/bash
set -e

echo "=== Comprehensive import path update ==="

OLD_PREFIX='"github.com/goquizvibe/'
BACKEND='github.com/goquizvibe/backend'

update_file() {
    local file="$1"
    local old="$2"
    local new="$3"
    if grep -q "$old" "$file" 2>/dev/null; then
        sed -i '' "s|$old|$new|g" "$file"
    fi
}

update_feature_imports() {
    local feature="$1"
    local feature_path="backend/feature/$feature"
    
    echo "Updating $feature imports..."
    
    # Process each Go file
    for file in $(find "$feature_path" -name "*.go" 2>/dev/null); do
        update_file "$file" "${OLD_PREFIX}handlers" "${BACKEND}/feature/${feature}/handlers"
        update_file "$file" "${OLD_PREFIX}services" "${BACKEND}/feature/${feature}/services"
        update_file "$file" "${OLD_PREFIX}middleware" "${BACKEND}/shared/middleware"
        update_file "$file" "${OLD_PREFIX}db" "${BACKEND}/shared/db"
        update_file "$file" "${OLD_PREFIX}models" "${BACKEND}/shared/models"
        update_file "$file" "${OLD_PREFIX}config" "${BACKEND}/shared/config"
        update_file "$file" "${OLD_PREFIX}locales" "${BACKEND}/shared/locales"
        update_file "$file" "${OLD_PREFIX}database" "${BACKEND}/shared/database"
        update_file "$file" "${OLD_PREFIX}custom_errors" "${BACKEND}/shared/custom_errors"
        update_file "$file" "${OLD_PREFIX}metrics" "${BACKEND}/shared/metrics"
        update_file "$file" "${OLD_PREFIX}types" "${BACKEND}/shared/types"
        update_file "$file" "${OLD_PREFIX}pkg" "${BACKEND}/shared/pkg"
        update_file "$file" "${OLD_PREFIX}infrastructure" "${BACKEND}/shared/infrastructure"
    done
}

# Update each feature
update_feature_imports "auth"
update_feature_imports "dashboard"
update_feature_imports "quiz"
update_feature_imports "admin"
update_feature_imports "learning_materials"
update_feature_imports "editor"
update_feature_imports "gamification"

# Update shared files
echo "Updating shared imports..."
for file in $(find backend/shared -name "*.go" 2>/dev/null); do
    update_file "$file" "${OLD_PREFIX}db" "${BACKEND}/shared/db"
    update_file "$file" "${OLD_PREFIX}models" "${BACKEND}/shared/models"
    update_file "$file" "${OLD_PREFIX}types" "${BACKEND}/shared/types"
    update_file "$file" "${OLD_PREFIX}config" "${BACKEND}/shared/config"
    update_file "$file" "${OLD_PREFIX}locales" "${BACKEND}/shared/locales"
    update_file "$file" "${OLD_PREFIX}middleware" "${BACKEND}/shared/middleware"
    update_file "$file" "${OLD_PREFIX}custom_errors" "${BACKEND}/shared/custom_errors"
    update_file "$file" "${OLD_PREFIX}database" "${BACKEND}/shared/database"
    update_file "$file" "${OLD_PREFIX}metrics" "${BACKEND}/shared/metrics"
done

# Update main.go
echo "Updating main.go..."
sed -i '' "s|${OLD_PREFIX}|${BACKEND}/|g" main.go

# Update test files
echo "Updating test imports..."
for file in $(find test -name "*.go" 2>/dev/null); do
    sed -i '' "s|${OLD_PREFIX}|${BACKEND}/|g" "$file"
done

# Update mocks
echo "Updating mock imports..."
for file in $(find mocks -name "*.go" 2>/dev/null); do
    sed -i '' "s|${OLD_PREFIX}|${BACKEND}/|g" "$file"
done

# Update go.work
echo "Updating go.work..."
sed -i '' 's|./microservices/typst|../microservices/typst|g' go.work
sed -i '' 's|./pkg|../pkg|g' go.work

# Update docker-compose path
echo "Updating docker-compose..."
sed -i '' 's|context: \.|context: ./backend|g' deployment/docker-compose.yml

echo "All import paths updated"