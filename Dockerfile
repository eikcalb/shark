# Dockerfile References: https://docs.docker.com/engine/reference/builder/

# Start from the latest golang base image
FROM golang:latest AS build

# Set the Current Working Directory inside the container
WORKDIR /app

# Add files to app folder
ADD . /app

# Build the Go app
RUN go build -o shark .

# Start a smaller build
FROM busybox:latest AS final

# Set the Current Working Directory inside the container
WORKDIR /app

# Expose ports to the outside world
EXPOSE 8080

# Copy file from build stage into main image
COPY --from=build /app .

# Command to run the executable
CMD [ "./shark"]
