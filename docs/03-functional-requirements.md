# DevOps Control Plane - Functional Requirements

**Versione:** 0.1  
**Data:** 2026-06-25  
**Owner iniziale:** Vincenzo Marzario  
**Repository:** `https://github.com/vincmarz/devops-control-plane`  
**Documenti precedenti:**  
- `docs/00-vision.md`  
- `docs/01-scope-mvp.md`  
- `docs/02-personas-use-cases.md`  
**Stato:** Draft iniziale / Functional Requirements

---

## 1. Scopo del documento

Questo documento definisce i **requisiti funzionali** del primo MVP di **DevOps Control Plane**.

Lo scopo è trasformare la vision, lo scope MVP e gli use cases in requisiti implementabili, tracciabili e verificabili.

Il documento descrive:

- funzionalità applicative;
- input e output attesi;
- regole di validazione;
- stati gestiti;
- integrazioni richieste;
- errori da intercettare;
- criteri di accettazione;
- priorità MVP.

Il documento sarà usato come base per:

- `docs/05-architecture.md`;
- `docs/10-data-model.md`;
- `docs/11-change-workflows.md`;
- `docs/13-api-design.md`;
- implementazione backend Go;
- implementazione adapter Argo CD, GitLab, Tekton e Kubernetes/OpenShift.

---

## 2. Convenzioni

## 2.1 Priorità

I requisiti sono classificati con le seguenti priorità:

| Priorità | Significato |
|---|---|
| MUST | Necessario per MVP |
| SHOULD | Importante ma non bloccante per primo rilascio |
| COULD | Utile in roadmap successiva |
| WON'T | Esplicitamente fuori scope MVP |

---

## 2.2 Stati generali

Quando possibile, ogni operazione deve produrre:

- esito positivo;
- esito negativo;
- messaggio tecnico;
- messaggio operativo leggibile;
- riferimento all’oggetto coinvolto;
- evento persistito in PostgreSQL, se collegato a una ChangeRequest.

---

## 2.3 Regola GitOps fondamentale

Ogni change permanente su workload applicativi deve essere rappresentato da una modifica Git.

Il sistema **non deve** usare modifiche runtime dirette come stato finale.

Esempi non ammessi come workflow finale:

```bash
oc edit deployment
oc patch deployment
oc set env deployment
oc scale deployment
```

Questi comandi possono essere documentati o usati solo per troubleshooting manuale, non come workflow applicativo permanente del DevOps Control Plane.

---

## 3. Aree funzionali

Le funzionalità MVP sono raggruppate nelle seguenti aree:

1. Application Discovery;
2. Application Detail;
3. GitLab Operations;
4. ChangeRequest Management;
5. Change Workflows;
6. Tekton Validation;
7. Argo CD Sync;
8. Evidence Collection;
9. Change History;
10. Error Handling;
11. Security and Secret Handling;
12. Web/HTML Foundation.

---

# 4. Application Discovery

## FR-APP-001 - Lista applicazioni Argo CD

### Priorità

MUST

### Descrizione

Il sistema deve recuperare da Argo CD API la lista delle Application visibili al DevOps Control Plane.

### Input

Nessun input obbligatorio nella prima versione.

Filtri opzionali futuri:

- progetto Argo CD;
- namespace target;
- sync status;
- health status;
- repository;
- label.

### Output minimo

Per ogni Application:

```yaml
name: demo-go-color-app
argocdNamespace: openshift-gitops
project: devops-ci-demo
targetNamespace: devops-ci-demo
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
targetRevision: main
path: apps/demo-go-color-app
syncStatus: Synced
healthStatus: Healthy
currentRevision: b8e1d6b
```

### Regole

- Il sistema deve normalizzare i campi principali.
- Il sistema deve gestire Application non accessibili o non autorizzate.
- Il sistema non deve esporre token Argo CD nei log.

### Acceptance criteria

- `GET /api/applications` restituisce la lista applicazioni.
- Ogni record contiene almeno nome, progetto, namespace, sync, health e revision.
- In caso di errore API, il backend restituisce un errore leggibile.

---

## FR-APP-002 - Sincronizzazione cache applicazioni in PostgreSQL

### Priorità

SHOULD

### Descrizione

Il sistema dovrebbe salvare uno snapshot delle Application note in PostgreSQL.

### Motivazione

La cache permette di:

- correlare ChangeRequest e Application;
- mantenere storico anche se Argo CD è temporaneamente non raggiungibile;
- mostrare dati recenti in modalità degraded.

### Acceptance criteria

