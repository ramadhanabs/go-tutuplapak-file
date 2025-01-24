# Stage 1: Build the binary
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .

# Stage 2: Create a minimal image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
COPY .env .
RUN chmod +x ./main 
EXPOSE 8080
