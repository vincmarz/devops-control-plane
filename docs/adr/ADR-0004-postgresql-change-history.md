# ADR-0004 - PostgreSQL as change history and evidence store

**Status:** Accepted  
**Date:** 2026-06-25  
**Project:** DevOps Control Plane  
**Owner iniziale:** Vincenzo Marzario

---

## Context

DevOps Control Plane deve conservare lo storico funzionale dei change.

GitLab conserva commit e merge request.  
Argo CD conserva history di sync.  
Tekton conserva stato PipelineRun/TaskRun finché le risorse sono presenti o tramite sistemi come Tekton Results.  
Kubernetes conserva eventi runtime con retention limitata.

Tuttavia, nessuno di questi sistemi da solo conserva una vista unica e funzionale del workflow end-to-end:

```text
ChangeRequest
  -> GitLab branch/commit/MR
  -> Tekton validation
  -> Argo CD sync
  -> Runtime evidence
```

Serve quindi un database interno per correlare tutte queste informazioni.

---

## Decision

DevOps Control Plane userà PostgreSQL come database interno per:

- Application snapshot/cache;
- ChangeRequest;
- ChangeEvent;
- Evidence;
- riferimenti GitLab;
- riferimenti Tekton;
- riferimenti Argo CD;
- stato runtime sintetico;
- audit trail operativo.

PostgreSQL non conserverà token o secret.

Il modello iniziale userà tabelle relazionali e colonne `jsonb` per payload estensibili.

---

## Consequences

### Positive

- Storico change persistente.
- Query semplici per audit e troubleshooting.
- Correlazione tra strumenti differenti.
- Supporto naturale a JSONB per evidenze flessibili.
- Tecnologia matura e nota.
- Buona base per dashboard e UI futura.

### Negative / Trade-off

- Introduce una dipendenza infrastrutturale.
- Richiede backup e restore.
- Richiede migration database.
- Richiede sanitizzazione rigorosa dei payload.
- Richiede gestione connessioni e readiness.

---

## Alternatives considered

### Solo Git history

Insufficiente perché Git non contiene stato Tekton, Argo CD e runtime.

### Solo Argo CD history

Insufficiente perché Argo CD non contiene motivazione completa del change, requestedBy, validazione Tekton e diff summary applicativo.

### File JSON locali

Scartati perché poco scalabili e meno adatti a query/audit.

### PostgreSQL

Scelto come change/evidence store per robustezza e flessibilità.

---

## Related documents

- `docs/10-data-model.md`
- `docs/12-evidence-model.md`
- `docs/13-api-design.md`
- `docs/04-non-functional-requirements.md`