- Il sistema può aggiornare o inserire Application in tabella `applications`.
- Ogni sync aggiorna `updated_at`.
- I dati sensibili non vengono salvati.

---

# 5. Application Detail

## FR-APP-010 - Dettaglio Application

### Priorità

MUST

### Descrizione

Il sistema deve mostrare il dettaglio operativo di una Application.

### Input

```text
applicationName
```

### Output minimo

- nome;
- progetto;
- namespace target;
- repository;
- path;
- target revision;
- current revision;
- sync status;
- health status;
- sync policy;
- conditions;
- destination server;
- resources summary.

### Acceptance criteria

- `GET /api/applications/{name}` restituisce dettaglio coerente con Argo CD.
- Lo stato `OrphanedResourceWarning` viene riportato ma non trattato automaticamente come errore applicativo se `healthStatus=Healthy`.

---

## FR-APP-011 - Resources gestite e orphaned resources

### Priorità

MUST

### Descrizione

Il sistema deve recuperare le risorse associate a una Application e distinguere:

- risorse gestite;
- risorse orphaned;
- risorse missing;
- risorse out of sync.

### Output esempio

```yaml
resources:
  - group: apps
    kind: Deployment
    namespace: devops-ci-demo
    name: demo-go-color-app
    status: Synced
    health: Healthy
    orphaned: false
  - group: image.openshift.io
    kind: ImageStream
    namespace: devops-ci-demo
    name: demo-go-color-app
    orphaned: true
```

### Regole

- Il group vuoto deve essere rappresentato come stringa vuota o `core` a livello UI, ma preservato correttamente nel backend.
- Il sistema deve spiegare che Service, ConfigMap, Secret e PVC appartengono al core API group Kubernetes.

### Acceptance criteria

- `GET /api/applications/{name}/resources` restituisce risorse gestite e orphaned.
- L’output distingue chiaramente risorse sane, missing, out of sync e orphaned.

---

## FR-APP-012 - History Argo CD

### Priorità

SHOULD

### Descrizione

Il sistema deve recuperare la history Argo CD di una Application.

### Output minimo

- history ID;
- revision;
- deployedAt;
- author, se disponibile;
- message, se disponibile.

### Acceptance criteria

- `GET /api/applications/{name}/history` restituisce history ordinata.
- Il sistema distingue history Argo CD da history Git.

---

# 6. GitLab Operations

## FR-GIT-001 - Configurazione repository GitLab

### Priorità

MUST

### Descrizione

Il sistema deve conoscere o ricavare i metadati GitLab necessari per operare sui repository GitOps.

### Dati necessari

```yaml
provider: gitlab
baseUrl: https://gitlab.example.local
projectId: "12345"
defaultBranch: main
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
path: apps/demo-go-color-app
```

### Acceptance criteria

- Ogni Application gestita può essere associata a un repository GitLab.
- Il sistema può leggere file nel path GitOps configurato.

---

## FR-GIT-002 - Lettura file repository

### Priorità

MUST

### Descrizione

Il sistema deve leggere un file dal repository GitLab usando branch/ref e path.

### Esempi file

```text
apps/demo-go-color-app/deployment.yaml
apps/demo-go-color-app/configmap.yaml
apps/demo-go-color-app/kustomization.yaml
```

### Acceptance criteria

- Il sistema recupera contenuto file da GitLab API.
- Il sistema gestisce file mancanti con errore leggibile.
- Il sistema non scarica interi repository se non necessario nell’MVP.

---

## FR-GIT-003 - Creazione branch

### Priorità

MUST

### Descrizione

Il sistema deve creare un branch dedicato per ogni ChangeRequest.

### Naming convention

```text
change/<change-id>-<short-description>
```

Esempio:

```text
change/CHG-2026-0001-update-replicas
```

### Regole

- Il branch deve partire dal branch target configurato, tipicamente `main`.
- Il nome branch deve essere normalizzato e sicuro.
- Se il branch esiste già, il sistema deve gestire il conflitto.

### Acceptance criteria

- Il branch viene creato tramite GitLab API.
- Il branch viene salvato nella ChangeRequest.
- L’evento `BranchCreated` viene registrato.

---

## FR-GIT-004 - Commit file modificati

### Priorità

MUST

### Descrizione

Il sistema deve creare commit GitLab con una o più modifiche file.

### Input

- branch;
- lista file;
- nuovo contenuto;
- commit message.

### Regole

- Il commit message deve includere Change ID.
- Il commit non deve contenere token o secret.
- Il sistema deve salvare SHA commit risultante.

### Esempio messaggio commit

