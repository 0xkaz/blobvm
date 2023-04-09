#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if ! [[ "$0" =~ scripts/watch.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi
## avalanchego
VERSION=1.9.3
GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)
DOWNLOAD_URL=https://github.com/ava-labs/avalanchego/releases/download/v${VERSION}/avalanchego-linux-${GOARCH}-v${VERSION}.tar.gz
DOWNLOAD_PATH=/tmp/avalanchego.tar.gz
if [[ ${GOOS} == "darwin" ]]; then
  DOWNLOAD_URL=https://github.com/ava-labs/avalanchego/releases/download/v${VERSION}/avalanchego-macos-v${VERSION}.zip
  DOWNLOAD_PATH=/tmp/avalanchego.zip
fi

rm -rf /tmp/avalanchego-v${VERSION}
rm -rf /tmp/avalanchego-build
rm -f ${DOWNLOAD_PATH}


echo "downloading avalanchego ${VERSION} at ${DOWNLOAD_URL}"
curl -L ${DOWNLOAD_URL} -o ${DOWNLOAD_PATH}



echo "extracting downloaded avalanchego"
if [[ ${GOOS} == "linux" ]]; then
  tar xzvf ${DOWNLOAD_PATH} -C /tmp
elif [[ ${GOOS} == "darwin" ]]; then
  unzip ${DOWNLOAD_PATH} -d /tmp/avalanchego-build
  mv /tmp/avalanchego-build/build /tmp/avalanchego-v${VERSION}
fi
find /tmp/avalanchego-v${VERSION}

AVALANCHEGO_PATH=/tmp/avalanchego-v${VERSION}/avalanchego
AVALANCHEGO_PLUGIN_DIR=/tmp/avalanchego-v${VERSION}/plugins


# blobvm

LAST=""
TMP=""
while [ true ]
do
  # TMP=`find . | grep '.go' | xargs stat -f %m  | sort -n | tail -n 1`
  TMP=`find . | egrep '\.go|\.sh$'| xargs stat -f %m  | sort -n | tail -n 1`
  if [ "$TMP" != "$LAST" ]; then
    LAST=$TMP
    echo "${LAST}"

    name="kM6h4LYe3AcEU1MB2UNg6ubzAiDAALZzpVrbX8zn3hXF6Avd8"
    mkdir -p ./build
    echo "Building blobvm in ./build/$name"
    go build -o ./build/$name ./cmd/blobvm
    cp ./build/$name $AVALANCHEGO_PLUGIN_DIR/
    echo "Building blob-cli in ./build/blob-cli"
    go build -o ./build/blob-cli ./cmd/blob-cli
    echo $AVALANCHEGO_PLUGIN_DIR

  fi
  sleep 1
done
