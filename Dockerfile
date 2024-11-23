ARG GO_VERSION=1.16.6

FROM golang:${GO_VERSION}-alpine as builder

RUN go env -w GOPROXY=direct
RUN apk add --no-cache git
RUN apk --no-cache add ca-certificates && update-ca-certificates



COPY ./ /src
WORKDIR /src
RUN cd /src
RUN go mod download

