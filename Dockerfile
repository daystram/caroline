FROM golang:1.17 AS builder
ARG VERSION=v0.0.0-development
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -ldflags="-X 'github.com/daystram/caroline/internal/config.version=${VERSION}'" -o . ./cmd/...

FROM ubuntu:20.04
RUN apt-get update && apt-get install -y ffmpeg libopus0 libopus-dev && apt-get clean
RUN apt-get update && apt-get install -y curl python && apt-get clean
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/download/2023.03.04/yt-dlp -o /usr/local/bin/yt-dlp && chmod a+rx /usr/local/bin/yt-dlp
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/caroline /usr/local/bin/
CMD ["/usr/local/bin/caroline"]
