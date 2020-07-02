FROM golang:latest AS build
WORKDIR /go/src
COPY . /go/src
RUN GIT_COMMIT=$(git rev-parse HEAD) && \
  CGO_ENABLED=0 GOOS=linux go build -a \
  -ldflags "-X github.com/rhd-gitops-examples/gitops-backend/pkg/health.GitRevision=${GIT_COMMIT}" ./cmd/backend-http

FROM registry.access.redhat.com/ubi8/ubi-minimal
WORKDIR /root/
COPY --from=build /go/src/backend-http .
EXPOSE 8080
ENTRYPOINT ["./backend-http"]
