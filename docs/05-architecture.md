# DevOps Control Plane - Architecture

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
**Stato:** Draft iniziale / Architecture

---

## 1. Scopo del documento

Questo documento descrive l’architettura iniziale del progetto **DevOps Control Plane**.

L’obiettivo è definire:

- componenti principali;
- responsabilità dei componenti;
- boundary tra dominio e integrazioni esterne;
- flussi principali;
- modello di deployment iniziale;
- dipendenze esterne;
- modalità di integrazione con GitLab, Argo CD, Tekton, Kubernetes/OpenShift e PostgreSQL;
- principi di sicurezza e osservabilità architetturale;
- evoluzione prevista dopo il primo MVP.

Il documento traduce la vision e i requisiti in una struttura tecnica implementabile.

---

## 2. Executive summary

**DevOps Control Plane** è un’applicazione backend scritta in Go che orchestra workflow GitOps standardizzati.

Il sistema non sostituisce GitLab, Argo CD, Tekton o OpenShift. Li coordina.

```text
GitLab       = gestione repository, branch, commit, merge request
Argo CD      = motore GitOps di sync e riconciliazione
Tekton       = motore di validazione workflow
OpenShift    = piattaforma runtime
PostgreSQL   = storico ChangeRequest, eventi ed evidenze
Go backend   = orchestrazione, API, workflow, adapter
```

Il principio architetturale fondamentale è:

```text
Ogni change permanente deve passare da Git.
```

DevOps Control Plane non deve diventare un pannello alternativo di modifica runtime del cluster.

---

## 3. Stack architetturale scelto

## 3.1 Backend

```text
Go
```

Responsabilità:

- API HTTP;
- orchestrazione workflow;
- gestione ChangeRequest;
- integrazione adapter;
- persistenza PostgreSQL;
- raccolta evidenze;
- predisposizione HTML templates.

---

## 3.2 Frontend MVP

```text
Go HTML templates + Bootstrap
```

La UI completa arriverà dopo la stabilizzazione dei workflow.

Nel primo MVP il backend può esporre pagine minime per:

- health/debug;
- lista applicazioni;
- lista change;
- dettaglio change.

Tuttavia, la priorità resta:

```text
API e workflow stabili prima di una UI ricca.
```

---

## 3.3 Database

```text
PostgreSQL
```

Responsabilità:

- Application snapshot/cache;
- ChangeRequest;
- ChangeEvent;
- Evidence;
- riferimenti GitLab;
- riferimenti Tekton;
- riferimenti Argo CD;
- audit trail operativo.

---

## 3.4 Git integration

```text
GitLab API
```

Responsabilità:

- leggere file dal repository;
- leggere branch e commit;
- creare branch;
- aggiornare file;
- creare commit;
- aprire merge request;
- leggere stato merge request.

Nota: il codice sorgente del progetto può essere ospitato su GitHub, ma l’integrazione funzionale MVP verso repository applicativi/GitOps target è GitLab API.

---

## 3.5 Argo CD integration

```text
Argo CD API
```

Responsabilità:

- lista Application;
- dettaglio Application;
- resources;
- orphaned resources;
- history;
- sync;
- stato operazione;
- polling fino a `Synced` e `Healthy`.

---

## 3.6 Tekton integration

```text
Kubernetes API diretta
```

Responsabilità:

- creare PipelineRun;
- monitorare PipelineRun;
- leggere TaskRun;
- raccogliere log;
- salvare evidenze validazione.

---

## 3.7 OpenShift/Kubernetes integration

```text
Kubernetes API diretta
```

Responsabilità:

- leggere Deployment;
- leggere ReplicaSet;
- leggere Pod;
- leggere ConfigMap;
- leggere Service;
- leggere Route, se disponibile tramite API OpenShift;
- raccogliere evidenze runtime;
- eventuale dry-run server-side tramite validation workflow.

---

## 4. Architettura logica alto livello

