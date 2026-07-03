# Architecture Decision Records

Questo folder contiene gli ADR del progetto DevOps Control Plane.

## ADR index

| ADR | Titolo | Stato |
|---|---|---|
| ADR-0001-git-source-of-truth.md | Git as source of truth | Accepted |
| ADR-0002-argocd-as-gitops-engine.md | Argo CD as GitOps engine | Accepted |
| ADR-0003-tekton-validation-engine.md | Tekton as validation engine | Accepted |
| ADR-0004-postgresql-change-history.md | PostgreSQL as change history and evidence store | Accepted |
| ADR-0005-api-first-before-web-ui.md | API-first before full Web UI | Accepted |
| ADR-0006-adapter-based-architecture.md | Adapter-based architecture | Accepted |
| ADR-0007-gitlab-api-as-git-provider.md | GitLab API as Git provider integration | Accepted |
| ADR-0008-kubernetes-api-for-tekton.md | Kubernetes API for Tekton integration | Accepted |
| ADR-0009-authn-authz-strategy.md | AuthN/AuthZ strategy | Accepted |
| ADR-0010-oauth-proxy-deployment-design.md | OAuth Proxy deployment design | Accepted |
| [ADR-0011](ADR-0011-multi-environment-model.md) | Multi-environment DevOps Control Plane model | Accepted |

## Regola

Ogni decisione architetturale significativa deve essere documentata con un ADR versionato in Git.

## Note operative

- Gli ADR devono essere numerati in modo progressivo.
- Il nome file deve includere il numero ADR e uno slug descrittivo.
- Il README deve essere aggiornato ogni volta che viene aggiunto, rinominato o rimosso un ADR.
- Gli ADR accettati non devono essere modificati retroattivamente per cambiarne il significato; eventuali nuove decisioni devono essere documentate con un nuovo ADR o con una revisione esplicita.
