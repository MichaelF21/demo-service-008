FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /demo-service-008 .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /demo-service-008 /usr/local/bin/demo-service-008
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/demo-service-008"]
