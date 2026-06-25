# DevOps Control Plane - Tekton Integration

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
**Stato:** Draft iniziale / Tekton Integration

---

## 1. Scopo del documento

Questo documento descrive come **DevOps Control Plane** deve integrarsi con **Tekton / OpenShift Pipelines**.

L’obiettivo è definire:

- responsabilità dell’integrazione Tekton;
- motivazione dell’uso di Tekton nel progetto;
- modello di interazione tramite Kubernetes API diretta;
- risorse Tekton coinvolte;
- modello `PipelineRun` / `TaskRun`;
- pipeline di validazione MVP;
- parametri, workspace e ServiceAccount;
- raccolta stato e log;
- evidenze prodotte;
- error handling;
- requisiti RBAC;
- ordine di implementazione dell’adapter Tekton.

Tekton non sostituisce GitLab o Argo CD. Nel DevOps Control Plane, Tekton ha il ruolo di **motore di validazione e automazione tecnica** dei change GitOps.

---

## 2. Ruolo di Tekton nell’architettura

Nel progetto DevOps Control Plane, Tekton ha il ruolo di **Validation Engine**.

Responsabilità Tekton:

- eseguire pipeline di validazione;
- clonare repository/branch GitOps;
- validare YAML;
- eseguire `kustomize build`, se applicabile;
- eseguire `oc apply --dry-run=server`;
- eseguire controlli anti-secret;
- verificare regole GitOps/AppProject quando possibile;
- produrre stato `Succeeded` o `Failed`;
- generare log tecnici associabili alla ChangeRequest.

Responsabilità DevOps Control Plane:

- creare `PipelineRun` tramite Kubernetes API;
- passare parametri alla PipelineRun;
- monitorare stato PipelineRun;
- raccogliere TaskRun collegate;
- raccogliere log principali;
- salvare evidenze in PostgreSQL;
- aggiornare lo stato della ChangeRequest;
- interpretare errori Tekton in modo operativo.

---

## 3. Perché usare Kubernetes API diretta

La scelta architetturale iniziale è:

```text
DevOps Control Plane -> Kubernetes API -> Tekton CRDs
```

Il backend Go non deve dipendere dalla CLI `tkn` o da wrapper shell.

### Vantaggi

- integrazione nativa con OpenShift/Kubernetes;
- nessuna dipendenza dalla shell del container;
- migliore controllo degli errori;
- migliore osservabilità via status CRD;
- possibilità di watch/polling programmatico;
- coerenza con il modello Kubernetes-native di Tekton.

### Regola

La CLI `tkn` resta utile per troubleshooting manuale, ma non è il meccanismo primario del prodotto.

---

## 4. Concetti Tekton rilevanti

## 4.1 Task

Una `Task` definisce un’unità riutilizzabile di lavoro.

Esempi:

- clone repository;
- validazione YAML;
- `kustomize build`;
- dry-run server-side;
- anti-secret check;
- generazione report.

---

## 4.2 TaskRun

Una `TaskRun` rappresenta l’esecuzione concreta di una `Task`.

Nel workflow DevOps Control Plane, le TaskRun sono generate dalla PipelineRun e devono essere lette per conoscere:

- task eseguite;
- stato task;
- durata;
- errore eventuale;
- log principali.

---

## 4.3 Pipeline

Una `Pipeline` definisce una sequenza o grafo di Task.

Per l’MVP, la pipeline principale sarà:

```text
validate-gitops-change
```

---

## 4.4 PipelineRun

Una `PipelineRun` istanzia ed esegue una `Pipeline`.

Nel DevOps Control Plane, ogni ChangeRequest che richiede validazione deve generare una PipelineRun dedicata.

Esempio naming:

```text
validate-gitops-change-chg-2026-0001
```

oppure con suffisso generato:

```text
validate-gitops-change-chg-2026-0001-abcde
```

---

## 5. Funzionalità MVP Tekton

## 5.1 Creare PipelineRun di validazione

### Obiettivo

DevOps Control Plane deve creare una PipelineRun per validare un branch GitLab relativo a una ChangeRequest.

### Input minimo

```yaml
changeId: CHG-2026-0001
applicationName: demo-go-color-app
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
revision: change/CHG-2026-0001-update-replicas
path: apps/demo-go-color-app
targetNamespace: devops-ci-demo
```

### Output atteso

```yaml
pipelineRunName: validate-gitops-change-chg-2026-0001-abcde
namespace: devops-control-plane
status: Running
```

---

## 5.2 Monitorare PipelineRun

### Obiettivo

DevOps Control Plane deve monitorare la PipelineRun fino allo stato finale.

