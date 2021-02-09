FROM golang:1.15-alpine AS build_base

# Set the Current Working Directory inside the container
WORKDIR /tmp/go-feather-slack-app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

COPY main.go /tmp/go-feather-slack-app/main.go
COPY src /tmp/go-feather-slack-app/src

RUN go mod download && \ 
    go mod vendor && \ 
    go build -o ./out/app .

# Start fresh from a smaller image
FROM alpine:3.13.1

RUN apk add ca-certificates

COPY --from=build_base /tmp/go-feather-slack-app/out/app /app/app

# Run the binary program produced by `go install`
ENTRYPOINT ["/app/app"]

