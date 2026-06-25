# DevOps Control Plane - Argo CD Integration

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
**Stato:** Draft iniziale / Argo CD Integration

---

## 1. Scopo del documento

Questo documento descrive come **DevOps Control Plane** deve integrarsi con **Argo CD / OpenShift GitOps**.

L’obiettivo è definire:

- responsabilità dell’integrazione Argo CD;
- funzionalità MVP da implementare;
- modello di autenticazione;
- modello dati normalizzato;
- operazioni API richieste;
- gestione degli stati `Synced`, `OutOfSync`, `Healthy`, `Degraded`;
- gestione di resources e orphaned resources;
- gestione delle sync operation;
- gestione errori noti;
- requisiti di sicurezza;
- mapping verso i requisiti funzionali già definiti.

DevOps Control Plane non sostituisce Argo CD. Argo CD resta il motore GitOps responsabile della riconciliazione tra stato desiderato Git e stato runtime nel cluster.

---

## 2. Ruolo di Argo CD nell’architettura

Nel progetto DevOps Control Plane, Argo CD ha il ruolo di **GitOps reconciliation engine**.

Responsabilità Argo CD:

- leggere manifest GitOps dal repository;
- confrontare stato Git con stato runtime cluster;
- segnalare `Synced` o `OutOfSync`;
- valutare health delle risorse;
- eseguire sync;
- mantenere history delle revisioni sincronizzate;
- mostrare conditions e warning, come `OrphanedResourceWarning`;
- impedire azioni non consentite dagli AppProject.

Responsabilità DevOps Control Plane:

- interrogare Argo CD;
- normalizzare le informazioni;
- mostrare stato e history;
- lanciare sync quando il workflow lo richiede;
- attendere stato finale;
- interpretare errori comuni;
- salvare evidenze in PostgreSQL;
- correlare stato Argo CD con GitLab, Tekton e runtime OpenShift.

---

## 3. Principi di integrazione

## 3.1 Argo CD resta il motore GitOps

DevOps Control Plane non applica direttamente manifest applicativi al cluster come workflow finale.

Il flusso corretto è:

```text
GitLab change
  -> merge/commit sul branch target
  -> Argo CD sync
  -> OpenShift runtime state
```

---

## 3.2 DevOps Control Plane usa Argo CD API

L’integrazione target deve usare **Argo CD API**, non parsing fragile dell’output CLI.

La CLI `argocd` resta utile per troubleshooting manuale e per verificare il comportamento atteso durante il lab, ma l’applicazione Go deve incapsulare l’accesso ad Argo CD in un adapter dedicato.

---

## 3.3 Argo CD history e Git history sono diverse

Il sistema deve distinguere tra:

- history GitLab, cioè commit e merge request;
- history Argo CD, cioè revisioni sincronizzate;
- history DevOps Control Plane, cioè ChangeRequest e workflow events.

Regola didattica:

```text
GitLab dice cosa è cambiato nei file.
Argo CD dice cosa è stato sincronizzato.
DevOps Control Plane dice perché, da chi e con quali evidenze.
```

---

## 3.4 OutOfSync non è sempre un errore applicativo

`OutOfSync` significa che lo stato runtime non coincide con lo stato desiderato Git.

Può indicare:

- drift non previsto;
- sync non ancora eseguita;
- sync fallita;
- rollback operativo temporaneo;
- risorsa runtime modificata manualmente;
- nuova revision Git non ancora applicata.

Il sistema deve mostrare `OutOfSync` in modo chiaro, ma deve evitare interpretazioni automatiche senza contesto.

---

## 4. Funzionalità MVP Argo CD

## 4.1 Lista Application

### Obiettivo

Recuperare la lista delle Application Argo CD visibili.

### Dati richiesti

Per ogni Application:

```yaml
name: demo-go-color-app
namespace: openshift-gitops
project: devops-ci-demo
destinationServer: https://kubernetes.default.svc
destinationNamespace: devops-ci-demo
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
targetRevision: main
path: apps/demo-go-color-app
syncStatus: Synced
healthStatus: Healthy
currentRevision: b8e1d6b
```

