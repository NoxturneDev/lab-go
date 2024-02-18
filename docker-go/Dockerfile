# Go Version
FROM golang:1.22.0-bookworm

WORKDIR /go/src

# Environment variables which CompileDaemon requires to run
ENV PROJECT_DIR=/go/src \
  GO111MODULE=on \
  CGO_ENABLED=0

# Basic setup of the container
COPY . /go/src

# Get CompileDaemon
RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

# The build flag sets how to build after a change has been detected in the source code
# The command flag sets how to run the app after it has been built
ENTRYPOINT CompileDaemon -build="go build -o api" -command="./api"
