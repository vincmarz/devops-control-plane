# DevOps Control Plane - MVP Scope

**Versione:** 0.1  
**Data:** 2026-06-25  
**Owner iniziale:** Vincenzo Marzario  
**Repository:** `https://github.com/vincmarz/devops-control-plane`  
**Documento precedente:** `docs/00-vision.md`  
**Stato:** Draft iniziale / Scope MVP

---

## 1. Scopo del documento

Questo documento definisce il perimetro del primo MVP del progetto **DevOps Control Plane**.

L’obiettivo è stabilire in modo chiaro:

- quali funzionalità implementare nella prima versione;
- quali funzionalità lasciare fuori dal primo MVP;
- quali integrazioni sono necessarie;
- quali workflow devono essere supportati;
- quali criteri useremo per dire che l’MVP è completato;
- quale sequenza di milestone seguire per non perdere informazioni e non costruire componenti premature.

Il documento deve essere usato come riferimento operativo durante la progettazione e l’implementazione del backend, delle API, degli adapter e dei primi workflow GitOps.

---

## 2. Definizione sintetica dell’MVP

Il primo MVP di **DevOps Control Plane** deve permettere a un operatore DevOps di:

1. visualizzare le applicazioni Argo CD gestite;
2. consultare stato GitOps e runtime essenziale;
3. creare un change GitOps semplice;
4. generare branch GitLab;
5. modificare manifest GitOps in modo controllato;
6. validare il change tramite Tekton;
7. produrre commit o merge request;
8. sincronizzare Argo CD dopo merge o commit approvato;
9. attendere `Synced` e `Healthy`;
10. raccogliere evidenze;
11. conservare uno storico interno del change in PostgreSQL.

L’MVP non deve essere un sostituto di Argo CD, GitLab o Tekton. Deve essere un livello di orchestrazione, governance e audit sopra questi strumenti.

---

## 3. Stack tecnologico vincolante per l’MVP

Le decisioni tecnologiche iniziali sono le seguenti.

### 3.1 Backend

```text
Go
```

Il backend sarà responsabile di:

- esposizione API HTTP;
- orchestrazione workflow;
- integrazione con GitLab API;
- integrazione con Argo CD API;
- integrazione con Kubernetes API per Tekton;
- persistenza su PostgreSQL;
- raccolta evidenze;
- predisposizione rendering HTML templates.

---

### 3.2 Frontend MVP

```text
Go HTML templates + Bootstrap
```

Il progetto predisporrà fin dall’inizio una struttura web server-side con template HTML e Bootstrap.

Tuttavia, la Web UI completa arriverà dopo la stabilizzazione dei workflow.

Nel primo MVP potranno essere presenti pagine minime di debug o consultazione, ma la priorità resta:

```text
API + workflow stabili prima della UI completa
```

---

### 3.3 Database

```text
PostgreSQL
```

PostgreSQL conserverà:

- applicazioni note;
- change request;
- eventi di workflow;
- riferimenti Git;
- riferimenti Tekton;
- riferimenti Argo CD;
- evidenze;
- stato finale del change.

---

### 3.4 Integrazione Git

```text
GitLab API
```

L’MVP userà GitLab API per:

- leggere file dal repository;
- leggere commit;
- creare branch;
- aggiornare file;
- creare commit;
- creare merge request;
- leggere stato merge request.

Nota: il repository sorgente del DevOps Control Plane può essere ospitato su GitHub nell’account `vincmarz`, ma l’applicazione implementerà l’integrazione target tramite GitLab API.

---

### 3.5 Integrazione Argo CD

```text
Argo CD API
```

L’MVP userà Argo CD API per:

- listare applicazioni;
- leggere dettaglio Application;
- leggere sync status;
- leggere health status;
- leggere current revision;
- leggere resources;
- leggere orphaned resources;
- leggere history;
- lanciare sync;
- monitorare stato fino a `Synced` e `Healthy`.

---

### 3.6 Integrazione Tekton

```text
Kubernetes API diretta
```

L’MVP userà Kubernetes API diretta per operare sulle risorse Tekton:

- creazione `PipelineRun`;
- osservazione stato `PipelineRun`;
- osservazione stato `TaskRun`;
- raccolta log;
- raccolta esito validazione.

---

## 4. Funzionalità incluse nell’MVP

## 4.1 F1 - Lista applicazioni Argo CD

### Descrizione

Il sistema deve visualizzare le applicazioni Argo CD disponibili per il progetto/ambiente configurato.

### Dati minimi da mostrare

- nome Application;
- namespace Argo CD;
- progetto Argo CD;
- namespace target;
- repository Git;
- path GitOps;
- target revision;
- sync status;
- health status;
- current revision.

### Output atteso

Esempio logico:

```text
Application          Project          Sync      Health    Revision
-------------------  ---------------  --------  --------  --------
demo-go-color-app    devops-ci-demo   Synced    Healthy   b8e1d6b
demo-app             devops-ci-demo   Synced    Healthy   62d3b18
```

### Acceptance criteria

La funzionalità è completa quando:

- il backend recupera la lista da Argo CD API;
- il backend normalizza i campi principali;
- il backend espone `GET /api/applications`;
- la risposta può essere salvata o sincronizzata in PostgreSQL;
- eventuali errori Argo CD sono restituiti in modo leggibile.

---

## 4.2 F2 - Dettaglio applicazione

### Descrizione

Il sistema deve mostrare il dettaglio operativo di una Application Argo CD.

### Dati minimi

- sync status;
- health status;
- current revision;
- repo URL;
- target revision;
- Git path;
- namespace target;
- resources gestite;
- orphaned resources;
- history Argo CD;
- ultimo commit Git rilevante.

### Acceptance criteria

La funzionalità è completa quando:

- il backend espone `GET /api/applications/{name}`;
- il backend espone `GET /api/applications/{name}/resources`;
- il backend distingue risorse gestite e orphaned;
- il backend espone almeno l’ultima revisione Git conosciuta;
- lo stato risulta coerente con quanto mostrato da Argo CD.

---

## 4.3 F3 - Change repliche

### Descrizione

Il sistema deve consentire un change GitOps per modificare il numero di repliche di un Deployment.

### Input utente

- Application;
- Deployment target;
- numero repliche desiderato;
- descrizione/motivazione del change.

### Azioni automatiche

1. leggere metadata applicazione;
2. leggere il file GitOps corretto tramite GitLab API;
3. creare branch;
4. modificare `spec.replicas`;
5. produrre diff;
6. creare commit o merge request;
7. lanciare validazione Tekton;
8. dopo merge, sincronizzare Argo CD;
9. raccogliere evidenze.

### Acceptance criteria

La funzionalità è completa quando:

- il sistema produce una modifica Git tracciata;
- il sistema non modifica direttamente il Deployment runtime;
- il deployment raggiunge il numero di repliche desiderato dopo sync Argo CD;
- la ChangeRequest viene salvata in PostgreSQL con stato finale `Completed` o `Failed`.

---

## 4.4 F4 - Change APP_VERSION

### Descrizione

Il sistema deve permettere di aggiornare il valore applicativo `APP_VERSION`.

### Modalità supportate

L’MVP deve supportare almeno una delle due modalità:

1. `APP_VERSION` definita come env inline nel `Deployment`;
2. `APP_VERSION` definita in una `ConfigMap` referenziata dal `Deployment`.

La modalità preferita, dopo il refactoring del Lab GO, è:

```text
ConfigMap-managed APP_VERSION
```

### Acceptance criteria

La funzionalità è completa quando:

- il sistema individua dove è definita `APP_VERSION`;
- il sistema modifica il file corretto;
- il sistema produce branch/commit/MR;
- Tekton valida il manifest;
- Argo CD sincronizza;
- il Deployment risultante espone la nuova versione.

---

## 4.5 F5 - Change PAGE_COLOR

### Descrizione

Il sistema deve permettere di aggiornare il valore applicativo `PAGE_COLOR`.

### Validazioni minime

Il valore deve essere validato come colore esadecimale, ad esempio:

```text
#28A745
#1E90FF
#FF0000
```

### Acceptance criteria

La funzionalità è completa quando:

