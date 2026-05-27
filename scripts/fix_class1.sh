#!/bin/bash
set -e

echo "=== Fix Class 1: Templ files referencing old 'pages' package ==="

# Fix Base reference in templ files - Base is from base_templ.go
# Templates need to import shared/ui and call ui.Base
# For now, let's just add the correct import path to each templ file

fix_templ_import() {
    local file="$1"
    local import_line="$2"

    # Check if line 11 (0-indexed 10) has the wrong import
    sed -i '' "11s|.*|${import_line}|" "$file"
}

# Landing templ
sed -i '' '11s|.*|import ui "github.com/goquizvibe/backend/shared/ui"|' backend/feature/auth/ui/landing_templ.go

# Login templ
sed -i '' '11s|.*|import ui "github.com/goquizvibe/backend/shared/ui"|' backend/feature/auth/ui/login_templ.go

# Register templ
sed -i '' '11s|.*|import ui "github.com/goquizvibe/backend/shared/ui"|' backend/feature/auth/ui/register_templ.go

# Dashboard templ
sed -i '' '14s|.*|import ui "github.com/goquizvibe/backend/shared/ui"|' backend/feature/dashboard/ui/dashboard_templ.go

# Quiz templ
sed -i '' '14s|.*|import ui "github.com/goquizvibe/backend/shared/ui"|' backend/feature/quiz/ui/quiz_templ.go

# Quiz info templ
sed -i '' '14s|.*|import ui "github.com/goquizvibe/backend/shared/ui"|' backend/feature/quiz/ui/quiz_info_templ.go

echo "=== Fix Class 2: Services referencing other feature services ==="

# dashboard.go imports - need to add imports for GamificationService, Authenticator, QuizSessionService, CacheService
# These come from: gamification/services, auth/services, quiz/services, infrastructure/cache

cat > /tmp/dashboard_imports.txt << 'EOF'
import (
	"context"

	"github.com/goquizvibe/backend/shared/db"
	"github.com/goquizvibe/backend/shared/infrastructure/interfaces"
	gamificationSvc "github.com/goquizvibe/backend/feature/gamification/services"
	authSvc "github.com/goquizvibe/backend/feature/auth/services"
	quizSvc "github.com/goquizvibe/backend/feature/quiz/services"
	cacheSvc "github.com/goquizvibe/backend/shared/infrastructure/cache"
	"github.com/goquizvibe/backend/feature/auth/services"
)

type DashboardService struct {
	queries      *db.Queries
	gamification  *gamificationSvc.GamificationService
	auth          *authSvc.AuthService
	quizSession   *quizSvc.QuizSessionService
	cache         *cacheSvc.CacheService
}
EOF

# Let's look at actual dashboard.go imports first
head -30 backend/feature/dashboard/services/dashboard.go