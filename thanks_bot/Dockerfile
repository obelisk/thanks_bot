# Build stage docker file
FROM golang:latest AS builder
WORKDIR /tmp/build/
COPY . .
RUN make thanks_bot

# Build a container to get certs
FROM alpine:latest as certs
RUN apk --update add ca-certificates

# Copy in to our new docker container
FROM scratch
COPY --from=builder /tmp/build/build/thanks_bot /root/thanks_bot
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/root/thanks_bot"]

