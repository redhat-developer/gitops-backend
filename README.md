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
  gitops-backend-secret --from-literal=token=GENERATE_ME
```

Then you can deploy the resources with:

```shell
$ kubectl apply -k deploy
```
