
#FROM golang:1.20.2-buster  as builder
FROM golang:1.20.2-buster  

WORKDIR /blobvm
COPY . .

RUN /blobvm/scripts/build.sh

CMD /blobvm/scripts/run.sh 1.9.7
