FROM golang:1.24.3 AS builder

WORKDIR /build
COPY . .

RUN go build -o /go/bin/api main.go
COPY config.yml /go/bin

EXPOSE 8080

CMD ["/go/bin/api"]
