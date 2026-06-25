# DevOps Control Plane - Change Workflows

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
**Stato:** Draft iniziale / Change Workflows

---

## 1. Scopo del documento

Questo documento descrive i **workflow operativi** del progetto **DevOps Control Plane**.

L’obiettivo è definire, in modo ripetibile e implementabile:

- il ciclo di vita di una ChangeRequest;
- gli stati del workflow;
- le transizioni ammesse;
- i workflow MVP;
- gli errori gestiti;
- le evidenze da raccogliere;
- i punti di integrazione con GitLab, Tekton, Argo CD e OpenShift;
- le policy di sicurezza e governance applicate durante il change;
- i criteri di completamento o fallimento.

Questo documento è pensato sia per guidare l’implementazione backend sia per diventare una guida didattica per operatori e colleghi meno esperti.

---

## 2. Principio fondamentale

Il DevOps Control Plane non modifica direttamente il runtime come stato finale permanente.

Ogni change applicativo deve passare da Git.

```text
Richiesta change
  -> GitLab branch / commit / MR
  -> Tekton validation
  -> Argo CD sync
  -> OpenShift runtime validation
  -> Evidence
  -> Change history
```

Comandi come:

```bash
oc edit deployment
oc patch deployment
oc set env deployment
oc scale deployment
```

non devono essere il workflow permanente del prodotto.

Possono essere usati solo per troubleshooting manuale o emergenze operative esplicitamente documentate.

---

## 3. Tipi di workflow MVP

I workflow MVP iniziali sono:

```text
WF-001 update-replicas
WF-002 update-app-version
WF-003 update-page-color
WF-004 update-configmap-values
WF-005 validate-only
WF-006 sync-only
WF-007 collect-evidence-only
```

Priorità implementativa consigliata:

1. `update-replicas`;
2. `validate-only`;
3. `sync-only`;
4. `collect-evidence-only`;
5. `update-app-version`;
6. `update-page-color`;
7. `update-configmap-values`.

---

## 4. Modello stati ChangeRequest

## 4.1 Stati principali

```text
Draft
Created
BranchCreated
FilesRead
FilesUpdated
CommitCreated
MergeRequestOpened
WaitingForMerge
Merged
ValidationRequested
ValidationRunning
ValidationSucceeded
ValidationFailed
SyncRequested
SyncRunning
SyncSucceeded
SyncFailed
EvidenceCollectionRunning
EvidenceCollected
Completed
Failed
Cancelled
```

---

## 4.2 Stati finali

Stati finali:

```text
Completed
Failed
Cancelled
```

Una ChangeRequest in stato finale non deve più avanzare automaticamente.

Eventuali retry devono creare:

- una nuova ChangeRequest; oppure
- un evento di retry esplicitamente tracciato, se il modello lo permetterà in futuro.

---

## 4.3 Stati non finali critici

Stati da monitorare con timeout o verifica periodica:

```text
ValidationRunning
WaitingForMerge
SyncRunning
EvidenceCollectionRunning
```

Regola:

```text
Nessuna ChangeRequest deve restare indefinitamente in uno stato intermedio senza timeout o azione richiesta.
```

---

## 5. Transizioni principali

## 5.1 Transizioni nominali end-to-end

```text
Created
  -> BranchCreated
  -> FilesRead
  -> FilesUpdated
  -> CommitCreated
  -> ValidationRequested
  -> ValidationRunning
  -> ValidationSucceeded
  -> MergeRequestOpened
  -> WaitingForMerge
  -> Merged
  -> SyncRequested
  -> SyncRunning
  -> SyncSucceeded
  -> EvidenceCollectionRunning
  -> EvidenceCollected
  -> Completed
```

---

## 5.2 Transizioni con commit diretto

Nel lab iniziale può essere supportato un flusso semplificato:

```text
Created
  -> BranchCreated
  -> FilesUpdated
  -> CommitCreated
  -> ValidationRequested
  -> ValidationSucceeded
  -> SyncRequested
  -> SyncSucceeded
  -> EvidenceCollected
  -> Completed
```

Nota:

Questo flusso è più veloce, ma meno adatto alla governance enterprise perché salta la review tramite MR.

---

## 5.3 Transizioni di fallimento

Qualunque stato operativo può transitare a `Failed` se l’errore è non recuperabile.

Esempi:

