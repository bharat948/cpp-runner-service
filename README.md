FROM golang:1.23

# Set working directory
WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the code into the container
COPY . ./

# Build the Go app
RUN go build -o /build/app/main .

# Expose service port
EXPOSE 8080

# Run the application command
CMD ["/build/app/main"]

# Test an Example Endpoint 
request.json file format
{
  "code": "Code Content",
  "test_input": "Test File",
   language:"parameter"
}
