# Use the official Go image as the base image
FROM golang:latest

# Set the working directory in the container to /app
WORKDIR /app

# Copy the local code into the container
COPY ./src .

# Install the required packages
RUN go get -d -v ./...

# Build the app
RUN go build -o main .

# Set the environment variable for the API endpoint
ENV API_ENDPOINT http://localhost:8080

# Expose port 8080 to the host so that the app can be accessed from outside the container
EXPOSE 8080

# Run the app when the container starts
CMD ["./main"]
