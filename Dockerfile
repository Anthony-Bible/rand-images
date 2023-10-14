FROM golang:1.21.2-bookworm as builder

LABEL org.opencontainers.image.source=https://github.com/anthony-bible/rand-images
WORKDIR /app
copy . .

RUN go build -o server

FROM alpine:3.14

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server /server
COPY --from=builder /app/images /images


CMD ["/server"]

