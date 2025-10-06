# Official Go Alpine Base Image for building the application
FROM golang:1.24-alpine AS builder


# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code into the container
COPY . .

# Build the Go binary
RUN go build -o api cmd/server/main.go

# Final Image Creation Stage using a lightweight Alpine image
FROM alpine:3.21

# Set the working directory
WORKDIR /root/

# Install runtime dependencies including tzdata
RUN apk add --no-cache libc6-compat bash tzdata

# Set timezone (optional)
ENV TZ=Asia/Ho_Chi_Minh

# Copy the built Go binary from the builder image
COPY --from=builder /app/api .

# Copy the .env file
COPY ./.env /root/.env

# Copy the wait-for-it.sh script into the container
COPY ./scripts/wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

# Copy the credentials folder
COPY ./credentials /root/credentials

# Expose the necessary port
EXPOSE 8015

# Set the entrypoint to wait for MariaDB to be ready before starting the application
CMD ["/wait-for-it.sh", "mongo-main:27017", "--", "./api"] 