```text
+-------------------------------------------------------------+
|                    DevOps Control Plane                     |
|                                                             |
|  +--------------------+       +--------------------------+  |
|  | HTTP API / Web     |       | Workflow Engine          |  |
|  | Go handlers        |-----> | Change orchestration     |  |
|  +--------------------+       +--------------------------+  |
|             |                              |                 |
|             |                              v                 |
|             |                 +--------------------------+  |
|             |                 | Domain Services          |  |
|             |                 | Application / Change     |  |
|             |                 +--------------------------+  |
|             |                              |                 |
|             v                              v                 |
|  +-------------------------------------------------------+  |
|  |                    Adapters Layer                     |  |
|  |                                                       |  |
|  |  GitLab  |  Argo CD  |  Kubernetes  |  Tekton         |  |
|  +-------------------------------------------------------+  |
|             |                              |                 |
|             v                              v                 |
|  +--------------------+       +--------------------------+  |
|  | PostgreSQL         |       | Evidence Collector       |  |
|  | Change store       |       | Runtime/validation data  |  |
|  +--------------------+       +--------------------------+  |
+-------------------------------------------------------------+

External systems:

GitLab API
Argo CD API
Kubernetes/OpenShift API
Tekton CRDs
PostgreSQL
```

---

## 5. Componenti principali

## 5.1 HTTP API Layer

### Responsabilità

- esporre endpoint REST;
- validare input HTTP;
- autenticare/autorizzare, quando implementato;
- inoltrare comandi al Workflow Engine;
- restituire risposte JSON;
- predisporre rendering HTML templates.

### Esempi endpoint

```text
GET  /healthz
GET  /readyz
GET  /api/applications
GET  /api/applications/{name}
GET  /api/changes
GET  /api/changes/{id}
POST /api/changes
POST /api/changes/{id}/validate
POST /api/changes/{id}/sync
```

### Regole

- gli handler HTTP non devono contenere logica GitLab/Argo/Tekton diretta;
- gli handler chiamano servizi applicativi;
- gli errori devono essere tradotti in risposte leggibili.

---

## 5.2 Domain Layer

### Responsabilità

Contiene il modello concettuale del sistema:

- Application;
- ChangeRequest;
- ChangeEvent;
- Evidence;
- WorkflowStatus;
- GitReference;
- ArgoStatus;
- TektonValidation;
- RuntimeEvidence.

### Regole

- non dipende da dettagli HTTP GitLab;
- non dipende da dettagli HTTP Argo CD;
- non dipende da implementazioni concrete Kubernetes client;
- deve essere testabile con unit test.

---

## 5.3 Workflow Engine

### Responsabilità

Coordina i workflow end-to-end.

Esempio workflow `update-replicas`:

```text
Create ChangeRequest
  -> read Application metadata
  -> create GitLab branch
  -> update YAML
  -> commit or MR
  -> create Tekton PipelineRun
  -> wait validation
  -> sync Argo CD
  -> wait Synced/Healthy
  -> collect evidence
  -> complete ChangeRequest
```

### Regole

- ogni passaggio crea un evento;
- ogni fallimento produce stato esplicito;
- nessun workflow deve restare appeso senza timeout;
- le operazioni lunghe devono essere asincrone o tracciate con polling.

---

## 5.4 Application Service

### Responsabilità

Gestisce le informazioni sulle Application Argo CD.

Funzioni:

- lista applicazioni;
- dettaglio applicazione;
- resources;
- orphaned resources;
- history;
- correlazione con Git repository/path;
- update cache PostgreSQL.

### Dipendenze

- Argo CD Adapter;
- GitLab Adapter, per ultimo commit;
- PostgreSQL Repository.

---

## 5.5 Change Service

### Responsabilità

Gestisce ChangeRequest e ChangeEvent.

Funzioni:

- creare ChangeRequest;
- aggiornare stato;
- aggiungere evento;
- associare Git branch/commit/MR;
- associare Tekton PipelineRun;
- associare Argo CD sync result;
- associare evidenze.

