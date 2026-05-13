.PHONY: dev build
.PHONY: dc-up templ-generate rustywind tw-build tw-watch templ-watch

# dev: start full development environment
# runs: containers + one-time generation + parallel file watchers
dev: dc-up
	go tool templ generate
	rustywind --write .
	@make -j 2 tw-watch templ-watch

# generate: one-time generation of all assets (no watching)
# runs: templ generation -> gettextgocodegen -> rustywind -> tailwind build
generate: gettext-generate templ-generate tw-build rustywind

# gettext-generate: generate locales/locales.go from .po files
gettext-generate:
	go tool gettextgocodegen -dir=locales -lang=ru

# build: one-time build without watchers
build: templ-generate gettext-generate go-run

# dc-up: start docker containers in detached mode
# podman compose up -d is idempotent - won't restart already running containers
dc-up:
	@echo "Starting docker containers..."
	podman compose up -d

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
