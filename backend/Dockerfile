FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY /backend .

RUN  go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server .

FROM alpine:3.23.4

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app .
COPY --from=builder /app/static ../static
COPY --from=builder /app/initial_data ../initial_data
COPY --from=builder /app/shared/locales ../locales
COPY ../deployment/.env  ../deployment/ 

EXPOSE 7890

CMD ["./server"]