### Dipendenze

- PostgreSQL Repository;
- Workflow Engine.

---

## 5.6 Evidence Service

### Responsabilità

Raccoglie e normalizza evidenze.

Tipi di evidenza iniziali:

- Git diff;
- Git commit/MR;
- Tekton PipelineRun status;
- Tekton TaskRun status;
- Argo CD Application status;
- Kubernetes Deployment status;
- Pod status;
- ConfigMap status;
- Route/health check.

### Regole

- non salvare secret;
- non salvare token;
- salvare payload JSON quando utile;
- salvare summary leggibile.

---

## 6. Adapter Layer

L’Adapter Layer isola il dominio dai dettagli tecnici delle integrazioni esterne.

---

## 6.1 GitLab Adapter

### Responsabilità

Interagisce con GitLab API.

### Operazioni MVP

```text
GetFile(projectId, ref, path)
CreateBranch(projectId, branch, ref)
CommitFiles(projectId, branch, message, actions)
CreateMergeRequest(projectId, sourceBranch, targetBranch, title, description)
GetCommit(projectId, sha)
ListCommits(projectId, ref, path)
GetMergeRequest(projectId, iid)
```

### Regole

- token GitLab mai loggato;
- errori GitLab normalizzati;
- branch conflict gestito;
- file not found gestito;
- commit SHA salvato nella ChangeRequest.

---

## 6.2 Argo CD Adapter

### Responsabilità

Interagisce con Argo CD API.

### Operazioni MVP

```text
ListApplications()
GetApplication(name)
GetApplicationResources(name)
GetApplicationHistory(name)
SyncApplication(name)
GetOperationState(name)
WaitForSyncedHealthy(name, timeout)
```

### Regole

- token Argo CD mai loggato;
- sync operation already in progress gestita;
- errori AppProject interpretati;
- stato OutOfSync non sempre equivale a errore applicativo;
- OrphanedResourceWarning va mostrato come warning, non automaticamente come failure.

---

## 6.3 Kubernetes Adapter

### Responsabilità

Interagisce con Kubernetes/OpenShift API.

### Operazioni MVP

```text
GetDeployment(namespace, name)
GetPodsByLabel(namespace, selector)
GetConfigMap(namespace, name)
GetService(namespace, name)
GetRoute(namespace, name)
GetEvents(namespace, selector)
```

### Regole

- permessi minimi;
- read-only per evidenze runtime, salvo creazione PipelineRun Tekton;
- non modificare workload applicativi come stato finale.

---

## 6.4 Tekton Adapter

### Responsabilità

Gestisce risorse Tekton tramite Kubernetes API.

### Operazioni MVP

```text
CreatePipelineRun(namespace, spec)
GetPipelineRun(namespace, name)
ListTaskRunsForPipelineRun(namespace, pipelineRunName)
GetTaskRunLogs(namespace, taskRunName)
WaitPipelineRun(namespace, name, timeout)
```

### Regole

- PipelineRun deve includere Change ID in label/annotation;
- timeout obbligatorio;
- log raccolti senza secret;
- stato finale associato a ChangeRequest.

---

## 6.5 PostgreSQL Repository Layer

### Responsabilità

Gestisce la persistenza.

Entità principali:

- applications;
- change_requests;
- change_events;
- evidences;
- integration_snapshots, se necessario.

### Regole

- transazioni per transizioni critiche;
- timestamp sempre presenti;
- payload JSONB per dati estensibili;
- nessun secret persistito.

---

## 7. Architettura package Go proposta

Struttura iniziale:

```text
devops-control-plane/
├── cmd/
│   └── devops-control-plane/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── router.go
│   │   ├── health_handlers.go
│   │   ├── application_handlers.go
│   │   └── change_handlers.go
│   ├── app/
│   │   ├── application_service.go
│   │   ├── change_service.go
│   │   └── evidence_service.go
│   ├── workflow/
│   │   ├── engine.go
│   │   ├── update_replicas.go
│   │   ├── update_app_version.go
│   │   └── update_configmap.go
│   ├── domain/
│   │   ├── application.go
│   │   ├── change_request.go
│   │   ├── change_event.go
│   │   └── evidence.go
│   ├── adapters/
│   │   ├── argocd/
│   │   ├── gitlab/
│   │   ├── kubernetes/
│   │   └── tekton/
│   ├── database/
│   │   ├── postgres.go
│   │   └── repositories.go
│   ├── config/
│   │   └── config.go
│   ├── logging/
│   │   └── logger.go
│   └── web/
│       ├── handlers/
│       ├── templates/
│       └── static/
├── migrations/
├── manifests/
├── pipelines/
├── docs/
└── README.md
```

### Regole package

- `domain` non importa adapter;
- `workflow` usa interfacce, non implementazioni concrete;
- `adapters` implementa interfacce;
- `api` chiama servizi applicativi;
- `database` incapsula SQL.

---

## 8. Flussi architetturali principali

## 8.1 Flow A - Application discovery

```text
User/API client
  -> HTTP API GET /api/applications
  -> Application Service
  -> Argo CD Adapter
  -> Argo CD API
  -> normalize Application list
  -> optional PostgreSQL cache update
  -> response JSON
```

### Output logico

```yaml
- name: demo-go-color-app
  project: devops-ci-demo
  syncStatus: Synced
  healthStatus: Healthy
  currentRevision: b8e1d6b
```

---

## 8.2 Flow B - Dettaglio Application

```text
User/API client
  -> GET /api/applications/{name}
  -> Application Service
  -> Argo CD Adapter GetApplication
  -> Argo CD Adapter Resources
  -> Argo CD Adapter History
  -> GitLab Adapter latest commit, se configurato
  -> merge normalized view
  -> response JSON
```

### Nota

Il sistema deve distinguere:

```text
Argo CD revision != sempre ultimo commit Git visualizzato
```

Esempio: rollback operativo Argo CD può rendere il cluster temporaneamente diverso da `main`.

---

## 8.3 Flow C - Change repliche

```text
POST /api/changes
  -> create ChangeRequest
  -> Workflow Engine update-replicas
  -> GitLab Adapter create branch
  -> GitLab Adapter read deployment.yaml
  -> YAML modifier update spec.replicas
  -> GitLab Adapter commit or MR
  -> Tekton Adapter create validation PipelineRun
  -> wait validation
  -> Argo CD Adapter sync
  -> wait Synced/Healthy
  -> Kubernetes Adapter collect runtime evidence
  -> Evidence Service persist evidence
  -> ChangeRequest Completed
```

### Regola fondamentale

Il Deployment runtime non viene scalato direttamente.

---

## 8.4 Flow D - Change ConfigMap value

```text
POST /api/changes
  -> changeType update-configmap-values
  -> read Application metadata
  -> verify configmap.yaml exists
  -> verify kustomization.yaml includes configmap.yaml
  -> verify AppProject allows ConfigMap, se disponibile
  -> create GitLab branch
  -> update configmap.yaml
  -> commit/MR
  -> Tekton validation
  -> Argo CD sync
  -> runtime evidence
```

### Rischio noto

Aggiornare una ConfigMap usata come env var non garantisce rollout automatico dei Pod.

Decisione futura da formalizzare in ADR:

```text
Come forzare/referenziarie rollout quando cambia ConfigMap?
```

Opzioni possibili:

- annotation checksum nel Pod template;
- annotation change-id nel Deployment;
- rollout restart controllato;
- esplicitare che alcune ConfigMap richiedono nuovo change del Deployment.

---

## 8.5 Flow E - Tekton validation

