FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["go", "run", "main.go"]
