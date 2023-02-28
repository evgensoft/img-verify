FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o img-verify

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/img-verify /app/img-verify

EXPOSE 8080

CMD ["./img-verify"]