### Stati logici

```text
ValidationRequested
ValidationRunning
ValidationSucceeded
ValidationFailed
ValidationTimeout
```

### Regole

- timeout obbligatorio;
- polling interval configurabile;
- stato finale salvato in PostgreSQL;
- messaggio errore salvato se fallisce;
- evidence Tekton salvata.

---

## 5.3 Raccogliere TaskRun

### Obiettivo

Associare alla ChangeRequest le TaskRun generate dalla PipelineRun.

### Dati minimi

```yaml
taskRunName: validate-gitops-change-chg-2026-0001-yaml-lint
pipelineRunName: validate-gitops-change-chg-2026-0001-abcde
taskName: yaml-lint
status: Succeeded
startTime: "2026-06-25T15:00:00+02:00"
completionTime: "2026-06-25T15:00:15+02:00"
```

---

## 5.4 Raccogliere log principali

### Obiettivo

Raccogliere log utili alla diagnosi e all’audit.

### Regole

- non salvare token;
- non salvare secret;
- limitare dimensione log;
- salvare summary leggibile;
- salvare log completi solo se sanitizzati o referenziati in storage adeguato.

---

## 6. Pipeline MVP: validate-gitops-change

## 6.1 Obiettivo

Validare un change GitOps prima della sync Argo CD.

La Pipeline deve dimostrare che:

- il branch Git esiste;
- i manifest sono leggibili;
- YAML è valido;
- Kustomize build funziona, se applicabile;
- il dry-run server-side non fallisce;
- non ci sono secret evidenti nei file modificati;
- le risorse introdotte sono coerenti con governance prevista.

---

## 6.2 Task candidate

### Task 1 - clone-repository

Responsabilità:

- clonare repository GitLab;
- checkout del branch ChangeRequest;
- rendere il workspace disponibile alle task successive.

Parametri:

```yaml
repo-url: https://gitlab.example.local/group/demo-app-gitops.git
revision: change/CHG-2026-0001-update-replicas
```

---

### Task 2 - show-context

Responsabilità:

- stampare contesto non sensibile;
- mostrare Change ID;
- mostrare branch;
- mostrare path GitOps.

### Nota

Non stampare token o credenziali.

---

### Task 3 - yaml-validate

Responsabilità:

- validare sintassi YAML;
- individuare errori di indentazione;
- fallire se file YAML non validi.

Motivazione:

Nel lab è stato osservato che un errore di indentazione `valueFrom/configMapKeyRef` può causare failure durante la sync Argo CD.

---

### Task 4 - kustomize-build

Responsabilità:

- eseguire `kustomize build` sul path applicativo, se presente `kustomization.yaml`;
- verificare che tutti i manifest referenziati esistano;
- individuare file non inclusi in `resources`.

Esempio problema da intercettare:

```text
configmap.yaml creato ma non aggiunto a kustomization.yaml
```

---

### Task 5 - server-side-dry-run

Responsabilità:

- eseguire dry-run server-side contro OpenShift/Kubernetes;
- intercettare errori schema Kubernetes;
- intercettare risorse non valide.

Esempio errore da intercettare:

```text
valueFrom: Invalid value: must specify configMapKeyRef, secretKeyRef, fieldRef or resourceFieldRef
```

---

### Task 6 - anti-secret-check

Responsabilità:

- cercare pattern sospetti nei file modificati;
- bloccare commit/promozione se vengono trovati secret evidenti.

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
config.json
```

---

### Task 7 - appproject-policy-check

Responsabilità:

- verificare se le risorse introdotte sono consentite dall’AppProject;
- almeno per MVP, intercettare casi noti come ConfigMap non autorizzata.

Esempio:

```text
resource :ConfigMap is not permitted in project devops-ci-demo
```

Nota: questa task può essere implementata in modo incrementale. Nel primissimo MVP può essere un controllo statico o opzionale.

---

### Task 8 - report

Responsabilità:

- produrre summary finale;
- elencare task eseguite;
- elencare file validati;
- elencare esito;
- produrre eventuale messaggio didattico.

---

## 7. PipelineRun template concettuale

Esempio concettuale non definitivo:

```yaml
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  generateName: validate-gitops-change-
  namespace: devops-control-plane
  labels:
    app.kubernetes.io/name: devops-control-plane
    devops-control-plane/change-id: CHG-2026-0001
    devops-control-plane/application: demo-go-color-app
