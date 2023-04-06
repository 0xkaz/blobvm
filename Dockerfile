
#FROM golang:1.20.2-buster  as builder
FROM golang:1.20.2-buster  

WORKDIR /blobvm
COPY . .

RUN /blobvm/scripts/build.sh 
RUN /blobvm/scripts/run2-pre.sh 1.7.13

EXPOSE 12352
EXPOSE 12353
EXPOSE 9650
EXPOSE 9652
EXPOSE 9654
EXPOSE 9656
EXPOSE 9658

# CMD /blobvm/scripts/run2.sh 1.7.13
CMD /blobvm/scripts/run2-next.sh 1.7.13