```text
Workflow Engine
  -> Tekton Adapter CreatePipelineRun
  -> Kubernetes API creates PipelineRun
  -> Tekton creates TaskRuns
  -> Tekton Adapter watches PipelineRun status
  -> Tekton Adapter collects TaskRuns/logs
  -> Evidence Service stores validation result
```

### Pipeline tasks candidate

```text
clone git branch
show diff
kustomize build
server-side dry-run
anti-secret check
AppProject policy check
report
```

---

## 8.6 Flow F - Argo CD sync and wait

```text
Workflow Engine
  -> Argo CD Adapter SyncApplication
  -> Argo CD starts operation
  -> poll operation state
  -> poll Application sync/health
  -> if Synced/Healthy: success
  -> else timeout/failure
  -> save evidence
```

### Errori noti da gestire

```text
another operation is already in progress
resource :ConfigMap is not permitted in project devops-ci-demo
manifest invalid
Application Degraded
Application OutOfSync after timeout
```

---

## 9. Data flow e persistenza

## 9.1 Dati transienti

Dati usati solo durante workflow:

- token API;
- HTTP response raw;
- log completi temporanei;
- file temporanei YAML;
- kube client session.

Questi dati non devono essere persistiti se contengono informazioni sensibili.

---

## 9.2 Dati persistenti

Dati da salvare:

- Application metadata normalizzati;
- ChangeRequest;
- ChangeEvent;
- branch;
- commit;
- MR;
- PipelineRun;
- sync result;
- evidence summary;
- payload JSONB sanitizzato.

---

## 9.3 Dati vietati in persistenza

Non salvare:

- GitLab token;
- Argo CD token;
- kubeconfig;
- password PostgreSQL;
- secret Kubernetes;
- dockerconfig auth;
- private key;
- output contenenti credenziali.

---

## 10. Modello di deployment iniziale

## 10.1 Deployment locale/lab

Nel primissimo sviluppo:

```text
Go process su bastion o workstation
PostgreSQL locale o container
accesso Argo CD API tramite route
accesso GitLab API via rete
accesso Kubernetes API tramite kubeconfig/token
```

---

## 10.2 Deployment OpenShift target

In OpenShift il sistema sarà eseguito come:

```text
Namespace: devops-control-plane
Deployment: devops-control-plane
Service: devops-control-plane
Route: devops-control-plane
ConfigMap: configurazione non sensibile
Secret: token e credenziali
ServiceAccount: devops-control-plane
Role/RoleBinding: permessi minimi
PostgreSQL: servizio dedicato o managed
```

---

## 10.3 Container runtime

Requisiti:

- container non privilegiato;
- porta HTTP configurabile;
- readiness probe su `/readyz`;
- liveness probe su `/healthz`;
- configurazione tramite env;
- nessun secret embedded nell’immagine.

---

## 11. Security architecture

## 11.1 Trust boundaries

```text
User/API client
  -> DevOps Control Plane
  -> External APIs GitLab/Argo/Kubernetes
  -> PostgreSQL
```

Boundary principali:

- ingresso HTTP;
- accesso database;
- accesso API esterne;
- accesso Secret;
- scrittura evidenze.

---

## 11.2 Secrets

Secret richiesti:

```text
GITLAB_TOKEN
ARGOCD_AUTH_TOKEN
DATABASE_URL or DATABASE_PASSWORD
KUBERNETES_SERVICE_ACCOUNT_TOKEN, se in cluster
```

Regole:

- token mai in Git;
- token mai in log;
- token mai in evidence;
- token montati da Secret Kubernetes/OpenShift;
- esempi template solo con placeholder.

---

## 11.3 RBAC Kubernetes/OpenShift

Il ServiceAccount del DevOps Control Plane deve avere permessi minimi.

Permessi iniziali indicativi:

- create/get/list/watch PipelineRun nel namespace Tekton previsto;
- get/list/watch TaskRun nel namespace Tekton previsto;
- get/list/watch Pod/log per evidenze Tekton, dove necessario;
- get/list Deployment/Pod/Service/ConfigMap/Route nei namespace applicativi autorizzati;
- nessun cluster-admin.

