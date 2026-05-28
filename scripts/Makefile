.PHONY: dev build
.PHONY: dc-up dc-down templ-generate rustywind tw-build tw-watch templ-watch
.PHONY: monitoring-up monitoring-down
.PHONY: typst-bundle editor-bundle

# dev: start full development environment
# runs: containers + one-time generation + parallel file watchers
dev: dc-up
	cd backend && go tool templ generate
	rustywind --write .
	@make -j 2 tw-watch templ-watch

# generate: one-time generation of all assets
# runs: editor-bundle -> gettext -> templ -> tailwind -> rustywind
generate: editor-bundle gettext-generate templ-generate tw-build rustywind

# gettext-generate: generate locales/locales.go from .po files
gettext-generate:
	cd ../backend && go tool gettextgocodegen -dir=shared/locales -lang=ru

# build: one-time build without watchers
build: templ-generate gettext-generate go-run

# dc-up: start docker containers in detached mode
dc-up:
	@echo "Starting docker containers..."
	cd ../deployment && podman compose up -d

# dc-down: stop docker containers
dc-down:
	@echo "Stopping docker containers..."
	cd ../deployment && podman compose down

# monitoring-up: start monitoring stack (Prometheus, Grafana, node-exporter)
monitoring-up:
	@echo "Starting monitoring containers..."
	cd ../deployment && podman compose -f docker-compose.yml up -d prometheus grafana node-exporter

# monitoring-down: stop monitoring containers
monitoring-down:
	@echo "Stopping monitoring containers..."
	cd ../deployment && podman compose -f docker-compose.yml stop prometheus grafana node-exporter

# templ-generate: generate Go templ files from .templ source files
templ-generate:
	cd ../backend && go tool templ generate

# rustywind: sort CSS classes in .templ files alphabetically
# must run after templ-generate to ensure all .templ files exist
rustywind:
	rustywind --write .

# tw-build: one-time Tailwind CSS build (no watch mode)
tw-build:
	cd ../ && npx @tailwindcss/cli -i backend/shared/ui/style/app.css -o backend/static/style/app.css

# tw-watch: Tailwind CSS with file watching and auto-rebuild
tw-watch:
	npx @tailwindcss/cli -i backend/shared/ui/style/app.css -o backend/static/style/app.css --watch

# templ-watch: Go templ with file watching and auto-regeneration
# runs main.go when templ files change
templ-watch:
	cd ../backend && go tool templ generate --watch --cmd="go run ."

go-run:
	cd ../backend && go run .

# typst-bundle: build typst.ts ES module bundle
typst-bundle:
	cd ../backend/static/scripts/editor && pnpm run typst-bundle

# editor-bundle: build editor IIFE bundle
editor-bundle:
	cd ../backend/static/scripts/editor && pnpm run editor-bundle

# sql-generate: generate SQLC code
sql-generate:
	cd ../backend && go tool sqlc generate

# wire-generate: regenerate DI wire code
wire-generate:
	cd ../backend && go generate ./...

# docker-build: build docker image
docker-build:
	cd ../deployment && podman build -t goquizvibe:latest -f ../Dockerfile ..