```text
BranchCreated -> Failed
ValidationRunning -> ValidationFailed
SyncRunning -> SyncFailed
EvidenceCollectionRunning -> Failed oppure Completed con warning
```

Regola:

- `ValidationFailed` non deve procedere automaticamente a sync;
- `SyncFailed` non deve diventare `Completed`;
- failure deve sempre generare `change_events` con `error_code` e messaggio leggibile.

---

## 6. Workflow WF-001 - update-replicas

## 6.1 Scopo

Modificare il numero di repliche di un Deployment tramite GitOps.

Il workflow deve modificare `spec.replicas` nel file GitOps, non scalare direttamente il Deployment runtime.

---

## 6.2 Input

```yaml
applicationName: demo-go-color-app
changeType: update-replicas
requestedBy: vmarzario
description: "Scale demo-go-color-app from 2 to 3 replicas"
payload:
  deploymentName: demo-go-color-app
  replicas: 3
```

---

## 6.3 Validazioni input

- `applicationName` obbligatorio;
- `changeType=update-replicas`;
- `replicas` deve essere intero;
- `replicas >= 0`;
- `deploymentName` opzionale se deducibile dalla Application;
- Application deve essere nota o recuperabile da Argo CD;
- repository GitLab deve essere configurato;
- path GitOps deve essere noto.

---

## 6.4 Flusso dettagliato

```text
1. Create ChangeRequest
2. Resolve Application metadata da Argo CD
3. Resolve repository GitLab e path GitOps
4. Check conflitti con change attivi
5. Create branch GitLab
6. Read deployment.yaml
7. Parse YAML
8. Find Deployment target
9. Update spec.replicas
10. Generate diff summary
11. Anti-secret check sui file modificati
12. Commit files su branch
13. Create Tekton PipelineRun validation
14. Wait validation
15. Open Merge Request oppure commit diretto secondo modalità
16. Wait merge oppure conferma merge
17. Trigger Argo CD sync
18. Wait Synced/Healthy
19. Collect runtime evidence
20. Mark Completed
```

---

## 6.5 Eventi attesi

```text
Created
GitRepositoryResolved
GitBranchCreated
GitFileRead
GitFilesUpdated
GitCommitCreated
ValidationRequested
ValidationRunning
ValidationSucceeded
GitMergeRequestOpened
GitMergeRequestMerged
SyncRequested
SyncRunning
SyncSucceeded
EvidenceCollected
Completed
```

---

## 6.6 Evidence richieste

- branch GitLab;
- commit SHA;
- MR URL, se presente;
- diff summary;
- Tekton PipelineRun;
- TaskRun summary;
- Argo CD sync revision;
- Deployment status;
- Pod status;
- numero repliche desiderato/osservato.

Esempio evidence runtime:

```json
{
  "deployment": "demo-go-color-app",
  "desiredReplicas": 3,
  "readyReplicas": 3,
  "availableReplicas": 3
}
```

---

## 6.7 Criteri di successo

Il workflow è `Completed` quando:

- file GitOps aggiornato;
- commit creato;
- validazione Tekton riuscita;
- MR mergiata o commit approvato;
- Argo CD `Synced`;
- Argo CD `Healthy`;
- Deployment runtime ha repliche attese;
- evidenze salvate.

---

## 6.8 Errori gestiti

| Errore | Stato | Codice |
|---|---|---|
| Application non trovata | Failed | ARGO_APPLICATION_NOT_FOUND |
| File deployment non trovato | Failed | GITLAB_FILE_NOT_FOUND |
| Branch già esistente | Failed | GITLAB_BRANCH_EXISTS |
| YAML non valido | ValidationFailed | VALIDATION_INVALID_YAML |
| Tekton fallisce | ValidationFailed | TEKTON_PIPELINERUN_FAILED |
| Sync Argo CD fallisce | SyncFailed | ARGO_SYNC_FAILED |
| Runtime non raggiunge repliche | SyncFailed | KUBERNETES_RUNTIME_NOT_READY |

---

## 7. Workflow WF-002 - update-app-version

## 7.1 Scopo

Aggiornare il valore applicativo `APP_VERSION` tramite GitOps.

Il valore può essere definito:

1. inline nel Deployment;
2. in una ConfigMap referenziata dal Deployment.

Modalità preferita:

```text
ConfigMap-managed APP_VERSION
```

---

## 7.2 Input

```yaml
applicationName: demo-go-color-app
changeType: update-app-version
requestedBy: vmarzario
description: "Update demo app version"
payload:
  APP_VERSION: v3-green
```

