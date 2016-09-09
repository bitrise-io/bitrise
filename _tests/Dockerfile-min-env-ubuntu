# This is the base Ubuntu environment sufficient for running the Bitrise CLI
FROM ubuntu

RUN apt-get update -qq

# Required for `bitrise setup`
RUN DEBIAN_FRONTEND=noninteractive apt-get -y install ca-certificates git
# Required for `deps`
RUN DEBIAN_FRONTEND=noninteractive apt-get -y install sudo
