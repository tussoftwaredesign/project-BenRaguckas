Following help provided:
https://hub.docker.com/_/golang


build:
```docker
docker build -t beanbon/go-rest-api:0.2.0 .
```
run:
```docker
docker run -it --rm --name go-rest-test -p 8080:8080 beanbon/go-rest-api:0.2.0
```
run + volume config.xml:
```docker
docker run -v /local/path/to/file1.txt:/container/path/to/file1.txt -it --rm --name go-rest-test -p 8080:8080 beanbon/go-rest-api:0.2.0
```