---

## 7.3 Flusso dettagliato

```text
1. Create ChangeRequest
2. Resolve Application metadata
3. Read Deployment manifest
4. Detect APP_VERSION source
5. If APP_VERSION inline: update deployment.yaml
6. If APP_VERSION in ConfigMap: read and update configmap.yaml
7. Verify kustomization.yaml includes modified file
8. Generate diff
9. Anti-secret check
10. Commit files
11. Tekton validation
12. Merge/commit approved
13. Argo CD sync
14. Wait Synced/Healthy
15. Runtime evidence
16. Completed
```

---

## 7.4 Regola ConfigMap e rollout

Se `APP_VERSION` è usata come env var da ConfigMap, i Pod già esistenti non acquisiscono automaticamente il nuovo valore finché non vengono ricreati.

Possibili strategie future:

- annotation checksum ConfigMap nel Pod template;
- annotation Change ID nel Deployment;
- rollout restart controllato;
- change separato sul Deployment.

Decisione da formalizzare in ADR dedicato.

---

## 7.5 Criteri di successo

- valore aggiornato nel file corretto;
- Tekton validation succeeded;
- Argo CD Synced/Healthy;
- runtime espone o riflette nuova versione, se verificabile;
- evidenze salvate.

---

## 8. Workflow WF-003 - update-page-color

## 8.1 Scopo

Aggiornare il valore applicativo `PAGE_COLOR` tramite GitOps.

---

## 8.2 Input

```yaml
applicationName: demo-go-color-app
changeType: update-page-color
requestedBy: vmarzario
description: "Update page color"
payload:
  PAGE_COLOR: "#28A745"
```

---

## 8.3 Validazioni input

`PAGE_COLOR` deve rispettare il formato:

```text
#[0-9A-Fa-f]{6}
```

Esempi validi:

```text
#1E90FF
#28A745
#FF0000
```

Esempi non validi:

```text
blue
28A745
#XYZ123
#123
```

---

## 8.4 Flusso dettagliato

```text
1. Create ChangeRequest
2. Validate PAGE_COLOR
3. Resolve Application metadata
4. Detect PAGE_COLOR source
5. Update Deployment or ConfigMap manifest
6. Verify kustomization.yaml
7. Generate diff
8. Anti-secret check
9. Commit files
10. Tekton validation
11. Merge/commit approved
12. Argo CD sync
13. Wait Synced/Healthy
14. Optional HTTP evidence su Route
15. Completed
```

---

## 8.5 Evidence opzionale HTTP

Se la Application espone una Route, il sistema può raccogliere:

- HTTP status code;
- `/healthz` response;
- pagina applicativa se endpoint non sensibile;
- valore colore esposto, se verificabile.

---

## 9. Workflow WF-004 - update-configmap-values

## 9.1 Scopo

Modificare una o più chiavi in una ConfigMap gestita via GitOps.

---

## 9.2 Input

```yaml
applicationName: demo-go-color-app
changeType: update-configmap-values
requestedBy: vmarzario
description: "Update application configuration"
payload:
  configMapName: demo-go-color-app-config
  values:
    APP_VERSION: v3-green
    PAGE_COLOR: "#28A745"
```

---

## 9.3 Validazioni

- ConfigMap deve esistere nel repository;
- ConfigMap deve essere inclusa in `kustomization.yaml`, se applicabile;
- AppProject deve consentire `ConfigMap`;
- valori non devono contenere secret;
- `PAGE_COLOR`, se presente, deve rispettare formato esadecimale;
- chiavi richieste devono essere presenti o policy deve consentire creazione chiavi.

---

## 9.4 Flusso dettagliato

```text
1. Create ChangeRequest
2. Resolve Application metadata
3. Read configmap.yaml
4. Read kustomization.yaml
5. Verify ConfigMap is included in Kustomize
6. Optional: verify AppProject allows ConfigMap
7. Update values
8. Generate diff
9. Anti-secret check
10. Commit files
11. Tekton validation
12. Merge/commit approved
13. Argo CD sync
14. Wait Synced/Healthy
15. Collect ConfigMap runtime evidence
16. Collect Deployment/Pod evidence
17. Completed
```

---

## 9.5 Errori specifici

