Following help provided:
https://hub.docker.com/_/golang


build:

    docker build -t go-rest:0.1.1 .

run:

    docker run -it --rm --name go-rest-test -p 8080:8080 go-rest:0.1.1