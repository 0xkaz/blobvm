#!/usr/bin/env bash
set -e

# e.g.,
# ./scripts/run.sh 1.7.13
#
# run without e2e tests
# ./scripts/run.sh 1.7.13
#
# to run E2E tests (terminates cluster afterwards)
# E2E=true ./scripts/run.sh 1.7.13
if ! [[ "$0" =~ scripts/run2-pre.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

VERSION=$1
if [[ -z "${VERSION}" ]]; then
  echo "Missing version argument!"
  echo "Usage: ${0} [VERSION]" >> /dev/stderr
  exit 255
fi

MODE=${MODE:-run}
E2E=${E2E:-false}
if [[ ${E2E} == true ]]; then
  MODE="test"
fi

AVALANCHE_LOG_LEVEL=${AVALANCHE_LOG_LEVEL:-INFO}

echo "Running with:"
echo VERSION: ${VERSION}
echo MODE: ${MODE}

BLOBVM_PATH=$(
  cd "$(dirname "${BASH_SOURCE[0]}")"
  cd .. && pwd
)
source "$BLOBVM_PATH"/scripts/constants.sh

############################
# download avalanchego
# https://github.com/ava-labs/avalanchego/releases
GOARCH=$(go env GOARCH)
GOOS=$(go env GOOS)

if [[ -f /tmp/avalanchego-v${VERSION}/avalanchego ]]; then
  echo "already: /tmp/avalanchego-v${VERSION}/avalanchego"
else 
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
fi
find /tmp/avalanchego-v${VERSION}

AVALANCHEGO_PATH=/tmp/avalanchego-v${VERSION}/avalanchego
AVALANCHEGO_PLUGIN_DIR=/tmp/avalanchego-v${VERSION}/plugins

############################

############################
echo "building blobvm"
AVALANCHEGO_BIN=/tmp/avalanchego-v${VERSION}/avalanchego
if [ ! -f "$AVALANCHEGO_BIN" ]; then

  go build \
  -o /tmp/avalanchego-v${VERSION}/plugins/kM6h4LYe3AcEU1MB2UNg6ubzAiDAALZzpVrbX8zn3hXF6Avd8 \
  ./cmd/blobvm
  
fi 
find /tmp/avalanchego-v${VERSION}
if [ ! -f "/tmp/blob-cli"]; then 
  echo "building blob-cli"
  go build -v -o /tmp/blob-cli ./cmd/blob-cli
fi
find /tmp/blob-cli

############################

############################
echo "creating allocations file"
cat <<EOF > /tmp/allocations.json
[
  {"address":"0xF9370fa73846393798C2d23aa2a4aBA7489d9810", "balance":10000000},
  {"address":"0x8Db3219F3f59b504BCF132EfB4B87Bf08c771d83", "balance":10000000},
  {"address":"0x162a5fadfdd769f9a665701348FbeEd12A4FFce7", "balance":10000000},
  {"address":"0x69fd199Aca8250d520F825d22F4ad9db4A58E9D9", "balance":10000000},
  {"address":"0x454474642C32b19E370d9A55c20431d85833cDD6", "balance":10000000},
  {"address":"0xeB4Fc761FAb7501abe8cD04b2d831a45E8913DdF", "balance":10000000},
  {"address":"0xD23cbfA7eA985213aD81223309f588A7E66A246A", "balance":10000000}
]
EOF

echo "creating VM genesis file"
rm -f /tmp/blobvm.genesis
/tmp/blob-cli genesis 1 /tmp/allocations.json \
--genesis-file /tmp/blobvm.genesis \
--airdrop-hash 0xccbf8e430b30d08b5b3342208781c40b373d1b5885c1903828f367230a2568da \
--airdrop-units 10000
############################

############################
echo "building e2e.test"
# to install the ginkgo binary (required for test build and run)
go install -v github.com/onsi/ginkgo/v2/ginkgo@v2.1.4
ACK_GINKGO_RC=true ginkgo build ./tests/e2e
./tests/e2e/e2e.test --help

# #################################
# # download avalanche-network-runner
# # https://github.com/ava-labs/avalanche-network-runner
# # TODO: use "go install -v github.com/ava-labs/avalanche-network-runner/cmd/avalanche-network-runner@v${NETWORK_RUNNER_VERSION}"
# NETWORK_RUNNER_VERSION=1.1.4
# DOWNLOAD_PATH=/tmp/avalanche-network-runner.tar.gz
# DOWNLOAD_URL=https://github.com/ava-labs/avalanche-network-runner/releases/download/v${NETWORK_RUNNER_VERSION}/avalanche-network-runner_${NETWORK_RUNNER_VERSION}_linux_amd64.tar.gz
# if [[ ${GOOS} == "darwin" ]]; then
#   DOWNLOAD_URL=https://github.com/ava-labs/avalanche-network-runner/releases/download/v${NETWORK_RUNNER_VERSION}/avalanche-network-runner_${NETWORK_RUNNER_VERSION}_darwin_amd64.tar.gz
# fi

# rm -f ${DOWNLOAD_PATH}
# rm -f /tmp/avalanche-network-runner

# echo "downloading avalanche-network-runner ${NETWORK_RUNNER_VERSION} at ${DOWNLOAD_URL}"
# curl -L ${DOWNLOAD_URL} -o ${DOWNLOAD_PATH}

# echo "extracting downloaded avalanche-network-runner"
# tar xzvf ${DOWNLOAD_PATH} -C /tmp
# /tmp/avalanche-network-runner -h

#################################
# download avalanche-network-runner
# https://github.com/ava-labs/avalanche-network-runner
ANR_REPO_PATH=github.com/ava-labs/avalanche-network-runner
ANR_VERSION=v1.3.5
# version set
go install -v ${ANR_REPO_PATH}@${ANR_VERSION}

#################################
# run "avalanche-network-runner" server
GOPATH=$(go env GOPATH)
if [[ -z ${GOBIN+x} ]]; then
  # no gobin set
  BIN=${GOPATH}/bin/avalanche-network-runner
else
  # gobin set
  BIN=${GOBIN}/avalanche-network-runner
fi
cp $BIN /tmp/avalanche-network-runner
/tmp/avalanche-network-runner -h

# # killall avalanche-network-runner || true
# killall avalanche-network-runner || true


# ############################
# # run "avalanche-network-runner" server
# echo "launch avalanche-network-runner in the background"
# /tmp/avalanche-network-runner \
# server \
# --log-level debug \
# --port=":12342" \
# --disable-grpc-gateway &
# PID=${!}

# ############################
# # By default, it runs all e2e test cases!
# # Use "--ginkgo.skip" to skip tests.
# # Use "--ginkgo.focus" to select tests.
# echo "running e2e tests"
# ./tests/e2e/e2e.test \
# --ginkgo.v \
# --network-runner-log-level debug \
# --network-runner-grpc-endpoint="0.0.0.0:12342" \
# --avalanchego-path=${AVALANCHEGO_PATH} \
# --avalanchego-plugin-dir=${AVALANCHEGO_PLUGIN_DIR} \
# --vm-genesis-path=/tmp/blobvm.genesis \
# --output-path=/tmp/avalanchego-v${VERSION}/output.yaml \
# --mode=${MODE}

# ############################
# # e.g., print out MetaMask endpoints
# if [[ -f "/tmp/avalanchego-v${VERSION}/output.yaml" ]]; then
#   echo "cluster is ready!"
#   cat /tmp/avalanchego-v${VERSION}/output.yaml
# else
#   echo "cluster is not ready in time... terminating ${PID}"
#   kill ${PID}
#   exit 255
# fi

# ############################
# if [[ ${MODE} == "test" ]]; then
#   # "e2e.test" already terminates the cluster for "test" mode
#   # just in case tests are aborted, manually terminate them again
#   echo "network-runner RPC server was running on PID ${PID} as test mode; terminating the process..."
#   pkill -P ${PID} || true
#   kill -2 ${PID}
#   pkill -9 -f kM6h4LYe3AcEU1MB2UNg6ubzAiDAALZzpVrbX8zn3hXF6Avd8 || true # in case pkill didn't work
# else
#   echo "network-runner RPC server is running on PID ${PID}..."
#   echo ""
#   echo "use the following command to terminate:"
#   echo ""
#   echo "pkill -P ${PID} && kill -2 ${PID} && pkill -9 -f kM6h4LYe3AcEU1MB2UNg6ubzAiDAALZzpVrbX8zn3hXF6Avd8"
#   echo ""
# fi