```text
CHG-2026-0001 Update demo-go-color-app replicas
```

### Acceptance criteria

- Il commit viene creato tramite GitLab API.
- La ChangeRequest registra commit SHA.
- L’evento `FilesUpdated` o `CommitCreated` viene registrato.

---

## FR-GIT-005 - Apertura Merge Request

### Priorità

SHOULD

### Descrizione

Il sistema dovrebbe aprire una Merge Request GitLab per change soggetti a review.

### Output minimo

- MR IID;
- MR URL;
- source branch;
- target branch;
- stato MR.

### Acceptance criteria

- Il sistema crea MR tramite GitLab API.
- La ChangeRequest registra MR IID e URL.
- L’evento `MergeRequestOpened` viene registrato.

---

## FR-GIT-006 - Lettura ultimo commit

### Priorità

MUST

### Descrizione

Il sistema deve recuperare l’ultimo commit rilevante per repository, branch e path.

### Uso

- dettaglio Application;
- correlazione con Argo CD revision;
- audit change.

### Acceptance criteria

- Il sistema espone ultimo commit Git per Application.
- Il sistema distingue commit Git da revision Argo CD quando non coincidono.

---

# 7. ChangeRequest Management

## FR-CHG-001 - Creare ChangeRequest

### Priorità

MUST

### Descrizione

Il sistema deve creare una ChangeRequest per ogni change GitOps.

### Input minimo

```yaml
applicationName: demo-go-color-app
changeType: update-replicas
requestedBy: vmarzario
description: Scale applicazione per test
payload:
  replicas: 3
```

### Output

```yaml
changeId: CHG-2026-0001
status: Created
```

### Acceptance criteria

- La ChangeRequest viene salvata in PostgreSQL.
- Viene generato un Change ID univoco.
- Viene registrato evento `Created`.

---

## FR-CHG-002 - Aggiornare stato ChangeRequest

### Priorità

MUST

### Descrizione

Il sistema deve aggiornare lo stato della ChangeRequest durante il workflow.

### Stati iniziali

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

### Acceptance criteria

- Ogni cambio stato crea un evento in `change_events`.
- Non sono consentite transizioni incoerenti senza errore esplicito.

---

## FR-CHG-003 - Consultare ChangeRequest

### Priorità

MUST

### Descrizione

Il sistema deve permettere di consultare ChangeRequest e relativo dettaglio.

### Endpoint previsti

```text
GET /api/changes
GET /api/changes/{id}
```

### Acceptance criteria

- Lista change ordinabile per data e stato.
- Dettaglio change include eventi, Git, Tekton, Argo CD ed evidenze.

---

# 8. Change Workflows

## FR-WF-001 - Workflow update-replicas

### Priorità

MUST

### Descrizione

Il sistema deve implementare un workflow per aggiornare `spec.replicas`.

### Input

```yaml
changeType: update-replicas
replicas: 3
```

### Regole

- `replicas` deve essere intero >= 0.
- Il file target deve contenere un Deployment supportato.
- Il sistema deve aggiornare solo il manifest GitOps.

### Acceptance criteria

- Branch creato.
- File aggiornato.
- Diff generato.
- Commit o MR creato.
- Sync Argo CD completata.
- Deployment runtime raggiunge numero repliche desiderato.

---

## FR-WF-002 - Workflow update-app-version

### Priorità

MUST

### Descrizione

Il sistema deve implementare un workflow per aggiornare `APP_VERSION`.

### Regole

- Il sistema deve individuare se `APP_VERSION` è inline oppure in ConfigMap.
- Se è in ConfigMap, deve aggiornare `configmap.yaml`.
- Se è inline, deve aggiornare `deployment.yaml`.
- Il sistema deve registrare quale modalità è stata usata.

### Acceptance criteria

- Il valore viene modificato nel file corretto.
- La validazione Tekton passa.
- Il runtime mostra la nuova versione dopo sync/rollout.

---

## FR-WF-003 - Workflow update-page-color

### Priorità

MUST

### Descrizione

Il sistema deve implementare un workflow per aggiornare `PAGE_COLOR`.

### Regole

- Il valore deve rispettare formato `#[0-9A-Fa-f]{6}`.
- Il sistema deve individuare file target.
- Il sistema deve impedire valori vuoti o malformati.

### Acceptance criteria

- Valori invalidi vengono rifiutati.
- Il change passa validazione.
- Argo CD sincronizza.
- L’applicazione resta Healthy.

---

## FR-WF-004 - Workflow update-configmap-values

### Priorità

SHOULD

