FROM golang:1.20-alpine
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main ./cmd/server
EXPOSE 8080
CMD ["/app/main"]