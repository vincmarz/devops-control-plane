# DevOps Control Plane - Data Model

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
- `docs/07-gitlab-integration.md`  
- `docs/08-tekton-integration.md`  
- `docs/09-security-rbac.md`  
**Stato:** Draft iniziale / Data Model

---

## 1. Scopo del documento

Questo documento definisce il **modello dati iniziale** del progetto **DevOps Control Plane**.

L’obiettivo è descrivere:

- entità principali;
- relazioni tra entità;
- tabelle PostgreSQL MVP;
- campi obbligatori e opzionali;
- convenzioni di naming;
- stati ChangeRequest;
- payload JSONB previsti;
- regole di auditabilità;
- regole di sicurezza sui dati;
- prime migration SQL;
- mapping tra modello dati, workflow e API.

Il modello dati deve supportare il primo obiettivo del prodotto:

```text
tracciare change GitOps end-to-end
```

Il flusso dati minimo da rappresentare è:

```text
Application
  -> ChangeRequest
  -> Git operation
  -> Tekton validation
  -> Argo CD sync
  -> Runtime evidence
  -> Change history
```

---

## 2. Principi del data model

## 2.1 PostgreSQL come change store

PostgreSQL è il database interno del DevOps Control Plane.

Responsabilità:

- conservare storico funzionale dei change;
- correlare GitLab, Argo CD, Tekton e OpenShift;
- salvare eventi di workflow;
- salvare evidenze tecniche;
- permettere audit e troubleshooting.

PostgreSQL non sostituisce:

- GitLab history;
- Argo CD history;
- Tekton Results;
- Kubernetes Events.

PostgreSQL conserva una vista funzionale e normalizzata del workflow.

---

## 2.2 Nessun secret nel database

Il database non deve contenere:

- token GitLab;
- token Argo CD;
- kubeconfig;
- password PostgreSQL;
- secret Kubernetes;
- `.dockerconfigjson`;
- private key;
- valori sensibili trovati in manifest.

Le evidenze devono essere sanitizzate prima della persistenza.

---

## 2.3 JSONB per payload estensibili

Il modello usa colonne `jsonb` per dati estensibili.

Esempi:

- payload iniziale ChangeRequest;
- dettaglio evento;
- snapshot Argo CD;
- summary Tekton;
- evidence runtime.

Regola:

```text
Usare colonne relazionali per campi interrogati spesso.
Usare jsonb per dati variabili e dettagli tecnici.
```

---

## 2.4 Event sourcing leggero

Ogni cambio stato rilevante deve generare un record in `change_events`.

Questo non è un event sourcing completo, ma consente di ricostruire la timeline del workflow.

Esempio:

```text
Created
BranchCreated
FilesUpdated
ValidationRequested
ValidationSucceeded
SyncRequested
SyncSucceeded
EvidenceCollected
Completed
```

---

## 3. Entità principali

## 3.1 Application

Rappresenta una Application Argo CD nota al DevOps Control Plane.

Origine principale:

```text
Argo CD API
```

Arricchimenti:

- GitLab repository metadata;
- namespace target;
- ultimo commit Git;
- stato sync/health corrente;
- path GitOps.

---

## 3.2 ChangeRequest

Rappresenta una richiesta di change GitOps.

Esempi:

- update repliche;
- update `APP_VERSION`;
- update `PAGE_COLOR`;
- update ConfigMap values;
- governance change futuro.

È l’entità centrale del sistema.

---

## 3.3 ChangeEvent

Rappresenta un evento della timeline di una ChangeRequest.

Esempi:

- branch creato;
- commit creato;
- validazione Tekton fallita;
- sync Argo CD completata;
- evidenza raccolta.

---

## 3.4 Evidence

Rappresenta una evidenza tecnica associata a una ChangeRequest.

Tipi:

- GitLab evidence;
- Tekton evidence;
- Argo CD evidence;
- Kubernetes/OpenShift runtime evidence;
- health check evidence;
- diff summary.

