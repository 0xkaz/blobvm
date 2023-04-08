#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

if ! [[ "$0" =~ scripts/watch.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi


LAST=""
TMP=""
while [ true ]
do
  TMP=`find . | grep '.go' | xargs stat -f %m  | sort -n | tail -n 1`
  
  if [ "$TMP" != "$LAST" ]; then
    LAST=$TMP
    echo "${LAST}"

    name="kM6h4LYe3AcEU1MB2UNg6ubzAiDAALZzpVrbX8zn3hXF6Avd8"
    mkdir -p ./build
    echo "Building blobvm in ./build/$name"
    go build -o ./build/$name ./cmd/blobvm

    echo "Building blob-cli in ./build/blob-cli"
    go build -o ./build/blob-cli ./cmd/blob-cli

  fi
  sleep 1
done
