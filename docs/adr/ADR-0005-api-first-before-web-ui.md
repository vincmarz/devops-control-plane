# ADR-0005 - API-first before full Web UI

**Status:** Accepted  
**Date:** 2026-06-25  
**Project:** DevOps Control Plane  
**Owner iniziale:** Vincenzo Marzario

---

## Context

DevOps Control Plane dovrà avere una UI semplice basata su Go HTML templates e Bootstrap.

Tuttavia, il valore principale del progetto non è inizialmente la UI, ma la stabilizzazione dei workflow:

- discovery Application;
- creazione ChangeRequest;
- GitLab branch/commit/MR;
- Tekton validation;
- Argo CD sync;
- evidence collection;
- storico change.

Costruire subito una UI ricca rischierebbe di spostare l'attenzione su elementi visuali prima di aver stabilizzato modello dati, adapter e workflow.

---

## Decision

Il progetto seguirà un approccio API-first.

Prima saranno implementati:

- backend Go;
- API HTTP/REST;
- data model PostgreSQL;
- adapter GitLab, Argo CD, Tekton e Kubernetes;
- workflow engine;
- evidence model.

La UI HTML/Bootstrap arriverà dopo o in forma minimale, usando le API già definite.

---

## Consequences

### Positive

- Workflow validabili anche senza UI.
- Test più semplici tramite curl/Postman/script.
- Separazione più chiara tra logica applicativa e presentazione.
- Possibilità di aggiungere UI progressivamente.
- API riutilizzabili da automazioni future.

### Negative / Trade-off

- Esperienza utente iniziale meno ricca.
- I newbie inizialmente useranno API o comandi guidati.
- Serve documentazione API più accurata.

---

## Alternatives considered

### UI-first

Scartata perché rischia di anticipare decisioni non ancora validate sul workflow.

### CLI-first

Possibile in futuro, ma non scelta come base perché il progetto deve diventare un control plane web/API.

### API-first

Scelta perché consente stabilità tecnica e iterazione controllata.

---

## Related documents

- `docs/05-architecture.md`
- `docs/13-api-design.md`
- `docs/11-change-workflows.md`
- `docs/04-non-functional-requirements.md`
