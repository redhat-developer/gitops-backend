FROM openshift/origin-release:golang-1.18 AS build
WORKDIR /go/src
COPY . /go/src
RUN GIT_COMMIT=$(git rev-parse HEAD) && \
  CGO_ENABLED=0 GOOS=linux go build -a -mod=readonly \
  -ldflags "-X github.com/redhat-developer/gitops-backend/pkg/health.GitRevision=${GIT_COMMIT}" ./cmd/backend-http

FROM registry.access.redhat.com/ubi8/ubi-minimal
WORKDIR /root/
COPY --from=build /go/src/backend-http .
EXPOSE 8080
ENTRYPOINT ["./backend-http"]
