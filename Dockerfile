FROM golang:1.15-alpine AS build_base

# Set the Current Working Directory inside the container
WORKDIR /tmp/go-feather-slack-app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

COPY main.go /tmp/go-feather-slack-app/main.go
COPY src /tmp/go-feather-slack-app/src

RUN go mod download
RUN go mod vendor

# Unit tests
RUN CGO_ENABLED=0 go test -v

# Build the Go app
RUN go build -o ./out/app .

# Start fresh from a smaller image
FROM alpine:3.9 
RUN apk add ca-certificates

COPY --from=build_base /tmp/go-feather-slack-app/out/app /app/app

# This container exposes port 8080 to the outside world
EXPOSE 8080

# Run the binary program produced by `go install`
CMD ["/app/app"]