Il dettaglio sarà formalizzato in:

```text
docs/09-security-rbac.md
```

---

## 12. Observability architecture

## 12.1 Logging

Log strutturati con campi:

```text
timestamp
level
component
requestId
changeId
applicationName
operation
status
errorCode
```

---

## 12.2 Health/readiness

Endpoint:

```text
GET /healthz
GET /readyz
```

`/healthz` controlla processo vivo.  
`/readyz` controlla almeno PostgreSQL e configurazione minima.

---

## 12.3 Metrics future

Endpoint futuro:

```text
GET /metrics
```

Metriche candidate:

- change created;
- change completed;
- change failed;
- workflow duration;
- GitLab API errors;
- Argo CD API errors;
- Tekton validation failures.

---

## 13. Error architecture

## 13.1 Error model

Ogni errore applicativo dovrebbe avere:

```yaml
code: ARGO_OPERATION_IN_PROGRESS
technicalMessage: another operation is already in progress
userMessage: Argo CD ha già una operazione in corso sulla Application.
recoverable: true
suggestedAction: Attendere o terminare l’operazione bloccata.
```

---

## 13.2 Famiglie errori

```text
GITLAB_*
ARGO_*
KUBERNETES_*
TEKTON_*
DATABASE_*
VALIDATION_*
SECURITY_*
WORKFLOW_*
```

---

## 13.3 Errori noti MVP

```text
ARGO_OPERATION_IN_PROGRESS
ARGO_RESOURCE_NOT_PERMITTED
ARGO_SYNC_FAILED
GITLAB_FILE_NOT_FOUND
GITLAB_BRANCH_EXISTS
GITLAB_COMMIT_FAILED
TEKTON_PIPELINERUN_FAILED
KUBE_RESOURCE_NOT_FOUND
VALIDATION_INVALID_YAML
VALIDATION_SECRET_DETECTED
WORKFLOW_CONFLICT_ACTIVE_CHANGE
```

---

## 14. Decisioni architetturali da registrare in ADR

I seguenti ADR devono essere creati:

```text
docs/adr/ADR-0001-git-source-of-truth.md
docs/adr/ADR-0002-argocd-as-gitops-engine.md
docs/adr/ADR-0003-tekton-validation-engine.md
docs/adr/ADR-0004-postgresql-change-history.md
docs/adr/ADR-0005-api-first-before-web-ui.md
docs/adr/ADR-0006-adapter-based-architecture.md
docs/adr/ADR-0007-gitlab-api-as-git-provider.md
docs/adr/ADR-0008-kubernetes-api-for-tekton.md
```

---

## 15. Deployment diagram testuale

```text
OpenShift Cluster

Namespace: devops-control-plane

+-----------------------------------------------------+
| Deployment/devops-control-plane                     |
|                                                     |
|  Container: devops-control-plane                    |
|  Port: 8080                                         |
|  Env from ConfigMap                                 |
|  Secrets from Secret                                |
|  ServiceAccount: devops-control-plane               |
+-----------------------------------------------------+
        |
        v
+------------------------------+
| Service/devops-control-plane |
+------------------------------+
        |
        v
+-----------------------------+
| Route/devops-control-plane  |
+-----------------------------+

External dependencies:

- PostgreSQL
- Argo CD API
- GitLab API
- Kubernetes API
- Tekton CRDs
```

---

## 16. Sequence diagram: update-replicas

