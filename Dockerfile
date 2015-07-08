FROM ubuntu:14.04.2

RUN apt-get update

RUN DEBIAN_FRONTEND=noninteractive apt-get -y install git mercurial golang

RUN mkdir /go
RUN mkdir /go/bin
RUN mkdir -p /go/src/github.com/bitrise-io/bitrise-cli
RUN export GOPATH=/go
ENV GOPATH /go
RUN export PATH=$PATH:$GOPATH/bin
ENV PATH $PATH:$GOPATH/bin

WORKDIR /go/src/github.com/bitrise-io/bitrise-cli

COPY . /go/src/github.com/bitrise-io/bitrise-cli

RUN go get ./...
RUN go install

CMD bitrise-cli --version
