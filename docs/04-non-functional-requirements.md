# DevOps Control Plane - Non-Functional Requirements

**Versione:** 0.1  
**Data:** 2026-06-25  
**Owner iniziale:** Vincenzo Marzario  
**Repository:** `https://github.com/vincmarz/devops-control-plane`  
**Documenti precedenti:**  
- `docs/00-vision.md`  
- `docs/01-scope-mvp.md`  
- `docs/02-personas-use-cases.md`  
- `docs/03-functional-requirements.md`  
**Stato:** Draft iniziale / Non-Functional Requirements

---

## 1. Scopo del documento

Questo documento definisce i **requisiti non funzionali** del progetto **DevOps Control Plane**.

I requisiti non funzionali descrivono le qualità attese del sistema, indipendentemente dalle singole funzionalità applicative. In particolare, questo documento copre:

- sicurezza;
- gestione credenziali;
- auditabilità;
- osservabilità;
- affidabilità;
- resilienza;
- performance;
- scalabilità;
- manutenibilità;
- configurabilità;
- portabilità;
- usabilità;
- compliance GitOps;
- gestione errori;
- backup e recovery;
- requisiti di deployment su OpenShift.

Il documento completa `docs/03-functional-requirements.md` e sarà usato come riferimento per architettura, data model, API design, deployment e implementazione.

---

## 2. Convenzioni

## 2.1 Classificazione requisiti

| Priorità | Significato |
|---|---|
| MUST | Obbligatorio per MVP o per sicurezza base |
| SHOULD | Importante, ma non bloccante per primo rilascio |
| COULD | Utile per evoluzioni successive |
| WON'T | Fuori scope per MVP |

---

## 2.2 Principio generale

DevOps Control Plane gestirà workflow GitOps e interagirà con sistemi critici come GitLab, Argo CD, Tekton e OpenShift.

Per questo motivo, anche nel primo MVP, il sistema deve essere progettato con attenzione a:

- protezione delle credenziali;
- tracciabilità dei change;
- comportamento prevedibile in caso di errore;
- logging utile ma non rischioso;
- separazione tra stato desiderato Git e stato runtime cluster;
- assenza di modifiche runtime permanenti fuori Git.

---

# 3. Security Requirements

## NFR-SEC-001 - Gestione sicura dei token

### Priorità

MUST

### Descrizione

Il sistema deve gestire in modo sicuro i token e le credenziali usate per accedere a:

- GitLab API;
- Argo CD API;
- Kubernetes/OpenShift API;
- PostgreSQL.

### Regole

- Nessun token deve essere salvato nel repository Git.
- Nessun token deve essere scritto nei log applicativi.
- Nessun token deve essere restituito tramite API.
- Nessun token deve essere incluso nelle evidenze.
- I token devono essere forniti tramite Secret Kubernetes/OpenShift o variabili ambiente sicure.
- Il file `.env` deve essere ammesso solo per sviluppo locale e deve restare ignorato da Git.

### Acceptance criteria

- `.gitignore` include `.env`, `*.env`, chiavi private e directory secret.
- Gli errori di autenticazione non stampano token.
- Le evidenze non contengono valori sensibili.

---

## NFR-SEC-002 - Principio del minimo privilegio

### Priorità

MUST

### Descrizione

Le credenziali usate dal DevOps Control Plane devono avere solo i privilegi necessari.

### Regole

Per GitLab:

- leggere repository configurati;
- creare branch;
- aggiornare file;
- creare commit;
- aprire merge request.

Per Argo CD:

- leggere Application;
- leggere resources/history;
- lanciare sync sulle Application autorizzate.

Per Kubernetes/OpenShift:

- creare/leggere PipelineRun nel namespace Tekton previsto;
- leggere stato TaskRun e log rilevanti;
- leggere risorse runtime necessarie alle evidenze;
- non avere privilegi cluster-admin per default.

### Acceptance criteria

- I permessi sono documentati in `docs/09-security-rbac.md`.
- Il deployment OpenShift usa ServiceAccount dedicata.
- I token personali non sono usati come soluzione stabile di produzione.

---

## NFR-SEC-003 - Separazione dati sensibili e non sensibili

### Priorità

MUST

### Descrizione

Il sistema deve distinguere chiaramente tra configurazioni non sensibili e dati sensibili.

### Regole

- Configurazioni applicative non sensibili possono stare in ConfigMap.
- Password, token, chiavi private e credenziali devono stare in Secret.
- DevOps Control Plane non deve proporre workflow che salvino secret in ConfigMap o in Git.

