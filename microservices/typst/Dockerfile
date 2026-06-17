# Must be run from project root, not from microservices/typst

FROM alpine:3.23.4 AS builder

RUN apk add go

WORKDIR /app

COPY backend/shared/pkg /app/pkg
COPY microservices/typst /app/typst

WORKDIR /app/typst

RUN --mount=type=cache,target=$HOME/go/pkg/mod \
--mount=type=cache,target=$HOME/.cache/go-build \
go mod download && go build -o /typst-server .

FROM alpine:3.23.4

RUN apk add typst

COPY --from=builder /typst-server /app/typst-server

WORKDIR /app

CMD ["/app/typst-server"]