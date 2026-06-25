# ADR-0001 - Git as source of truth

**Status:** Accepted  
**Date:** 2026-06-25  
**Project:** DevOps Control Plane  
**Owner iniziale:** Vincenzo Marzario

---

## Context

DevOps Control Plane nasce per agevolare i workflow GitOps, non per sostituire GitOps.

Il sistema deve permettere agli operatori DevOps e Platform Engineer di gestire change applicativi in modo guidato, tracciabile e auditabile.

Nel contesto Kubernetes/OpenShift esistono due possibilità operative:

1. modificare direttamente il runtime cluster;
2. modificare lo stato desiderato in Git e lasciare che Argo CD riconcili il cluster.

Esempi di modifica runtime diretta:

```bash
oc edit deployment
oc patch deployment
oc set env deployment
oc scale deployment
```

Questi comandi sono utili in troubleshooting o emergenza, ma se usati come modalità standard creano problemi di governance:

- il cluster può divergere da Git;
- la modifica non è facilmente revisionabile;
- il rollback è meno chiaro;
- lo storico operativo è frammentato;
- la responsabilità del change è meno evidente;
- Argo CD può riportare il runtime allo stato Git precedente.

Per questo progetto, il principio GitOps deve rimanere centrale.

---

## Decision

Tutti i change applicativi permanenti gestiti dal DevOps Control Plane devono essere rappresentati in Git.

Il flusso standard è:

```text
ChangeRequest
  -> GitLab branch
  -> manifest update
  -> commit / merge request
  -> Tekton validation
  -> Argo CD sync
  -> OpenShift runtime validation
  -> evidence collection
```

DevOps Control Plane non deve modificare direttamente Deployment, ConfigMap, Service o Route come stato finale permanente.

Le modifiche runtime dirette sono considerate fuori dal workflow applicativo normale e rimangono solo attività manuali di troubleshooting/emergenza.

---

## Consequences

### Positive

- Coerenza con GitOps.
- Auditabilità più forte.
- Review tramite branch, commit e Merge Request.
- Rollback più chiaro tramite revert Git.
- Maggiore integrazione con Argo CD.
- Possibilità di correlare change, commit, validation, sync ed evidenze.
- Maggiore valore formativo per utenti newbie.

### Negative / Trade-off

- Workflow più lungo rispetto a un patch runtime diretto.
- Richiede disponibilità GitLab.
- Richiede gestione branch, commit e MR.
- Richiede gestione dei conflitti Git.
- Richiede validazione manifest prima della sync.

---

## Alternatives considered

### Direct runtime patching

Scartata come flusso standard perché rompe il modello GitOps e riduce l'auditabilità.

### Mixed mode Git + runtime patch

Scartata per MVP perché rischia ambiguità tra stato desiderato e stato reale.

### Git as source of truth

Scelta perché coerente con Argo CD, GitOps e obiettivi di governance.

---

## Related documents

- `docs/00-vision.md`
- `docs/03-functional-requirements.md`
- `docs/05-architecture.md`
- `docs/11-change-workflows.md`
- `docs/12-evidence-model.md`
