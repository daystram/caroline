FROM golang:1.17 AS builder
ARG CGO_ENABLED=0
WORKDIR /app
COPY . .
RUN ["go", "mod", "tidy"]
RUN ["go", "build", "-o", ".", "./cmd/..."]

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/caroline /usr/local/bin/
CMD ["/usr/local/bin/caroline"]
