# DevOps Control Plane - API Design

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
- `docs/10-data-model.md`  
- `docs/11-change-workflows.md`  
- `docs/12-evidence-model.md`  
**Stato:** Draft iniziale / API Design

---

## 1. Scopo del documento

Questo documento definisce il **design iniziale delle API HTTP/REST** del progetto **DevOps Control Plane**.

L’obiettivo è descrivere:

- convenzioni API;
- endpoint MVP;
- request e response JSON;
- codici HTTP;
- modello errori;
- mapping tra API, workflow e database;
- endpoint per Application Argo CD;
- endpoint per ChangeRequest;
- endpoint per GitLab evidence;
- endpoint per Tekton validation;
- endpoint per Argo CD sync;
- endpoint per runtime evidence;
- endpoint health/readiness;
- requisiti di sicurezza e sanitizzazione.

Questo documento guiderà l’implementazione del backend Go e deve rimanere allineato a:

- `docs/10-data-model.md`;
- `docs/11-change-workflows.md`;
- `docs/12-evidence-model.md`.

---

## 2. Principi API

## 2.1 API-first, UI later

Il progetto parte con un backend Go e API HTTP stabili.

La UI HTML con Bootstrap potrà usare queste stesse API, ma non deve guidare prematuramente le scelte architetturali.

Principio:

```text
Workflow e API stabili prima della UI completa.
```

---

## 2.2 JSON come formato primario

Le API applicative devono usare JSON.

Header attesi:

```http
Content-Type: application/json
Accept: application/json
```

---

## 2.3 Risposte consistenti

Tutte le risposte API dovrebbero avere una struttura prevedibile.

Per risposta singola:

```json
{
  "data": {},
  "meta": {},
  "error": null
}
```

Per lista:

```json
{
  "data": [],
  "meta": {
    "count": 0,
    "limit": 50,
    "offset": 0
  },
  "error": null
}
```

Per errore:

```json
{
  "data": null,
  "meta": {},
  "error": {
    "code": "ARGO_OPERATION_IN_PROGRESS",
    "message": "Argo CD ha già una operazione in corso sulla Application.",
    "technicalMessage": "another operation is already in progress",
    "suggestedAction": "Attendere il completamento dell'operazione oppure verificarla manualmente.",
    "recoverable": true
  }
}
```

---

## 2.4 Nessun secret nelle API

Le API non devono mai restituire:

- token GitLab;
- token Argo CD;
- password PostgreSQL;
- kubeconfig;
- Secret Kubernetes raw;
- header Authorization;
- valori sensibili non redatti.

---

## 3. Versioning API

Per MVP, usare prefisso:

```text
/api/v1
```

Esempio:

```text
GET /api/v1/applications
GET /api/v1/changes
```

Per semplicità durante il lab, può essere supportato anche alias senza versione:

```text
/api/applications
/api/changes
```

Raccomandazione:

```text
Implementare /api/v1 come canonical path.
```

---

## 4. Health e readiness

## 4.1 GET /healthz

### Scopo

Verifica che il processo sia vivo.

### Request

```http
GET /healthz
```

### Response 200

```json
{
  "status": "ok",
  "service": "devops-control-plane",
  "timestamp": "2026-06-25T15:00:00+02:00"
}
```

### Note

`/healthz` non deve dipendere da sistemi esterni complessi.

---

## 4.2 GET /readyz

### Scopo

Verifica che il servizio sia pronto a ricevere traffico.

Controlli minimi:

- configurazione caricata;
- connessione PostgreSQL disponibile.

### Request

```http
GET /readyz
```

### Response 200

```json
{
  "status": "ready",
  "checks": {
    "database": "ok",
    "configuration": "ok"
  }
}
```

### Response 503

```json
{
  "status": "not-ready",
  "checks": {
    "database": "failed",
    "configuration": "ok"
  }
}
```

---

## 5. Applications API

## 5.1 GET /api/v1/applications

### Scopo

Restituisce la lista delle Application Argo CD visibili al DevOps Control Plane.

### Query params MVP

```text
project
namespace
syncStatus
healthStatus
limit
offset
```

### Request

```http
GET /api/v1/applications?project=devops-ci-demo&limit=50&offset=0
```

### Response 200

```json
{
  "data": [
    {
      "name": "demo-go-color-app",
      "argocdNamespace": "openshift-gitops",
      "project": "devops-ci-demo",
      "targetNamespace": "devops-ci-demo",
      "repoUrl": "https://gitlab.example.local/group/demo-app-gitops.git",
      "targetRevision": "main",
      "path": "apps/demo-go-color-app",
      "syncStatus": "Synced",
      "healthStatus": "Healthy",
      "currentRevision": "b8e1d6b"
    }
  ],
  "meta": {
    "count": 1,
    "limit": 50,
    "offset": 0
  },
  "error": null
}
```

### Mapping

- Argo CD Adapter: `ListApplications`;
- tabella opzionale: `applications`.

---

## 5.2 GET /api/v1/applications/{name}

### Scopo

Restituisce dettaglio operativo di una Application.

### Request

```http
GET /api/v1/applications/demo-go-color-app
```

### Response 200

```json
{
  "data": {
    "name": "demo-go-color-app",
    "argocdNamespace": "openshift-gitops",
    "project": "devops-ci-demo",
    "syncPolicy": "Automated",
    "syncStatus": "Synced",
    "healthStatus": "Healthy",
    "currentRevision": "b8e1d6b",
    "source": {
      "repoUrl": "https://gitlab.example.local/group/demo-app-gitops.git",
      "targetRevision": "main",
      "path": "apps/demo-go-color-app"
    },
    "destination": {
      "server": "https://kubernetes.default.svc",
      "namespace": "devops-ci-demo"
    },
    "conditions": []
  },
  "meta": {},
  "error": null
}
```

### Errori

- `ARGO_APPLICATION_NOT_FOUND`;
- `ARGO_GET_FAILED`;
- `ARGO_AUTH_FAILED`.

---

## 5.3 GET /api/v1/applications/{name}/resources

### Scopo

Restituisce risorse gestite e orphaned resources.

### Request

```http
GET /api/v1/applications/demo-go-color-app/resources
```

### Response 200

```json
{
  "data": [
    {
      "group": "apps",
      "kind": "Deployment",
      "namespace": "devops-ci-demo",
      "name": "demo-go-color-app",
      "status": "Synced",
      "health": "Healthy",
      "orphaned": false
    },
    {
      "group": "image.openshift.io",
      "kind": "ImageStream",
      "namespace": "devops-ci-demo",
      "name": "demo-go-color-app",
      "status": null,
      "health": null,
      "orphaned": true
    }
  ],
  "meta": {
    "count": 2
  },
  "error": null
}
```

---

## 5.4 GET /api/v1/applications/{name}/history

### Scopo

Restituisce history Argo CD della Application.

### Response 200

```json
{
  "data": [
    {
      "id": 1,
      "revision": "4d6dfd4",
      "deployedAt": "2026-06-24T18:00:00+02:00",
      "sourcePath": "apps/demo-go-color-app"
    }
  ],
  "meta": {
    "count": 1
  },
  "error": null
}
```

---

## 6. Changes API

## 6.1 GET /api/v1/changes

### Scopo

Restituisce lista ChangeRequest.

### Query params

```text
applicationName
status
changeType
limit
offset
sort
```

### Request

```http
GET /api/v1/changes?applicationName=demo-go-color-app&limit=50
```

### Response 200

```json
{
  "data": [
    {
      "id": "1d871b5c-8a7a-4c11-9420-d2bc070b4a12",
      "changeNumber": "CHG-2026-0001",
      "applicationName": "demo-go-color-app",
      "changeType": "update-replicas",
      "status": "Completed",
      "requestedBy": "vmarzario",
      "createdAt": "2026-06-25T15:00:00+02:00",
      "completedAt": "2026-06-25T15:20:00+02:00"
    }
  ],
  "meta": {
    "count": 1,
    "limit": 50,
    "offset": 0
  },
  "error": null
}
```

---

## 6.2 POST /api/v1/changes

### Scopo

Crea una nuova ChangeRequest.

### Request update-replicas

```json
{
  "applicationName": "demo-go-color-app",
  "changeType": "update-replicas",
  "requestedBy": "vmarzario",
  "description": "Scale demo-go-color-app from 2 to 3 replicas",
  "payload": {
    "deploymentName": "demo-go-color-app",
    "replicas": 3
  }
}
```

### Response 201

```json
{
  "data": {
    "id": "1d871b5c-8a7a-4c11-9420-d2bc070b4a12",
    "changeNumber": "CHG-2026-0001",
    "applicationName": "demo-go-color-app",
    "changeType": "update-replicas",
    "status": "Created"
  },
  "meta": {},
  "error": null
}
```

### Validazioni

- `applicationName` obbligatorio;
- `changeType` obbligatorio;
- `payload` coerente con il change type;
- nessun change concorrente incompatibile.

### Errori

- `VALIDATION_INVALID_REQUEST`;
- `WORKFLOW_CONFLICT_ACTIVE_CHANGE`;
- `ARGO_APPLICATION_NOT_FOUND`.

---

## 6.3 GET /api/v1/changes/{id}

### Scopo

Restituisce dettaglio ChangeRequest con timeline ed evidenze sintetiche.

### Response 200

```json
{
  "data": {
    "id": "1d871b5c-8a7a-4c11-9420-d2bc070b4a12",
    "changeNumber": "CHG-2026-0001",
    "applicationName": "demo-go-color-app",
    "changeType": "update-replicas",
    "status": "Completed",
    "requestedBy": "vmarzario",
    "description": "Scale demo-go-color-app from 2 to 3 replicas",
    "git": {
      "provider": "gitlab",
      "sourceBranch": "change/CHG-2026-0001-update-replicas",
      "targetBranch": "main",
      "commitShortSha": "b8e1d6b",
      "mergeRequestUrl": "https://gitlab.example.local/group/repo/-/merge_requests/42"
    },
    "tekton": {
      "pipelineRunName": "validate-gitops-change-chg-2026-0001-abcde",
      "status": "Succeeded"
    },
    "argocd": {
      "syncStatus": "Synced",
      "healthStatus": "Healthy",
      "syncRevision": "b8e1d6b"
    },
    "runtime": {
      "namespace": "devops-ci-demo",
      "status": "Ready"
    },
    "createdAt": "2026-06-25T15:00:00+02:00",
    "completedAt": "2026-06-25T15:20:00+02:00"
  },
  "meta": {},
  "error": null
}
```

---

## 6.4 GET /api/v1/changes/{id}/events

### Scopo

Restituisce timeline completa della ChangeRequest.

### Response 200

```json
{
  "data": [
    {
      "eventType": "Created",
      "previousStatus": null,
      "newStatus": "Created",
      "message": "ChangeRequest created",
      "source": "workflow",
      "createdAt": "2026-06-25T15:00:00+02:00"
    },
    {
      "eventType": "ValidationSucceeded",
      "previousStatus": "ValidationRunning",
      "newStatus": "ValidationSucceeded",
      "message": "Tekton validation succeeded",
      "source": "tekton",
      "createdAt": "2026-06-25T15:05:00+02:00"
    }
  ],
  "meta": {
    "count": 2
  },
  "error": null
}
```

---

## 7. Change Workflow Action API

## 7.1 POST /api/v1/changes/{id}/run

### Scopo

Avvia workflow automatico end-to-end.

### Nota MVP

Endpoint utile, ma implementazione consigliata dopo gli step espliciti.

### Request

```json
{
  "mode": "auto",
  "dryRun": false
}
```

### Response 202

```json
{
  "data": {
    "changeNumber": "CHG-2026-0001",
    "status": "BranchCreated",
    "message": "Workflow started"
  },
  "meta": {},
  "error": null
}
```

---

## 7.2 POST /api/v1/changes/{id}/create-branch

### Scopo

Crea branch GitLab per la ChangeRequest.

### Response 200

```json
{
  "data": {
    "sourceBranch": "change/CHG-2026-0001-update-replicas",
    "status": "BranchCreated"
  },
  "meta": {},
  "error": null
}
```

---

## 7.3 POST /api/v1/changes/{id}/update-files

### Scopo

Applica modifiche ai file GitOps e crea commit.

### Response 200

```json
{
  "data": {
    "status": "CommitCreated",
    "commitSha": "b8e1d6b2fc88b41909f87d505739bc855de41516",
    "commitShortSha": "b8e1d6b",
    "filesChanged": [
      "apps/demo-go-color-app/deployment.yaml"
    ]
  },
  "meta": {},
  "error": null
}
```

---

## 7.4 POST /api/v1/changes/{id}/validate

### Scopo

Crea e monitora una PipelineRun Tekton di validazione.

### Request

```json
{
  "wait": true,
  "timeoutSeconds": 900
}
```

### Response 200

```json
{
  "data": {
    "status": "ValidationSucceeded",
    "pipelineRunName": "validate-gitops-change-chg-2026-0001-abcde",
    "tektonStatus": "Succeeded"
  },
  "meta": {},
  "error": null
}
```

### Response 202 se asincrono

```json
{
  "data": {
    "status": "ValidationRunning",
    "pipelineRunName": "validate-gitops-change-chg-2026-0001-abcde"
  },
  "meta": {},
  "error": null
}
```

---

## 7.5 POST /api/v1/changes/{id}/open-merge-request

### Scopo

Apre Merge Request GitLab.

### Response 200

```json
{
  "data": {
    "status": "MergeRequestOpened",
    "mergeRequestIid": 42,
    "mergeRequestUrl": "https://gitlab.example.local/group/repo/-/merge_requests/42"
  },
  "meta": {},
  "error": null
}
```

---

## 7.6 POST /api/v1/changes/{id}/sync

### Scopo

Lancia sync Argo CD per la Application collegata alla ChangeRequest.

### Request

```json
{
  "wait": true,
  "timeoutSeconds": 600
}
```

### Response 200

```json
{
  "data": {
    "status": "SyncSucceeded",
    "application": "demo-go-color-app",
    "syncStatus": "Synced",
    "healthStatus": "Healthy",
    "revision": "b8e1d6b"
  },
  "meta": {},
  "error": null
}
```

---

## 7.7 POST /api/v1/changes/{id}/collect-evidence

### Scopo

Raccoglie evidenze runtime e workflow.

### Response 200

```json
{
  "data": {
    "status": "EvidenceCollected",
    "evidencesCreated": 4
  },
  "meta": {},
  "error": null
}
```

---

## 8. Evidence API

## 8.1 GET /api/v1/changes/{id}/evidence

### Scopo

Restituisce tutte le evidenze associate a una ChangeRequest.

### Response 200

```json
{
  "data": [
    {
      "id": "c77ce0b7-e24f-4095-9c9f-7b4cf3eb42b0",
      "evidenceType": "gitlab",
      "name": "gitlab-commit-CHG-2026-0001",
      "summary": "GitLab commit b8e1d6b created on branch change/CHG-2026-0001-update-replicas",
      "sanitized": true,
      "createdAt": "2026-06-25T15:01:00+02:00"
    }
  ],
  "meta": {
    "count": 1
  },
  "error": null
}
```

---

## 8.2 GET /api/v1/changes/{id}/evidence/{type}

### Scopo

Restituisce evidenze filtrate per tipo.

Tipi supportati:

```text
gitlab
tekton
argocd
runtime
security
workflow-summary
error-summary
```

### Request

```http
GET /api/v1/changes/CHG-2026-0001/evidence/argocd
```

---

## 9. Validation API

## 9.1 GET /api/v1/changes/{id}/validation

### Scopo

Restituisce stato validazione Tekton della ChangeRequest.

### Response 200

```json
{
  "data": {
    "provider": "tekton",
    "namespace": "devops-control-plane",
    "pipelineRunName": "validate-gitops-change-chg-2026-0001-abcde",
    "status": "Succeeded",
    "startedAt": "2026-06-25T15:00:00+02:00",
    "completedAt": "2026-06-25T15:03:00+02:00"
  },
  "meta": {},
  "error": null
}
```

---

## 9.2 GET /api/v1/changes/{id}/validation/taskruns

### Scopo

Restituisce TaskRun collegate alla PipelineRun di validazione.

---

## 9.3 GET /api/v1/changes/{id}/validation/logs

### Scopo

Restituisce excerpt log sanitizzati.

### Regole

- limitare dimensione;
- redigere secret;
- non esporre token;
- indicare se il log è troncato.

---

## 10. Git metadata API

## 10.1 GET /api/v1/applications/{name}/git/commits

### Scopo

Restituisce commit Git rilevanti per Application/path.

### Query params

```text
limit
ref
path
```

---

## 10.2 GET /api/v1/changes/{id}/git

### Scopo

Restituisce Git metadata della ChangeRequest.

### Response 200

```json
{
  "data": {
    "provider": "gitlab",
    "projectId": "12345",
    "repoUrl": "https://gitlab.example.local/group/demo-app-gitops.git",
    "sourceBranch": "change/CHG-2026-0001-update-replicas",
    "targetBranch": "main",
    "commitSha": "b8e1d6b2fc88b41909f87d505739bc855de41516",
    "mergeRequestIid": 42,
    "mergeRequestUrl": "https://gitlab.example.local/group/repo/-/merge_requests/42"
  },
  "meta": {},
  "error": null
}
```

---

## 11. Sync API

## 11.1 POST /api/v1/applications/{name}/sync

### Scopo

Lancia sync Argo CD diretta sulla Application.

### Nota

Questo endpoint va usato con cautela: per change completi è preferibile usare `/changes/{id}/sync`, in modo da mantenere correlazione con ChangeRequest.

### Request

```json
{
  "wait": true,
  "timeoutSeconds": 600
}
```

---

## 11.2 GET /api/v1/applications/{name}/operation

### Scopo

Restituisce operation state Argo CD corrente.

---

## 12. Runtime API

## 12.1 GET /api/v1/applications/{name}/runtime

### Scopo

Restituisce stato runtime OpenShift/Kubernetes normalizzato.

### Response 200

```json
{
  "data": {
    "namespace": "devops-ci-demo",
    "deployment": {
      "name": "demo-go-color-app",
      "desiredReplicas": 3,
      "readyReplicas": 3,
      "availableReplicas": 3
    },
    "pods": [
      {
        "name": "demo-go-color-app-abcde",
        "phase": "Running",
        "ready": true,
        "restartCount": 0
      }
    ]
  },
  "meta": {},
  "error": null
}
```

---

## 13. Error model API

## 13.1 HTTP status mapping

| HTTP | Uso |
|---|---|
| 200 | Operazione riuscita |
| 201 | Risorsa creata |
| 202 | Operazione asincrona avviata |
| 400 | Request non valida |
| 401 | Non autenticato |
| 403 | Non autorizzato |
| 404 | Risorsa non trovata |
| 409 | Conflitto workflow o branch |
| 422 | Validazione fallita |
| 500 | Errore interno |
| 502 | Errore gateway verso API esterna |
| 503 | Servizio non pronto |
| 504 | Timeout verso API esterna |

---

## 13.2 Error response standard

```json
{
  "data": null,
  "meta": {
    "requestId": "req-123"
  },
  "error": {
    "code": "GITLAB_FILE_NOT_FOUND",
    "message": "Il file GitOps richiesto non esiste nel repository o nel branch indicato.",
    "technicalMessage": "404 File Not Found",
    "suggestedAction": "Verificare projectId, branch e path del file.",
    "recoverable": true
  }
}
```

---

## 13.3 Error codes principali

```text
VALIDATION_INVALID_REQUEST
WORKFLOW_CONFLICT_ACTIVE_CHANGE
GITLAB_FILE_NOT_FOUND
GITLAB_BRANCH_EXISTS
GITLAB_COMMIT_FAILED
ARGO_APPLICATION_NOT_FOUND
ARGO_OPERATION_IN_PROGRESS
ARGO_RESOURCE_NOT_PERMITTED
ARGO_SYNC_FAILED
TEKTON_PIPELINERUN_FAILED
TEKTON_PIPELINERUN_TIMEOUT
KUBERNETES_RUNTIME_NOT_READY
SECURITY_SECRET_DETECTED
DATABASE_ERROR
```

---

## 14. Sicurezza API

## 14.1 MVP

Nel primo MVP/lab:

- non esporre API pubblicamente;
- usare rete controllata;
- non loggare token;
- non restituire secret;
- validare input;
- usare timeout;
- usare RBAC OpenShift per il backend.

---

## 14.2 Futuro

Evoluzioni:

- autenticazione OpenShift OAuth;
- autorizzazione per ruolo;
- mapping utente reale;
- audit user action;
- rate limiting;
- CSRF protection per UI HTML;
- session management.

---

## 15. Pagination e sorting

## 15.1 Pagination

Parametri standard:

```text
limit
offset
```

Default consigliati:

```text
limit=50
offset=0
```

Massimo consigliato:

```text
limit=200
```

---

## 15.2 Sorting

Parametro:

```text
sort
```

Esempi:

```text
sort=createdAt:desc
sort=name:asc
```

---

## 16. Request ID e correlation

Ogni richiesta dovrebbe avere un `requestId`.

Se il client non lo fornisce, il backend lo genera.

Header opzionale:

```http
X-Request-ID: req-123
```

Il `requestId` deve comparire in:

- log;
- error response;
- eventuali eventi diagnostici.

---

## 17. API implementation order MVP

Ordine consigliato:

1. `GET /healthz`;
2. `GET /readyz`;
3. `GET /api/v1/applications`;
4. `GET /api/v1/applications/{name}`;
5. `GET /api/v1/applications/{name}/resources`;
6. `GET /api/v1/changes`;
7. `POST /api/v1/changes`;
8. `GET /api/v1/changes/{id}`;
9. `GET /api/v1/changes/{id}/events`;
10. `POST /api/v1/changes/{id}/create-branch`;
11. `POST /api/v1/changes/{id}/update-files`;
12. `POST /api/v1/changes/{id}/validate`;
13. `POST /api/v1/changes/{id}/sync`;
14. `POST /api/v1/changes/{id}/collect-evidence`;
15. `GET /api/v1/changes/{id}/evidence`.

---

## 18. Mapping package Go

Package suggeriti:

```text
internal/api/router.go
internal/api/health_handlers.go
internal/api/application_handlers.go
internal/api/change_handlers.go
internal/api/evidence_handlers.go
internal/api/validation_handlers.go
internal/api/runtime_handlers.go
internal/api/errors.go
internal/api/response.go
```

---

## 19. Checklist API MVP

Le API MVP sono considerate pronte quando:

- health/readiness funzionano;
- error response è consistente;
- lista Application funziona;
- creazione ChangeRequest funziona;
- dettaglio ChangeRequest mostra Git/Tekton/Argo/evidence;
- workflow step-by-step è invocabile;
- evidenze sono consultabili;
- input invalidi producono 400/422;
- conflitti producono 409;
- token e secret non sono mai restituiti;
- ogni chiamata ha logging con requestId.

---

## 20. Relazione con altri documenti

Questo documento alimenta:

- `internal/api/`;
- `internal/app/`;
- `internal/workflow/`;
- `docs/11-change-workflows.md`;
- `docs/12-evidence-model.md`;
- `docs/10-data-model.md`;
- future OpenAPI specification.

---

## 21. Messaggio chiave

Le API del DevOps Control Plane devono rendere il workflow GitOps chiaro e controllabile.

Ogni API deve aiutare a vedere o avanzare uno step:

```text
Application discovery
Change creation
Git change
Tekton validation
Argo CD sync
Runtime validation
Evidence review
```

Il design deve restare semplice, coerente, sicuro e facilmente spiegabile anche a chi sta imparando GitOps, Tekton e Argo CD.
