# DevOps Control Plane - Vision

**Versione:** 0.1  
**Data:** 2026-06-25  
**Owner iniziale:** Vincenzo Marzario  
**Repository previsto:** `https://github.com/vincmarz/devops-control-plane`  
**Stato:** Draft iniziale / Visione di progetto

---

## 1. Contesto

Negli ambienti OpenShift moderni, un flusso DevOps/GitOps maturo coinvolge più componenti specializzati:

- **Git** come sorgente dello stato desiderato;
- **Argo CD / OpenShift GitOps** come motore di riconciliazione tra Git e cluster;
- **Tekton / OpenShift Pipelines** come motore di build, validazione e automazione;
- **OpenShift/Kubernetes** come piattaforma runtime;
- **repository applicativi e GitOps** come punto di tracciabilità tecnica dei cambiamenti.

Questi strumenti sono potenti, ma richiedono competenze operative importanti. Un cambio apparentemente semplice, come aumentare le repliche, aggiornare `APP_VERSION`, cambiare `PAGE_COLOR` o spostare variabili in una `ConfigMap`, richiede spesso una sequenza di attività:

1. individuazione del manifest corretto;
2. modifica YAML;
3. validazione locale;
4. commit o branch;
5. eventuale merge request;
6. validazione Tekton;
7. sync Argo CD;
8. wait health/sync;
9. verifica runtime;
10. raccolta evidenze;
11. eventuale rollback.

Il progetto **DevOps Control Plane** nasce per standardizzare questi workflow e renderli tracciabili, ripetibili e comprensibili anche per colleghi meno esperti.

---

## 2. Problema da risolvere

Oggi i change GitOps possono essere eseguiti manualmente tramite una combinazione di:

- comandi `git`;
- comandi `argocd`;
- comandi `oc`;
- comandi `tkn`;
- modifica manuale di file YAML;
- raccolta manuale di evidenze.

Questo approccio funziona, ma presenta alcuni rischi:

- comandi lunghi e difficili da ricordare;
- errori di copia/incolla;
- YAML errati o non validati prima del commit;
- modifiche runtime non registrate in Git;
- AppProject Argo CD non allineati alle risorse introdotte;
- evidenze raccolte in modo non uniforme;
- difficoltà nel ricostruire chi ha fatto cosa, quando e perché;
- difficoltà nel proporre rollback chiari e auditabili.

Il DevOps Control Plane vuole ridurre questi rischi fornendo un livello di orchestrazione e governance sopra Git, Argo CD, Tekton e OpenShift.

---

## 3. Obiettivo del progetto

L’obiettivo del progetto è realizzare un’applicazione interna chiamata **DevOps Control Plane** che permetta di:

- visualizzare le applicazioni gestite da Argo CD;
- mostrare stato di sync, health, revisione corrente e orphaned resources;
- leggere l’ultimo commit Git associato allo stato desiderato;
- guidare change GitOps semplici tramite workflow standardizzati;
- generare branch Git;
- creare commit o merge request;
- validare YAML e policy tramite Tekton;
- sincronizzare Argo CD dopo merge o commit approvato;
- attendere lo stato `Synced` e `Healthy`;
- raccogliere evidenze tecniche;
- mantenere uno storico change interno più leggibile di Git/Argo CD da soli.

---

## 4. Principi guida

### 4.1 Git resta la sorgente dello stato desiderato

Il DevOps Control Plane non deve sostituire Git come fonte dello stato desiderato.

Un change permanente deve essere rappresentato da una modifica Git:

- branch;
- commit;
- merge request, ove previsto;
- merge su branch target;
- sync Argo CD.

Il portale non deve applicare modifiche runtime come stato finale tramite comandi imperativi quali:

```bash
oc edit deployment
oc patch deployment
oc set env deployment
oc scale deployment
```

Questi comandi possono essere utili per troubleshooting o test controllati, ma non devono rappresentare il workflow definitivo GitOps.

---

### 4.2 Argo CD resta il motore GitOps

Il DevOps Control Plane non implementa un proprio motore di deploy.

Argo CD rimane responsabile di:

