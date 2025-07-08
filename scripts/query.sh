#!/bin/bash

set -o pipefail
set -eu

A=${A?must be set in the environment to one of rv or ta}

base64url_encode() {
  if [ "$(uname)" == "Darwin" ]
  then
    _base64="base64"
  else
    _base64="base64 -w0"
  fi

  ${_base64} | tr '+/' '-_' | tr -d '=';
}

# ref-value query
function rv_query() {
cat << EOF | base64url_encode
{
  "0": "tag:arm.com,2023:cca_platform#1.0.0",
  "1": {
    "0": 2,
    "1": {
      "0": [
        {
          "0": "7f454c4602010100000000000000000003003e00010000005058000000000000"
        }
      ]
    }
  }
}
EOF
}

# ta query
function ta_query() {
cat << EOF | base64url_encode
{
  "0": "tag:arm.com,2023:cca_platform#1.0.0",
  "1": {
    "0": 1,
    "1": {
      "1": [
        {
          "1": "0107060504030201000f0e0d0c0b0a090817161514131211101f1e1d1c1b1a1918"
        }
      ]
    }
  }
}
EOF
}

if [ "${A}" == "rv" ]; then
  q=$(rv_query)
elif [ "${A}" == "ta" ]; then
  q=$(ta_query)
fi

curl http://localhost:8080/endorsement-distribution/v1/coserv/$q -s \
  --header 'Accept: application/coserv+cbor; profile="tag:arm.com,2023:cca_platform#1.0.0"' \
  | jq '.' 