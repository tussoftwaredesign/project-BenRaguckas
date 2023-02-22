FROM golang:1.19-alpine

WORKDIR /usr/src/app

# To replace with ENV variables perhaps ?
ENV MINIO_HOST "10.128.0.11"
ENV MINIO_PORT "9000"
ENV MINIO_CRED_ID ""
ENV MINIO_CRED_KEY ""
ENV RMQ_HOST "10.128.0.12"
ENV RMQ_PORT "5672"
ENV RMQ_CRED_ID ""
ENV RMQ_CRED_KEY ""



# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app ./src/...

#   Copy the config


# EXPOSE 8080/TCP
EXPOSE 8080

CMD [ "echo test" ]
ENTRYPOINT ["sh", "-c", "app \
    -minio_host ${MINIO_HOST} \
    -minio_port ${MINIO_PORT} \
    # -minio_cred_id ${MINIO_CRED_ID} \
    # -minio_cred_key ${MINIO_CRED_KEY} \
    -rmq_host ${RMQ_HOST} \
    -rmq_port ${RMQ_PORT} \
    # -rmq_cred_id ${RMQ_CRED_ID} \
    # -rmq_cred_key ${RMQ_CRED_KEY} \
    "]