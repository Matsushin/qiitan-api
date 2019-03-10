FROM golang:1.11.0-alpine3.8

ENV GOPATH /go
ENV CGO_ENABLED 0
WORKDIR /go/src/github.com/Matsushin/qiitan-api/

RUN apk update && apk upgrade && \
    apk add --no-cache git && \
    go get github.com/pilu/fresh

# 最初にonにするとfreshのgetができなくなるのでこのタイミングでonにする
ENV GO111MODULE on
COPY ./ /go/src/github.com/Matsushin/qiitan-api/

EXPOSE 8080
CMD ["go", "run", "main.go"]