### Descrizione

Il sistema deve permettere modifica di chiavi in una ConfigMap GitOps.

### Regole

- La ConfigMap deve essere presente nel repo.
- Se il path usa Kustomize, la ConfigMap deve essere inclusa in `kustomization.yaml`.
- L’AppProject deve consentire `ConfigMap`.
- Il sistema deve segnalare se la modifica ConfigMap richiede rollout Pod.

### Acceptance criteria

- Il sistema modifica la ConfigMap corretta.
- Il sistema valida la policy AppProject.
- Il sistema raccoglie evidenza runtime.

---

# 9. Tekton Validation

## FR-TKN-001 - Creare PipelineRun di validazione

### Priorità

MUST

### Descrizione

Il sistema deve creare una `PipelineRun` Tekton tramite Kubernetes API.

### Input minimo

- repository;
- branch;
- commit;
- path GitOps;
- Change ID.

### Acceptance criteria

- La PipelineRun viene creata nel namespace configurato.
- Il nome PipelineRun viene salvato nella ChangeRequest.
- L’evento `ValidationRequested` viene registrato.

---

## FR-TKN-002 - Monitorare PipelineRun

### Priorità

MUST

### Descrizione

Il sistema deve monitorare stato PipelineRun fino a successo, fallimento o timeout.

### Stati gestiti

- Running;
- Succeeded;
- Failed;
- Timeout;
- Unknown.

### Acceptance criteria

- Stato finale salvato in PostgreSQL.
- Errori leggibili nel dettaglio ChangeRequest.
- Log principali associati come evidence.

---

## FR-TKN-003 - Raccogliere TaskRun e log

### Priorità

SHOULD

### Descrizione

Il sistema dovrebbe raccogliere TaskRun generate e log principali.

### Acceptance criteria

- Ogni TaskRun rilevante è associata alla PipelineRun.
- I log non contengono token.
- I log salvati sono consultabili come evidenza.

---

# 10. Argo CD Sync

## FR-ARGO-001 - Lanciare sync Application

### Priorità

MUST

### Descrizione

Il sistema deve invocare Argo CD API per sincronizzare una Application.

### Regole

- Il sistema deve verificare se esiste una operation già in corso.
- Il sistema deve gestire errore `another operation is already in progress`.
- Il sistema deve registrare revision sincronizzata.

### Acceptance criteria

- La sync viene invocata tramite Argo CD API.
- Stato `SyncRequested` e `SyncRunning` vengono registrati.
- Errore di sync viene registrato come `SyncFailed`.

---

## FR-ARGO-002 - Attendere Synced/Healthy

### Priorità

MUST

### Descrizione

Il sistema deve attendere che l’Application raggiunga stato desiderato.

### Condizioni

- `syncStatus=Synced`;
- `healthStatus=Healthy`.

### Acceptance criteria

- Il sistema distingue `OutOfSync`, `Degraded`, `Progressing` e timeout.
- Il sistema salva stato finale.
- Il sistema produce messaggio operativo leggibile.

---

## FR-ARGO-003 - Interpretare errori AppProject

### Priorità

SHOULD

### Descrizione

Il sistema deve interpretare errori Argo CD legati agli AppProject.

### Esempio

```text
resource :ConfigMap is not permitted in project devops-ci-demo
```

### Messaggio operativo

```text
La risorsa ConfigMap non è autorizzata dall’AppProject. Aggiungere group: "" kind: ConfigMap alla namespaceResourceWhitelist.
```

### Acceptance criteria

- Il sistema riconosce almeno errore ConfigMap non permessa.
- Il sistema suggerisce azione correttiva.

---

# 11. Evidence Collection

## FR-EVD-001 - Creare evidence summary

### Priorità

MUST

### Descrizione

Il sistema deve creare un riepilogo evidenze per ogni ChangeRequest.

### Contenuto minimo

- Change ID;
- application;
- type;
- branch;
- commit;
- MR;
- Tekton PipelineRun;
- Argo CD sync/health;
- Deployment status;
- Pods status;
- timestamp.

### Acceptance criteria

- Evidence summary salvato in PostgreSQL.
- Evidence consultabile via API.

---

## FR-EVD-002 - Raccogliere evidenze runtime

### Priorità

SHOULD

### Descrizione

Il sistema dovrebbe raccogliere stato runtime da Kubernetes/OpenShift.

### Risorse target

- Deployment;
- ReplicaSet;
- Pod;
- Service;
- Route;
- ConfigMap.

### Acceptance criteria

- Le evidenze runtime non contengono secret.
- Gli output sono associati alla ChangeRequest.

