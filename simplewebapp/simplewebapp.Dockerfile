FROM golang:1.12 as build_env

WORKDIR /go/src/github.com/devsamurai/gke/simplewebapp

COPY ./webapp ./webapp
COPY go.mod ./
COPY go.sum ./

ENV PATH="${PATH}:$GOPATH/bin"
ENV GO111MODULE=on

RUN export GOPROXY="https://proxy.golang.org" && export GO111MODULE=on && go test ./...
RUN export GOPROXY="https://proxy.golang.org" && export GO111MODULE=on && CGO_ENABLED=0 go build -o simplewebapp webapp/simplewebapp.go

FROM alpine:latest
WORKDIR /app
COPY --from=build_env /go/src/github.com/devsamurai/gke/simplewebapp /app

EXPOSE 80
ENTRYPOINT ["./simplewebapp"]
