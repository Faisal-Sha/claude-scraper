FROM golang:1.23.8

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o product-analysis-service ./main.go

EXPOSE 50052

CMD ["./product-analysis-service"]
