# DevOps Control Plane - GitLab Integration

**Versione:** 0.1  
**Data:** 2026-06-25  
**Owner iniziale:** Vincenzo Marzario  
**Repository:** `https://github.com/vincmarz/devops-control-plane`  
**Documenti precedenti:**  
- `docs/00-vision.md`  
- `docs/01-scope-mvp.md`  
- `docs/02-personas-use-cases.md`  
- `docs/03-functional-requirements.md`  
- `docs/04-non-functional-requirements.md`  
- `docs/05-architecture.md`  
- `docs/06-argocd-integration.md`  
**Stato:** Draft iniziale / GitLab Integration

---

## 1. Scopo del documento

Questo documento descrive come **DevOps Control Plane** deve integrarsi con **GitLab**.

L’obiettivo è definire:

- responsabilità dell’integrazione GitLab;
- funzionalità MVP da implementare;
- modello di autenticazione;
- modello dati normalizzato;
- operazioni API richieste;
- gestione repository, branch, file, commit e merge request;
- convenzioni di naming;
- gestione errori;
- requisiti di sicurezza;
- evidenze GitLab da salvare;
- mapping verso ChangeRequest, Argo CD e Tekton.

Nel modello DevOps Control Plane, GitLab rappresenta il sistema che ospita i repository GitOps target e diventa il punto di origine dei change dichiarativi applicati da Argo CD.

---

## 2. Ruolo di GitLab nell’architettura

GitLab ha il ruolo di **Git provider e change collaboration platform**.

Responsabilità GitLab:

- ospitare repository GitOps;
- esporre branch e commit;
- permettere lettura e modifica file tramite API;
- gestire merge request;
- mantenere history Git;
- fornire link auditabili a commit, branch e MR.

Responsabilità DevOps Control Plane:

- leggere file GitOps;
- creare branch per ChangeRequest;
- modificare file YAML in modo controllato;
- creare commit;
- aprire merge request quando previsto;
- leggere stato MR;
- leggere ultimo commit;
- correlare commit/MR con ChangeRequest, Tekton validation e Argo CD sync;
- impedire che token o secret finiscano in Git.

---

## 3. Principi di integrazione

## 3.1 GitLab è la sorgente del change dichiarativo

DevOps Control Plane deve produrre modifiche Git, non modifiche runtime dirette.

Flusso corretto:

```text
ChangeRequest
  -> GitLab branch
  -> file update
  -> commit / merge request
  -> Tekton validation
  -> merge su branch target
  -> Argo CD sync
```

---

## 3.2 DevOps Control Plane usa GitLab API

L’integrazione target deve usare **GitLab REST API**.

Il backend Go non deve dipendere dalla CLI `git` per il workflow principale.

La CLI `git` può essere usata per sviluppo, troubleshooting e test manuali, ma il prodotto deve usare l’adapter GitLab API.

---

## 3.3 Ogni change deve essere tracciabile

Ogni ChangeRequest deve salvare almeno:

- project ID GitLab;
- repository URL;
- source branch;
- target branch;
- commit SHA;
- merge request IID/URL, se presente;
- autore logico/richiedente;
- messaggio commit;
- diff summary;
- stato MR;
- timestamp.

---

## 3.4 Commit diretto o Merge Request

MVP può supportare due modalità:

```text
Mode A - Commit diretto su branch target
Mode B - Branch + Merge Request
```

Modalità raccomandata:

```text
Branch + Merge Request
```

Motivo:

- review più chiara;
- audit migliore;
- validazione Tekton prima del merge;
- rollback più controllato;
- separazione tra proposta e applicazione finale.

Nel lab iniziale può essere utile supportare anche commit diretto per velocizzare test controllati.

---

## 4. Funzionalità MVP GitLab

## 4.1 Configurazione repository GitLab

### Obiettivo

Associare una Application Argo CD a un repository GitLab e a un path GitOps.

### Dati minimi

```yaml
applicationName: demo-go-color-app
repoProvider: gitlab
gitlabBaseUrl: https://gitlab.example.local
gitlabProjectId: "12345"
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
defaultBranch: main
path: apps/demo-go-color-app
```

