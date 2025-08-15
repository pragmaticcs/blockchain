# syntax=docker.io/docker/dockerfile:1.7-labs
FROM golang:1.23.4

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY --exclude=data . ./ 

RUN mkdir data

RUN mkdir bin

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/blockchain

EXPOSE 8081

ARG DATA_DIR="data/satoshi"

COPY $DATA_DIR data

CMD ./bin/blockchain --dataDir ./data node start --daemon

