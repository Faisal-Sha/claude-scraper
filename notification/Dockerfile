FROM golang:1.23.8

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o notification-service ./main.go

EXPOSE 50053

CMD ["./notification-service"]
