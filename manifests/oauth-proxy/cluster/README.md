# OAuth Proxy cluster-scoped resources

This directory contains cluster-scoped templates required by the OAuth Proxy runtime rollout.

They are intentionally not included in the namespaced Kustomize overlay.

Preferred command:

```bash
oc adm policy add-role-to-user \
  system:auth-delegator \
  system:serviceaccount:devops-control-plane:devops-control-plane-oauth-proxy
```

Alternatively, a ClusterRoleBinding template is provided for review.
Apply it only with appropriate cluster-admin approval.
