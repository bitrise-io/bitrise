FROM golang:1.21

ENV PROJ_NAME bitrise

RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get -y install curl git mercurial rsync ruby sudo

RUN mkdir -p /go/src/github.com/bitrise-io/$PROJ_NAME
COPY . /go/src/github.com/bitrise-io/$PROJ_NAME

RUN go install github.com/bitrise-io/stepman@latest

WORKDIR /go/src/github.com/bitrise-io/$PROJ_NAME
RUN go install
RUN bitrise setup

CMD bitrise version
