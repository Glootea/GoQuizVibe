FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY /backend .

RUN --mount=type=cache,target=$HOME/go/pkg/mod \
--mount=type=cache,target=$HOME/.cache/go-build \
    go mod download && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

FROM alpine:3.23.4

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server ./bin/server
COPY --from=builder /app/static static
COPY --from=builder /app/initial_data initial_data
COPY --from=builder /app/shared/locales shared/locales
COPY ../deployment/.env  ../deployment/ 
COPY --from=builder app/shared/sql/migrations shared/sql/migrations

EXPOSE 7890

ENTRYPOINT ["./bin/server"] 