---

## 3.5 IntegrationSnapshot

Entità opzionale per salvare snapshot raw o normalizzati da sistemi esterni.

Uso futuro:

- snapshot Application Argo CD;
- snapshot PipelineRun;
- snapshot Deployment runtime.

Per MVP può essere sostituita da `evidences.payload`.

---

## 4. Diagramma relazionale logico

```text
applications
    |
    | 1:N
    v
change_requests
    |
    | 1:N
    v
change_events

change_requests
    |
    | 1:N
    v
evidences

change_requests
    |
    | 1:1 logical
    v
GitLab branch / commit / merge request fields

change_requests
    |
    | 1:1 logical
    v
Tekton validation fields

change_requests
    |
    | 1:1 logical
    v
Argo CD sync fields
```

---

## 5. Tabelle MVP

Per il primo MVP sono previste queste tabelle:

```text
applications
change_requests
change_events
evidences
```

Tabelle future possibili:

```text
users
application_repositories
application_environments
workflow_locks
integration_snapshots
audit_exports
```

---

## 6. Tabella applications

## 6.1 Scopo

Conserva le applicazioni note, normalmente scoperte da Argo CD API.

## 6.2 Campi

```sql
CREATE TABLE applications (
    id                  uuid PRIMARY KEY,
    name                text NOT NULL,
    argocd_namespace    text NOT NULL,
    argocd_project      text,
    target_namespace    text,

    repo_provider       text,
    repo_url            text,
    repo_project_id     text,
    repo_default_branch text,
    repo_path           text,

    target_revision     text,
    current_revision    text,
    sync_status         text,
    health_status       text,

    last_seen_at        timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),

    UNIQUE (argocd_namespace, name)
);
```

## 6.3 Note campi

### `name`

Nome Argo CD Application.

Esempio:

```text
demo-go-color-app
```

### `argocd_namespace`

Namespace in cui vive la Application Argo CD.

Esempio:

```text
openshift-gitops
```

### `repo_provider`

Per MVP:

```text
gitlab
```

### `repo_project_id`

Project ID GitLab numerico o path URL-encoded.

### `sync_status` e `health_status`

Snapshot ultimo dello stato Argo CD.

---

## 6.4 Indici consigliati

```sql
CREATE INDEX idx_applications_name ON applications(name);
CREATE INDEX idx_applications_argocd_project ON applications(argocd_project);
CREATE INDEX idx_applications_target_namespace ON applications(target_namespace);
CREATE INDEX idx_applications_repo_project_id ON applications(repo_project_id);
```

---

## 7. Tabella change_requests

## 7.1 Scopo

Conserva la richiesta di change e lo stato corrente del workflow.

È la tabella centrale del sistema.

## 7.2 Campi

```sql
CREATE TABLE change_requests (
    id                      uuid PRIMARY KEY,
    change_number           text UNIQUE NOT NULL,

    application_id          uuid REFERENCES applications(id),
    application_name        text NOT NULL,

    change_type             text NOT NULL,
    status                  text NOT NULL,
    requested_by            text,
    description             text,

    request_payload         jsonb,

    -- GitLab fields
    git_provider            text,
    gitlab_project_id       text,
    repo_url                text,
    source_branch           text,
    target_branch           text,
    commit_sha              text,
    commit_short_sha        text,
    merge_request_iid       integer,
    merge_request_url       text,
    merge_request_state     text,

    -- Tekton fields
    tekton_namespace        text,
    tekton_pipeline_name    text,
    tekton_pipelinerun_name text,
    tekton_status           text,
    tekton_started_at       timestamptz,
    tekton_completed_at     timestamptz,

    -- Argo CD fields
    argocd_application      text,
    argocd_project          text,
    argocd_sync_revision    text,
    argocd_sync_status      text,
    argocd_health_status    text,
    argocd_operation_phase  text,

    -- Runtime summary
    runtime_namespace       text,
    runtime_status          text,

    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now(),
    completed_at            timestamptz
);
```