- il sistema impedisce valori non validi;
- il sistema aggiorna il file GitOps corretto;
- il change passa la validazione Tekton;
- Argo CD applica la modifica;
- l’applicazione resta `Healthy`.

---

## 4.6 F6 - Change ConfigMap values

### Descrizione

Il sistema deve permettere di modificare chiavi/valori in una ConfigMap GitOps.

### Esempio

```yaml
data:
  PAGE_COLOR: "#28A745"
  APP_VERSION: "v2-green"
```

### Validazioni minime

- la ConfigMap deve esistere nel repository;
- la ConfigMap deve essere inclusa in `kustomization.yaml`, se il path usa Kustomize;
- l’AppProject deve consentire `ConfigMap`;
- il Deployment deve referenziare correttamente le chiavi, se applicabile.

### Nota importante

La modifica di una ConfigMap non sempre causa automaticamente rollout dei Pod.

Se la ConfigMap è usata come variabile ambiente, i Pod devono essere ricreati per acquisire i nuovi valori. Nel futuro sarà necessario decidere una strategia, per esempio:

- aggiornare annotation nel Pod template;
- gestire checksum ConfigMap;
- proporre rollout restart controllato;
- rendere esplicita la necessità di un nuovo rollout.

Questa decisione sarà formalizzata in un ADR dedicato.

---

## 4.7 F7 - Validazione YAML con Tekton

### Descrizione

Il sistema deve delegare a Tekton la validazione automatica del change prima della promozione.

### PipelineRun prevista

Nome logico:

```text
validate-gitops-change
```

### Task minime

- clone repository/branch;
- controllo presenza file modificati;
- YAML lint, se disponibile;
- `kustomize build`, se applicabile;
- `oc apply --dry-run=server`;
- anti-secret check;
- verifica policy AppProject, se applicabile;
- report finale.

### Acceptance criteria

La funzionalità è completa quando:

- il backend crea una `PipelineRun` tramite Kubernetes API;
- il backend segue lo stato della `PipelineRun`;
- il backend raccoglie esito e log principali;
- la ChangeRequest viene aggiornata con risultato Tekton.

---

## 4.8 F8 - Sync Argo CD dopo merge o commit approvato

### Descrizione

Dopo il merge della MR o dopo il commit approvato, il sistema deve sincronizzare Argo CD.

### Azioni

- chiamare sync tramite Argo CD API;
- attendere completamento operazione;
- attendere stato `Synced`;
- attendere stato `Healthy`;
- raccogliere eventuali errori.

### Acceptance criteria

La funzionalità è completa quando:

- il sistema sa distinguere `Synced`, `OutOfSync`, `Healthy`, `Degraded`;
- il sistema salva l’esito della sync;
- il sistema salva la revisione Argo CD applicata;
- il sistema gestisce errori di sync in modo leggibile.

---

## 4.9 F9 - Raccolta evidenze

### Descrizione

Il sistema deve raccogliere evidenze tecniche per ogni change.

### Evidenze minime

- stato Application Argo CD;
- revisione Git;
- link branch/MR/commit;
- output validazione Tekton;
- stato Deployment;
- stato Pod;
- stato Service/Route, se applicabile;
- eventuale health check HTTP;
- log sintetico del workflow.

### Acceptance criteria

La funzionalità è completa quando:

- ogni ChangeRequest ha un set di evidenze associato;
- le evidenze sono consultabili via API;
- le evidenze restano disponibili anche dopo il completamento del change.

---

## 4.10 F10 - Storico change interno

### Descrizione

Il sistema deve mantenere una history funzionale dei change.

### Dati minimi

- Change ID;
- Application;
- tipo change;
- richiedente;
- stato;
- branch;
- commit;
- MR;
- PipelineRun;
- revisione Argo CD;
- timestamp creazione/completamento;
- esito finale;
- rollback suggerito.

### Acceptance criteria

La funzionalità è completa quando:

- il backend espone `GET /api/changes`;
- il backend espone `GET /api/changes/{id}`;
- ogni change registra eventi di avanzamento;
- un operatore può ricostruire cosa è successo senza leggere manualmente Git, Argo CD e Tekton separatamente.

---

## 5. Funzionalità escluse dal primo MVP

