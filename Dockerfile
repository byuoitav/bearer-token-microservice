FROM golang:1.7.1-alpine

RUN apk --no-cache update && \
    apk --no-cache add python py-pip py-setuptools ca-certificates groff less && \
    pip --no-cache-dir install awscli && \
    rm -rf /var/cache/apk/*

RUN mkdir -p /go/src/github.com/byuoitav
ADD . /go/src/github.com/byuoitav/bearer-token-microservice

WORKDIR /go/src/github.com/byuoitav/bearer-token-microservice
RUN go get -d -v
RUN go install -v

CMD ["/go/bin/bearer-token-microservice"]

EXPOSE 12000
