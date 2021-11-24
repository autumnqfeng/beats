FROM golang:1.16.10-alpine3.13 AS builder

WORKDIR /root/go/src/github.com/elastic/beats

ENV GOPROXY https://goproxy.cn
ENV GOPATH /root/go
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN cd filebeat/ && CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o filebeat

FROM alpine:3.13 AS final

WORKDIR /filebeat
COPY --from=builder /root/go/src/github.com/elastic/beats/filebeat/filebeat /filebeat/

ENTRYPOINT ["/filebeat/filebeat"]
