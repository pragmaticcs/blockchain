FROM golang:1.19

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /blockchain

EXPOSE 8080 

ARG DATA_DIR=""
ARG WALLET_NAME="satoshi"

CMD ["/blockchain", "--listenAddr", ":8080", "--dataDir", ${DATA_DIR}, "--walletName", ${WALLET_NAME}, "node", "start"]

