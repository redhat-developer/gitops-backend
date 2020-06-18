FROM golang:latest AS build
WORKDIR /go/src
COPY . /go/src
RUN go build ./cmd/backend-http

FROM registry.access.redhat.com/ubi8/ubi-minimal
WORKDIR /root/
COPY --from=build /go/src/backend-http .
EXPOSE 8080
ENTRYPOINT ["./backend-http"]