### Acceptance criteria

- I workflow ConfigMap includono warning didattico.
- Il controllo anti-secret blocca o segnala pattern sospetti.

---

## NFR-SEC-004 - Anti-secret scanning

### Priorità

SHOULD

### Descrizione

Il sistema dovrebbe eseguire un controllo anti-secret sui file modificati prima di commit/MR.

### Pattern minimi da intercettare

- `token`;
- `password`;
- `secret`;
- `auth`;
- `PRIVATE KEY`;
- `BEGIN RSA`;
- `ghp_`;
- `github_pat_`;
- `AKIA`;
- `ASIA`;
- file `config.json` Docker auth;
- file dockerconfig.

### Acceptance criteria

- Se il controllo trova pattern sospetti, il workflow passa in stato `ValidationFailed` o richiede review.
- Il messaggio spiega quale file contiene il possibile secret.

---

# 4. Auditability Requirements

## NFR-AUD-001 - Tracciabilità completa della ChangeRequest

### Priorità

MUST

### Descrizione

Ogni ChangeRequest deve essere completamente tracciabile.

### Dati minimi

- Change ID;
- applicazione;
- tipo change;
- richiedente;
- payload iniziale;
- branch GitLab;
- commit;
- merge request;
- PipelineRun Tekton;
- sync Argo CD;
- revisione applicata;
- stato finale;
- timestamp principali.

### Acceptance criteria

- Ogni stato produce un evento in `change_events`.
- Ogni ChangeRequest ha timeline consultabile.
- Gli eventi sono ordinabili temporalmente.

---

## NFR-AUD-002 - Evidenze persistenti

### Priorità

MUST

### Descrizione

Le evidenze principali devono essere persistite in PostgreSQL o in storage referenziato dal database.

### Tipi evidenza

- summary change;
- diff Git;
- esito Tekton;
- stato Argo CD;
- stato Deployment;
- stato Pod;
- stato ConfigMap, se coinvolta;
- stato Route/health check, se disponibile.

### Acceptance criteria

- Ogni ChangeRequest completata o fallita ha evidenze associate.
- Le evidenze non contengono token o secret.
- Le evidenze restano disponibili dopo riavvio applicazione.

---

## NFR-AUD-003 - Differenza tra history Git, Argo CD e ChangeRequest

### Priorità

SHOULD

### Descrizione

Il sistema deve rendere chiara la differenza tra:

- Git history;
- Argo CD history;
- ChangeRequest history interna.

### Regola didattica

Git descrive cosa è cambiato nei file.  
Argo CD descrive cosa è stato sincronizzato sul cluster.  
DevOps Control Plane descrive perché il change è stato fatto, da chi, con quale workflow e con quali evidenze.

### Acceptance criteria

- Il dettaglio ChangeRequest mostra riferimenti a commit Git, revision Argo CD e PipelineRun Tekton separatamente.

---

# 5. Observability Requirements

## NFR-OBS-001 - Logging strutturato

### Priorità

MUST

### Descrizione

Il backend deve produrre log strutturati.

### Campi minimi consigliati

- timestamp;
- level;
- component;
- requestId;
- changeId, se applicabile;
- applicationName, se applicabile;
- operation;
- status;
- errorCode, se applicabile.

### Regole

- Non loggare token.
- Non loggare password.
- Non loggare payload completi se contengono dati sensibili.

### Acceptance criteria

- Ogni richiesta API genera log start/end o equivalente.
- Ogni errore produce log con codice e messaggio.

---

## NFR-OBS-002 - Health e readiness endpoint

### Priorità

MUST

### Descrizione

Il servizio deve esporre endpoint di health e readiness.

### Endpoint minimi

```text
GET /healthz
GET /readyz
```

### Regole

`/healthz` verifica che il processo sia vivo.  
`/readyz` verifica dipendenze minime, almeno PostgreSQL.

### Acceptance criteria

- `/healthz` restituisce `200 OK` quando il processo è attivo.
- `/readyz` restituisce errore se PostgreSQL non è disponibile.

---

## NFR-OBS-003 - Metriche base

### Priorità

SHOULD

### Descrizione

Il sistema dovrebbe esporre metriche base in formato compatibile con Prometheus.

### Metriche candidate

- numero ChangeRequest create;
- numero ChangeRequest completed/failed;
- durata workflow;
- errori GitLab API;
- errori Argo CD API;
- errori Tekton;
- durata sync Argo CD;
- durata validazione Tekton.

### Acceptance criteria

- Endpoint `/metrics` disponibile in roadmap MVP avanzata.

