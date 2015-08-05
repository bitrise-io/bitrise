FROM ubuntu:14.04.2

RUN apt-get update

RUN DEBIAN_FRONTEND=noninteractive apt-get -y install git mercurial golang

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
RUN go install github.com/tools/godep
RUN godep restore
# install
RUN go install

# include _temp/bin in the PATH
ENV PATH /go/src/github.com/bitrise-io/bitrise/_temp/bin:$PATH

CMD bitrise --version
