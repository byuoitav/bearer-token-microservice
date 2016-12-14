FROM golang:1.7.1-alpine

RUN apk update && apk upgrade && apk add git

RUN mkdir -p /go/src/github.com/byuoitav
ADD . /go/src/github.com/byuoitav/bearer-token-microservice

WORKDIR /go/src/github.com/byuoitav/bearer-token-microservice
RUN go get -d -v
RUN go install -v

CMD ["/go/bin/ftp-microservice"]

EXPOSE 12000
