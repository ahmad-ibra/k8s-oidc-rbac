# k8s-oidc-rbac
POCs described here: https://app.excalidraw.com/s/172R1vSdAWD/15dsrtpnz9u

Purpose is to test out RBAC with a OIDC provider.

The `/oath-server` directory contains a go server that handles the OAUTH hanshake with Okta in lue of a frontend

The `/pod-service` directory contains a go server that will run in a kind cluster and is used to create, get, and list pods in a cluster.
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

## K8s Backend Setup

First we need to setup our kind cluster with OIDC enabled
```
cd pod-service
kind create cluster --config kind-okta-oidc.yaml
```

Once we have a kind cluster up and running, we can run the pod-service backend on it with the following command. This will also apply all our roles and bindings that we need:
```
make kind-deploy
```

Assuming it started up correctly, in another terminal, we can then issue requests as a cluster admin like this:
```
curl -X GET \
http://localhost:8000/api/v1/pods
```

To impersonate a user, generate the access token by visiting `localhost:8080`, then issue the following request:
```
curl -X GET \
-H "Authorization: Bearer <Access Token>" \
http://localhost:8000/api/v1/pods
```

If we need to remove the pod from the cluster, just run `make undeploy`
