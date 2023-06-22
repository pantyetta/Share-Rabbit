FROM golang:1.20-alpine

WORKDIR /usr/src/golang

# go.mod go.sum Copy local => docker
COPY ./go/go.mod ./go/go.sum ./
RUN go mod download && go mod verify

COPY ./go/. .
RUN go build -v -o /usr/local/bin/golang ./...

CMD [ "golang" ]