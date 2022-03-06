FROM golang:1.17 AS builder
WORKDIR /app
COPY . .
RUN ["go", "mod", "tidy"]
RUN ["go", "build", "-o", ".", "./cmd/..."]

FROM ubuntu:20.04
RUN apt-get update && apt-get install -y ffmpeg libopus0 libopus-dev && apt-get clean
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/caroline /usr/local/bin/
CMD ["/usr/local/bin/caroline"]
