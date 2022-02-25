FROM golang:1.17 as builder
#FROM golang:1.17-alpine3.15

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o img-verify

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/img-verify /app/img-verify

EXPOSE 8080

CMD ["./img-verify"]