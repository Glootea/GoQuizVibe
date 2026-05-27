#!/bin/bash
set -e

echo "=== Comprehensive fix for remaining build issues ==="

# 1. Fix package names in infrastructure
# Change package services to package cache in infrastructure/cache
sed -i '' 's/^package services$/package cache/' backend/shared/infrastructure/cache/cache_service.go
sed -i '' 's/^package services$/package cache/' backend/shared/infrastructure/cache/cache_helpers.go
sed -i '' 's/^package services$/package cache/' backend/shared/infrastructure/cache/images_helpers.go

# Change package services to package storage in infrastructure/storage
sed -i '' 's/^package services$/package storage/' backend/shared/infrastructure/storage/storage.go

# 2. Create proper time_provider package
cat > backend/shared/infrastructure/time/time.go << 'EOF'
package time

import "time"

type TimeProvider interface {
	Now() time.Time
}

type RealTimeProvider struct{}

func (RealTimeProvider) Now() time.Time { return time.Now() }
EOF

# 3. Create proper interfaces package
cat > backend/shared/infrastructure/interfaces/interfaces.go << 'EOF'
package interfaces

import (
	"context"

	"github.com/goquizvibe/backend/shared/models"
)

type Authenticator interface {
	ValidateToken(token string) (*models.AuthClaims, error)
}

type CacheServiceInterface interface {
	Get(ctx context.Context, key string, dest any) bool
	Set(ctx context.Context, key string, value any) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
EOF

# 4. Fix imports in infrastructure files
# Fix cache_service imports
sed -i '' \
    -e 's|"github.com/goquizvibe/backend/shared/metrics"|"github.com/goquizvibe/backend/shared/metrics"|g' \
    backend/shared/infrastructure/cache/cache_service.go

# 5. Fix imports across all files to use correct paths
# auth_helpers needs Authenticator from interfaces
sed -i '' \
    -e 's|"github.com/goquizvibe/backend/shared"|"github.com/goquizvibe/backend/shared/infrastructure/interfaces"|g' \
    backend/feature/auth/services/auth_helpers.go

# 6. Fix gamification to use time package
sed -i '' \
    -e 's|services.RealTimeProvider{}|timeProvider.RealTimeProvider{}|g' \
    backend/feature/gamification/services/gamification.go

# 7. Fix dashboard service imports
cat > backend/feature/dashboard/services/dashboard.go.new << 'DASHBOARD_EOF'
package services

import (
	"context"

	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/infrastructure/interfaces"
	"github.com/goquizvibe/backend/shared/infrastructure/time"
	"github.com/goquizvibe/backend/feature/gamification/services"
	"github.com/goquizvibe/backend/feature/auth/services"
	"github.com/goquizvibe/backend/feature/quiz/services"
	"github.com/goquizvibe/backend/feature/learning_materials/services"
	"github.com/goquizvibe/backend/shared/infrastructure/cache"
)

type DashboardService struct {
	queries           *db.Queries
	gamification      *services.GamificationService
	auth              *services.AuthService
	quizSession       *services.QuizSessionService
	cache             *cache.CacheService
}

func NewDashboardService(
	queries *db.Queries,
	gamification *services.GamificationService,
	auth *services.AuthService,
	quizSession *services.QuizSessionService,
	cache *cache.CacheService,
) *DashboardService {
	return &DashboardService{
		queries:      queries,
		gamification: gamification,
		auth:         auth,
		quizSession:  quizSession,
		cache:        cache,
	}
}

func (s *DashboardService) GetUserStats(ctx context.Context, userID interface{}) (interface{}, error) {
	// Implementation here
	return nil, nil
}

func (s *DashboardService) GetRecentAttempts(ctx context.Context, limit int32) ([]interface{}, error) {
	// Implementation here
	return nil, nil
}
DASHBOARD_EOF

echo "Created new dashboard.go - need to restore original logic"
mv backend/feature/dashboard/services/dashboard.go.new backend/feature/dashboard/services/dashboard.go.new 2>/dev/null || true

# Let me instead just fix the imports properly without rewriting

echo "Done with initial fixes"