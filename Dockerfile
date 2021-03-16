FROM golang:latest AS build_step
LABEL stage=okto-builder
ENV GO111MODULE=on
WORKDIR  /go/src
COPY . .
RUN go get -d -v ./...
EXPOSE 8080
EXPOSE 8000
RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o /go/build/proxy /go/src/cmd/proxy/main.go
RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w" -o /go/build/repeater /go/src/cmd/repeater/main.go

FROM alpine
WORKDIR /app
COPY --from=build_step /go/build/proxy /app/proxy
COPY --from=build_step /go/build/repeater /app/repeater
COPY ./certs /app/certs
RUN chmod +x /app/proxy
RUN chmod +x /app/repeater
EXPOSE 8080
EXPOSE 8000
CMD /app/proxy & /app/repeater