### Endpoint DevOps Control Plane

```text
GET /api/applications
```

### Requisiti collegati

```text
FR-APP-001
FR-APP-002
```

---

## 4.2 Dettaglio Application

### Obiettivo

Mostrare il dettaglio operativo di una Application.

### Dati richiesti

- metadata;
- project;
- source repo;
- source path;
- target revision;
- destination;
- sync policy;
- sync status;
- health status;
- current revision;
- conditions;
- operation state;
- resources summary.

### Endpoint DevOps Control Plane

```text
GET /api/applications/{name}
```

### Requisiti collegati

```text
FR-APP-010
FR-APP-011
FR-APP-012
```

---

## 4.3 Resources e orphaned resources

### Obiettivo

Recuperare le risorse collegate alla Application, distinguendo risorse gestite e risorse orphaned.

### Dati normalizzati

```yaml
resources:
  - group: ""
    kind: Service
    namespace: devops-ci-demo
    name: demo-go-color-app
    status: Synced
    health: Healthy
    orphaned: false
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

### Nota core API group

Risorse come `Service`, `ConfigMap`, `Secret` e `PersistentVolumeClaim` appartengono al core API group Kubernetes.

Nel modello Kubernetes hanno:

```yaml
apiVersion: v1
kind: Service
```

oppure:

```yaml
apiVersion: v1
kind: ConfigMap
```

Per questo motivo, nel campo `group`, il valore può essere vuoto.

### Endpoint DevOps Control Plane

```text
GET /api/applications/{name}/resources
```

---

## 4.4 Argo CD history

### Obiettivo

Recuperare la history di sync della Application.

### Dati normalizzati

```yaml
history:
  - id: 1
    revision: 4d6dfd4
    deployedAt: "2026-06-24T18:00:00+02:00"
    source: apps/demo-go-color-app
  - id: 2
    revision: 1cf3c91
    deployedAt: "2026-06-24T18:20:00+02:00"
    source: apps/demo-go-color-app