### Regole

- `gitlabProjectId` può essere ID numerico o path URL-encoded.
- `defaultBranch` deve essere configurabile.
- `path` deve puntare al path GitOps Argo CD.

---

## 4.2 Lettura file repository

### Obiettivo

Leggere file GitOps dal repository.

### File principali MVP

```text
apps/demo-go-color-app/deployment.yaml
apps/demo-go-color-app/configmap.yaml
apps/demo-go-color-app/kustomization.yaml
apps/demo-go-color-app/service.yaml
apps/demo-go-color-app/route.yaml
```

### Uso

- individuare valore `replicas`;
- individuare `APP_VERSION`;
- individuare `PAGE_COLOR`;
- verificare ConfigMap;
- verificare `kustomization.yaml`;
- generare diff;
- produrre evidenza.

### Regole

- Il contenuto file letto via API può essere base64 encoded e deve essere decodificato.
- File mancanti devono produrre errore leggibile.
- Path file deve essere URL-encoded quando richiesto dall’API.

---

## 4.3 Creazione branch

### Obiettivo

Creare un branch per la ChangeRequest.

### Naming convention

```text
change/<change-id>-<change-type>
```

Esempio:

```text
change/CHG-2026-0001-update-replicas
```

### Regole

- Il branch nasce da `defaultBranch` o da ref configurata.
- Il nome deve essere normalizzato.
- Il sistema deve gestire branch già esistente.
- Il branch deve essere salvato nella ChangeRequest.

---

## 4.4 Commit file modificati

### Obiettivo

Creare commit su branch GitLab con modifiche GitOps prodotte dal workflow.

### Commit message convention

```text
<change-id> <change-type> <application-name>
```

Esempio:

```text
CHG-2026-0001 update-replicas demo-go-color-app
```

### Descrizione commit estesa

Il body del commit può includere:

```text
Application: demo-go-color-app
Change-Type: update-replicas
Requested-By: vmarzario
Change-Id: CHG-2026-0001
```

### Regole

- Il commit deve includere Change ID.
- Secret scanning deve avvenire prima del commit o nella Tekton validation.
- Il commit SHA risultante deve essere salvato in PostgreSQL.

---

## 4.5 Apertura Merge Request

### Obiettivo

Aprire MR per review e merge del change.

### Titolo MR

```text
[CHG-2026-0001] update-replicas demo-go-color-app
```

### Descrizione MR suggerita

```text
## Change Summary

Application: demo-go-color-app
Change Type: update-replicas
Requested By: vmarzario

## Files Changed

- apps/demo-go-color-app/deployment.yaml

## Validation

Tekton validation: pending

## GitOps Flow

After merge, Argo CD will sync the application.

## Rollback

Recommended rollback: revert this merge commit or create a revert MR.
```

### Regole

- MR source branch = branch ChangeRequest.
- MR target branch = default branch, tipicamente `main`.
- MR IID e URL devono essere salvati.
- Lo stato MR deve essere interrogabile.

---

## 4.6 Lettura ultimo commit

### Obiettivo

Recuperare ultimo commit rilevante per repository/path/branch.

### Uso

- dettaglio Application;
- correlazione con Argo CD revision;
- storico change;
- troubleshooting OutOfSync.

### Regole

- Se viene passato un path, GitLab deve filtrare commit relativi al path quando supportato.
- Il commit short SHA e full SHA devono essere gestiti.

---

## 4.7 Lettura stato MR

### Obiettivo

Capire se una MR è ancora aperta, merged, closed o bloccata.

### Stati minimi

```text
opened
merged
closed
locked
```

### Uso nel workflow

- se MR è `opened`, il change resta in attesa merge;
- se MR è `merged`, il sistema può procedere a sync Argo CD;
- se MR è `closed`, il change può passare a `Cancelled` o `Failed`, secondo policy.

---

## 5. GitLab Adapter

## 5.1 Responsabilità

Il componente `GitLabAdapter` incapsula tutte le interazioni con GitLab API.

Il resto dell’applicazione non deve conoscere endpoint HTTP GitLab, header token, encoding file o codici errore raw.

---