spec:
  pipelineRef:
    name: validate-gitops-change
  params:
    - name: change-id
      value: CHG-2026-0001
    - name: repo-url
      value: https://gitlab.example.local/group/demo-app-gitops.git
    - name: revision
      value: change/CHG-2026-0001-update-replicas
    - name: gitops-path
      value: apps/demo-go-color-app
    - name: target-namespace
      value: devops-ci-demo
  workspaces:
    - name: shared-workspace
      persistentVolumeClaim:
        claimName: devops-control-plane-workspace
  serviceAccountName: pipeline
```

### Nota

Il manifest definitivo sarà creato in `pipelines/validate-gitops-change.yaml` quando inizierà l’implementazione.

---

## 8. Tekton Adapter

## 8.1 Responsabilità

Il componente `TektonAdapter` incapsula la gestione delle risorse Tekton tramite Kubernetes API.

Il resto dell’applicazione non deve conoscere dettagli CRD, GVR, client-go dynamic client o YAML raw.

---

## 8.2 Interfaccia logica MVP

Interfaccia concettuale:

```go
type TektonAdapter interface {
    CreatePipelineRun(ctx context.Context, namespace string, req CreatePipelineRunRequest) (*PipelineRunRef, error)
    GetPipelineRun(ctx context.Context, namespace string, name string) (*PipelineRunStatus, error)
    ListTaskRunsForPipelineRun(ctx context.Context, namespace string, pipelineRunName string) ([]TaskRunStatus, error)
    GetTaskRunLogs(ctx context.Context, namespace string, taskRunName string) ([]TaskLog, error)
    WaitPipelineRun(ctx context.Context, namespace string, name string, opts WaitOptions) (*PipelineRunStatus, error)
}
```

Nota: l’interfaccia è indicativa. La forma finale sarà definita durante l’implementazione Go.

---

## 8.3 Posizione package proposta

```text
internal/adapters/tekton/
├── client.go
├── models.go
├── mapper.go
├── pipelineruns.go
├── taskruns.go
├── logs.go
├── wait.go
└── errors.go
```

### File `client.go`

Responsabilità:

- inizializzare Kubernetes client;
- gestire namespace Tekton;
- gestire timeout;
- configurare dynamic client o typed client.

### File `pipelineruns.go`

Responsabilità:

- creare PipelineRun;
- leggere PipelineRun;
- serializzare/deserializzare spec e status.

### File `taskruns.go`

Responsabilità:

- trovare TaskRun collegate a una PipelineRun;
- leggere status e conditions.

### File `logs.go`

Responsabilità:

- raccogliere log dai Pod/containers collegati alle TaskRun;
- sanitizzare log;
- limitare dimensione.

### File `wait.go`

Responsabilità:

- polling stato PipelineRun;
- timeout;
- mapping condizioni finali.

### File `errors.go`

Responsabilità:

- normalizzare errori Kubernetes/Tekton.

---

## 9. Modello dati normalizzato

## 9.1 PipelineRunRef

```yaml
name: validate-gitops-change-chg-2026-0001-abcde
namespace: devops-control-plane
uid: "..."
changeId: CHG-2026-0001
applicationName: demo-go-color-app
```

---

## 9.2 PipelineRunStatus

```yaml
name: validate-gitops-change-chg-2026-0001-abcde
namespace: devops-control-plane
status: Succeeded
reason: Succeeded
message: "Tasks Completed: 8 completed, 0 failed"
startTime: "2026-06-25T15:00:00+02:00"
completionTime: "2026-06-25T15:03:00+02:00"
conditions:
  - type: Succeeded
    status: "True"
    reason: Succeeded
```

---

## 9.3 TaskRunStatus

```yaml
name: validate-gitops-change-chg-2026-0001-yaml-validate
namespace: devops-control-plane
taskName: yaml-validate
pipelineRunName: validate-gitops-change-chg-2026-0001-abcde
status: Succeeded
reason: Succeeded
message: "All YAML files are valid"
```

---

## 9.4 TaskLog

```yaml
taskRunName: validate-gitops-change-chg-2026-0001-yaml-validate
container: step-yaml-validate
logExcerpt: "YAML validation completed successfully"
truncated: false
```

---

## 10. ChangeRequest mapping

## 10.1 Campi ChangeRequest

```yaml
changeId: CHG-2026-0001
validation:
  provider: tekton
  namespace: devops-control-plane
  pipelineRunName: validate-gitops-change-chg-2026-0001-abcde
  status: Succeeded
  startedAt: "2026-06-25T15:00:00+02:00"
  completedAt: "2026-06-25T15:03:00+02:00"
