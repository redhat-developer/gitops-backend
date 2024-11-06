FROM golang:1.22 AS build
WORKDIR /go/src
COPY . /go/src
RUN GIT_COMMIT=$(git rev-parse HEAD) && \
  CGO_ENABLED=0 GOOS=linux go build -a -mod=readonly \
  -ldflags "-X github.com/redhat-developer/gitops-backend/pkg/health.GitRevision=${GIT_COMMIT}" ./cmd/backend-http

FROM scratch
WORKDIR /
COPY --from=build /go/src/backend-http .
EXPOSE 8080
ENTRYPOINT ["./backend-http"]
