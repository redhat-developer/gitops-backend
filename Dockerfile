FROM golang:1.23 AS build
WORKDIR /go/src
COPY . /go/src
RUN GIT_COMMIT=$(git rev-parse HEAD) && \
  GOEXPERIMENT=strictfipsruntime CGO_ENABLED=1 GOOS=linux go build -a -mod=mod -o /tmp/backend-http \
  -ldflags "-X github.com/redhat-developer/gitops-backend/pkg/health.GitRevision=${GIT_COMMIT}" -tags strictfipsruntime ./cmd/backend-http

FROM registry.access.redhat.com/ubi8/ubi-minimal
WORKDIR /
COPY --from=build /go/src/backend-http .
EXPOSE 8080
ENTRYPOINT ["./backend-http"]