---

## 7.3 Change number

Formato proposto:

```text
CHG-YYYY-NNNN
```

Esempio:

```text
CHG-2026-0001
```

Per MVP, il contatore può essere gestito dal database o dal codice applicativo.

Opzione consigliata iniziale:

- UUID come chiave tecnica;
- `change_number` come identificativo leggibile.

---

## 7.4 Change type

Valori MVP:

```text
update-replicas
update-app-version
update-page-color
update-configmap-values
```

Valori futuri:

```text
governance-appproject-change
rollback-git-revert
sync-only
validate-only
```

---

## 7.5 Status

Stati iniziali:

```text
Draft
Created
BranchCreated
FilesUpdated
ValidationRequested
ValidationRunning
ValidationSucceeded
ValidationFailed
MergeRequestOpened
Merged
SyncRequested
SyncRunning
SyncSucceeded
SyncFailed
EvidenceCollected
Completed
Failed
Cancelled
```

---

## 7.6 Request payload

Esempio `update-replicas`:

```json
{
  "replicas": 3,
  "deploymentName": "demo-go-color-app"
}
```

Esempio `update-configmap-values`:

```json
{
  "configMapName": "demo-go-color-app-config",
  "values": {
    "APP_VERSION": "v3-green",
    "PAGE_COLOR": "#28A745"
  }
}
```

---

## 7.7 Indici consigliati

```sql
CREATE INDEX idx_change_requests_application_id ON change_requests(application_id);
CREATE INDEX idx_change_requests_application_name ON change_requests(application_name);
CREATE INDEX idx_change_requests_status ON change_requests(status);
CREATE INDEX idx_change_requests_change_type ON change_requests(change_type);
CREATE INDEX idx_change_requests_created_at ON change_requests(created_at DESC);
CREATE INDEX idx_change_requests_source_branch ON change_requests(source_branch);
CREATE INDEX idx_change_requests_commit_sha ON change_requests(commit_sha);
```

---

## 8. Tabella change_events

## 8.1 Scopo

Conserva la timeline di ogni ChangeRequest.

## 8.2 Campi

```sql
CREATE TABLE change_events (
    id                  uuid PRIMARY KEY,
    change_request_id   uuid NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,

    event_type          text NOT NULL,
    previous_status     text,
    new_status          text,
    message             text,
    technical_message   text,
    error_code          text,
    source              text,
    payload             jsonb,

    created_at          timestamptz NOT NULL DEFAULT now()
);
```

---

## 8.3 Event type

Esempi:

```text
Created
GitRepositoryResolved
GitBranchCreated
GitFileRead
GitFilesUpdated
GitCommitCreated
GitMergeRequestOpened
ValidationRequested
ValidationRunning
ValidationSucceeded
ValidationFailed
SyncRequested
SyncRunning
SyncSucceeded
SyncFailed
EvidenceCollected
Completed
Failed
Cancelled
```

---

## 8.4 Source

Origine evento:

```text
system
gitlab
argocd
tekton
kubernetes
user
workflow
```

---

## 8.5 Payload esempio

```json
{
  "branch": "change/CHG-2026-0001-update-replicas",
  "projectId": "12345",
  "ref": "main"
}
```

Errore:

```json
{
  "operation": "CreateBranch",
  "recoverable": true,
  "suggestedAction": "Verificare se il branch è già collegato alla stessa ChangeRequest."
}
```

---

## 8.6 Indici consigliati

```sql
CREATE INDEX idx_change_events_change_request_id ON change_events(change_request_id);
CREATE INDEX idx_change_events_event_type ON change_events(event_type);
CREATE INDEX idx_change_events_created_at ON change_events(created_at DESC);
CREATE INDEX idx_change_events_source ON change_events(source);
```

