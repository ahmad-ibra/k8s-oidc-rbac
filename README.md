# k8s-rbac
POCs described here: https://app.excalidraw.com/s/172R1vSdAWD/15dsrtpnz9u

## Okta Setup

- following this guide: https://developer.okta.com/blog/2021/11/08/k8s-api-server-oidc#what-youll-need-to-get-started

- created this okta account: https://dev-04886319-admin.okta.com/admin/getting-started

- next followed this guide: https://developer.okta.com/blog/2021/10/08/secure-access-to-aws-eks#configure-your-okta-org
    - created k8s-creator-group and k8s-user-group
    - created user with email ahmad.ibrahim+creator@spectrocloud.com that is in both reader and creater groups
    - created user with email ahmad.ibrahim+reader@spectrocloud.com that is in the reader group
    - both users passwords are welcome2Spectr0!
    - created an app integration named k8s
        - client ID: 0oalgq0kxjvutVEaR5d7
    - using default authorization server
        - audience: api://default
        - issuer: https://dev-04886319.okta.com/oauth2/default

## Cluster Setup

- referenced this repo: https://github.com/int128/kind-oidc

Create the kind cluster by running:
```
kind create cluster --config cluster-okta.yaml
```

Create roles:
```

```