```text
User
  -> DevOps Control Plane API: POST /api/changes update-replicas
DevOps Control Plane
  -> PostgreSQL: create ChangeRequest
DevOps Control Plane
  -> Argo CD API: get Application metadata
DevOps Control Plane
  -> GitLab API: create branch
DevOps Control Plane
  -> GitLab API: read deployment.yaml
DevOps Control Plane
  -> DevOps Control Plane: update replicas in YAML
DevOps Control Plane
  -> GitLab API: commit or open MR
DevOps Control Plane
  -> Kubernetes API: create Tekton PipelineRun
Tekton
  -> GitLab: clone branch
Tekton
  -> Kubernetes API: dry-run validation
DevOps Control Plane
  -> Kubernetes API: watch PipelineRun
DevOps Control Plane
  -> Argo CD API: sync Application
DevOps Control Plane
  -> Argo CD API: wait Synced/Healthy
DevOps Control Plane
  -> Kubernetes API: collect runtime evidence
DevOps Control Plane
  -> PostgreSQL: save evidence and complete ChangeRequest
```

---

## 17. Sequence diagram: update ConfigMap value

```text
User
  -> DevOps Control Plane API: POST /api/changes update-configmap-values
DevOps Control Plane
  -> PostgreSQL: create ChangeRequest
DevOps Control Plane
  -> Argo CD API: get Application and AppProject metadata
DevOps Control Plane
  -> GitLab API: read configmap.yaml
DevOps Control Plane
  -> GitLab API: read kustomization.yaml
DevOps Control Plane
  -> DevOps Control Plane: validate ConfigMap is allowed
DevOps Control Plane
  -> GitLab API: create branch
DevOps Control Plane
  -> DevOps Control Plane: modify configmap.yaml
DevOps Control Plane
  -> GitLab API: commit or MR
DevOps Control Plane
  -> Kubernetes API: create Tekton PipelineRun
DevOps Control Plane
  -> Argo CD API: sync Application
DevOps Control Plane
  -> Kubernetes API: collect ConfigMap/Deployment/Pod evidence
DevOps Control Plane
  -> PostgreSQL: complete ChangeRequest
```

---

## 18. MVP architecture boundaries

## 18.1 In scope

- single backend Go service;
- PostgreSQL change store;
- GitLab API adapter;
- Argo CD API adapter;
- Kubernetes/Tekton adapter;
- basic HTML template support;
- one OpenShift deployment model;
- one or few configured environments.

---

## 18.2 Out of scope for MVP

- microservices architecture;
- event bus;
- full UI portal;
- multi-tenant RBAC enterprise;
- full policy engine;
- ITSM integration;
- multi-provider Git;
- automatic AppProject generation;
- automatic production promotion workflow.

---

## 19. Implementation priorities from architecture

Ordine raccomandato:

1. project skeleton Go;
2. config loader;
3. structured logger;
4. `/healthz` and `/readyz`;
5. PostgreSQL connection;
6. initial migrations;
7. domain models;
8. Application Service;
9. Argo CD Adapter read-only;
10. GitLab Adapter read-only;
11. ChangeRequest Service;
12. Workflow Engine base;
13. update-replicas workflow;
14. Tekton Adapter;
15. Argo CD sync integration;
16. Evidence Service;
17. minimal HTML pages.

---

## 20. Architecture validation checklist

Prima di iniziare implementazione estesa, verificare:

- il repository contiene documentazione base;
- le decisioni principali sono o saranno tracciate in ADR;
- il backend Go parte localmente;
- `/healthz` funziona;
- `/readyz` verifica PostgreSQL;
- i token sono configurati fuori Git;
- il data model iniziale è definito;
- gli adapter hanno interfacce chiare;
- il primo workflow target è `update-replicas`.

---

## 21. Messaggio chiave

L’architettura del DevOps Control Plane deve restare semplice, modulare e coerente con GitOps.

Il valore non è costruire un nuovo orchestratore complesso, ma creare un livello ordinato che renda più sicuro e tracciabile il flusso:

```text
GitLab change
  -> Tekton validation
  -> Argo CD sync
  -> OpenShift runtime
  -> PostgreSQL evidence/history
```

Ogni componente deve avere una responsabilità chiara e ogni workflow deve produrre evidenze utili per operatori, reviewer, platform engineer e auditor.
