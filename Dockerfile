FROM golang:latest as builder

WORKDIR /build
COPY . .
RUN go mod download && \
        CGO_ENABLED=0 go build -o runtime-metrics cmd/runtime-metrics/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /build/runtime-metrics .
CMD ["./runtime-metrics"]