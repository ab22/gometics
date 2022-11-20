FROM golang:latest as builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o runtime-metrics -ldflags="-X main.commit=$(git rev-parse HEAD)" cmd/runtime-metrics/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /build/runtime-metrics .
CMD ["./runtime-metrics"]
