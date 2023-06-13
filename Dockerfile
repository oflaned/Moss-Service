FROM golang:latest
WORKDIR /app
COPY . .
RUN go build -o microservice
CMD ["./microservice"]