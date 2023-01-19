FROM golang:latest

WORKDIR /app

COPY . .
RUN go build ./src/main.go
CMD ["./main"]