# k8s-oidc-rbac
POCs described here: https://app.excalidraw.com/s/172R1vSdAWD/15dsrtpnz9u

Purpose is to test out RBAC with a OIDC provider.

## Github OAuth App
### Setup 
First I created the [k8s-rbac-demo](https://github.com/k8s-rbac-demo) organization. Within that organization, I setup 2 teams: 
- `k8s-creator`: members (ahm.ibr+creator@hotmail.com)
- `k8s-lister`: members (ahm.ibr+lister@hotmail.com and ahm.ibr+creator@hotmail.com)

Next, I created an OAuth Application named [K8s RBAC Demo](https://github.com/settings/applications/2816314)

### Running Dex on Kind
We use dex to handle the OAuth handshake for us with our github OAuth application.
To setup a kind cluster, and install dex on it, run:
```
make install-dex
```

Next, to start the dex-ui, in a separate terminal run:
```
make start-dex-ui
```

This will start the dex-ui on [localhost:5555](http://localhost:5555).
Visit that URL to kick off the OAuth flow. Login as either of the created users to get their access tokens.

## K8s Backend Setup
Run the pod-service backend on our kind cluster with the following command. This will also apply all our roles and bindings that we need:
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