## 5.2 Interfaccia logica MVP

Interfaccia concettuale:

```go
type GitLabAdapter interface {
    GetProject(ctx context.Context, projectID string) (*GitLabProject, error)
    GetFile(ctx context.Context, projectID string, ref string, filePath string) (*RepositoryFile, error)
    ListCommits(ctx context.Context, projectID string, opts ListCommitsOptions) ([]GitCommit, error)
    GetBranch(ctx context.Context, projectID string, branch string) (*GitBranch, error)
    CreateBranch(ctx context.Context, projectID string, branch string, ref string) (*GitBranch, error)
    CommitFiles(ctx context.Context, projectID string, branch string, message string, actions []CommitAction) (*GitCommit, error)
    CreateMergeRequest(ctx context.Context, projectID string, req CreateMergeRequestRequest) (*MergeRequest, error)
    GetMergeRequest(ctx context.Context, projectID string, iid int) (*MergeRequest, error)
}
```

Nota: l’interfaccia è indicativa. La forma finale sarà definita durante l’implementazione Go.

---

## 5.3 Posizione package proposta

```text
internal/adapters/gitlab/
├── client.go
├── models.go
├── mapper.go
├── files.go
├── branches.go
├── commits.go
├── merge_requests.go
└── errors.go
```

### File `client.go`

Responsabilità:

- configurazione HTTP client;
- base URL;
- header autenticazione;
- timeout;
- gestione status code comuni.

### File `files.go`

Responsabilità:

- lettura file;
- decodifica base64;
- gestione file not found.

### File `branches.go`

Responsabilità:

- lettura branch;
- creazione branch;
- gestione branch exists.

### File `commits.go`

Responsabilità:

- commit multi-file;
- list commit;
- mapping commit metadata.

### File `merge_requests.go`

Responsabilità:

- create MR;
- get MR;
- mapping stato MR.

### File `errors.go`

Responsabilità:

- conversione errori GitLab in error model interno.

---

## 6. Modello dati normalizzato

## 6.1 GitRepository

```yaml
provider: gitlab
baseUrl: https://gitlab.example.local
projectId: "12345"
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
defaultBranch: main
```

---

## 6.2 RepositoryFile

```yaml
filePath: apps/demo-go-color-app/deployment.yaml
fileName: deployment.yaml
ref: main
blobId: abcdef
commitId: 123456
lastCommitId: 789abc
encoding: base64
contentDecoded: "apiVersion: apps/v1..."
contentSha256: "..."
```

---

## 6.3 GitBranch

```yaml
name: change/CHG-2026-0001-update-replicas
protected: false
merged: false
canPush: true
commit:
  id: b8e1d6b2fc88b41909f87d505739bc855de41516
  shortId: b8e1d6b
  title: "Previous commit title"
```

---

## 6.4 GitCommit

```yaml
id: b8e1d6b2fc88b41909f87d505739bc855de41516
shortId: b8e1d6b
title: "CHG-2026-0001 update-replicas demo-go-color-app"
message: "..."
authorName: "DevOps Control Plane"
committedDate: "2026-06-25T14:00:00+02:00"
webUrl: "https://gitlab.example.local/group/repo/-/commit/b8e1d6b"
```

---

## 6.5 MergeRequest

```yaml
iid: 42
id: 123456
title: "[CHG-2026-0001] update-replicas demo-go-color-app"
state: opened
sourceBranch: change/CHG-2026-0001-update-replicas
targetBranch: main
webUrl: "https://gitlab.example.local/group/repo/-/merge_requests/42"
mergeStatus: unchecked
```

---

## 7. Autenticazione verso GitLab

## 7.1 Token based authentication

Configurazione prevista:

```text
GITLAB_BASE_URL=https://gitlab.example.local
GITLAB_TOKEN=<from Secret>
GITLAB_TIMEOUT_SECONDS=30
GITLAB_DEFAULT_PROJECT_ID=
GITLAB_DEFAULT_BRANCH=main
```

### Regole

- `GITLAB_TOKEN` deve arrivare da Secret;
- il token non deve essere loggato;
- il token non deve essere restituito via API;
- il token non deve essere salvato in PostgreSQL;
- errori 401/403 devono essere normalizzati.

