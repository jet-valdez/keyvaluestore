# Stage 1: Compile the binary in a containerized Golang environment
#
FROM golang:1.22-alpine as build

# Copy the source files from the host
COPY . /go/src/keyvaluestore

# Set the working directory to the same place we copied the code
WORKDIR /go/src/keyvaluestore

# Download the dependency code
RUN go get github.com/gorilla/mux github.com/lib/pq

# Build the binary!
RUN CGO_ENABLED=0 GOOS=linux go build -o keyvaluestore


# Stage 2: Build the Key-Value Store image proper
#
# Use a "scratch" image, which contains no distribution files
FROM scratch as image

# Copy the binary from the build container
COPY --from=build /go/src/keyvaluestore/keyvaluestore .

# Tell Docker we'll be using port 8080
EXPOSE 8080

# Tell Docker to execute this command on a "docker run"
CMD ["/keyvaluestore"]