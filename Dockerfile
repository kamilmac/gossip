FROM golang:1.6.2-alpine
COPY . /app/gossip
WORKDIR /app/gossip
RUN go build main.go
CMD ["./main"]