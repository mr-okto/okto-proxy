#!/bin/sh

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
CERTS_DIR=$DIR/../certs/
sudo cp $CERTS_DIR/ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates