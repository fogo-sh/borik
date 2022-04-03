FROM golang:1.18-buster

WORKDIR /build
COPY . .
RUN go mod download &&\
    go mod verify &&\
    apt-get update &&\
    apt-get install -y libmagickwand-dev &&\
    go build

ENTRYPOINT ["/build/borik"]
