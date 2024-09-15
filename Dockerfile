FROM golang:1.22.0-alpine

WORKDIR /app

COPY ./tender/go.mod ./
COPY ./tender/go.sum ./
COPY .env .env
RUN go mod download

COPY ./tender .
RUN go build -o main cmd/main.go
RUN go mod tidy

EXPOSE 8080

CMD ["/app/main"]