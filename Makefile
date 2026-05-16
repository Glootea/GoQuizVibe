.PHONY: dev build
.PHONY: dc-up dc-down templ-generate rustywind tw-build tw-watch templ-watch
.PHONY: monitoring-up monitoring-down
.PHONY: typst-copy-wasm typst-bundle editor-bundle

# dev: start full development environment
# runs: containers + one-time generation + parallel file watchers
dev: dc-up
	go tool templ generate
	rustywind --write .
	@make -j 2 tw-watch templ-watch

# generate: one-time generation of all assets
# runs: typst-copy-wasm -> typst-bundle -> editor-bundle -> gettext -> templ -> tailwind -> rustywind
generate: typst-copy-wasm typst-bundle editor-bundle gettext-generate templ-generate tw-build rustywind

# gettext-generate: generate locales/locales.go from .po files
gettext-generate:
	go tool gettextgocodegen -dir=locales -lang=ru

# build: one-time build without watchers
build: templ-generate gettext-generate go-run

# dc-up: start docker containers in detached mode
dc-up:
	@echo "Starting docker containers..."
	podman compose up -d

# dc-down: stop docker containers
dc-down:
	@echo "Stopping docker containers..."
	podman compose down

# monitoring-up: start monitoring stack (Prometheus, Grafana, node-exporter)
monitoring-up:
	@echo "Starting monitoring containers..."
	podman compose -f docker-compose.yml up -d prometheus grafana node-exporter

# monitoring-down: stop monitoring containers
monitoring-down:
	@echo "Stopping monitoring containers..."
	podman compose -f docker-compose.yml stop prometheus grafana node-exporter

# templ-generate: generate Go templ files from .templ source files
templ-generate:
	go tool templ generate

# rustywind: sort CSS classes in .templ files alphabetically
# must run after templ-generate to ensure all .templ files exist
rustywind:
	rustywind --write .

# tw-build: one-time Tailwind CSS build (no watch mode)
tw-build:
	tailwindcss -i ./pages/styles/app.css -o ./static/style/app.css

# tw-watch: Tailwind CSS with file watching and auto-rebuild
tw-watch:
	tailwindcss -i ./pages/styles/app.css -o ./static/style/app.css --watch

# templ-watch: Go templ with file watching and auto-regeneration
# runs main.go when templ files change
templ-watch:
	go tool templ generate --watch --cmd="go run main.go"

go-run:
	go run main.go

# typst-copy-wasm: copy WASM modules to static/wasm/
typst-copy-wasm:
	@echo "Copying WASM modules..."
	@mkdir -p static/wasm
	cp /Users/glootea/Documents/Dev/Projects/GoQuizVibe.worktrees/typst/static/scripts/editor/node_modules/.pnpm/@myriaddreamin+typst-ts-web-compiler@0.7.0-rc2/node_modules/@myriaddreamin/typst-ts-web-compiler/pkg/*.wasm static/wasm/
	cp /Users/glootea/Documents/Dev/Projects/GoQuizVibe.worktrees/typst/static/scripts/editor/node_modules/.pnpm/@myriaddreamin+typst-ts-renderer@0.7.0-rc2/node_modules/@myriaddreamin/typst-ts-renderer/pkg/*.wasm static/wasm/
	@echo "WASM modules copied"
	@ls -la static/wasm/

# typst-bundle: build typst.ts ES module bundle
typst-bundle:
	cd static/scripts/editor && pnpm run typst-bundle

# editor-bundle: build editor IIFE bundle
editor-bundle:
	cd static/scripts/editor && pnpm run editor-bundle