---

## 9. Tabella evidences

## 9.1 Scopo

Conserva evidenze tecniche associate alla ChangeRequest.

## 9.2 Campi

```sql
CREATE TABLE evidences (
    id                  uuid PRIMARY KEY,
    change_request_id   uuid NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,

    evidence_type       text NOT NULL,
    name                text,
    summary             text,
    content             text,
    payload             jsonb,
    external_ref        text,

    sanitized           boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now()
);
```

---

## 9.3 Evidence type

Valori iniziali:

```text
gitlab
tekton
argocd
kubernetes-runtime
health-check
diff-summary
workflow-summary
security-validation
```

---

## 9.4 Evidence GitLab esempio

```json
{
  "provider": "gitlab",
  "projectId": "12345",
  "repoUrl": "https://gitlab.example.local/group/demo-app-gitops.git",
  "sourceBranch": "change/CHG-2026-0001-update-replicas",
  "targetBranch": "main",
  "commitSha": "b8e1d6b2fc88b41909f87d505739bc855de41516",
  "mergeRequestIid": 42,
  "mergeRequestUrl": "https://gitlab.example.local/group/repo/-/merge_requests/42",
  "filesChanged": [
    "apps/demo-go-color-app/deployment.yaml"
  ]
}
```

---

## 9.5 Evidence Tekton esempio

```json
{
  "provider": "tekton",
  "namespace": "devops-control-plane",
  "pipelineRunName": "validate-gitops-change-chg-2026-0001-abcde",
  "status": "Succeeded",
  "taskRuns": [
    {
      "name": "yaml-validate",
      "status": "Succeeded"
    },
    {
      "name": "server-side-dry-run",
      "status": "Succeeded"
    }
  ]
}
```

---

## 9.6 Evidence Argo CD esempio

```json
{
  "application": "demo-go-color-app",
  "project": "devops-ci-demo",
  "syncStatus": "Synced",
  "healthStatus": "Healthy",
  "revision": "b8e1d6b",
  "operationPhase": "Succeeded"
}
```

---

## 9.7 Evidence runtime esempio

```json
{
  "namespace": "devops-ci-demo",
  "deployment": {
    "name": "demo-go-color-app",
    "availableReplicas": 3,
    "readyReplicas": 3,
    "updatedReplicas": 3
  },
  "pods": [
    {
      "name": "demo-go-color-app-abcde",
      "phase": "Running",
      "ready": true
    }
  ]
}
```

---

## 9.8 Indici consigliati

```sql
CREATE INDEX idx_evidences_change_request_id ON evidences(change_request_id);
CREATE INDEX idx_evidences_evidence_type ON evidences(evidence_type);
CREATE INDEX idx_evidences_created_at ON evidences(created_at DESC);
```

---

## 10. Tabelle future

## 10.1 users

Quando sarà introdotta autenticazione reale:

```sql
CREATE TABLE users (
    id              uuid PRIMARY KEY,
    username        text UNIQUE NOT NULL,
    display_name    text,
    email           text,
    provider        text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);
```

Per MVP, `requested_by` può restare testo.

---

## 10.2 workflow_locks

Per impedire change concorrenti sullo stesso file/applicazione:

```sql
CREATE TABLE workflow_locks (
    id                  uuid PRIMARY KEY,
    application_id      uuid REFERENCES applications(id),
    resource_key        text NOT NULL,
    change_request_id   uuid REFERENCES change_requests(id),
    created_at          timestamptz NOT NULL DEFAULT now(),
    expires_at          timestamptz
);
```

Esempio `resource_key`:

```text
demo-go-color-app:apps/demo-go-color-app/configmap.yaml
```

---

## 10.3 integration_snapshots

Per salvare snapshot raw o semi-normalizzati:

