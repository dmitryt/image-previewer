FROM golang:1.14 AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o main .

# Build a small image
FROM alpine:latest

ENV PORT 8082
ENV CACHE_DIR .cache
ENV CACHE_SIZE 10

COPY --from=builder /build/main /
EXPOSE 8082

# Command to run
ENTRYPOINT ["/main"]