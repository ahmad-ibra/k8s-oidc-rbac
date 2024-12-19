FROM golang:1.22 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o pod-service main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/pod-service /pod-service
ENTRYPOINT ["/pod-service"]
