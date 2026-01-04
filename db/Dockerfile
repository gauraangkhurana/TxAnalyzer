# Lightweight base image - alpine
FROM alpine:latest

# Install SQLite
RUN apk update && apk add sqlite

# Make Directory for bind mounts
RUN mkdir -p /app/data
RUN mkdir -p /app/scripts

# Set the workdir
WORKDIR /db

# Set the default to sql cmd line
ENTRYPOINT ["sh", "-c", "sqlite3 /app/data/database.db < /app/scripts/init.sql && sqlite3 /app/data/database.db"]
