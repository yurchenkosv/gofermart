FROM golang:1.18 AS builder
WORKDIR WORKDIR /build
COPY . .
RUN go mod download && \
    GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o /build/gofermart ./cmd/gophermart/

FROM alpine:latest
RUN apk --no-cache add ca-certificates && \
    wget -O - $(wget -O - https://api.github.com/repos/powerman/dockerize/releases/latest | \
    grep -i /dockerize-$(uname -s)-$(uname -m)\" | cut -d\" -f4) | \
    install /dev/stdin /usr/local/bin/dockerize
WORKDIR /app
COPY --from=builder /build/gofermart ./
COPY db ./db/
CMD ["./gofermart"]