| Errore | Stato | Codice |
|---|---|---|
| configmap.yaml mancante | Failed | GITLAB_FILE_NOT_FOUND |
| ConfigMap non inclusa in Kustomize | ValidationFailed | VALIDATION_KUSTOMIZE_RESOURCE_MISSING |
| ConfigMap non permessa da AppProject | SyncFailed | ARGO_RESOURCE_NOT_PERMITTED |
| Possibile secret nei valori | ValidationFailed | VALIDATION_SECRET_DETECTED |
| Pod non aggiornati | Completed con warning oppure SyncFailed | KUBERNETES_RUNTIME_NOT_UPDATED |

---

## 10. Workflow WF-005 - validate-only

## 10.1 Scopo

Eseguire validazione Tekton su branch o commit senza avviare sync Argo CD.

Utile per:

- verifiche preventive;
- review MR;
- test workflow;
- troubleshooting manifest.

---

## 10.2 Input

```yaml
applicationName: demo-go-color-app
changeType: validate-only
payload:
  repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
  revision: change/CHG-2026-0001-update-replicas
  path: apps/demo-go-color-app
```

---

## 10.3 Flusso

```text
1. Create ChangeRequest or validation record
2. Resolve Git metadata
3. Create Tekton PipelineRun
4. Wait validation
5. Collect Tekton evidence
6. Completed or ValidationFailed
```

---

## 11. Workflow WF-006 - sync-only

## 11.1 Scopo

Lanciare sync Argo CD senza creare nuovo commit Git.

Utile quando:

- Git è già aggiornato;
- Argo CD è OutOfSync;
- MR è stata mergiata manualmente;
- si vuole riconciliare lo stato.

---

## 11.2 Precondizioni

- Application esiste;
- repository Git è già nello stato desiderato;
- non ci sono validation failure note;
- non esiste operation Argo CD già in corso.

---

## 11.3 Flusso

```text
1. Create ChangeRequest sync-only
2. Read Argo CD Application
3. If operation in progress: fail or wait
4. Sync Application
5. Wait Synced/Healthy
6. Collect evidence
7. Completed
```

---

## 11.4 Errori noti

```text
ARGO_OPERATION_IN_PROGRESS
ARGO_SYNC_FAILED
ARGO_RESOURCE_NOT_PERMITTED
ARGO_APPLICATION_DEGRADED
ARGO_APPLICATION_OUTOFSYNC_TIMEOUT
```

---

## 12. Workflow WF-007 - collect-evidence-only

## 12.1 Scopo

Raccogliere evidenze di una Application senza modificare Git o lanciare sync.

Utile per:

- audit;
- troubleshooting;
- verifica post-change;
- baseline runtime.

---

## 12.2 Flusso

```text
1. Create evidence collection request
2. Read Argo CD status
3. Read resources/orphaned resources
4. Read Deployment/Pod/Service/Route
5. Optional health check HTTP
6. Save evidence
7. Completed
```

---

## 13. Gestione conflitti

## 13.1 Change concorrenti

Il sistema deve evitare workflow concorrenti pericolosi sulla stessa Application o sullo stesso file.

Regola MVP:

```text
Bloccare una nuova ChangeRequest se esiste già una ChangeRequest attiva sulla stessa Application e stesso file target.
```

---

## 13.2 Esempio conflitto

Change A:

```text
update-configmap-values su apps/demo-go-color-app/configmap.yaml
```

Change B:

```text
update-page-color su apps/demo-go-color-app/configmap.yaml
```

Azione consigliata:

```text
Bloccare Change B fino a completamento o cancellazione di Change A.
```

Codice errore:

```text
WORKFLOW_CONFLICT_ACTIVE_CHANGE
```

---

## 14. Gestione rollback

## 14.1 Principio

Rollback definitivo deve preferibilmente passare da Git.

```text
git revert / revert MR
  -> merge
  -> Argo CD sync
```

---

## 14.2 Rollback Argo CD operativo

Rollback Argo CD può riportare il cluster a una revision precedente, ma non modifica Git.

Il sistema deve spiegare che può risultare:

```text
OutOfSync from main
Healthy
```

Significato:

```text
runtime sano ma non allineato al target Git corrente.
```

---

## 14.3 Roadmap rollback

MVP:

- mostrare rollback hint;
- mostrare commit/MR da revertire;
- non implementare rollback automatico complesso.

Futuro:

- workflow `rollback-git-revert`;
- creazione automatica MR di revert;
- sync post-revert;
- evidence rollback.