```sql
CREATE TABLE integration_snapshots (
    id                  uuid PRIMARY KEY,
    change_request_id   uuid REFERENCES change_requests(id),
    application_id      uuid REFERENCES applications(id),
    provider            text NOT NULL,
    snapshot_type       text NOT NULL,
    payload             jsonb NOT NULL,
    created_at          timestamptz NOT NULL DEFAULT now()
);
```

Provider:

```text
gitlab
argocd
tekton
kubernetes
```

---

## 11. Errori e stati

## 11.1 Error codes

Gli error code sono salvati in `change_events.error_code`.

Famiglie:

```text
GITLAB_*
ARGO_*
TEKTON_*
KUBERNETES_*
DATABASE_*
VALIDATION_*
SECURITY_*
WORKFLOW_*
```

Esempi:

```text
GITLAB_FILE_NOT_FOUND
GITLAB_BRANCH_EXISTS
ARGO_OPERATION_IN_PROGRESS
ARGO_RESOURCE_NOT_PERMITTED
TEKTON_PIPELINERUN_FAILED
VALIDATION_SECRET_DETECTED
WORKFLOW_CONFLICT_ACTIVE_CHANGE
```

---

## 11.2 Status finale

Stati finali:

```text
Completed
Failed
Cancelled
```

Stati non finali:

```text
Created
BranchCreated
ValidationRunning
SyncRunning
```

Regola:

```text
Una ChangeRequest in stato non finale deve avere updated_at recente o un timeout gestito.
```

---

## 12. Sicurezza dati

## 12.1 Campi vietati

Non creare colonne per:

- `gitlab_token`;
- `argocd_token`;
- `kubeconfig`;
- `secret_value`;
- `password` non strettamente necessaria.

---

## 12.2 Payload sanitizzato

Ogni `payload` e `content` deve essere sanitizzato se deriva da log o output esterno.

Regola:

```text
Se non sei sicuro che un payload sia sicuro, non salvarlo integralmente.
```

---

## 12.3 Evidences sanitized flag

La colonna `sanitized` indica se l’evidenza è stata filtrata.

Per MVP deve essere sempre:

```text
true
```

Se in futuro si salva un riferimento esterno a log raw, la gestione deve essere documentata.

---

## 13. Migration iniziale proposta

File suggerito:

```text
migrations/000001_init.up.sql
```

Contenuto:

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE applications (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name                text NOT NULL,
    argocd_namespace    text NOT NULL,
    argocd_project      text,
    target_namespace    text,
    repo_provider       text,
    repo_url            text,
    repo_project_id     text,
    repo_default_branch text,
    repo_path           text,
    target_revision     text,
    current_revision    text,
    sync_status         text,
    health_status       text,
    last_seen_at        timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    UNIQUE (argocd_namespace, name)
);

CREATE TABLE change_requests (
    id                      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    change_number           text UNIQUE NOT NULL,
    application_id          uuid REFERENCES applications(id),
    application_name        text NOT NULL,
    change_type             text NOT NULL,
    status                  text NOT NULL,
    requested_by            text,
    description             text,
    request_payload         jsonb,
    git_provider            text,
    gitlab_project_id       text,
    repo_url                text,
    source_branch           text,
    target_branch           text,
    commit_sha              text,
    commit_short_sha        text,
    merge_request_iid       integer,
    merge_request_url       text,
    merge_request_state     text,
    tekton_namespace        text,
    tekton_pipeline_name    text,
    tekton_pipelinerun_name text,
    tekton_status           text,
    tekton_started_at       timestamptz,
    tekton_completed_at     timestamptz,
    argocd_application      text,
    argocd_project          text,
    argocd_sync_revision    text,
    argocd_sync_status      text,
    argocd_health_status    text,
    argocd_operation_phase  text,
    runtime_namespace       text,
    runtime_status          text,
    created_at              timestamptz NOT NULL DEFAULT now(),
    updated_at              timestamptz NOT NULL DEFAULT now(),
    completed_at            timestamptz
);

