#!/bin/sh

docker build -t okto-proxy .
docker run -p 8080:8080 -p 8000:8000 --name okto-proxy-app okto-proxy &
sleep 1
curl -x http://127.0.0.1:8080 http://mail.ru >/dev/null
docker rm --force okto-proxy-app
docker image prune --force --filter label=stage=okto-builder
docker run -p 8080:8080 -p 8000:8000 --name okto-proxy-app okto-proxy
