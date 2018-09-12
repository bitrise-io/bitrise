FROM ubuntu:16.04

ENV PROJ_NAME envman
ENV BITRISE_CLI_VERSION 1.21.0
ENV GO_VERSION go1.10.3.linux-amd64.tar.gz

RUN apt-get update

RUN DEBIAN_FRONTEND=noninteractive apt-get -y install \
    # Requiered for Bitrise CLI
    git \
    mercurial \
    curl \
    wget \
    rsync \
    sudo \
    expect \
    build-essential

#
# Install Bitrise CLI
RUN curl -L https://github.com/bitrise-io/bitrise/releases/download/$BITRISE_CLI_VERSION/bitrise-$(uname -s)-$(uname -m) > /usr/local/bin/bitrise
RUN chmod +x /usr/local/bin/bitrise
RUN bitrise setup

#
# Install Go
#  from official binary package
RUN wget -q https://storage.googleapis.com/golang/$GO_VERSION -O go-bins.tar.gz \
    && tar -C /usr/local -xvzf go-bins.tar.gz \
    && rm go-bins.tar.gz

# ENV setup
ENV PATH $PATH:/usr/local/go/bin
# Go Workspace dirs & envs
# From the official Golang Dockerfile
#  https://github.com/docker-library/golang
ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH
# 755 because Ruby complains if 777 (warning: Insecure world writable dir ... in PATH)
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 755 "$GOPATH"

RUN mkdir -p /go/src/github.com/bitrise-io/$PROJ_NAME
COPY . /go/src/github.com/bitrise-io/$PROJ_NAME

WORKDIR /go/src/github.com/bitrise-io/$PROJ_NAME

# install
RUN go install
CMD $PROJ_NAME --version
