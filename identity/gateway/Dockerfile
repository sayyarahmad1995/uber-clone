FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/identity-gateway ./cmd/identity-gateway

FROM alpine:3.20
COPY --from=builder /out/identity-gateway /identity-gateway
EXPOSE 8081
ENTRYPOINT ["/identity-gateway"]