---

# 6. Reliability Requirements

## NFR-REL-001 - Fallimento non ambiguo

### Priorità

MUST

### Descrizione

Il sistema deve evitare stati ambigui in caso di errore.

### Regole

- Se fallisce GitLab branch creation, lo stato deve indicare il fallimento.
- Se fallisce commit, lo stato deve indicare il fallimento.
- Se fallisce Tekton, lo stato deve essere `ValidationFailed`.
- Se fallisce Argo CD sync, lo stato deve essere `SyncFailed`.
- Se fallisce evidence collection, il change può essere completato con warning solo se runtime e sync sono corretti.

### Acceptance criteria

- Nessuna ChangeRequest resta indefinitamente in stato intermedio senza timeout.
- Ogni errore produce evento e messaggio leggibile.

---

## NFR-REL-002 - Idempotenza operazioni critiche

### Priorità

SHOULD

### Descrizione

Le operazioni critiche dovrebbero essere progettate per supportare retry controllati.

### Esempi

- recupero Application Argo CD;
- polling PipelineRun;
- polling sync Argo CD;
- raccolta evidenze;
- lettura file GitLab.

### Regole

Le operazioni che creano oggetti, come branch, commit o PipelineRun, devono gestire duplicati e retry con attenzione.

### Acceptance criteria

- Retry non crea branch duplicati senza controllo.
- Retry non crea commit duplicati senza controllo.
- Retry non crea PipelineRun duplicate senza riferimento alla ChangeRequest.

---

## NFR-REL-003 - Timeout espliciti

### Priorità

MUST

### Descrizione

Ogni integrazione esterna deve avere timeout.

### Dipendenze coinvolte

- GitLab API;
- Argo CD API;
- Kubernetes API;
- PostgreSQL;
- health check HTTP verso applicazioni.

### Acceptance criteria

- Nessuna chiamata esterna resta appesa indefinitamente.
- I timeout sono configurabili.
- Il timeout produce stato e evento coerente.

---

# 7. Performance Requirements

## NFR-PERF-001 - Performance API base

### Priorità

SHOULD

### Descrizione

Le API principali devono rispondere in tempi accettabili per uso operativo.

### Target iniziali indicativi

- `GET /healthz`: < 100 ms;
- `GET /readyz`: < 500 ms;
- `GET /api/applications`: < 5 secondi per ambiente piccolo/lab;
- `GET /api/changes`: < 2 secondi per centinaia di change;
- dettagli ChangeRequest: < 2 secondi.

### Nota

Questi target sono indicativi per MVP/lab e andranno rivisti in ambienti più grandi.

---

## NFR-PERF-002 - Operazioni asincrone per workflow lunghi

### Priorità

MUST

### Descrizione

Workflow lunghi come validazione Tekton e sync Argo CD non devono bloccare request HTTP per tempi eccessivi.

### Regole

- Una richiesta crea o avvia il workflow.
- Il client legge stato tramite polling/API.
- Lo stato è persistito in PostgreSQL.

### Acceptance criteria

- La creazione ChangeRequest risponde rapidamente.
- Lo stato del workflow è consultabile successivamente.

---

# 8. Scalability Requirements

## NFR-SCL-001 - Scalabilità MVP

### Priorità

SHOULD

### Descrizione

Il sistema deve essere progettato per supportare crescita graduale.

### Target iniziali MVP

- decine di Application;
- centinaia di ChangeRequest;
- pochi workflow concorrenti;
- uno o pochi cluster/ambienti.

### Acceptance criteria

- Il data model non assume una sola applicazione.
- Il codice non assume un solo namespace hardcoded, salvo default configurabile.
- Le integrazioni hanno configurazione per ambiente.

---

## NFR-SCL-002 - Concorrenza controllata

### Priorità

SHOULD

### Descrizione

Il sistema deve evitare di lanciare workflow concorrenti incoerenti sulla stessa Application.

### Regola iniziale

Per MVP, può essere accettabile impedire più ChangeRequest attive sulla stessa Application e stesso file target.

### Acceptance criteria

- Il sistema rileva ChangeRequest attive potenzialmente conflittuali.
- Il sistema segnala conflitto prima di creare branch o commit.

---

# 9. Maintainability Requirements

## NFR-MNT-001 - Architettura modulare ad adapter

### Priorità

MUST

### Descrizione

Il codice deve essere organizzato per separare dominio e integrazioni esterne.

### Adapter previsti

- Argo CD adapter;
- GitLab adapter;
- Kubernetes adapter;
- Tekton adapter;
- PostgreSQL repository;
- Evidence collector.

