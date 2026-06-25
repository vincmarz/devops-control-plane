# ADR-0002 - Argo CD as GitOps engine

**Status:** Accepted  
**Date:** 2026-06-25  
**Project:** DevOps Control Plane  
**Owner iniziale:** Vincenzo Marzario

---

## Context

DevOps Control Plane deve orchestrare change GitOps su OpenShift/Kubernetes.

Nel lab e nel modello architetturale esiste già Argo CD / OpenShift GitOps come componente responsabile della riconciliazione tra repository Git e stato runtime del cluster.

DevOps Control Plane potrebbe, teoricamente:

- applicare manifest direttamente con Kubernetes API;
- eseguire `oc apply` direttamente;
- delegare la riconciliazione ad Argo CD;
- implementare un proprio motore GitOps.

Implementare un motore GitOps proprietario sarebbe complesso e non necessario.

Argo CD offre già:

- Application model;
- sync status;
- health status;
- resource tracking;
- history;
- operation state;
- AppProject governance;
- gestione warning come `OrphanedResourceWarning`;
- riconciliazione dichiarativa.

---

## Decision

DevOps Control Plane userà Argo CD come motore GitOps ufficiale.

DevOps Control Plane non applicherà direttamente i manifest applicativi al cluster come modalità standard.

Il ruolo del DevOps Control Plane sarà:

- leggere Application Argo CD;
- mostrare sync/health/resources/history;
- lanciare sync quando il workflow lo richiede;
- attendere `Synced` e `Healthy`;
- interpretare errori comuni;
- salvare evidenze Argo CD;
- correlare Argo CD con GitLab, Tekton e runtime OpenShift.

---

## Consequences

### Positive

- Coerenza con GitOps.
- Evita duplicazione di funzionalità già presenti in Argo CD.
- Mantiene Argo CD come fonte autorevole dello stato GitOps.
- Permette di usare AppProject come governance boundary.
- Facilita troubleshooting tramite sync status, health status e history.
- Permette integrazione naturale con rollback e drift detection.

### Negative / Trade-off

- DevOps Control Plane dipende da Argo CD API.
- Errori o indisponibilità Argo CD bloccano sync controllate.
- Occorre gestire operation già in corso.
- Occorre interpretare differenze tra Git history e Argo CD history.
- Alcuni rollback Argo CD non modificano Git e possono produrre `OutOfSync` rispetto a `main`.

---

## Alternatives considered

### Direct Kubernetes apply

Scartata perché bypassa Argo CD e riduce coerenza GitOps.

### Implementare motore GitOps interno

Scartata per complessità e duplicazione funzionale.

### Argo CD as GitOps engine

Scelta perché è già il componente specializzato per riconciliazione GitOps.

---

## Related documents

- `docs/05-architecture.md`
- `docs/06-argocd-integration.md`
- `docs/11-change-workflows.md`
- `docs/12-evidence-model.md`
- `docs/13-api-design.md`