- leggere lo stato desiderato da Git;
- confrontarlo con lo stato runtime del cluster;
- applicare le differenze;
- mostrare stato `Synced` / `OutOfSync`;
- mostrare stato `Healthy` / `Degraded`;
- gestire history, sync e rollback operativo.

Il DevOps Control Plane userà l’**Argo CD API** per interrogare e orchestrare queste funzioni.

---

### 4.3 Tekton valida e produce evidenze

Tekton sarà usato come motore di validazione e automazione.

Esempi di validazioni:

- clone del branch Git;
- verifica sintassi YAML;
- `kustomize build`;
- `oc apply --dry-run=server`;
- controllo anti-secret;
- verifica AppProject/risorse ammesse;
- raccolta log ed evidenze.

Il DevOps Control Plane interagirà con Tekton tramite **Kubernetes API diretta**, creando e osservando risorse `PipelineRun` e `TaskRun`.

---

### 4.4 PostgreSQL conserva lo storico funzionale

Git e Argo CD mantengono history tecniche, ma non sempre rispondono facilmente a domande come:

- chi ha richiesto il change?
- perché è stato fatto?
- quale workflow è stato eseguito?
- quale validazione Tekton è stata associata?
- quale evidenza è stata raccolta?
- quale rollback è consigliato?

Per questo il DevOps Control Plane userà **PostgreSQL** come database interno per Change Request, eventi, evidenze e stato workflow.

---

### 4.5 API e workflow prima della Web UI completa

Il progetto parte con backend Go, API e workflow stabili.

Il supporto a **HTML templates + Bootstrap** sarà previsto fin dall’inizio, ma la Web UI operativa arriverà dopo, quando i workflow saranno chiari e validati.

Questo evita di costruire schermate sopra processi ancora instabili.

---

## 5. Obiettivi MVP

Il primo MVP deve coprire queste funzionalità.

### 5.1 Lista applicazioni Argo CD

Il sistema deve mostrare una lista di applicazioni Argo CD con:

- nome Application;
- progetto Argo CD;
- namespace target;
- repository Git;
- path GitOps;
- target revision;
- sync status;
- health status;
- current revision.

---

### 5.2 Dettaglio applicazione

Per ogni applicazione, il sistema deve mostrare:

- `sync status`;
- `health status`;
- `current revision`;
- `orphaned resources`;
- resources gestite;
- history Argo CD;
- ultimo commit Git rilevante.

---

### 5.3 Change semplici

Il sistema deve supportare form/workflow per change standard:

- modifica numero repliche;
- modifica `APP_VERSION`;
- modifica `PAGE_COLOR`;
- modifica valori in `ConfigMap`.

Ogni change deve produrre:

- branch Git;
- diff;
- commit o merge request;
- validazione;
- sync;
- evidenza;
- storico change.

---

### 5.4 GitLab API

Il sistema deve integrare Git tramite **GitLab API**.

Funzioni previste:

- leggere file dal repository;
- creare branch;
- aggiornare file;
- creare commit;
- aprire merge request;
- leggere commit history;
- recuperare diff.

Nota: il codice del progetto DevOps Control Plane può essere ospitato su GitHub nell’account `vincmarz`, mentre il prodotto implementerà adapter GitLab API per gestire repository applicativi/GitOps target.

---

### 5.5 Argo CD API

Il sistema deve integrare Argo CD tramite API.

Funzioni previste:

- lista Application;
- dettaglio Application;
- resources;
- orphaned resources;
- history;
- sync;
- polling fino a `Synced` e `Healthy`;
- lettura errori sync.

---

### 5.6 Tekton tramite Kubernetes API

Il sistema deve interagire con Tekton tramite Kubernetes API diretta.

Funzioni previste:

- creare `PipelineRun` di validazione;
- leggere stato `PipelineRun`;
- leggere stato `TaskRun`;
- raccogliere log;
- salvare evidenze.

---

### 5.7 Raccolta evidenze

Per ogni change, il sistema deve raccogliere almeno:

- `changeId`;
- utente/richiedente;
- timestamp;
- applicazione;
- tipo change;
- repository;
- branch;
- commit;
- merge request;
- PipelineRun Tekton;
- stato Argo CD;
- stato Deployment;
- stato Pod;
- eventuale Route/health check;
- log sintetici.

---

## 6. Non-obiettivi iniziali

Nel primo MVP non sono inclusi:

- gestione multi-cluster avanzata;
- workflow approval complessi;
- integrazione ITSM;
- gestione Secret enterprise;
- RBAC enterprise completo;
- supporto Helm avanzato;
- supporto multi-provider Git;
- replica completa delle funzionalità UI di Argo CD;
- sostituzione di Tekton Dashboard o Argo CD UI.

---

## 7. Architettura logica

Componenti principali:

```text
User / DevOps Operator
        |
        v
DevOps Control Plane
        |
        +--> Argo CD API Adapter
        |
        +--> GitLab API Adapter
        |
        +--> Tekton Kubernetes Adapter
        |
        +--> OpenShift/Kubernetes Adapter
        |
        v
PostgreSQL Change Store / Evidence Store
```

---

## 8. Stack tecnologico iniziale

### Backend

```text
Go
```

### Frontend MVP

```text
Go HTML templates + Bootstrap
```

La UI completa arriverà dopo la stabilizzazione dei workflow.

### Database

```text
PostgreSQL
```

### Git integration

```text
GitLab API
```

### Argo CD integration

```text
Argo CD API
```

### Tekton integration

```text
Kubernetes API diretta
```

---

## 9. Workflow target di un change

Esempio: modifica `APP_VERSION`.

```text
1. Utente seleziona applicazione.
2. DevOps Control Plane legge stato Argo CD.
3. DevOps Control Plane legge file GitOps da GitLab.
4. Utente inserisce nuovo valore APP_VERSION.
5. DevOps Control Plane crea branch.
6. DevOps Control Plane modifica file YAML.
7. DevOps Control Plane crea commit o merge request.
8. Tekton valida il branch.
9. Dopo merge, DevOps Control Plane avvia sync Argo CD.
10. DevOps Control Plane attende Synced/Healthy.
11. DevOps Control Plane raccoglie evidenze.
12. ChangeRequest passa a Completed.
```

---

## 10. Governance GitOps

Il DevOps Control Plane deve distinguere tra:

- **change applicativi**, per esempio `deployment.yaml`, `configmap.yaml`, `kustomization.yaml`;
- **change di governance**, per esempio `AppProject`, whitelist risorse, permessi Argo CD.

Esempio pratico:

Se un’applicazione introduce una `ConfigMap`, l’`AppProject` deve consentire:

```yaml
- group: ""
  kind: ConfigMap
```

Il sistema dovrà rilevare prerequisiti di questo tipo e segnalarli prima della sync.

---

## 11. Roadmap iniziale

### Milestone 0.1 - Project bootstrap

- repository creato;
- struttura directory;
- documentazione iniziale;
- backend Go minimale;
- endpoint `/healthz`;
- configurazione PostgreSQL;
- migrations iniziali.

### Milestone 0.2 - Argo CD discovery

- integrazione Argo CD API;
- lista applicazioni;
- dettaglio applicazione;
- sync/health/revision;
- orphaned resources.

### Milestone 0.3 - GitLab read operations

- lettura file repository;
- lettura ultimo commit;
- branch metadata;
- diff base.

### Milestone 0.4 - ChangeRequest database

- modello `ChangeRequest`;
- eventi change;
- stato workflow;
- API CRUD minima.

### Milestone 0.5 - First change workflow

- change repliche;
- branch GitLab;
- update YAML;
- commit/MR;
- validazione Tekton;
- sync Argo CD;
- evidenze.

---

## 12. Messaggio chiave

DevOps Control Plane deve aiutare il team DevOps a eseguire change GitOps in modo guidato, sicuro e tracciabile.

Il progetto non deve sostituire Git, Argo CD, Tekton o OpenShift.

Deve invece essere il livello di orchestrazione, governance e audit che rende questi strumenti più facili da usare e più affidabili nei processi operativi quotidiani.