---

## 7.2 Tipi token

Per MVP/lab può essere usato un token tecnico, ma in produzione è preferibile usare token con privilegi minimi e scadenza definita.

Opzioni possibili:

- Project Access Token, se il workflow opera su un singolo progetto;
- Group Access Token, se serve accesso a più repository di un gruppo;
- Service account / bot user token, se disponibile e governato;
- Personal Access Token solo per sviluppo o lab controllato.

### Regola di sicurezza

Usare sempre il token con scope più restrittivo compatibile con il workflow.

---

## 7.3 Scope minimi indicativi

Per le funzioni MVP servono capacità di:

- leggere repository;
- leggere file;
- creare branch;
- creare commit;
- aprire MR.

In molti scenari GitLab, lo scope `api` abilita l’uso completo delle API richieste. Per ambienti più restrittivi, valutare token/project role con permessi minimi compatibili.

---

## 8. Workflow GitLab

## 8.1 Flow A - Lettura file

```text
Application Service / Workflow
  -> GitLab Adapter GetFile(projectId, ref, path)
  -> GitLab API repository files
  -> decode content
  -> return RepositoryFile
```

### Errori

- project not found;
- file not found;
- ref not found;
- unauthorized;
- timeout.

---

## 8.2 Flow B - Creazione branch ChangeRequest

```text
Workflow Engine
  -> build branch name
  -> GitLab Adapter CreateBranch(projectId, branch, ref=main)
  -> save branch in ChangeRequest
  -> add event BranchCreated
```

### Errori

- branch already exists;
- ref not found;
- protected branch policy;
- unauthorized.

---

## 8.3 Flow C - Commit multi-file

Esempio change ConfigMap:

```text
Workflow Engine
  -> read configmap.yaml
  -> modify values
  -> read kustomization.yaml, se necessario
  -> build commit actions
  -> GitLab Adapter CommitFiles
  -> save commit SHA
  -> add event FilesUpdated / CommitCreated
```

### Commit actions logiche

```yaml
- action: update
  filePath: apps/demo-go-color-app/configmap.yaml
  content: "..."
- action: update
  filePath: apps/demo-go-color-app/kustomization.yaml
  content: "..."
```

---

## 8.4 Flow D - Apertura Merge Request

```text
Workflow Engine
  -> GitLab Adapter CreateMergeRequest
  -> save MR IID and URL
  -> add event MergeRequestOpened
```

### Regole

- MR descrizione include Change ID;
- MR descrizione include file modificati;
- MR descrizione include stato validazione Tekton, se già disponibile;
- MR descrizione include rollback hint.

---

## 8.5 Flow E - Polling Merge Request

```text
Workflow Engine / Scheduler
  -> GitLab Adapter GetMergeRequest
  -> if state=merged: continue sync Argo CD
  -> if state=closed: cancel/fail change
  -> if state=opened: wait
```

### Nota MVP

Il polling MR può essere implementato dopo il primo workflow end-to-end. Nel primissimo MVP può essere manuale: l’utente conferma che la MR è stata mergiata.

---

## 9. Mapping ChangeRequest -> GitLab

## 9.1 Campi ChangeRequest

```yaml
changeId: CHG-2026-0001
applicationName: demo-go-color-app
changeType: update-replicas
git:
  provider: gitlab
  projectId: "12345"
  repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
  sourceBranch: change/CHG-2026-0001-update-replicas
  targetBranch: main
  commitSha: b8e1d6b2fc88b41909f87d505739bc855de41516
  commitShortSha: b8e1d6b
  mergeRequestIid: 42
  mergeRequestUrl: https://gitlab.example.local/group/repo/-/merge_requests/42
```

---

## 9.2 Change events GitLab

Eventi minimi:

```text
GitRepositoryResolved
GitFileRead
GitBranchCreated
GitFilesUpdated
GitCommitCreated
GitMergeRequestOpened
GitMergeRequestMerged
GitMergeRequestClosed
GitOperationFailed
```

---

## 10. File modification strategy

## 10.1 YAML modifications