```

### Endpoint DevOps Control Plane

```text
GET /api/applications/{name}/history
```

### Nota importante

La history Argo CD non modifica Git. Un rollback operativo Argo CD può riportare temporaneamente il cluster a una revision precedente, ma Git resta sulla branch configurata, ad esempio `main`.

---

## 4.5 Sync Application

### Obiettivo

Avviare una sync Argo CD su una Application.

### Endpoint DevOps Control Plane

```text
POST /api/applications/{name}/sync
```

### Regole

- la sync deve essere invocata solo dopo commit/MR approvato o merge;
- il sistema deve registrare evento `SyncRequested`;
- il sistema deve gestire operation già in corso;
- il sistema deve salvare revision sincronizzata;
- il sistema deve produrre evidenza.

### Requisiti collegati

```text
FR-ARGO-001
FR-ARGO-002
```

---

## 4.6 Wait Synced/Healthy

### Obiettivo

Attendere che la Application raggiunga stato finale desiderato.

### Condizioni di successo

```text
syncStatus = Synced
healthStatus = Healthy
```

### Stati intermedi

```text
OutOfSync
Progressing
Unknown
Missing
Suspended
```

### Stati di errore

```text
Degraded
SyncFailed
OperationFailed
Timeout
```

### Regole

- timeout obbligatorio;
- polling configurabile;
- ogni cambio significativo crea evento;
- timeout produce errore leggibile;
- `Healthy` senza `Synced` non basta per considerare il workflow completato;
- `Synced` senza `Healthy` non basta per considerare il workflow completato.

---

## 5. Argo CD Adapter

## 5.1 Responsabilità

Il componente `ArgoCDAdapter` incapsula tutte le chiamate Argo CD API.

Il resto dell’applicazione non deve conoscere dettagli HTTP, endpoint specifici, token o payload raw Argo CD.

---

## 5.2 Interfaccia logica MVP

Interfaccia concettuale:

```go
type ArgoCDAdapter interface {
    ListApplications(ctx context.Context) ([]ApplicationSummary, error)
    GetApplication(ctx context.Context, name string) (*ApplicationDetail, error)
    GetApplicationResources(ctx context.Context, name string) ([]ApplicationResource, error)
    GetApplicationHistory(ctx context.Context, name string) ([]ApplicationHistoryItem, error)
    SyncApplication(ctx context.Context, name string, opts SyncOptions) (*SyncResult, error)
    GetOperationState(ctx context.Context, name string) (*OperationState, error)
    WaitForSyncedHealthy(ctx context.Context, name string, opts WaitOptions) (*ApplicationStatus, error)
}
```

Nota: l’interfaccia è indicativa. La forma finale sarà definita durante l’implementazione Go.

---

## 5.3 Posizione package proposta

```text
internal/adapters/argocd/
├── client.go
├── models.go
├── mapper.go
├── errors.go
└── wait.go
```

### File `client.go`

Responsabilità:

- configurazione HTTP client;
- autenticazione;
- chiamate API;
- timeout;
- retry controllati, se previsti.

### File `models.go`

Responsabilità:

- strutture DTO Argo CD raw o parziali;
- strutture normalizzate interne.

### File `mapper.go`

Responsabilità:

- convertire payload Argo CD in domain model DevOps Control Plane.

### File `errors.go`

Responsabilità:

- mapping errori Argo CD in error model interno.

### File `wait.go`

Responsabilità:

- polling sync/health;
- polling operation state;
- gestione timeout.

---

## 6. Modello dati normalizzato

## 6.1 ApplicationSummary

```yaml
name: demo-go-color-app
namespace: openshift-gitops
project: devops-ci-demo
targetNamespace: devops-ci-demo
repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
targetRevision: main
path: apps/demo-go-color-app
syncStatus: Synced
healthStatus: Healthy
currentRevision: b8e1d6b
```

---

## 6.2 ApplicationDetail

```yaml
name: demo-go-color-app
namespace: openshift-gitops
project: devops-ci-demo
syncPolicy: Automated
pruneEnabled: true
selfHealEnabled: false
syncStatus: Synced
healthStatus: Healthy
currentRevision: b8e1d6b
source:
  repoUrl: https://gitlab.example.local/group/demo-app-gitops.git
  targetRevision: main
  path: apps/demo-go-color-app
destination:
  server: https://kubernetes.default.svc
  namespace: devops-ci-demo
conditions:
  - type: OrphanedResourceWarning
    message: Application has 2 orphaned resources
