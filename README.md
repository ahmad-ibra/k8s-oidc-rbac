# k8s-rbac
POCs described here: https://app.excalidraw.com/s/172R1vSdAWD/15dsrtpnz9u

Purpose is to test out RBAC with a OIDC provider.

The `/kind` directory shows how to setup the cluster.

The `/oath-server` directory contains a go server that handles the OAUTH hanshake with Okta in lue of a frontend

The `/k8s-backend` directory contains a go server that will run in a cluster and is used to create, get, and list pods in a cluster.
This server will impersonate users that invoke its endpoints. K8s native RBAC will ensure only authorized users can create, get, and list pods.

## Okta Setup

Following this guide, i ran the steps below: https://developer.okta.com/blog/2021/11/08/k8s-api-server-oidc#what-youll-need-to-get-started

- created this okta account: https://dev-04886319-admin.okta.com/admin/getting-started

- next followed this guide: https://developer.okta.com/blog/2021/10/08/secure-access-to-aws-eks#configure-your-okta-org
    - created k8s-creator-group and k8s-user-group
    - created user with email ahmad.ibrahim+creator@spectrocloud.com that is in both reader and creater groups
    - created user with email ahmad.ibrahim+reader@spectrocloud.com that is in the reader group
    - created user with email ahmad.ibrahim@spectrocloud.com that is in no group
    - both users passwords are welcome2Spectr0!
    - created an app integration named k8s
        - client ID: 0oalgq0kxjvutVEaR5d7
    - using default authorization server
        - audience: api://default
        - issuer: https://dev-04886319.okta.com/oauth2/default

Run the oath-server:
```
cd oath-server
go run oath-server/main.go
```

Visit `localhost:8080` to kick off the OAuth flow.

## Cluster Setup

Follow the steps below to configure a kind cluster where the api-server uses OIDC.
Also, this guide will setup the roles and role-bidings that'll help us test out RBAC.


Create the kind cluster by running:
```
kind create cluster --config kind/cluster-okta.yaml
```

Apply the roles and role-bindings:
```
kubectl apply -f kind/roles.yaml

```
