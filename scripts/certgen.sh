#!/bin/sh

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
CERTS_DIR=$DIR/../certs/
mkdir $CERTS_DIR
openssl genrsa -out $CERTS_DIR/ca.key 2048
openssl req -new -x509 -days 3650 -key $CERTS_DIR/ca.key -out $CERTS_DIR/ca.crt -subj "/CN=golang proxy CA"