```

---

## 6.3 ApplicationResource

```yaml
group: apps
kind: Deployment
namespace: devops-ci-demo
name: demo-go-color-app
status: Synced
health: Healthy
orphaned: false
message: ""
```

Per core API group:

```yaml
group: ""
kind: ConfigMap
namespace: devops-ci-demo
name: demo-go-color-app-config
status: Synced
health: Unknown
orphaned: false
```

---

## 6.4 ApplicationHistoryItem

```yaml
id: 3
revision: b8e1d6b2fc88b41909f87d505739bc855de41516
deployedAt: "2026-06-25T11:04:50+02:00"
sourcePath: apps/demo-go-color-app
```

---

## 6.5 OperationState

```yaml
phase: Failed
message: "one or more synchronization tasks completed unsuccessfully"
startedAt: "2026-06-25T11:04:50+02:00"
finishedAt: "2026-06-25T11:04:53+02:00"
syncRevision: b8e1d6b2fc88b41909f87d505739bc855de41516
```

---

## 7. Autenticazione verso Argo CD

## 7.1 Token based authentication

Il DevOps Control Plane deve autenticarsi ad Argo CD tramite token.

Configurazione prevista:

```text
ARGOCD_BASE_URL=https://openshift-gitops-server-openshift-gitops.apps.example.local
ARGOCD_AUTH_TOKEN=<from Secret>
ARGOCD_INSECURE_TLS=false
ARGOCD_TIMEOUT_SECONDS=30
```

### Regole

- `ARGOCD_AUTH_TOKEN` deve arrivare da Secret;
- il token non deve essere loggato;
- il token non deve essere esposto via API;
- errori HTTP 401/403 devono essere normalizzati.

---

## 7.2 Gestione TLS

In ambienti lab può essere necessario gestire certificati non trusted o route reencrypt.

Configurazione proposta:

```text
ARGOCD_INSECURE_TLS=false
ARGOCD_CA_CERT_PATH=/var/run/secrets/argocd/ca.crt
```

Regola:

- in produzione preferire CA/certificati trusted;
- opzione insecure solo per lab controllati;
- se insecure è abilitato, loggare warning senza stampare dati sensibili.

---

## 8. Mapping stati Argo CD

## 8.1 Sync status

| Stato Argo CD | Significato operativo |
|---|---|
| Synced | Runtime allineato allo stato desiderato Git |
| OutOfSync | Runtime diverso dallo stato desiderato Git |
| Unknown | Stato non determinato |

---

## 8.2 Health status

| Stato Argo CD | Significato operativo |
|---|---|
| Healthy | Risorse considerate sane |
| Progressing | Risorse in aggiornamento |
| Degraded | Risorse in errore o non sane |
| Missing | Risorsa attesa mancante |
| Suspended | Risorsa sospesa |
| Unknown | Stato non determinato |

---

## 8.3 Interpretazione combinata

| Sync | Health | Interpretazione |
|---|---|---|
| Synced | Healthy | Stato desiderato raggiunto |
| OutOfSync | Healthy | Applicazione sana ma non allineata a Git |
| Synced | Degraded | Manifest applicato ma runtime non sano |
| OutOfSync | Degraded | Drift o sync incompleta con runtime non sano |
| Unknown | Unknown | Stato non determinabile |

---

## 9. Conditions e warning

## 9.1 OrphanedResourceWarning

Argo CD può segnalare:

```text
OrphanedResourceWarning
```

Questo warning indica che nel namespace target esistono risorse non gestite dalla Application.

Esempio:

```text
ImageStream demo-go-color-app orphaned
```

### Regola DevOps Control Plane

- Mostrare il warning.
- Non considerarlo automaticamente errore se `Health=Healthy` e `Sync=Synced`.
- Salvare il warning nelle evidenze.
- Collegarlo alla sezione didattica del dettaglio Application.

---

## 9.2 Resource not permitted by AppProject

Errore tipico:

```text
resource :ConfigMap is not permitted in project devops-ci-demo
```

Interpretazione:

```text
La risorsa è presente nel repository GitOps, ma l’AppProject non consente alla Application di gestirla.
```

Messaggio operativo:

```text
Aggiungere group: "" kind: ConfigMap alla namespaceResourceWhitelist dell’AppProject devops-ci-demo.
```

### Regola DevOps Control Plane

- Riconoscere l’errore.
- Classificarlo come `ARGO_RESOURCE_NOT_PERMITTED`.
- Suggerire correzione.
- Registrare evento `SyncFailed`.

---

## 10. Sync operation management

## 10.1 Avvio sync

Prima di avviare una sync, il sistema deve:

1. leggere stato attuale Application;
2. verificare operation in corso;
3. registrare evento `SyncRequested`;
4. invocare sync;
5. registrare evento `SyncRunning`.

---

## 10.2 Operation already in progress

Errore tipico:

```text
another operation is already in progress
```

Gestione:

- classificare come `ARGO_OPERATION_IN_PROGRESS`;
- non lanciare sync duplicata;
- suggerire wait o terminate manuale, se previsto;
- registrare evento nel workflow.

Nota MVP:

DevOps Control Plane, nel primo MVP, dovrebbe limitarsi a segnalare e attendere. La terminazione automatica di operation può essere valutata in futuro con policy esplicita.

---

## 10.3 Polling operation state

Il sistema deve leggere periodicamente lo stato operation.

Stati possibili:

```text
Running
Succeeded
Failed
Error
Terminating
```

Regole:

- timeout obbligatorio;
- polling interval configurabile;
- ogni transizione significativa può generare evento;
- failure deve salvare messaggio Argo CD.

---

## 10.4 Wait Synced/Healthy

Dopo sync succeeded, il sistema deve attendere:

```text
syncStatus=Synced
healthStatus=Healthy
```

Se entro il timeout lo stato non viene raggiunto:

- registrare `SyncFailed` o `SyncTimeout`;
- salvare stato finale osservato;
- salvare resources non sane o OutOfSync;
- generare messaggio operativo.

---

## 11. Rollback Argo CD

## 11.1 Regola generale

Il rollback Argo CD è operativo e non modifica Git.

Nel modello DevOps Control Plane, il rollback definitivo deve essere preferibilmente gestito tramite GitLab:

```text
git revert / MR di revert
  -> merge
  -> Argo CD sync
