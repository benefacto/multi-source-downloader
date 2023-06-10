# Choose a base image that includes Go
FROM golang:1.20

# Set your work directory in the docker container
WORKDIR /app

# Copy go module and sum files to the work directory
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download 

# Copy the source from the current directory to the work directory in the container
COPY . .

# Build the application
RUN go build -o main ./cmd/main.go

# Command to run the application when the docker container starts
CMD ["/app/main"]
