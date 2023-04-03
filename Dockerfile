
#FROM golang:1.20.2-buster  as builder
FROM golang:1.20.2-buster  

WORKDIR /blobvm
COPY . .

RUN /blobvm/scripts/build.sh 
RUN /blobvm/scripts/run2-pre.sh 1.9.7
EXPOSE 12352
EXPOSE 12353
# CMD /blobvm/scripts/run2.sh 1.9.7
CMD /blobvm/scripts/run2-next.sh 1.9.7