DevOps Control Plane dovrà modificare YAML in modo sicuro.

Possibili approcci:

1. parsing YAML strutturato;
2. patch mirata su oggetti noti;
3. manipolazione testuale controllata solo per MVP/lab.

### Raccomandazione

Per workflow stabili, preferire parsing YAML e update strutturato.

### Rischio

Manipolazioni testuali possono rompere indentazione YAML, come già osservato nel lab con `valueFrom` e `configMapKeyRef`.

---

## 10.2 Diff generation

Il sistema deve produrre un diff leggibile prima del commit o nella ChangeRequest.

### Uso

- review;
- evidenza;
- audit;
- troubleshooting.

### Regola

Il diff non deve includere secret.

---

## 10.3 File locking/concurrency

MVP può impedire ChangeRequest concorrenti sulla stessa Application e stesso file.

### Esempio

Se esiste già un change aperto su:

```text
apps/demo-go-color-app/configmap.yaml
```

un nuovo change sullo stesso file deve essere bloccato o segnalato come conflitto.

---

## 11. Error model GitLab

## 11.1 Codici errore interni

```text
GITLAB_AUTH_FAILED
GITLAB_FORBIDDEN
GITLAB_PROJECT_NOT_FOUND
GITLAB_FILE_NOT_FOUND
GITLAB_REF_NOT_FOUND
GITLAB_BRANCH_EXISTS
GITLAB_BRANCH_CREATE_FAILED
GITLAB_COMMIT_FAILED
GITLAB_MR_CREATE_FAILED
GITLAB_MR_NOT_FOUND
GITLAB_RATE_LIMITED
GITLAB_TIMEOUT
GITLAB_UNKNOWN_ERROR
```

---

## 11.2 Struttura errore normalizzata

```yaml
code: GITLAB_FILE_NOT_FOUND
technicalMessage: "404 File Not Found"
userMessage: "Il file GitOps richiesto non esiste nel repository o nel branch indicato."
suggestedAction: "Verificare projectId, branch e path del file."
recoverable: true
```

---

## 11.3 Branch già esistente

Errore:

```text
GITLAB_BRANCH_EXISTS
```

Possibili azioni:

- riusare branch se la ChangeRequest è la stessa;
- fallire se branch appartiene ad altro change;
- generare nome branch alternativo;
- richiedere intervento manuale.

MVP consigliato:

```text
fallire con messaggio chiaro se branch esiste e non è collegato alla stessa ChangeRequest.
```

---

## 12. Evidence GitLab

## 12.1 Evidence minima

```yaml
provider: gitlab
projectId: "12345"
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
sourceBranch: change/CHG-2026-0001-update-replicas
targetBranch: main
commitSha: b8e1d6b2fc88b41909f87d505739bc855de41516
commitShortSha: b8e1d6b
mergeRequestIid: 42
mergeRequestUrl: https://gitlab.example.local/group/repo/-/merge_requests/42
filesChanged:
  - apps/demo-go-color-app/deployment.yaml
```

---

## 12.2 Evidence diff summary

```yaml
files:
  - path: apps/demo-go-color-app/deployment.yaml
    changeType: update
    summary: "replicas changed from 2 to 3"
```

---

## 12.3 Evidence in caso di errore

```yaml
provider: gitlab
operation: CreateBranch
errorCode: GITLAB_BRANCH_EXISTS
message: "Branch change/CHG-2026-0001-update-replicas already exists"
```

---

## 13. Configurazione GitLab

Variabili previste:

```text
GITLAB_BASE_URL=https://gitlab.example.local
GITLAB_TOKEN=<from secret>
GITLAB_TIMEOUT_SECONDS=30
GITLAB_DEFAULT_BRANCH=main
GITLAB_DEFAULT_PROJECT_ID=
GITLAB_INSECURE_TLS=false
GITLAB_CA_CERT_PATH=
```

### Regole

- token da Secret;
- base URL da ConfigMap/env;
- TLS insecure solo per lab;
- timeout obbligatorio;
- project ID per Application salvato in PostgreSQL o configurazione.

---

## 14. API interne DevOps Control Plane collegate a GitLab

## 14.1 Git metadata

