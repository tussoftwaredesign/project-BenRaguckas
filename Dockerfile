FROM golang:1.19-alpine

WORKDIR /usr/src/app

ENV MINIO_SERV="localhost:9090"
ENV MRQ_SERV="localhost:15762"
ENV MINIO_CRED_ID="minio"
ENV MINIO_CRED_KEY="minio123"
ENV RMQ_CRED_ID="guest"
ENV RMQ_CRED_KEY="guest"



# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app ./src/...

#   Copy the config


# EXPOSE 8080/TCP
EXPOSE 8080

ENTRYPOINT ["sh", "-c", "app \
    -minio_serv $MINIO_SERV \
    -rmq_serv $RMQ_SERV \
    -minio_cred_id ${MINIO_CRED_ID} \
    -minio_cred_key ${MINIO_CRED_KEY} \
    -rmq_cred_id ${RMQ_CRED_ID} \
    -rmq_cred_key ${RMQ_CRED_KEY} \
    "]