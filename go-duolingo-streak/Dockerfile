FROM golang:latest
LABEL authors="galih"

ENV PROJECT_DIR=/app \
    GO111MODULE=on \
    CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go get github.com/githubnemo/CompileDaemon
RUN go install github.com/githubnemo/CompileDaemon

COPY ./ ./


EXPOSE 8080
#CMD ["./bin/server/server"]
#
ENTRYPOINT CompileDaemon -build="go build -o /bin/server" -command="/bin/server" -build-dir=/app

