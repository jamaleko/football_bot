FROM golang:1.26.1

WORKDIR /app

COPY . .

CMD ["./bot"]
