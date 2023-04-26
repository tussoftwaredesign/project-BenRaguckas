FROM golang:1.19-alpine

WORKDIR /usr/src/app

ENV MINIO_SERV="localhost:9000"
ENV MRQ_SERV="localhost:5762"
ENV MONGO_SERV="localhost:27017"
ENV MINIO_CRED_ID="minio"
ENV MINIO_CRED_KEY="minio123"
ENV RMQ_CRED_ID="guest"
ENV RMQ_CRED_KEY="guest"
ENV MONGO_CRED_ID="mongo"
ENV MONGO_CRED_KEY="mongo123"

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

#   Copy the config
COPY . .
RUN go build -v -o /usr/local/bin/app ./src/...

# EXPOSE 8080/TCP
EXPOSE 8080

ENTRYPOINT ["sh", "-c", "app \
    -minio_serv $MINIO_SERV \
    -rmq_serv $RMQ_SERV \
    -mongo_serv $MONGO_SERV \
    -minio_cred_id ${MINIO_CRED_ID} \
    -minio_cred_key ${MINIO_CRED_KEY} \
    -rmq_cred_id ${RMQ_CRED_ID} \
    -rmq_cred_key ${RMQ_CRED_KEY} \
    -mongo_cred_id ${MONGO_CRED_ID} \
    -mongo_cred_key ${MONGO_CRED_KEY} \
    "]