```

---

## 10.2 Change events Tekton

Eventi minimi:

```text
ValidationRequested
ValidationRunning
ValidationSucceeded
ValidationFailed
ValidationTimeout
TektonPipelineRunCreated
TektonTaskRunCollected
TektonLogsCollected
```

---

## 11. Stati e conditions Tekton

## 11.1 Mapping condition Succeeded

| Tekton condition | Stato DevOps Control Plane |
|---|---|
| Succeeded=True | ValidationSucceeded |
| Succeeded=False | ValidationFailed |
| Succeeded=Unknown | ValidationRunning |
| Timeout exceeded | ValidationTimeout |

---

## 11.2 Regole

- `Succeeded=True` è successo.
- `Succeeded=False` è fallimento.
- `Succeeded=Unknown` è running/pending.
- assenza condition oltre timeout è timeout.
- errori Kubernetes API sono errori adapter o infrastrutturali.

---

## 12. Evidence Tekton

## 12.1 Evidence minima

```yaml
provider: tekton
namespace: devops-control-plane
pipelineRunName: validate-gitops-change-chg-2026-0001-abcde
status: Succeeded
reason: Succeeded
taskRuns:
  - name: clone-repository
    status: Succeeded
  - name: yaml-validate
    status: Succeeded
  - name: kustomize-build
    status: Succeeded
  - name: server-side-dry-run
    status: Succeeded