### Acceptance criteria

- La logica workflow non dipende direttamente da dettagli HTTP GitLab o Argo CD.
- Gli adapter sono testabili separatamente.

---

## NFR-MNT-002 - Documentazione sempre aggiornata

### Priorità

MUST

### Descrizione

Ogni milestone deve aggiornare la documentazione relativa.

### Regole

- Nuove decisioni architetturali devono generare o aggiornare ADR.
- Nuovi endpoint devono essere documentati in `docs/13-api-design.md`.
- Nuovi workflow devono essere documentati in `docs/11-change-workflows.md`.

### Acceptance criteria

- Non si chiude una milestone senza aggiornamento documentale.

---

## NFR-MNT-003 - Naming coerente

### Priorità

SHOULD

### Descrizione

Il progetto deve usare naming coerente per:

- ChangeRequest;
- ChangeEvent;
- Evidence;
- Application;
- Workflow;
- Adapter;
- stati.

### Acceptance criteria

- Gli stessi termini sono usati in documentazione, codice e database.

---

# 10. Configurability Requirements

## NFR-CFG-001 - Configurazione esterna

### Priorità

MUST

### Descrizione

Il sistema deve leggere configurazione da variabili ambiente o file configurazione non sensibili.

### Parametri minimi

```text
HTTP_ADDR
DATABASE_URL
ARGOCD_BASE_URL
GITLAB_BASE_URL
TEKTON_NAMESPACE
KUBERNETES_NAMESPACE
EVIDENCE_BASE_PATH
LOG_LEVEL
```

### Parametri sensibili

```text
ARGOCD_AUTH_TOKEN
GITLAB_TOKEN
DATABASE_PASSWORD
```

Questi devono arrivare da Secret.

### Acceptance criteria

- Il servizio fallisce in modo leggibile se manca una configurazione obbligatoria.
- I valori sensibili non vengono stampati.

---

## NFR-CFG-002 - Supporto ambienti multipli nel tempo

### Priorità

COULD

### Descrizione

Il sistema dovrebbe essere predisposto per ambienti multipli, anche se l’MVP parte con uno scenario singolo.

### Esempi futuri

- dev;
- test;
- prod;
- lab.

---

# 11. Usability Requirements

## NFR-UX-001 - Messaggi didattici per newbie

### Priorità

SHOULD

### Descrizione

Il sistema deve fornire messaggi chiari e didattici per operazioni e errori comuni.

### Esempi

- spiegare perché Git è source of truth;
- spiegare cosa significa `OutOfSync`;
- spiegare cosa significa `Healthy`;
- spiegare perché ConfigMap richiede autorizzazione AppProject;
- spiegare cosa fa Tekton validation.

### Acceptance criteria

- Gli errori noti hanno messaggi operativi.
- I workflow mostrano step comprensibili.

---

## NFR-UX-002 - Nessuna UI complessa prima dei workflow stabili

### Priorità

MUST

### Descrizione

La UI completa non deve essere costruita prima della stabilizzazione dei workflow backend.

### Regola

Il primo obiettivo è stabilizzare API, workflow, adapter e data model.

### Acceptance criteria

- Le pagine HTML iniziali, se presenti, sono semplici e non guidano decisioni architetturali premature.

---

# 12. GitOps Compliance Requirements

## NFR-GITOPS-001 - Nessun bypass GitOps

### Priorità

MUST

### Descrizione

Il sistema non deve modificare direttamente risorse applicative runtime come stato finale permanente.

### Regola

Ogni change applicativo deve passare da GitLab e Argo CD.

### Acceptance criteria

- Workflow repliche modifica Git, non il Deployment runtime.
- Workflow APP_VERSION modifica Git, non `oc set env`.
- Workflow ConfigMap modifica Git, non `oc edit configmap`.

---

## NFR-GITOPS-002 - Gestione drift

### Priorità

SHOULD

### Descrizione

Il sistema deve rilevare Application OutOfSync e spiegare il possibile drift.

### Acceptance criteria

- Application OutOfSync viene mostrata chiaramente.
- Il sistema non considera automaticamente OutOfSync come errore se deriva da rollback operativo temporaneo documentato, ma deve segnalarlo.

---

# 13. Backup and Recovery Requirements

## NFR-BCK-001 - Backup PostgreSQL

### Priorità

SHOULD

### Descrizione

PostgreSQL contiene storico change ed evidenze. Deve essere previsto un meccanismo di backup.

### MVP

Nel primo MVP è sufficiente documentare la necessità di backup.

