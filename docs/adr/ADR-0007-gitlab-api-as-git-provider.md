# ADR-0007 - GitLab API as Git provider integration

**Status:** Accepted  
**Date:** 2026-06-25  
**Project:** DevOps Control Plane  
**Owner iniziale:** Vincenzo Marzario

---

## Context

DevOps Control Plane deve creare e gestire change GitOps.

Le operazioni richieste includono:

- leggere file;
- creare branch;
- creare commit;
- aprire Merge Request;
- leggere commit e stato MR.

Esistono diverse possibilità:

- usare CLI `git` nel backend;
- fare clone locale dei repository;
- usare GitLab REST API;
- supportare subito più provider Git.

Per MVP è stato deciso che il target funzionale sarà GitLab API.

Il repository del codice DevOps Control Plane può essere ospitato su GitHub, ma i repository GitOps target del workflow MVP vengono trattati come GitLab.

---

## Decision

DevOps Control Plane userà GitLab API come integrazione Git primaria per l’MVP.

Il backend non dovrà dipendere dalla CLI `git` per il workflow principale.

Operazioni MVP:

```text
GetFile
ListCommits
CreateBranch
CommitFiles
CreateMergeRequest
GetMergeRequest
```

---

## Consequences

### Positive

- Nessun clone locale obbligatorio nel backend.
- Operazioni Git tracciabili via API.
- Branch, commit e MR possono essere creati in modo controllato.
- Migliore integrazione con ChangeRequest e evidence model.
- Più semplice gestire errori specifici GitLab.

### Negative / Trade-off

- Dipendenza da GitLab API.
- Token GitLab richiesto.
- API GitLab può avere rate limit o differenze tra versioni.
- Supporto multi-provider rimandato.
- Alcune operazioni Git avanzate potrebbero essere più semplici via CLI.

---

## Alternatives considered

### CLI git

Scartata come base prodotto perché richiede gestione workspace locale, credenziali git e cleanup.

### Clone repository nel backend

Scartata per MVP perché più complessa e meno necessaria per modifiche file mirate.

### Multi-provider subito

Rimandato per non aumentare la complessità iniziale.

### GitLab API

Scelta perché soddisfa le operazioni MVP e mantiene l’integrazione controllata.

---

## Related documents

- `docs/07-gitlab-integration.md`
- `docs/11-change-workflows.md`
- `docs/12-evidence-model.md`
- `docs/13-api-design.md`