```

## 11.2 Supporto MVP

Nel primo MVP:

- leggere history Argo CD: SHOULD;
- mostrare revision precedenti: SHOULD;
- eseguire rollback Argo CD automatico: COULD / fuori dal flusso principale;
- proporre rollback GitLab: SHOULD in roadmap.

## 11.3 Nota didattica

Se un rollback Argo CD porta il cluster a una revision precedente, Git può restare su `main` corrente.

In questa situazione la Application può risultare:

```text
OutOfSync from main
Healthy
```

Questo significa:

```text
runtime sano, ma non allineato al target Git corrente.
```

---

## 12. AppProject awareness

## 12.1 Perché serve

Molti errori di sync derivano da policy AppProject.

Esempio:

- Application introduce ConfigMap;
- AppProject consente Service, Deployment e Route;
- AppProject non consente ConfigMap;
- sync fallisce.

---

## 12.2 Requisito MVP

Il sistema dovrebbe rilevare almeno alcuni errori AppProject noti, tra cui:

```text
resource :ConfigMap is not permitted in project ...
```

Nel futuro può interrogare direttamente AppProject tramite Kubernetes API o Argo CD API, se disponibile nel modello scelto.

---

## 12.3 Messaggio operativo

```text
La risorsa ConfigMap non è autorizzata dall’AppProject.
Verificare namespaceResourceWhitelist e aggiungere:

- group: ""
  kind: ConfigMap
```

---

## 13. Error model Argo CD

## 13.1 Codici errore interni

```text
ARGO_AUTH_FAILED
ARGO_FORBIDDEN
ARGO_APPLICATION_NOT_FOUND
ARGO_LIST_FAILED
ARGO_GET_FAILED
ARGO_RESOURCES_FAILED
ARGO_HISTORY_FAILED
ARGO_SYNC_FAILED
ARGO_SYNC_TIMEOUT
ARGO_OPERATION_IN_PROGRESS
ARGO_RESOURCE_NOT_PERMITTED
ARGO_APPLICATION_DEGRADED
ARGO_APPLICATION_OUTOFSYNC_TIMEOUT
ARGO_UNKNOWN_ERROR
```

---

## 13.2 Struttura errore normalizzata

```yaml
code: ARGO_RESOURCE_NOT_PERMITTED
technicalMessage: "resource :ConfigMap is not permitted in project devops-ci-demo"
userMessage: "La ConfigMap non è consentita dall’AppProject devops-ci-demo."
suggestedAction: "Aggiungere group: \"\" kind: ConfigMap alla namespaceResourceWhitelist."
recoverable: true
```

---

## 14. Evidence Argo CD

Per ogni sync o dettaglio change, il sistema deve poter salvare evidenze Argo CD.

## 14.1 Evidence minima

```yaml
application: demo-go-color-app
project: devops-ci-demo
syncStatus: Synced
healthStatus: Healthy
revision: b8e1d6b
syncPolicy: Automated
conditions:
  - type: OrphanedResourceWarning
    message: Application has 2 orphaned resources
resources:
  - kind: Deployment
    name: demo-go-color-app
    status: Synced
    health: Healthy
