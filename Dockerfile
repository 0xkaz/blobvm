
#FROM golang:1.20.2-buster  as builder
FROM golang:1.20.2-buster  

WORKDIR /blobvm
COPY . .

# RUN /blobvm/scripts/build.sh 
# RUN MODE=test E2E=true /blobvm/scripts/run.sh 1.7.13
RUN MODE=test E2E=true /blobvm/scripts/run.sh 1.9.3

EXPOSE 12352
EXPOSE 12353
EXPOSE 9650
EXPOSE 9652
EXPOSE 9654
EXPOSE 9656
EXPOSE 9658


# # 
# RUN export CGO_CFLAGS="-O -D__BLST_PORTABLE__"
# RUN export VERSION=1.9.3
# RUN echo $VERSION 

# RUN export GOARCH=$(go env GOARCH)
# RUN export GOOS=$(go env GOOS)
# RUN export DOWNLOAD_URL=https://github.com/ava-labs/avalanchego/releases/download/v${VERSION}/avalanchego-linux-${GOARCH}-v${VERSION}.tar.gz
# RUN export DOWNLOAD_PATH=/tmp/avalanchego.tar.gz
# RUN curl -L ${DOWNLOAD_URL} -o ${DOWNLOAD_PATH}
# RUN tar xzvf ${DOWNLOAD_PATH} -C /tmp
# RUN find /tmp/avalanchego-v${VERSION}
# RUN export AVALANCHEGO_PATH=/tmp/avalanchego-v${VERSION}/avalanchego
# RUN export AVALANCHEGO_PLUGIN_DIR=/tmp/avalanchego-v${VERSION}/plugins
# RUN echo "building blobvm"
# RUN go build \
# -o /tmp/avalanchego-v${VERSION}/plugins/kM6h4LYe3AcEU1MB2UNg6ubzAiDAALZzpVrbX8zn3hXF6Avd8 \
# ./cmd/blobvm
# RUN find /tmp/avalanchego-v${VERSION}
# RUN echo "building blob-cli"
# RUN go build -v -o /tmp/blob-cli ./cmd/blob-cli

# # 
# RUN echo "creating allocations file"


# # RUN export AVALANCHEGO_PATH=/tmp/avalanchego-v${VERSION}/avalanchego
# # RUN export AVALANCHEGO_PLUGIN_DIR=/tmp/avalanchego-v${VERSION}/plugins
# # RUN echo "building avalanchego"
# # RUN mkdir -p /tmp/avalanchego-src
# # WORKDIR /tmp/avalanchego-src
# # RUN git clone –depth 1 https://github.com/ava-labs/avalanchego.git

# # RUN /tmp/avalanchego-src
# WORKDIR /myvm
# # CMD ["/bin/bash", "/myvm/examples/weavedbvm/scripts/run.sh"]



# # CMD /blobvm/scripts/run.sh 1.7.13
# # CMD /blobvm/scripts/run.sh 1.9.3

