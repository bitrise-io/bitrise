FROM golang:1.7

ENV PROJ_NAME bitrise

RUN apt-get update

RUN DEBIAN_FRONTEND=noninteractive apt-get -y install curl git mercurial rsync ruby sudo

# From the official Golang Dockerfile
#  https://github.com/docker-library/golang/blob/master/1.4/Dockerfile
RUN mkdir -p /go/src /go/bin && chmod -R 777 /go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

# Install required (testing) tools
#  Install dependencies
RUN go get -u github.com/tools/godep
#  Check for unhandled errors
RUN go get -u github.com/kisielk/errcheck
#  Go lint
RUN go get -u github.com/golang/lint/golint

RUN mkdir -p /go/src/github.com/bitrise-io/$PROJ_NAME
COPY . /go/src/github.com/bitrise-io/$PROJ_NAME

WORKDIR /go/src/github.com/bitrise-io/$PROJ_NAME
# install
RUN go install

# setup (downloads envman & stepman)
RUN bitrise setup
RUN $HOME/.bitrise/tools/stepman setup -c https://github.com/bitrise-io/bitrise-steplib.git

CMD bitrise version
