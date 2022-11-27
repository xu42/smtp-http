FROM golang:1.19.3-alpine

RUN mkdir /app

ADD main.go /app

WORKDIR /app

RUN go build -o main main.go

CMD ["/app/main"]