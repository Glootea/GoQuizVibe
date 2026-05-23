FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY pkg /app/pkg

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

FROM alpine:3.23.4

RUN apk add --no-cache typst ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app .
COPY --from=builder /app/static ./static
COPY --from=builder /app/initial_data ./initial_data
COPY --from=builder /app/locales ./locales

EXPOSE 7890

CMD ["./server"]