CREATE TABLE change_events (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    change_request_id   uuid NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,
    event_type          text NOT NULL,
    previous_status     text,
    new_status          text,
    message             text,
    technical_message   text,
    error_code          text,
    source              text,
    payload             jsonb,
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE evidences (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    change_request_id   uuid NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,
    evidence_type       text NOT NULL,
    name                text,
    summary             text,
    content             text,
    payload             jsonb,
    external_ref        text,
    sanitized           boolean NOT NULL DEFAULT true,
    created_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_applications_name ON applications(name);
CREATE INDEX idx_applications_argocd_project ON applications(argocd_project);
CREATE INDEX idx_applications_target_namespace ON applications(target_namespace);
CREATE INDEX idx_applications_repo_project_id ON applications(repo_project_id);

CREATE INDEX idx_change_requests_application_id ON change_requests(application_id);
CREATE INDEX idx_change_requests_application_name ON change_requests(application_name);
CREATE INDEX idx_change_requests_status ON change_requests(status);
CREATE INDEX idx_change_requests_change_type ON change_requests(change_type);
CREATE INDEX idx_change_requests_created_at ON change_requests(created_at DESC);
CREATE INDEX idx_change_requests_source_branch ON change_requests(source_branch);
CREATE INDEX idx_change_requests_commit_sha ON change_requests(commit_sha);

CREATE INDEX idx_change_events_change_request_id ON change_events(change_request_id);
CREATE INDEX idx_change_events_event_type ON change_events(event_type);
CREATE INDEX idx_change_events_created_at ON change_events(created_at DESC);
CREATE INDEX idx_change_events_source ON change_events(source);

CREATE INDEX idx_evidences_change_request_id ON evidences(change_request_id);
CREATE INDEX idx_evidences_evidence_type ON evidences(evidence_type);
CREATE INDEX idx_evidences_created_at ON evidences(created_at DESC);
```

---

## 14. Migration rollback proposta

File suggerito:

```text
migrations/000001_init.down.sql
```

Contenuto:

```sql
DROP TABLE IF EXISTS evidences;
DROP TABLE IF EXISTS change_events;
DROP TABLE IF EXISTS change_requests;
DROP TABLE IF EXISTS applications;
```

Nota:

In contesti production-like, il rollback distruttivo deve essere usato con estrema cautela.

---

## 15. Query utili MVP

## 15.1 Lista change recenti

```sql
SELECT
    change_number,
    application_name,
    change_type,
    status,
    requested_by,
    created_at,
    completed_at
FROM change_requests
ORDER BY created_at DESC
LIMIT 50;
```

---

## 15.2 Timeline di una ChangeRequest

```sql
SELECT
    event_type,
    previous_status,
    new_status,
    message,
    error_code,
    source,
    created_at
FROM change_events
WHERE change_request_id = $1
ORDER BY created_at ASC;
```

---

## 15.3 Evidenze di una ChangeRequest

```sql
SELECT
    evidence_type,
    name,
    summary,
    sanitized,
    created_at
FROM evidences
WHERE change_request_id = $1
ORDER BY created_at ASC;
```

---

## 15.4 Change attivi per Application

```sql
SELECT
    change_number,
    application_name,
    change_type,
    status,
    source_branch,
    created_at
FROM change_requests
WHERE application_name = $1
  AND status NOT IN ('Completed', 'Failed', 'Cancelled')
ORDER BY created_at ASC;
```

---

## 16. Mapping verso API

## 16.1 `GET /api/changes`

Origine principale:

```text
change_requests
```

Campi restituiti:

- change number;
- application;
- type;
- status;
- requested by;
- created at;
- completed at.

---

## 16.2 `GET /api/changes/{id}`

Origine:

```text
change_requests
change_events
evidences
```

Restituisce:

- dettaglio ChangeRequest;
- timeline eventi;
- riferimenti GitLab;
- stato Tekton;
- stato Argo CD;
- evidenze.

---

## 16.3 `GET /api/applications`

Origine:

```text
Argo CD API live
applications cache opzionale
```

Per MVP, può interrogare Argo CD live e aggiornare `applications`.

---

## 17. Mapping verso workflow

## 17.1 update-replicas

Campi principali:

```text
change_requests.change_type = update-replicas
change_requests.request_payload.replicas
change_requests.source_branch
change_requests.commit_sha
change_requests.tekton_pipelinerun_name
change_requests.argocd_sync_status
change_requests.runtime_status
```

---

## 17.2 update-configmap-values

Campi principali:

```text
change_requests.change_type = update-configmap-values
change_requests.request_payload.configMapName
change_requests.request_payload.values
change_requests.source_branch
change_requests.commit_sha
evidences.evidence_type = diff-summary
```

---

## 18. Regole di consistenza

## 18.1 ChangeRequest senza Application ID

Per MVP, `application_id` può essere nullable perché alcune ChangeRequest potrebbero nascere prima della cache Application.

Tuttavia `application_name` è obbligatorio.

---

## 18.2 Stato finale e completed_at

Regola:

```text
Se status IN ('Completed', 'Failed', 'Cancelled'), completed_at deve essere valorizzato.
```

Questa regola può essere applicata nel codice all’inizio e poi con constraint/check in futuro.

---

## 18.3 Stato Tekton

Se `tekton_pipelinerun_name` è valorizzato, `tekton_namespace` dovrebbe essere valorizzato.

---

## 18.4 Stato Argo CD

Se `argocd_application` è valorizzato, `argocd_sync_status` e `argocd_health_status` dovrebbero rappresentare l’ultimo stato osservato.

---

## 19. Backup e retention

## 19.1 Backup

PostgreSQL conserva storico change ed evidenze.

Per MVP:

```text
Documentare backup necessario.
```

Per futuro:

- backup schedulato;
- retention policy;
- restore testato;
- export audit.

---

## 19.2 Retention evidenze

Le evidenze possono crescere.

Per MVP:

- salvare summary e payload sanitizzati;
- evitare log enormi;
- limitare `content` a excerpt.

Per futuro:

- storage esterno;
- retention configurabile;
- compressione evidenze;
- Tekton Results integration.

---

## 20. Data model validation checklist

Il data model MVP è considerato pronto quando:

- le quattro tabelle principali sono definite;
- ogni ChangeRequest ha Change ID leggibile;
- gli eventi sono persistiti;
- le evidenze sono persistite;
- nessun secret è previsto nel modello;
- GitLab, Tekton e Argo CD hanno campi minimi di correlazione;
- le query principali sono supportate;
- le migration up/down sono disponibili;
- il modello alimenta API e workflow.

---

## 21. Relazione con altri documenti

Questo documento alimenta:

- `migrations/000001_init.up.sql`;
- `migrations/000001_init.down.sql`;
- `internal/domain/`;
- `internal/database/`;
- `docs/11-change-workflows.md`;
- `docs/12-evidence-model.md`;
- `docs/13-api-design.md`;
- ADR `ADR-0004-postgresql-change-history.md`.

---

## 22. Messaggio chiave

Il data model del DevOps Control Plane deve permettere di ricostruire ogni change in modo semplice.

Domande a cui il database deve rispondere:

```text
Chi ha richiesto il change?
Su quale applicazione?
Che cosa voleva modificare?
Quale branch GitLab è stato creato?
Quale commit o MR ha rappresentato il change?
Quale PipelineRun Tekton ha validato il change?
Quale revisione Argo CD è stata sincronizzata?
Quale stato runtime è stato osservato?
Quali evidenze sono disponibili?
```

Il database non deve sostituire Git, Argo CD o Tekton. Deve correlare le informazioni prodotte da questi strumenti in uno storico funzionale unico, leggibile e auditabile.
