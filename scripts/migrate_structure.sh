#!/bin/bash

set -e

echo "=== Step 1: Create directory structure ==="
mkdir -p backend/feature/{auth,dashboard,quiz,admin,learning_materials,editor,gamification}/{handlers,services,ui}
mkdir -p backend/shared/{db,di,config,models,types,locales,middleware,custom_errors,database,metrics,mocks,infrastructure/{cache,storage},ui}
mkdir -p deployment/k8s/{base,overlays/k3s}
mkdir -p scripts

echo "=== Step 2: Move microservices ==="
mv microservices deployment/

echo "=== Step 3: Move deployment configs ==="
mv k8s deployment/
mv monitoring deployment/
mv nginx.conf deployment/
mv docker-compose.yml deployment/

echo "=== Step 4: Move scripts ==="
mv Makefile scripts/
mv run_bg.py scripts/
mv kill_bg.sh scripts/

echo "=== Step 5: Move shared infrastructure ==="
mv db backend/shared/
mv di backend/shared/
mv config backend/shared/
mv models backend/shared/
mv types backend/shared/
mv locales backend/shared/
mv middleware backend/shared/
mv custom_errors backend/shared/
mv database backend/shared/
mv metrics backend/shared/
mv mocks backend/shared/

echo "=== Step 6: Move infrastructure services ==="
mv services/cache_service.go backend/shared/infrastructure/cache/
mv services/storage.go backend/shared/infrastructure/storage/

echo "=== Step 7: Move shared pages (ui) ==="
mv pages/base.templ backend/shared/ui/
mv pages/base_templ.go backend/shared/ui/
mv pages/error.templ backend/shared/ui/
mv pages/error_templ.go backend/shared/ui/
mv pages/helper.go backend/shared/ui/
mv pages/translator.go backend/shared/ui/
mv pages/colors.go backend/shared/ui/

echo "=== Step 8: Move feature code ==="

# Auth
mv handlers/auth.go backend/feature/auth/handlers/
mv handlers/auth.go backend/feature/auth/handlers/
mv services/auth.go backend/feature/auth/services/
mv services/auth_helpers.go backend/feature/auth/services/

# Dashboard
mv handlers/dashboard.go backend/feature/dashboard/handlers/
mv services/dashboard.go backend/feature/dashboard/services/

# Quiz
mv handlers/quiz.go backend/feature/quiz/handlers/
mv services/quiz.go backend/feature/quiz/services/
mv services/quiz_session.go backend/feature/quiz/services/
mv services/quiz_timer.go backend/feature/quiz/services/

# Admin
mv handlers/admin.go backend/feature/admin/handlers/
mv services/admin.go backend/feature/admin/services/

# Learning Materials
mv handlers/learning_materials.go backend/feature/learning_materials/handlers/
mv services/learning_materials.go backend/feature/learning_materials/services/
mv services/typst_grpc_client.go backend/feature/learning_materials/services/

# Editor
mv handlers/editor.go backend/feature/editor/handlers/

# Gamification
mv services/gamification.go backend/feature/gamification/services/

# Pages
mv pages/quiz.templ backend/feature/quiz/ui/
mv pages/quiz_templ.go backend/feature/quiz/ui/
mv pages/quiz_info.templ backend/feature/quiz/ui/
mv pages/quiz_info_templ.go backend/feature/quiz/ui/
mv pages/question_card.templ backend/feature/quiz/ui/
mv pages/question_card_templ.go backend/feature/quiz/ui/
mv pages/timer_component.templ backend/feature/quiz/ui/
mv pages/timer_component_templ.go backend/feature/quiz/ui/
mv pages/session_conflict.templ backend/feature/quiz/ui/
mv pages/session_conflict_templ.go backend/feature/quiz/ui/

mv pages/landing.templ backend/feature/auth/ui/
mv pages/landing_templ.go backend/feature/auth/ui/
mv pages/login.templ backend/feature/auth/ui/
mv pages/login_templ.go backend/feature/auth/ui/
mv pages/register.templ backend/feature/auth/ui/
mv pages/register_templ.go backend/feature/auth/ui/

mv pages/dashboard.templ backend/feature/dashboard/ui/
mv pages/dashboard_templ.go backend/feature/dashboard/ui/

mv pages/admin backend/feature/admin/ui/

mv pages/editor backend/feature/editor/ui/

mv pages/components backend/shared/ui/components

echo "Migration of files completed"