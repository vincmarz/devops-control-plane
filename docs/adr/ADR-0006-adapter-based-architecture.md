# ADR-0006 - Adapter-based architecture

**Status:** Accepted  
**Date:** 2026-06-25  
**Project:** DevOps Control Plane  
**Owner iniziale:** Vincenzo Marzario

---

## Context

DevOps Control Plane deve integrarsi con diversi sistemi esterni:

- GitLab;
- Argo CD;
- Kubernetes/OpenShift;
- Tekton;
- PostgreSQL.

Se la logica di workflow chiamasse direttamente API HTTP o client Kubernetes, il codice diventerebbe accoppiato e difficile da testare.

Serve una separazione tra:

- dominio applicativo;
- workflow orchestration;
- integrazioni tecniche esterne.

---

## Decision

Il progetto userà una architettura basata su adapter.

Adapter iniziali:

```text
GitLab Adapter
Argo CD Adapter
Kubernetes Adapter
Tekton Adapter
PostgreSQL Repository
Evidence Service
```

La logica di dominio e workflow userà interfacce, non implementazioni concrete.

Struttura indicativa:

```text
internal/domain
internal/app
internal/workflow
internal/adapters/gitlab
internal/adapters/argocd
internal/adapters/kubernetes
internal/adapters/tekton
internal/database
```

---

## Consequences

### Positive

- Codice più modulare.
- Adapter testabili separatamente.
- Workflow testabile con fake adapter.
- Minore coupling con API esterne.
- Possibile sostituire o estendere provider in futuro.
- Migliore isolamento errori.

### Negative / Trade-off

- Più file e interfacce da mantenere.
- Overhead iniziale di design.
- Richiede disciplina nel non far trapelare dettagli adapter nel dominio.

---

## Alternatives considered

### Chiamate dirette negli handler HTTP

Scartata perché accoppia troppo API, workflow e integrazioni.

### Chiamate dirette nel workflow senza interfacce

Scartata perché complica test e manutenzione.

### Adapter-based architecture

Scelta perché adatta a un sistema di orchestrazione multi-integrazione.

---

## Related documents

- `docs/05-architecture.md`
- `docs/06-argocd-integration.md`
- `docs/07-gitlab-integration.md`
- `docs/08-tekton-integration.md`
- `docs/13-api-design.md`
