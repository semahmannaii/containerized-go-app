FROM golang:1.20.1-alpine3.17

WORKDIR /app

COPY . .

RUN go get -d -v ./...

RUN go build -o containerized-go-app .

EXPOSE 8000

CMD ["./containerized-go-app"]