# Start from the official Go image
FROM golang:1.20-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first, then download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -ldflags="-s -w" -o myapp.exe ./cmd/web

# Start a new image to run the Go app
FROM alpine:latest

# Copy the built Go binary from the builder image
COPY --from=builder /app/myapp.exe /myapp.exe

# Set the entrypoint to the Go app
ENTRYPOINT ["/myapp.exe"]