---

# 12. Error Handling

## FR-ERR-001 - Errori leggibili

### Priorità

MUST

### Descrizione

Il sistema deve trasformare errori tecnici in messaggi operativi comprensibili.

### Esempi

- ConfigMap non permessa da AppProject;
- YAML non valido;
- valueFrom non valido;
- operation Argo CD già in corso;
- token non valido;
- repository file non trovato;
- PipelineRun fallita.

### Acceptance criteria

- Ogni errore ha codice interno.
- Ogni errore ha messaggio tecnico e messaggio operativo.
- Gli errori collegati a ChangeRequest sono salvati negli eventi.

---

## FR-ERR-002 - Failure non distruttivo

### Priorità

MUST

### Descrizione

Un fallimento non deve lasciare il sistema in stato ambiguo.

### Regole

- Se GitLab branch creation fallisce, ChangeRequest passa a `Failed`.
- Se Tekton fallisce, ChangeRequest passa a `ValidationFailed`.
- Se Argo CD sync fallisce, ChangeRequest passa a `SyncFailed`.
- Ogni stato deve avere messaggio e timestamp.

---

# 13. Security and Secret Handling

## FR-SEC-001 - Gestione token

### Priorità

MUST

### Descrizione

Il sistema deve leggere token e credenziali da variabili ambiente o Secret montate.

### Regole

- Nessun token in Git.
- Nessun token nei log.
- Nessun token nelle evidenze.
- `.env` deve restare ignorato da Git.

---

## FR-SEC-002 - Anti-secret check

### Priorità

SHOULD

### Descrizione

Il sistema dovrebbe integrare un controllo anti-secret sui file modificati prima del commit/MR.

### Acceptance criteria

- Il controllo intercetta pattern comuni.
- Se rileva possibile secret, blocca il workflow o richiede review esplicita.

---

# 14. Web/HTML Foundation

## FR-WEB-001 - Predisposizione web server

### Priorità

SHOULD

### Descrizione

Il backend Go deve predisporre una struttura per HTML templates e Bootstrap.

### Pagine minime future

- dashboard applicazioni;
- dettaglio applicazione;
- lista change;
- dettaglio change;
- form change repliche;
- form change ConfigMap.

### Nota

La UI completa non è bloccante per i primi workflow API/backend.

---

# 15. Traceability Matrix

| Use Case | Requisiti principali |
|---|---|
| UC-001 Lista applicazioni | FR-APP-001, FR-APP-002 |
| UC-002 Dettaglio applicazione | FR-APP-010, FR-APP-011, FR-APP-012 |
| UC-003 Change repliche | FR-CHG-001, FR-GIT-003, FR-GIT-004, FR-WF-001, FR-ARGO-001, FR-EVD-001 |
| UC-004 Change APP_VERSION | FR-WF-002, FR-GIT-002, FR-GIT-004, FR-TKN-001, FR-ARGO-001 |
| UC-005 Change PAGE_COLOR | FR-WF-003, FR-GIT-002, FR-GIT-004, FR-TKN-001 |
| UC-006 Change ConfigMap | FR-WF-004, FR-ARGO-003, FR-GIT-002, FR-GIT-004 |
| UC-007 Validazione Tekton | FR-TKN-001, FR-TKN-002, FR-TKN-003 |
| UC-008 Sync Argo CD | FR-ARGO-001, FR-ARGO-002, FR-ARGO-003 |
| UC-009 Raccolta evidenze | FR-EVD-001, FR-EVD-002 |
| UC-010 Storico change | FR-CHG-002, FR-CHG-003 |

---

## 16. Criterio di completamento del documento

Questo documento sarà considerato stabile quando:

- tutti i requisiti MVP avranno priorità assegnata;
- ogni requisito avrà acceptance criteria;
- gli use cases principali saranno coperti;
- i requisiti saranno collegati al data model e all’API design;
- le decisioni architetturali correlate saranno riportate negli ADR.

---

## 17. Messaggio chiave

I requisiti funzionali del DevOps Control Plane devono mantenere il progetto concentrato su un MVP piccolo ma utile.

Il sistema deve prima dimostrare di saper eseguire un workflow completo:

```text
Application discovery
  -> ChangeRequest
  -> GitLab branch/commit/MR
  -> Tekton validation
  -> Argo CD sync
  -> Runtime validation
  -> Evidence collection
  -> Change history
```

Solo dopo la stabilizzazione di questo flusso sarà opportuno estendere UI, workflow avanzati e integrazioni enterprise.
