# Build image
FROM golang:1.11.2 as build

# Copy the local package files to the container
ADD . /src
WORKDIR /src
ENV VERSION=dev-build

# Build the project inside the container
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
    -o /app /src/main.go

# Create our deploy image
FROM scratch

# Certs for ssl
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy our static executable.
COPY --from=build /app /app

# Run the server
ENTRYPOINT ["/app"]