Sono escluse dal primo MVP:

- gestione multi-cluster avanzata;
- gestione multi-tenant enterprise;
- approval workflow complessi;
- integrazione ITSM;
- supporto ServiceNow/Jira;
- gestione Secret applicativi;
- rotazione credenziali;
- Helm values management avanzato;
- gestione completa AppProject;
- editor YAML visuale avanzato;
- AI assistant integrato;
- RBAC fine-grained per organizzazione;
- audit compliance completo;
- sostituzione UI Argo CD;
- sostituzione Tekton Dashboard;
- supporto GitHub/Gitea/Bitbucket come provider target.

Queste funzionalità potranno essere considerate in roadmap successive.

---

## 6. Requisiti non funzionali MVP

### 6.1 Sicurezza

- I token GitLab devono essere gestiti come Secret.
- I token Argo CD devono essere gestiti come Secret.
- Le password PostgreSQL devono essere gestite come Secret.
- Nessun token deve essere scritto nei log applicativi.
- Nessun token deve essere salvato in Git.
- Tutti i change devono essere tracciati.

---

### 6.2 Auditabilità

Ogni ChangeRequest deve avere:

- utente/richiedente;
- timestamp;
- input iniziale;
- eventi workflow;
- risultato validazione;
- risultato sync;
- evidenze;
- stato finale.

---

### 6.3 Ripetibilità

Un workflow deve essere ripetibile e documentabile.

Se una validazione fallisce, il sistema deve registrare:

- fase fallita;
- errore;
- suggerimento tecnico, se disponibile;
- log o payload rilevante.

---

### 6.4 Separazione responsabilità

- GitLab gestisce repository, branch, commit e MR.
- Argo CD gestisce riconciliazione GitOps.
- Tekton gestisce validazione workflow.
- PostgreSQL gestisce storico funzionale.
- DevOps Control Plane orchestra e correla.

---

## 7. Modello stati ChangeRequest

Stati iniziali previsti:

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

### Regola generale

Ogni transizione deve aggiungere un record in `change_events`.

---

## 8. API MVP previste

### Applications

```text
GET /api/applications
GET /api/applications/{name}
GET /api/applications/{name}/resources
GET /api/applications/{name}/history
GET /api/applications/{name}/git/commits
```

### Change Requests

```text
POST /api/changes
GET  /api/changes
GET  /api/changes/{id}
POST /api/changes/{id}/create-branch
POST /api/changes/{id}/update-files
POST /api/changes/{id}/validate
POST /api/changes/{id}/open-merge-request
POST /api/changes/{id}/sync
POST /api/changes/{id}/collect-evidence
POST /api/changes/{id}/cancel
```

### Health

```text
GET /healthz
GET /readyz
```

---

## 9. Milestone MVP

## 9.1 Milestone 0.1 - Project bootstrap

### Obiettivo

Preparare repository, struttura progetto e backend minimo.

### Deliverable

- struttura directory;
- `docs/00-vision.md`;
- `docs/01-scope-mvp.md`;
- Go module;
- endpoint `/healthz`;
- configurazione base;
- Dockerfile iniziale;
- connessione PostgreSQL;
- migration iniziale.

### Criterio di completamento

Il servizio parte localmente ed espone:

```text
GET /healthz -> ok
```

---

## 9.2 Milestone 0.2 - PostgreSQL change store

### Obiettivo

Creare modello dati iniziale per ChangeRequest.

### Deliverable

- migration `applications`;
- migration `change_requests`;
- migration `change_events`;
- migration `evidences`;
- repository Go per persistenza;
- API base per change.

### Criterio di completamento

È possibile creare e leggere una ChangeRequest via API.

---

## 9.3 Milestone 0.3 - Argo CD API discovery

### Obiettivo

Integrare Argo CD API per listare applicazioni e stato.

### Deliverable

- Argo CD client;
- lista Application;
- dettaglio Application;
- sync/health/revision;
- resources;
- orphaned resources.

### Criterio di completamento

`GET /api/applications` restituisce almeno una Application reale con stato coerente con Argo CD.

---

## 9.4 Milestone 0.4 - GitLab read operations

### Obiettivo

