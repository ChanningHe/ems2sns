FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/ems2sns ./cmd/ems2sns

# ---

FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/ems2sns /app/ems2sns
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

ENTRYPOINT ["/app/ems2sns"]
