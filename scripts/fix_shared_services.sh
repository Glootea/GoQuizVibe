#!/bin/bash
set -e

echo "=== Fix shared services packages ==="

# Move shared time_provider to shared/infrastructure
mkdir -p backend/shared/infrastructure/time
mv backend/shared/time_provider.go backend/shared/infrastructure/time/

# Move shared interfaces to shared/infrastructure
mkdir -p backend/shared/infrastructure/interfaces
mv backend/shared/interfaces.go backend/shared/infrastructure/interfaces/

# Fix imports for time_provider
find backend -name "*.go" -exec sed -i '' \
    -e 's|"github.com/goquizvibe/backend/shared/time_provider"|"github.com/goquizvibe/backend/shared/infrastructure/time"|g' \
    {} \;

# Fix imports for interfaces
find backend -name "*.go" -exec sed -i '' \
    -e 's|"github.com/goquizvibe/backend/shared/interfaces"|"github.com/goquizvibe/backend/shared/infrastructure/interfaces"|g' \
    {} \;

# Fix gamification service - it needs TimeProvider
# The RealTimeProvider is in infrastructure/time
sed -i '' \
    -e 's|services.RealTimeProvider{}|time.RealTimeProvider{}|g' \
    -e 's|24:15: undefined: TimeProvider|fixed|g' \
    backend/feature/gamification/services/gamification.go 2>/dev/null || true

# Add time import to gamification.go
if ! grep -q '"time"' backend/feature/gamification/services/gamification.go; then
    sed -i '' '/^import (/a\	"time"' backend/feature/gamification/services/gamification.go
fi

# Fix dashboard service imports
sed -i '' \
    -e 's|services.GamificationService|shared.GamificationService|g' \
    -e 's|services.Authenticator|shared.Authenticator|g' \
    -e 's|services.QuizSessionService|quizSession.QuizSessionService|g' \
    -e 's|services.CacheService|cacheService.CacheService|g' \
    -e 's|services.GetUserIDFromRequest|authHelpers.GetUserIDFromRequest|g' \
    -e 's|services.GetOrFetch|cacheService.GetOrFetch|g' \
    backend/feature/dashboard/services/dashboard.go 2>/dev/null || true

# Fix quiz services imports
find backend/feature/quiz/services -name "*.go" -exec sed -i '' \
    -e 's|services.CacheService|cacheService.CacheService|g' \
    -e 's|services.GamificationService|gamificationService.GamificationService|g' \
    -e 's|services.Authenticator|authService.Authenticator|g' \
    {} \;

# Fix quiz_session.go
sed -i '' \
    -e 's|services.GamificationService|gamificationService.GamificationService|g' \
    -e 's|services.CacheService|cacheService.CacheService|g' \
    -e 's|services.Authenticator|authService.Authenticator|g' \
    backend/feature/quiz/services/quiz_session.go 2>/dev/null || true

# Fix quiz_timer.go
sed -i '' \
    -e 's|services.CacheService|cacheService.CacheService|g' \
    backend/feature/quiz/services/quiz_timer.go 2>/dev/null || true

# Fix question_schema imports
sed -i '' \
    -e 's|services.PromptGenerator|promptGenerator.PromptGenerator|g' \
    -e 's|services.NewQuestionSchema|cacheService.NewQuestionSchema|g' \
    backend/feature/admin/services/question_schema.go 2>/dev/null || true

echo "Done with initial fixes"