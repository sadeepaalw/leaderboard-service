# Start from the official Golang image
FROM golang:1.24.3-alpine AS builder
WORKDIR /app
COPY . .
RUN cd cmd/server && go build -o /app/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server ./server
EXPOSE 8080
CMD ["./server"] 