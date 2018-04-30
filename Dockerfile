FROM golang:1.10.1

RUN mkdir -p /go/src
COPY . /go/src
RUN go get -d ./...

WORKDIR /go/src

EXPOSE 8085

CMD ["go", "run", "main.go" ]
