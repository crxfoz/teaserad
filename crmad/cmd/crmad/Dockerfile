FROM golang:1.18.3-alpine3.16 AS builder

WORKDIR /go/src/app
COPY . .

RUN apk update && apk add git

RUN go install -v ./...

FROM alpine

WORKDIR /go/bin

COPY --from=builder /go/bin/crmad ./crmad

EXPOSE 6060

ENV WAIT_VERSION 2.7.2
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/$WAIT_VERSION/wait /wait
RUN chmod +x /wait

CMD /wait && ./crmad
