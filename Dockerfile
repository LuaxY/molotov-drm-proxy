FROM golang:alpine AS builder
WORKDIR /proxy
COPY . .
RUN GOOS=linux go build -o proxy ./cmd/proxy

FROM alpine AS proxy
WORKDIR /
COPY --from=builder /proxy/proxy proxy
ENTRYPOINT ["/proxy"]