### Futuro

- backup schedulato;
- retention;
- restore testato.

---

## NFR-BCK-002 - Ricostruibilità da Git

### Priorità

MUST

### Descrizione

Il repository sorgente deve permettere di ricostruire applicazione e documentazione.

### Acceptance criteria

- Manifest, codice e documentazione sono versionati.
- Secret reali non sono versionati.
- Template Secret possono essere versionati solo con placeholder.

---

# 14. Deployment Requirements on OpenShift

## NFR-OCP-001 - Containerizzazione

### Priorità

MUST

### Descrizione

Il servizio deve essere eseguibile come container.

### Acceptance criteria

- È presente un Dockerfile/Containerfile.
- Il container espone la porta HTTP configurata.
- Il container non richiede privilegi elevati.

---

## NFR-OCP-002 - Deployment OpenShift

### Priorità

SHOULD

### Descrizione

Il progetto deve prevedere manifest OpenShift/Kubernetes per deploy.

### Risorse previste

- Namespace o Project dedicato;
- Deployment;
- Service;
- Route;
- ConfigMap;
- Secret template;
- ServiceAccount;
- Role/RoleBinding minimi.

### Acceptance criteria

- I manifest non contengono secret reali.
- I permessi sono minimi e documentati.

---

# 15. Testing Requirements

## NFR-TST-001 - Test unitari per logica dominio

### Priorità

SHOULD

### Descrizione

La logica dominio deve essere testabile senza chiamate reali a GitLab, Argo CD o Kubernetes.

### Acceptance criteria

- Workflow state transitions testabili.
- Validazione input testabile.
- Generazione branch name testabile.

---

## NFR-TST-002 - Test adapter con mock

### Priorità

SHOULD

### Descrizione

Gli adapter esterni devono poter essere testati con mock/fake client.

### Acceptance criteria

- GitLab adapter può essere testato senza GitLab reale.
- Argo CD adapter può essere testato senza Argo CD reale.
- Tekton adapter può essere testato senza cluster reale.

---

## NFR-TST-003 - Test manuali evidence-aligned

### Priorità

MUST

### Descrizione

Per le milestone MVP devono esistere test manuali ripetibili con evidenze.

### Acceptance criteria

- Ogni milestone ha una checklist.
- Ogni checklist produce output verificabile.

---

# 16. Documentation Requirements

## NFR-DOC-001 - Documentazione progressiva

### Priorità

MUST

### Descrizione

La documentazione deve crescere insieme al progetto.

### Documenti previsti

- vision;
- scope MVP;
- personas/use cases;
- functional requirements;
- non-functional requirements;
- architecture;
- data model;
- workflows;
- API design;
- security/RBAC;
- ADR.

### Acceptance criteria

- Ogni documento è versionato in Git.
- Ogni documento ha scopo e stato.

---

## NFR-DOC-002 - ADR per decisioni importanti

### Priorità

MUST

### Descrizione

Ogni decisione architetturale significativa deve essere registrata in ADR.

### Esempi ADR

- Git come source of truth;
- Argo CD come motore GitOps;
- Tekton come motore validazione;
- PostgreSQL come change store;
- API/workflow prima della UI completa.

---

# 17. Compliance Matrix MVP

| Area | Requisito chiave | Priorità |
|---|---|---|
| Security | Token non in Git/log/evidenze | MUST |
| Audit | ChangeRequest tracciata end-to-end | MUST |
| Observability | Health/readiness endpoint | MUST |
| Reliability | Stati di errore non ambigui | MUST |
| Performance | Workflow lunghi asincroni | MUST |
| Maintainability | Architettura ad adapter | MUST |
| Configurability | Config esterna | MUST |
| GitOps | Nessun bypass GitOps | MUST |
| OpenShift | Container non privilegiato | MUST |
| Documentation | Documentazione versionata | MUST |

---

## 18. Criterio di completamento del documento

Questo documento sarà considerato stabile quando:

- ogni area non funzionale ha almeno un requisito MVP;
- i requisiti MUST sono collegati ad architettura e deployment;
- le implicazioni di sicurezza sono riportate in `docs/09-security-rbac.md`;
- le decisioni principali sono riportate in ADR.

---

## 19. Messaggio chiave

DevOps Control Plane non deve essere solo funzionale: deve essere sicuro, tracciabile, osservabile e coerente con GitOps.

Il valore del progetto non sta solo nell’automatizzare comandi, ma nel rendere i change:

```text
guidati
ripetibili
validati
tracciati
auditable
coerenti con GitOps
```
