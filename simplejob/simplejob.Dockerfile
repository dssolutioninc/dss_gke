FROM golang:1.12 as build_env

WORKDIR /go/src/github.com/itdevsamurai/gke/simplejob

COPY ./job ./job
COPY go.mod ./
COPY go.sum ./

ENV PATH="${PATH}:$GOPATH/bin"
ENV GO111MODULE=on

RUN export GOPROXY="https://proxy.golang.org" && export GO111MODULE=on && CGO_ENABLED=0 go build -o simplejob job/simplejob.go

FROM alpine:latest
WORKDIR /app
COPY --from=build_env /go/src/github.com/devsamurai/gke/simplejob /job