```text
GET /api/applications/{name}/git/commits
GET /api/applications/{name}/git/files
```

---

## 14.2 Change workflow Git operations

```text
POST /api/changes/{id}/create-branch
POST /api/changes/{id}/update-files
POST /api/changes/{id}/open-merge-request
GET  /api/changes/{id}/git
```

---

## 14.3 Evidence

```text
GET /api/changes/{id}/evidence/gitlab
```

---

## 15. Security requirements specifici

- Non loggare `GITLAB_TOKEN`.
- Non restituire token via API.
- Non salvare token in PostgreSQL.
- Non includere token nelle evidenze.
- Usare token con privilegi minimi.
- Preferire Project Access Token o service account rispetto a token personali per produzione.
- Impostare scadenza token e rotazione.
- Non salvare `.env` in Git.
- Eseguire anti-secret check sui file modificati.

---

## 16. Testing strategy

## 16.1 Unit test

Testare:

- generazione branch name;
- normalizzazione path;
- mapping file GitLab;
- decodifica base64;
- mapping commit;
- mapping MR;
- mapping errori GitLab.

---

## 16.2 Test con fake GitLab client

Il Workflow Engine deve poter essere testato senza GitLab reale usando fake adapter.

Scenario:

```text
CreateBranch -> success
GetFile -> returns deployment.yaml
CommitFiles -> returns commit SHA
CreateMergeRequest -> returns MR IID
```

---

## 16.3 Integration test manuale MVP

Scenario minimo:

1. configurare GitLab base URL e token;
2. leggere file `deployment.yaml`;
3. creare branch test;
4. creare commit su branch;
5. aprire MR;
6. leggere stato MR;
7. verificare che ChangeRequest contenga branch, commit e MR.

---

## 16.4 Casi errore da simulare

- token errato;
- project ID errato;
- file non trovato;
- branch già esistente;
- commit fallito;
- MR non creabile;
- timeout API;
- rate limit.

---

## 17. MVP implementation order per GitLab adapter

Ordine consigliato:

1. configurazione GitLab;
2. HTTP client base;
3. autenticazione token;
4. `GetFile`;
5. decode base64;
6. `ListCommits`;
7. `CreateBranch`;
8. `CommitFiles`;
9. `CreateMergeRequest`;
10. `GetMergeRequest`;
11. mapping errori;
12. evidence builder.

---

## 18. Checklist di completamento integrazione GitLab

La prima implementazione GitLab sarà considerata pronta quando:

- il backend legge configurazione GitLab;
- il token è letto da Secret/env e non loggato;
- un file GitOps può essere letto;
- un branch ChangeRequest può essere creato;
- un commit può essere creato su branch;
- una MR può essere aperta;
- l’ultimo commit può essere letto;
- gli errori GitLab principali sono normalizzati;
- le evidenze GitLab sono salvabili in PostgreSQL;
- i test manuali sono documentati.

---

## 19. Relazione con altri documenti

Questo documento alimenta:

- `docs/10-data-model.md`, per entità GitRepository, GitCommit, MergeRequest;
- `docs/11-change-workflows.md`, per branch/commit/MR workflow;
- `docs/13-api-design.md`, per endpoint Git e ChangeRequest;
- `docs/09-security-rbac.md`, per gestione token GitLab;
- ADR specifico `ADR-0007-gitlab-api-as-git-provider.md`.

---

## 20. Riferimenti tecnici GitLab

Riferimenti da consultare durante l’implementazione:

- GitLab Repository Files API;
- GitLab Branches API;
- GitLab Commits API;
- GitLab Merge Requests API;
- GitLab Personal Access Tokens;
- GitLab Project Access Tokens.

---

## 21. Messaggio chiave

L’integrazione GitLab è il punto che consente a DevOps Control Plane di restare coerente con GitOps.

Il sistema non deve cambiare direttamente il cluster.

Deve invece produrre change Git tracciati:

```text
branch
  -> file update
  -> commit
  -> merge request
  -> validation
  -> merge
  -> Argo CD sync
```

In questo modo ogni modifica è leggibile, revisionabile, auditabile e reversibile.
