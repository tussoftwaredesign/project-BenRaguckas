Following help provided:
https://hub.docker.com/_/golang


build:

    docker build -t go-rest-api:0.1.2 .

run:

    docker run -it --rm --name go-rest-test -p 8080:8080 go-rest-api:0.1.2