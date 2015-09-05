FROM golang:1.5-wheezy

RUN apt-get update

RUN DEBIAN_FRONTEND=noninteractive apt-get -y install git mercurial curl rsync ruby

# From the official Golang Dockerfile
#  https://github.com/docker-library/golang/blob/master/1.4/Dockerfile
RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p /go/src/github.com/bitrise-io/bitrise
COPY . /go/src/github.com/bitrise-io/bitrise

WORKDIR /go/src/github.com/bitrise-io/bitrise
# godep
RUN go get -u github.com/tools/godep
RUN godep restore
# install
RUN go install

# setup (downloads envman & stepman)
RUN bitrise setup --minimal

CMD bitrise --version
