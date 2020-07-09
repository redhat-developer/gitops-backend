# gitops-backend

This is a PoC for getting files from a remote Git API.

## Installation

First of all, generate a GitHub access token.

You will need to create a secret in the correct namespace:

```shell
$ kubectl create -f deploy/namespace.yaml
```

```shell
$ kubectl create secret -n pipelines-app-delivery \
  pipelines-app-gitops --from-literal=token=GENERATE_ME
```

Then you can deploy the resources with:

```shell
$ kubectl apply -k deploy
```

## Usage

Deploy the `Deployment` and `Service` into your cluster.

You can fetch with `https://...?url=https://github.com/org/repo.git?secretNS=my-ns&secretName=my-secret`

You'll need to provide your OpenShift authentication token (`oc whoami --show-token`) as a Bearer token.

```
$ curl -H "Authorization: $(oc whoami --show-token)" https://...?url=https://github.com/org/repo.git?secretNS=my-ns&secretName=my-secret
```

This token is used to authenticate the Kube client request to load the secret by name/namespace to authenticate the call to the upstream Git provider.

The `token` field in the named secret will be extracted and used to authenticate
the request to the upstream Git hosting service.
