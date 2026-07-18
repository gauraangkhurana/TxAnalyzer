# Use golang base image
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates gcc musl-dev
WORKDIR /app

# Copy over go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of the source code
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk add --no-cache git ca-certificates

# Copy over the built binary
COPY --from=builder /app/main .

ENTRYPOINT ["./main"]
