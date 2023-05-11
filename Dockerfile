# Use a minimal base image
FROM golang:1.19-alpine as builder

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the application source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o aws-ec2-port-scanner

# Use a minimal base image for the final image
FROM alpine:latest

# Copy the built binary from the builder stage
COPY --from=builder /app/aws-ec2-port-scanner /app/

# Set the working directory
WORKDIR /app

# Run the application
ENTRYPOINT ["./aws-ec2-port-scanner"]

