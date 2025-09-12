FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN adduser -D -s /bin/sh appuser

WORKDIR /root/

COPY --from=builder /app/main .

RUN chown appuser:appuser main

USER appuser

EXPOSE 8080

CMD ["./main"]
