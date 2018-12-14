FROM golang:1.11.1-alpine AS builder

RUN apk add bash ca-certificates git

WORKDIR /application

ENV GO111MODULE=on
#ENV PATH_TO_PS_CONFIG=/application/config/config.yml

COPY go.mod go.sum ./
RUN go mod download

# Copy all files in currend directiry into home directory
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -o $GOPATH/bin/qilin_api .

ENTRYPOINT "$GOPATH/bin/qilin_api"