---

## 15. Evidence collection standard

## 15.1 Evidence minime per change completo

Ogni change completato deve avere:

- ChangeRequest summary;
- GitLab branch/commit/MR;
- diff summary;
- Tekton PipelineRun status;
- Tekton TaskRun summary;
- Argo CD sync/health/revision;
- Kubernetes runtime status;
- errori o warning;
- timestamp.

---

## 15.2 Evidence minime in caso di fallimento

Ogni change fallito deve avere:

- fase fallita;
- error code;
- messaggio tecnico;
- messaggio operativo;
- suggested action;
- stato ultimo osservato;
- log essenziale sanitizzato.

---

## 16. Error handling standard

## 16.1 Struttura errore workflow

```json
{
  "code": "ARGO_OPERATION_IN_PROGRESS",
  "technicalMessage": "another operation is already in progress",
  "userMessage": "Argo CD ha già una operazione in corso sulla Application.",
  "suggestedAction": "Attendere il completamento dell'operazione oppure verificarla manualmente.",
  "recoverable": true
}
```

---

## 16.2 Regole

- ogni errore deve avere `error_code`;
- ogni errore deve creare `change_events`;
- non loggare token;
- non salvare secret;
- non perdere lo stato precedente;
- non lasciare workflow appesi.

---

## 17. Timeout

## 17.1 Timeout consigliati MVP

```text
GitLab API call: 30s
Argo CD API call: 30s
Tekton PipelineRun: 900s
Argo CD sync wait: 600s
Runtime readiness wait: 300s
HTTP health check: 10s
```

---

## 17.2 Timeout behavior

Se scatta timeout:

- aggiornare stato;
- registrare evento;
- salvare stato ultimo osservato;
- salvare evidence parziale;
- proporre azione manuale.

---

## 18. API orchestration mapping

## 18.1 Create change

```text
POST /api/changes
```

Crea ChangeRequest in stato `Created`.

---

## 18.2 Execute workflow step-by-step

Endpoint possibili:

```text
POST /api/changes/{id}/create-branch
POST /api/changes/{id}/update-files
POST /api/changes/{id}/validate
POST /api/changes/{id}/open-merge-request
POST /api/changes/{id}/sync
POST /api/changes/{id}/collect-evidence
```

---

## 18.3 Execute workflow automatico

Endpoint futuro:

```text
POST /api/changes/{id}/run
```

Nota:

Per MVP è consigliabile implementare prima step espliciti, più facili da validare e da spiegare ai newbie.

---

## 19. Workflow didattico per newbie

Ogni workflow dovrebbe mostrare:

```text
Step corrente
Strumento usato
Perché serve
File coinvolti
Evidenza prodotta
Stato finale
```

Esempio:

```text
Step: Tekton validation
Strumento: OpenShift Pipelines / Tekton
Perché: validare i manifest prima della sync Argo CD
File: apps/demo-go-color-app/deployment.yaml
Evidenza: PipelineRun validate-gitops-change-chg-2026-0001 succeeded
```

---

## 20. Checklist workflow MVP

Un workflow MVP è considerato pronto quando:

- crea ChangeRequest;
- crea eventi per ogni step;
- modifica Git e non runtime diretto;
- crea branch/commit/MR;
- esegue validazione Tekton;
- blocca sync se validazione fallisce;
- lancia sync Argo CD solo se consentito;
- attende Synced/Healthy;
- raccoglie evidenze;
- salva stato finale;
- gestisce timeout;
- non salva secret;
- produce messaggi comprensibili.

---

## 21. Relazione con altri documenti

Questo documento alimenta:

- `docs/13-api-design.md`;
- `docs/12-evidence-model.md`;
- `internal/workflow/`;
- `internal/app/change_service.go`;
- `internal/adapters/gitlab/`;
- `internal/adapters/argocd/`;
- `internal/adapters/tekton/`;
- `migrations/000001_init.up.sql`;
- ADR sui workflow e sulle scelte di governance.

---

## 22. Messaggio chiave

I workflow sono il cuore del DevOps Control Plane.

Il prodotto deve aiutare l’operatore a fare la cosa giusta:

```text
modificare Git
validare con Tekton
sincronizzare con Argo CD
verificare OpenShift
salvare evidenze
```

Non deve diventare un sistema che nasconde il GitOps, ma un sistema che lo rende più chiaro, guidato, ripetibile e auditabile.
