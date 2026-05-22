FROM alpine:3.23.4 AS builder

RUN apk add go

WORKDIR /app

COPY pkg /app/pkg
COPY microservices/typst /app/typst

WORKDIR /app/typst
RUN go mod download

RUN go build -o /typst-server .

FROM alpine:3.23.4

RUN apk add typst

COPY --from=builder /typst-server /app/typst-server

WORKDIR /app

CMD ["/app/typst-server"]