Integrare GitLab API in sola lettura.

### Deliverable

- GitLab client;
- lettura file;
- lettura commit;
- lettura branch;
- mapping Application -> repository/path.

### Criterio di completamento

Il sistema mostra ultimo commit e file GitOps principali per una Application.

---

## 9.5 Milestone 0.5 - Primo workflow change repliche

### Obiettivo

Implementare il primo change end-to-end.

### Deliverable

- creazione ChangeRequest;
- creazione branch GitLab;
- modifica `replicas`;
- commit o MR;
- stato change aggiornato;
- sync Argo CD;
- evidenza finale.

### Criterio di completamento

Un operatore può modificare le repliche di una applicazione tramite DevOps Control Plane senza usare manualmente `git`, `argocd` o `oc` come workflow principale.

---

## 9.6 Milestone 0.6 - Tekton validation

### Obiettivo

Integrare validazione tramite Tekton.

### Deliverable

- template `PipelineRun`;
- creazione via Kubernetes API;
- polling stato;
- raccolta log;
- salvataggio esito.

### Criterio di completamento

Ogni change può essere validato da una PipelineRun Tekton prima di essere promosso.

---

## 10. Acceptance criteria globali MVP

L’MVP è considerato completato quando:

1. il servizio Go parte e si connette a PostgreSQL;
2. il servizio espone endpoint health;
3. il servizio lista Application Argo CD;
4. il servizio mostra dettaglio Application;
5. il servizio legge dati GitLab almeno in modalità read;
6. il servizio crea ChangeRequest;
7. il servizio esegue almeno un workflow di change repliche;
8. il servizio registra eventi change;
9. il servizio invoca una validazione Tekton;
10. il servizio sincronizza Argo CD;
11. il servizio raccoglie evidenze;
12. il servizio mantiene uno storico consultabile.

---

## 11. Rischi principali

### 11.1 Complessità workflow

Rischio:

```text
Costruire troppi workflow prima di stabilizzare il primo.
```

Mitigazione:

```text
Partire dal change repliche, poi estendere a APP_VERSION, PAGE_COLOR e ConfigMap.
```

---

### 11.2 Integrazioni premature

Rischio:

```text
Implementare UI o funzioni enterprise prima che gli adapter siano affidabili.
```

Mitigazione:

```text
API e workflow prima della UI completa.
```

---

### 11.3 Drift GitOps

Rischio:

```text
Permettere modifiche runtime dirette che non passano da Git.
```

Mitigazione:

```text
Bloccare modifiche runtime come workflow permanente.
```

---

### 11.4 Gestione credenziali

Rischio:

```text
Token GitLab o Argo CD esposti in log o repository.
```

Mitigazione:

```text
Secret, masking log, anti-secret check e revisione sicurezza.
```

---

## 12. Ordine consigliato di implementazione

Ordine raccomandato:

1. backend Go minimale;
2. configurazione e logging;
3. PostgreSQL connection;
4. migrations;
5. health endpoints;
6. modello ChangeRequest;
7. Argo CD API adapter;
8. GitLab API adapter read-only;
9. ChangeRequest API;
10. workflow repliche senza Tekton;
11. Tekton validation;
12. evidence collector;
13. workflow APP_VERSION;
14. workflow PAGE_COLOR;
15. workflow ConfigMap;
16. pagine HTML minime;
17. UI Bootstrap completa.

---

## 13. Regola di avanzamento progetto

Non si passa alla milestone successiva se:

- la milestone corrente non ha documentazione aggiornata;
- non esiste almeno un test manuale ripetibile;
- non è chiaro come raccogliere evidenza;
- non è chiaro come tornare indietro;
- ci sono token o secret nel repository.

---

## 14. Messaggio chiave

Il primo MVP del DevOps Control Plane deve essere piccolo ma solido.

La priorità non è costruire subito una UI ricca, ma stabilizzare i workflow end-to-end:

```text
GitLab branch/commit/MR
        -> Tekton validation
        -> Argo CD sync
        -> OpenShift runtime validation
        -> evidence
        -> PostgreSQL history
```

Quando questi workflow saranno affidabili, la Web UI potrà essere costruita sopra una base stabile.
