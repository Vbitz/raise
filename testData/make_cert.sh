#!/bin/bash

set -ex

rm -f server.csr server.key server.crt ca.key ca.crt

openssl genrsa -out server.key 2048
openssl genrsa -out ca.key 2048

openssl req -new -x509 -days 365 -key ca.key -subj "/CN=Raise Root CA" -out ca.crt

openssl req -newkey rsa:2048 -nodes -keyout server.key -subj "/CN=*.localhost" -out server.csr
openssl x509 -req -extfile <(printf "subjectAltName=DNS:localhost") -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt