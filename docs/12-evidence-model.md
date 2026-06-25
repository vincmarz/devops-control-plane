# DevOps Control Plane - Evidence Model

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
**Stato:** Draft iniziale / Evidence Model

---

## 1. Scopo del documento

Questo documento definisce il **modello delle evidenze** del progetto **DevOps Control Plane**.

L’obiettivo è descrivere:

- cosa si intende per evidenza;
- quali evidenze devono essere raccolte;
- quando devono essere raccolte;
- come devono essere normalizzate;
- come devono essere salvate in PostgreSQL;
- quali dati non devono mai essere salvati;
- come correlare evidenze GitLab, Tekton, Argo CD e Kubernetes/OpenShift;
- come rendere le evidenze utili per audit, troubleshooting e formazione dei newbie.

Il modello delle evidenze è centrale per il valore del DevOps Control Plane: non basta eseguire un change, bisogna poter ricostruire cosa è successo, quando, con quali strumenti e con quale esito.

---

## 2. Definizione di evidenza

Nel DevOps Control Plane, una **evidenza** è un’informazione tecnica, normalizzata e sanitizzata, associata a una ChangeRequest.

Una evidenza deve aiutare a rispondere a domande come:

```text
Quale branch GitLab è stato creato?
Quale commit ha rappresentato il change?
Quale PipelineRun Tekton ha validato il change?
Quale revisione Argo CD è stata sincronizzata?
L’applicazione è Synced e Healthy?
Il Deployment ha raggiunto le repliche attese?
I Pod sono Running e Ready?
Il change ha generato warning o errori?
```

---

## 3. Principi del modello evidence

## 3.1 Evidence first

Ogni workflow deve produrre evidenze.

Regola:

```text
Se un workflow modifica Git, valida con Tekton, sincronizza con Argo CD o legge runtime OpenShift, deve produrre evidenza.
```

---

## 3.2 Evidence sanitizzata

Le evidenze devono essere utili ma sicure.

Non devono contenere:

- token;
- password;
- secret values;
- kubeconfig;
- header Authorization;
- private key;
- `.dockerconfigjson`;
- output raw di Secret Kubernetes;
- log completi non filtrati.

---

## 3.3 Evidence leggibile

Ogni evidenza deve avere:

- tipo;
- nome;
- summary leggibile;
- payload tecnico normalizzato;
- timestamp;
- riferimento esterno, se disponibile.

---

## 3.4 Evidence correlabile

Ogni evidenza deve essere associata a:

```text
change_request_id
```

Quando possibile, deve includere anche:

- `changeNumber`;
- `applicationName`;
- `provider`;
- `namespace`;
- `resourceName`;
- `revision`;
- `status`.

---

## 4. Tipi di evidenza MVP

Tipi principali:

```text
gitlab
tekton
argocd
kubernetes-runtime
health-check
diff-summary
workflow-summary
security-validation
error-summary
```

---

## 5. Tabella evidences

Il data model principale è definito in `docs/10-data-model.md`.

Tabella MVP:

```sql
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
```

---

## 6. Evidence GitLab

## 6.1 Scopo

Dimostrare quale modifica Git ha rappresentato il change.

---

## 6.2 Quando raccoglierla

- dopo creazione branch;
- dopo commit;
- dopo apertura MR;
- dopo merge MR;
- in caso di errore GitLab.

---

## 6.3 Campi minimi

```json
{
  "provider": "gitlab",
  "projectId": "12345",
  "repoUrl": "https://gitlab.example.local/group/demo-app-gitops.git",
  "sourceBranch": "change/CHG-2026-0001-update-replicas",
  "targetBranch": "main",
  "commitSha": "b8e1d6b2fc88b41909f87d505739bc855de41516",
  "commitShortSha": "b8e1d6b",
  "mergeRequestIid": 42,
  "mergeRequestUrl": "https://gitlab.example.local/group/repo/-/merge_requests/42",
  "mergeRequestState": "opened",
  "filesChanged": [
    "apps/demo-go-color-app/deployment.yaml"
  ]
}
```

---

## 6.4 Summary suggerito

```text
GitLab change created on branch change/CHG-2026-0001-update-replicas with commit b8e1d6b.
```

---

## 7. Evidence diff-summary

## 7.1 Scopo

Mostrare cosa è cambiato nei manifest GitOps senza costringere l’utente a leggere il diff completo.

---

## 7.2 Quando raccoglierla

- dopo modifica file;
- prima del commit o subito dopo il commit;
- prima della validazione Tekton.

---

## 7.3 Payload esempio update-replicas

```json
{
  "files": [
    {
      "path": "apps/demo-go-color-app/deployment.yaml",
      "changeType": "update",
      "summary": "spec.replicas changed from 2 to 3",
      "fields": [
        {
          "field": "spec.replicas",
          "oldValue": 2,
          "newValue": 3
        }
      ]
    }
  ]
}
```

---

## 7.4 Payload esempio ConfigMap

```json
{
  "files": [
    {
      "path": "apps/demo-go-color-app/configmap.yaml",
      "changeType": "update",
      "summary": "ConfigMap values updated",
      "fields": [
        {
          "field": "data.APP_VERSION",
          "oldValue": "v2-green",
          "newValue": "v3-green"
        },
        {
          "field": "data.PAGE_COLOR",
          "oldValue": "#1E90FF",
          "newValue": "#28A745"
        }
      ]
    }
  ]
}
```

### Nota sicurezza

Se il campo modificato sembra contenere secret, il valore non deve essere salvato integralmente.

Esempio:

```json
{
  "field": "data.PASSWORD",
  "oldValue": "***REDACTED***",
  "newValue": "***REDACTED***",
  "redacted": true
}
```

---

## 8. Evidence Tekton

## 8.1 Scopo

Dimostrare che il change è stato validato tecnicamente prima della sync Argo CD.

---

## 8.2 Quando raccoglierla

- alla creazione della PipelineRun;
- durante il polling, se utile;
- a fine PipelineRun;
- in caso di fallimento;
- dopo raccolta TaskRun/log.

---

## 8.3 Payload successo

```json
{
  "provider": "tekton",
  "namespace": "devops-control-plane",
  "pipelineName": "validate-gitops-change",
  "pipelineRunName": "validate-gitops-change-chg-2026-0001-abcde",
  "status": "Succeeded",
  "reason": "Succeeded",
  "startTime": "2026-06-25T15:00:00+02:00",
  "completionTime": "2026-06-25T15:03:00+02:00",
  "taskRuns": [
    {
      "taskName": "clone-repository",
      "status": "Succeeded"
    },
    {
      "taskName": "yaml-validate",
      "status": "Succeeded"
    },
    {
      "taskName": "kustomize-build",
      "status": "Succeeded"
    },
    {
      "taskName": "server-side-dry-run",
      "status": "Succeeded"
    },
    {
      "taskName": "anti-secret-check",
      "status": "Succeeded"
    }
  ]
}
```

---

## 8.4 Payload fallimento

```json
{
  "provider": "tekton",
  "namespace": "devops-control-plane",
  "pipelineRunName": "validate-gitops-change-chg-2026-0001-abcde",
  "status": "Failed",
  "reason": "Failed",
  "failedTask": "server-side-dry-run",
  "message": "Deployment.apps demo-go-color-app is invalid: valueFrom is incomplete",
  "errorCode": "TEKTON_PIPELINERUN_FAILED"
}
```

---

## 8.5 Log evidence

I log Tekton devono essere salvati come excerpt sanitizzato.

Esempio:

```json
{
  "taskRunName": "validate-gitops-change-chg-2026-0001-yaml-validate",
  "container": "step-yaml-validate",
  "logExcerpt": "YAML validation completed successfully",
  "truncated": false,
  "sanitized": true
}
```

---

## 9. Evidence Argo CD

## 9.1 Scopo

Dimostrare quale stato Argo CD è stato osservato e quale revision è stata sincronizzata.

---

## 9.2 Quando raccoglierla

- prima della sync, come baseline;
- dopo richiesta sync;
- al completamento sync;
- al timeout o fallimento;
- quando vengono rilevati warning o orphaned resources.

---

## 9.3 Payload successo

```json
{
  "provider": "argocd",
  "application": "demo-go-color-app",
  "project": "devops-ci-demo",
  "syncStatus": "Synced",
  "healthStatus": "Healthy",
  "revision": "b8e1d6b",
  "operationPhase": "Succeeded",
  "conditions": []
}
```

---

## 9.4 Payload warning orphaned resources

```json
{
  "provider": "argocd",
  "application": "demo-go-color-app",
  "syncStatus": "Synced",
  "healthStatus": "Healthy",
  "conditions": [
    {
      "type": "OrphanedResourceWarning",
      "message": "Application has orphaned resources"
    }
  ],
  "orphanedResources": [
    {
      "group": "image.openshift.io",
      "kind": "ImageStream",
      "namespace": "devops-ci-demo",
      "name": "demo-go-color-app"
    }
  ]
}
```

---

## 9.5 Payload errore AppProject

```json
{
  "provider": "argocd",
  "application": "demo-go-color-app",
  "syncStatus": "OutOfSync",
  "healthStatus": "Healthy",
  "operationPhase": "Failed",
  "errorCode": "ARGO_RESOURCE_NOT_PERMITTED",
  "message": "resource :ConfigMap is not permitted in project devops-ci-demo",
  "suggestedAction": "Aggiungere group: \"\" kind: ConfigMap alla namespaceResourceWhitelist dell'AppProject."
}
```

---

## 10. Evidence Kubernetes/OpenShift runtime

## 10.1 Scopo

Dimostrare lo stato runtime osservato dopo sync o durante un controllo evidence-only.

---

## 10.2 Risorse target

- Deployment;
- ReplicaSet;
- Pod;
- Service;
- Route;
- ConfigMap;
- Event, se utile;
- health endpoint applicativo, se disponibile.

---

## 10.3 Payload Deployment

```json
{
  "provider": "kubernetes",
  "namespace": "devops-ci-demo",
  "kind": "Deployment",
  "name": "demo-go-color-app",
  "desiredReplicas": 3,
  "readyReplicas": 3,
  "availableReplicas": 3,
  "updatedReplicas": 3,
  "observedGenerationMatches": true
}
```

---

## 10.4 Payload Pod summary

```json
{
  "provider": "kubernetes",
  "namespace": "devops-ci-demo",
  "kind": "PodSummary",
  "selector": "app=demo-go-color-app",
  "pods": [
    {
      "name": "demo-go-color-app-abcde",
      "phase": "Running",
      "ready": true,
      "restartCount": 0
    }
  ]
}
```

---

## 10.5 Payload ConfigMap runtime

```json
{
  "provider": "kubernetes",
  "namespace": "devops-ci-demo",
  "kind": "ConfigMap",
  "name": "demo-go-color-app-config",
  "keys": [
    "APP_VERSION",
    "PAGE_COLOR"
  ],
  "values": {
    "APP_VERSION": "v3-green",
    "PAGE_COLOR": "#28A745"
  }
}
```

### Nota sicurezza

Per ConfigMap applicative non sensibili è possibile salvare valori. Tuttavia, se una chiave appare sensibile, il valore deve essere redatto.

---

## 11. Evidence health-check

## 11.1 Scopo

Dimostrare che l’applicazione risponde a un endpoint applicativo.

---

## 11.2 Endpoint preferiti

```text
/healthz
/readyz
/
```

---

## 11.3 Payload esempio

```json
{
  "provider": "http",
  "url": "https://demo-go-color-app-devops-ci-demo.apps.example.local/healthz",
  "statusCode": 200,
  "responseExcerpt": "ok",
  "durationMs": 42,
  "timestamp": "2026-06-25T15:10:00+02:00"
}
```

### Regole

- non salvare pagine contenenti dati sensibili;
- limitare response excerpt;
- timeout breve;
- registrare errore se endpoint non raggiungibile.

---

## 12. Evidence security-validation

## 12.1 Scopo

Dimostrare che sono stati eseguiti controlli di sicurezza minimi sul change.

---

## 12.2 Payload successo

```json
{
  "provider": "devops-control-plane",
  "check": "anti-secret-check",
  "status": "Passed",
  "filesScanned": [
    "apps/demo-go-color-app/deployment.yaml",
    "apps/demo-go-color-app/configmap.yaml"
  ],
  "matches": []
}
```

---

## 12.3 Payload fallimento

```json
{
  "provider": "devops-control-plane",
  "check": "anti-secret-check",
  "status": "Failed",
  "errorCode": "VALIDATION_SECRET_DETECTED",
  "matches": [
    {
      "file": "apps/demo-go-color-app/configmap.yaml",
      "pattern": "password",
      "value": "***REDACTED***"
    }
  ]
}
```

---

## 13. Evidence workflow-summary

## 13.1 Scopo

Fornire una sintesi finale umana della ChangeRequest.

---

## 13.2 Payload esempio Completed

```json
{
  "changeNumber": "CHG-2026-0001",
  "application": "demo-go-color-app",
  "changeType": "update-replicas",
  "finalStatus": "Completed",
  "summary": "Replicas updated from 2 to 3 and application reached Synced/Healthy state.",
  "git": {
    "branch": "change/CHG-2026-0001-update-replicas",
    "commitShortSha": "b8e1d6b"
  },
  "tekton": {
    "pipelineRun": "validate-gitops-change-chg-2026-0001-abcde",
    "status": "Succeeded"
  },
  "argocd": {
    "syncStatus": "Synced",
    "healthStatus": "Healthy"
  },
  "runtime": {
    "readyReplicas": 3
  }
}
```

---

## 13.3 Payload esempio Failed

```json
{
  "changeNumber": "CHG-2026-0001",
  "application": "demo-go-color-app",
  "changeType": "update-configmap-values",
  "finalStatus": "Failed",
  "failedPhase": "Argo CD sync",
  "errorCode": "ARGO_RESOURCE_NOT_PERMITTED",
  "userMessage": "La ConfigMap non è consentita dall'AppProject devops-ci-demo.",
  "suggestedAction": "Aggiornare namespaceResourceWhitelist dell'AppProject."
}
```

---

## 14. Error-summary evidence

## 14.1 Scopo

Normalizzare errori complessi in formato leggibile e auditabile.

---

## 14.2 Payload standard

```json
{
  "code": "TEKTON_PIPELINERUN_FAILED",
  "phase": "Validation",
  "technicalMessage": "PipelineRun validate-gitops-change-chg-2026-0001 failed",
  "userMessage": "La validazione Tekton del change è fallita.",
  "suggestedAction": "Consultare TaskRun e log associati alla ChangeRequest.",
  "recoverable": true
}
```

---

## 15. Naming evidenze

## 15.1 Nome evidenza

Formato consigliato:

```text
<provider>-<purpose>-<change-number>
```

Esempi:

```text
gitlab-commit-CHG-2026-0001
tekton-validation-CHG-2026-0001
argocd-sync-CHG-2026-0001
runtime-deployment-CHG-2026-0001
workflow-summary-CHG-2026-0001
```

---

## 16. Sanitizzazione

## 16.1 Funzione di sanitizzazione

Ogni adapter o evidence builder deve applicare una funzione di sanitizzazione prima del salvataggio.

Pattern minimi:

```text
token
password
secret
auth
PRIVATE KEY
BEGIN RSA
ghp_
github_pat_
AKIA
ASIA
.dockerconfigjson
Authorization:
Bearer
```

---

## 16.2 Redaction

Formato redaction:

```text
***REDACTED***
```

Esempio:

```json
{
  "key": "DATABASE_PASSWORD",
  "value": "***REDACTED***"
}
```

---

## 17. Dimensionamento e retention

## 17.1 Limiti consigliati MVP

```text
summary: breve, umano
content: excerpt limitato
payload: JSONB normalizzato
logExcerpt: massimo configurabile, ad esempio 10 KiB
```

---

## 17.2 Dati grandi

Per dati grandi:

- salvare summary;
- salvare external reference;
- evitare dump completo in PostgreSQL;
- valutare storage esterno o Tekton Results in futuro.

---

## 18. Evidence lifecycle

## 18.1 Creazione

Una evidenza viene creata da:

- Workflow Engine;
- Evidence Service;
- Adapter GitLab;
- Adapter Tekton;
- Adapter Argo CD;
- Adapter Kubernetes.

---

## 18.2 Aggiornamento

Per MVP, preferire evidenze append-only.

Regola:

```text
Non sovrascrivere evidenze storiche, aggiungere una nuova evidenza.
```

---

## 18.3 Cancellazione

Per MVP, non prevedere cancellazione applicativa ordinaria.

Future policy:

- retention temporale;
- export audit;
- pruning evidenze non critiche;
- conservazione evidenze finali.

---

## 19. Query evidence utili

## 19.1 Evidenze per ChangeRequest

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

## 19.2 Evidenze per tipo

```sql
SELECT
    name,
    summary,
    created_at
FROM evidences
WHERE change_request_id = $1
  AND evidence_type = $2
ORDER BY created_at ASC;
```

---

## 19.3 Error summary

```sql
SELECT
    name,
    payload,
    created_at
FROM evidences
WHERE change_request_id = $1
  AND evidence_type = 'error-summary'
ORDER BY created_at DESC;
```

---

## 20. API mapping evidence

Endpoint previsti:

```text
GET /api/changes/{id}/evidence
GET /api/changes/{id}/evidence/gitlab
GET /api/changes/{id}/evidence/tekton
GET /api/changes/{id}/evidence/argocd
GET /api/changes/{id}/evidence/runtime
GET /api/changes/{id}/evidence/security
```

---

## 21. Checklist evidence MVP

Una ChangeRequest completata deve avere almeno:

- `workflow-summary`;
- `gitlab`;
- `diff-summary`;
- `tekton`;
- `argocd`;
- `kubernetes-runtime`.

Una ChangeRequest fallita deve avere almeno:

- `workflow-summary`;
- `error-summary`;
- evidenza della fase fallita;
- stato ultimo osservato;
- suggested action.

---

## 22. Relazione con altri documenti

Questo documento alimenta:

- `internal/app/evidence_service.go`;
- `internal/domain/evidence.go`;
- `internal/database/repositories.go`;
- `docs/13-api-design.md`;
- `docs/11-change-workflows.md`;
- `migrations/000001_init.up.sql`;
- future policy di retention e audit export.

---

## 23. Messaggio chiave

Le evidenze sono ciò che trasforma un’automazione tecnica in un processo governabile.

Il DevOps Control Plane deve poter dire:

```text
Ho modificato Git.
Ho validato con Tekton.
Ho sincronizzato con Argo CD.
Ho verificato OpenShift.
Ho salvato una prova sicura e leggibile di ogni passaggio.
```

Senza evidenze, il sistema sarebbe solo un wrapper di comandi.

Con evidenze, diventa un control plane DevOps auditabile e utile anche per formazione, troubleshooting e governance.
