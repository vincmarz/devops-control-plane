# ADR-0008 - Kubernetes API for Tekton integration

**Status:** Accepted  
**Date:** 2026-06-25  
**Project:** DevOps Control Plane  
**Owner iniziale:** Vincenzo Marzario

---

## Context

Tekton è installato su Kubernetes/OpenShift come insieme di Custom Resource Definition.

PipelineRun e TaskRun sono risorse Kubernetes-native.

DevOps Control Plane deve:

- creare PipelineRun;
- leggere PipelineRun;
- monitorare status;
- listare TaskRun correlate;
- raccogliere log dai Pod associati;
- salvare evidenze.

Possibili approcci:

- eseguire CLI `tkn` dal backend;
- eseguire `oc`/`kubectl` dal backend;
- usare Kubernetes API diretta;
- usare eventuali librerie Tekton specifiche.

---

## Decision

DevOps Control Plane integrerà Tekton tramite Kubernetes API diretta.

Il backend Go userà client Kubernetes, typed o dynamic client, per gestire risorse Tekton e risorse runtime necessarie.

La CLI `tkn` e la CLI `oc` restano strumenti di troubleshooting manuale, non dipendenze runtime del prodotto.

---

## Consequences

### Positive

- Integrazione Kubernetes-native.
- Nessuna dipendenza da CLI nel container backend.
- Migliore controllo errori e timeout.
- Possibilità di watch/polling programmatico.
- Migliore testabilità con fake client.
- Coerenza con RBAC Kubernetes.

### Negative / Trade-off

- Richiede conoscenza delle CRD Tekton.
- Richiede gestione GVR o client typed.
- Richiede RBAC esplicito.
- Raccolta log richiede mapping TaskRun -> Pod.
- Più codice rispetto a invocare `tkn`.

---

## Alternatives considered

### CLI tkn

Scartata come dipendenza runtime perché meno controllabile e meno testabile.

### CLI oc/kubectl

Scartata per lo stesso motivo e per rischio di parsing fragile.

### Kubernetes API diretta

Scelta perché nativa, robusta e coerente con OpenShift.

---

## Related documents

- `docs/08-tekton-integration.md`
- `docs/09-security-rbac.md`
- `docs/05-architecture.md`
- `docs/13-api-design.md`
