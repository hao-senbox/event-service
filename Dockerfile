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
CMD ["/wait-for-it.sh", "event_db:27017", "--", "./api"] 