```

---

## 12.2 Evidence in caso di errore

```yaml
provider: tekton
pipelineRunName: validate-gitops-change-chg-2026-0001-abcde
status: Failed
failedTask: server-side-dry-run
message: "Deployment.apps demo-go-color-app is invalid: valueFrom is incomplete"
errorCode: TEKTON_PIPELINERUN_FAILED
```

---

## 12.3 Log evidence

I log dovrebbero essere salvati come:

- excerpt sintetico;
- riferimento a TaskRun;
- eventuale payload JSON sanitizzato;
- non come dump illimitato.

---

## 13. Error model Tekton

## 13.1 Codici errore interni

```text
TEKTON_AUTH_FAILED
TEKTON_FORBIDDEN
TEKTON_NAMESPACE_NOT_FOUND
TEKTON_PIPELINE_NOT_FOUND
TEKTON_PIPELINERUN_CREATE_FAILED
TEKTON_PIPELINERUN_NOT_FOUND
TEKTON_PIPELINERUN_FAILED
TEKTON_PIPELINERUN_TIMEOUT
TEKTON_TASKRUN_LIST_FAILED
TEKTON_LOG_COLLECTION_FAILED
TEKTON_UNKNOWN_ERROR
```

---

## 13.2 Struttura errore normalizzata

```yaml
code: TEKTON_PIPELINERUN_FAILED
technicalMessage: "PipelineRun validate-gitops-change-chg-2026-0001 failed"
userMessage: "La validazione Tekton del change è fallita."
suggestedAction: "Consultare TaskRun e log associati alla ChangeRequest."
recoverable: true
```

---

## 13.3 Errori YAML/dry-run noti

Esempio:

```text
valueFrom: Invalid value: must specify configMapKeyRef, secretKeyRef, fieldRef or resourceFieldRef
```

Messaggio operativo:

```text
Il Deployment contiene un blocco valueFrom incompleto o indentato male. Verificare che configMapKeyRef sia correttamente indentato sotto valueFrom.
```

---

## 14. Configurazione Tekton

Variabili previste:

```text
TEKTON_NAMESPACE=devops-control-plane
TEKTON_PIPELINE_NAME=validate-gitops-change
TEKTON_SERVICE_ACCOUNT=pipeline
TEKTON_TIMEOUT_SECONDS=900
TEKTON_POLL_INTERVAL_SECONDS=5
TEKTON_WORKSPACE_CLAIM=devops-control-plane-workspace
```

### Regole

- namespace configurabile;
- timeout obbligatorio;
- ServiceAccount dedicato o controllato;
- PVC/workspace configurabile;
- nessun token GitLab in chiaro nel manifest.

---

## 15. RBAC e sicurezza

## 15.1 ServiceAccount DevOps Control Plane

Il ServiceAccount dell’applicazione deve poter:

- creare PipelineRun nel namespace Tekton configurato;
- leggere PipelineRun;
- list/watch PipelineRun;
- leggere TaskRun;
- list/watch TaskRun;
- leggere Pod e log associati alle TaskRun, se necessario.

---

## 15.2 ServiceAccount PipelineRun

La PipelineRun deve avere un ServiceAccount adeguato per:

- clonare repository, se usa secret Git;
- eseguire `oc apply --dry-run=server`, se previsto;
- leggere AppProject o risorse governance, se previsto;
- non avere privilegi eccessivi.

---

## 15.3 Secret handling

- Git credentials per la pipeline devono stare in Secret.
- Token GitLab non devono essere stampati nei log.
- Token non devono essere salvati in evidenza.
- Il report finale deve essere sanitizzato.

---

## 16. API interne DevOps Control Plane collegate a Tekton

## 16.1 Validation endpoints

```text
POST /api/changes/{id}/validate
GET  /api/changes/{id}/validation
GET  /api/changes/{id}/validation/taskruns
GET  /api/changes/{id}/validation/logs
```

---

## 16.2 Evidence endpoint

```text
GET /api/changes/{id}/evidence/tekton
```

---

## 17. Testing strategy

## 17.1 Unit test

Testare:

- costruzione PipelineRun spec;
- mapping PipelineRun status;
- mapping TaskRun status;
- mapping conditions;
- mapping errori;
- timeout wait loop;
- sanitizzazione log.

---

## 17.2 Test con fake Kubernetes client

Il Workflow Engine deve poter essere testato senza cluster reale.

Scenario fake:

```text
CreatePipelineRun -> returns PipelineRunRef
GetPipelineRun -> Running
GetPipelineRun -> Succeeded
ListTaskRuns -> returns task status
GetTaskRunLogs -> returns sanitized logs
```

---

## 17.3 Integration test manuale MVP

Scenario minimo:

1. creare namespace `devops-control-plane`;
2. applicare Pipeline `validate-gitops-change`;
3. configurare Secret Git, se necessario;
4. creare PipelineRun via DevOps Control Plane;
5. verificare PipelineRun Running;
6. verificare TaskRun create;
7. verificare stato finale Succeeded;
8. verificare evidenze salvate nella ChangeRequest.

---

## 17.4 Casi errore da simulare

- namespace Tekton non esistente;
- Pipeline non trovata;
- ServiceAccount senza permessi;
- repository non clonabile;
- YAML non valido;
- dry-run fallito;
- secret rilevato;
- PipelineRun timeout;
- log non leggibili.

---

## 18. MVP implementation order per Tekton adapter

Ordine consigliato:

1. configurazione Tekton;
2. Kubernetes client base;
3. modello `CreatePipelineRunRequest`;
4. creazione PipelineRun;
5. lettura PipelineRun;
6. mapping condition `Succeeded`;
7. wait con timeout;
8. list TaskRuns collegate;
9. raccolta log essenziale;
10. evidence builder;
11. mapping errori;
12. integrazione con ChangeRequest Service.

---

## 19. Checklist di completamento integrazione Tekton

La prima integrazione Tekton sarà considerata pronta quando:

- il backend legge configurazione Tekton;
- il Kubernetes client è inizializzato;
- una PipelineRun può essere creata;
- la PipelineRun contiene label/annotation con Change ID;
- lo stato PipelineRun viene monitorato;
- TaskRun collegate sono visibili;
- log principali sono raccoglibili o referenziabili;
- timeout è gestito;
- errori noti sono normalizzati;
- l’esito validazione aggiorna la ChangeRequest;
- evidence Tekton è salvabile in PostgreSQL.

---

## 20. Relazione con altri documenti

Questo documento alimenta:

- `docs/10-data-model.md`, per entità TektonValidation, PipelineRun e TaskRun;
- `docs/11-change-workflows.md`, per lo step validation;
- `docs/12-evidence-model.md`, per evidenze Tekton;
- `docs/13-api-design.md`, per endpoint validation;
- `docs/09-security-rbac.md`, per ServiceAccount e Role/RoleBinding;
- ADR specifico `ADR-0003-tekton-validation-engine.md`;
- ADR specifico `ADR-0008-kubernetes-api-for-tekton.md`.

---

## 21. Riferimenti tecnici Tekton/OpenShift Pipelines

Riferimenti da consultare durante l’implementazione:

- Tekton Tasks and Pipelines documentation;
- Tekton PipelineRuns documentation;
- Tekton Pipelines documentation;
- Red Hat OpenShift Pipelines documentation;
- Tekton Results documentation, per evoluzione futura della raccolta evidenze.

---

## 22. Messaggio chiave

Nel DevOps Control Plane, Tekton serve a trasformare la validazione dei change GitOps da attività manuale a processo ripetibile, osservabile e auditabile.

Il flusso target è:

```text
GitLab branch
  -> Tekton PipelineRun
  -> TaskRun validation
  -> evidence
  -> Argo CD sync solo se validation succeeded
```

Tekton non decide cosa deve essere deployato. Valida che il change GitOps sia tecnicamente corretto prima che Argo CD lo applichi al cluster.