```

---

## 14.2 Evidence in caso di errore

```yaml
application: demo-go-color-app
syncStatus: OutOfSync
healthStatus: Healthy
operationPhase: Failed
errorCode: ARGO_RESOURCE_NOT_PERMITTED
message: "resource :ConfigMap is not permitted in project devops-ci-demo"
```

---

## 15. Configurazione Argo CD

Variabili previste:

```text
ARGOCD_BASE_URL=https://openshift-gitops-server-openshift-gitops.apps.example.local
ARGOCD_AUTH_TOKEN=<from secret>
ARGOCD_INSECURE_TLS=false
ARGOCD_CA_CERT_PATH=
ARGOCD_TIMEOUT_SECONDS=30
ARGOCD_POLL_INTERVAL_SECONDS=5
ARGOCD_DEFAULT_NAMESPACE=openshift-gitops
```

### Regole

- token da Secret;
- base URL da ConfigMap o env;
- TLS insecure solo per lab;
- timeout obbligatorio.

---

## 16. API interne DevOps Control Plane collegate ad Argo CD

## 16.1 Applications

```text
GET /api/applications
GET /api/applications/{name}
GET /api/applications/{name}/resources
GET /api/applications/{name}/history
```

---

## 16.2 Sync

```text
POST /api/applications/{name}/sync
GET  /api/applications/{name}/operation
POST /api/changes/{id}/sync
```

---

## 16.3 Evidence

```text
GET /api/changes/{id}/evidence/argocd
```

---

## 17. Security requirements specifici

- Non loggare `ARGOCD_AUTH_TOKEN`.
- Non restituire token nelle API.
- Non salvare token in PostgreSQL.
- Non includere token nelle evidenze.
- Usare timeout su tutte le chiamate.
- Limitare il token Argo CD alle Application/progetti richiesti, se possibile.
- Preferire token applicativo/servizio rispetto a credenziali admin.

---

## 18. Testing strategy

## 18.1 Unit test

Testare:

- mapping Application raw -> normalized model;
- mapping resources;
- mapping health/sync status;
- mapping errori noti;
- wait loop con fake client.

---

## 18.2 Integration test manuale MVP

Scenario minimo:

1. configurare Argo CD base URL e token;
2. chiamare `GET /api/applications`;
3. verificare `demo-go-color-app` presente;
4. chiamare dettaglio;
5. chiamare resources;
6. verificare orphaned resources;
7. lanciare sync su Application già Synced;
8. verificare gestione stato finale.

---

## 18.3 Casi errore da simulare

- token errato;
- Application inesistente;
- operation già in corso;
- risorsa non permessa da AppProject;
- Application Degraded;
- timeout wait Synced/Healthy.

---

## 19. MVP implementation order per Argo CD adapter

Ordine consigliato:

1. configurazione Argo CD;
2. HTTP client base;
3. autenticazione token;
4. `ListApplications`;
5. `GetApplication`;
6. mapping sync/health/revision;
7. resources;
8. history;
9. sync;
10. operation state;
11. wait Synced/Healthy;
12. mapping errori;
13. evidence builder.

---

## 20. Checklist di completamento integrazione Argo CD

La prima implementazione Argo CD sarà considerata pronta quando:

- il backend legge configurazione Argo CD;
- il token è letto da Secret/env e non loggato;
- la lista Application funziona;
- il dettaglio Application funziona;
- resources e orphaned resources sono visibili;
- history è disponibile o gestita come funzione opzionale;
- sync può essere invocata;
- wait Synced/Healthy funziona con timeout;
- errori noti sono normalizzati;
- evidence Argo CD è salvabile in PostgreSQL;
- i test manuali sono documentati.

---

## 21. Relazione con documenti successivi

Questo documento alimenta:

- `docs/10-data-model.md`, per entità Application, ArgoStatus e Evidence;
- `docs/11-change-workflows.md`, per i workflow sync e wait;
- `docs/13-api-design.md`, per endpoint applicativi;
- `docs/09-security-rbac.md`, per token e permessi;
- ADR specifico `ADR-0002-argocd-as-gitops-engine.md`.

---

## 22. Messaggio chiave

L’integrazione Argo CD deve essere robusta, ma semplice.

DevOps Control Plane deve usare Argo CD per ciò che Argo CD fa meglio:

```text
confrontare Git e cluster
sincronizzare manifest
mostrare health
mostrare history
segnalare drift e warning
```

DevOps Control Plane aggiunge valore correlando queste informazioni con GitLab, Tekton, OpenShift runtime e storico ChangeRequest.
