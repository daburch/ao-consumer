FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -v -o bin/ao-consumer ./cmd/ao-consumer

EXPOSE 8080

CMD [ "/app/bin/ao-consumer" ]