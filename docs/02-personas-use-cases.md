# DevOps Control Plane - Personas e Use Cases

**Versione:** 0.1  
**Data:** 2026-06-25  
**Owner iniziale:** Vincenzo Marzario  
**Repository:** `https://github.com/vincmarz/devops-control-plane`  
**Documento precedente:** `docs/01-scope-mvp.md`  
**Stato:** Draft iniziale / Personas e Use Cases

---

## 1. Scopo del documento

Questo documento descrive le **personas** e gli **use cases** principali del progetto **DevOps Control Plane**.

Lo scopo è chiarire:

- chi userà il sistema;
- quali problemi operativi deve risolvere;
- quali workflow devono essere supportati nel primo MVP;
- quali informazioni devono essere visibili a ogni tipologia di utente;
- quali output devono essere generati per rendere il processo GitOps tracciabile, ripetibile e comprensibile anche per colleghi meno esperti.

Il documento è propedeutico alla definizione dei requisiti funzionali dettagliati in `docs/03-functional-requirements.md`.

---

## 2. Principi di progettazione orientati agli utenti

DevOps Control Plane deve essere progettato considerando tre esigenze principali:

1. **guidare l’operatore** nei change GitOps standard;
2. **evitare errori manuali** su Git, YAML, Argo CD e Tekton;
3. **creare uno storico funzionale** più leggibile della sola somma di Git history, Argo CD history e Tekton logs.

Il sistema deve quindi essere:

- didattico per i newbie;
- affidabile per gli operatori DevOps;
- utile per reviewer e platform engineer;
- compatibile con il modello GitOps;
- orientato alla raccolta evidenze.

---

## 3. Personas principali

## 3.1 Persona P1 - DevOps Operator

### Descrizione

Il **DevOps Operator** è l’utente operativo principale del sistema.

Esegue change GitOps ricorrenti su applicazioni OpenShift già gestite da Argo CD.

### Obiettivi

Il DevOps Operator vuole:

- vedere velocemente lo stato delle applicazioni;
- eseguire change standard senza ricordare tutti i comandi;
- ridurre il rischio di errori YAML;
- generare branch, commit o merge request in modo guidato;
- lanciare validazioni automatiche;
- sincronizzare Argo CD;
- raccogliere evidenze finali.

### Esempi di bisogni

- “Voglio aumentare le repliche dell’applicazione da 2 a 3.”
- “Voglio cambiare `APP_VERSION` da `v2-green` a `v3-green`.”
- “Voglio cambiare `PAGE_COLOR` senza modificare manualmente YAML.”
- “Voglio sapere se Argo CD è Synced e Healthy dopo il change.”

### Informazioni richieste

- lista Application Argo CD;
- stato sync/health;
- current revision;
- repository e path GitOps;
- ultimo commit;
- stato validazione Tekton;
- stato runtime del Deployment;
- evidenze post-change.

---

## 3.2 Persona P2 - Platform Engineer

### Descrizione

Il **Platform Engineer** gestisce la piattaforma OpenShift, Argo CD, Tekton e le regole di governance.

Non è interessato solo al change applicativo, ma anche alla coerenza tra applicazioni, AppProject, policy e runtime.

### Obiettivi

Il Platform Engineer vuole:

- verificare che le Application siano coerenti con gli AppProject;
- identificare problemi di governance prima della sync;
- vedere risorse non consentite o orphaned;
- evitare drift GitOps;
- validare che i workflow rispettino le regole della piattaforma.

### Esempi di bisogni

- “L’applicazione introduce una ConfigMap: l’AppProject la consente?”
- “La Application ha orphaned resources: sono attese o sono drift?”
- “Il change sta tentando di modificare una risorsa non ammessa?”
- “Il workflow ha usato Git o ha modificato direttamente il cluster?”

### Informazioni richieste

- AppProject associato;
- namespaceResourceWhitelist;
- risorse gestite dalla Application;
- orphaned resources;
- errori di sync Argo CD;
- evidenze di validazione;
- eventi ChangeRequest.

---

## 3.3 Persona P3 - Newbie / Junior DevOps

### Descrizione

Il **Newbie** è un collega con conoscenze iniziali di GitOps, OpenShift, Argo CD e Tekton.

Ha bisogno di un sistema che non nasconda completamente i dettagli, ma li presenti in modo guidato e comprensibile.

### Obiettivi

Il Newbie vuole:

- capire cosa succede durante un change;
- vedere quali file vengono modificati;
- comprendere perché si usa Git e non `oc edit`;
- capire la differenza tra Git, Argo CD, Tekton e OpenShift;
- imparare dai workflow standardizzati.

### Esempi di bisogni

- “Perché devo modificare Git invece del Deployment direttamente?”
- “Cosa significa `OutOfSync`?”
- “Cosa significa `Healthy`?”
- “Perché una ConfigMap deve essere autorizzata nell’AppProject?”
- “Quale commit ha prodotto questo stato?”

### Informazioni richieste

- spiegazioni passo-passo;
- diff Git leggibile;
- stato workflow;
- messaggi di errore interpretati;
- comandi equivalenti, quando utile;
- link a evidenze.

---

## 3.4 Persona P4 - Reviewer / Approver

### Descrizione

Il **Reviewer** verifica che un change sia corretto prima del merge o della promozione.

Può essere un senior DevOps, un technical lead o un platform owner.

### Obiettivi

Il Reviewer vuole:

- capire cosa cambia;
- verificare il diff;
- vedere il risultato della validazione Tekton;
- approvare o respingere una merge request;
- valutare il rischio operativo;
- sapere come fare rollback.

### Esempi di bisogni

- “Quali file sono stati modificati?”
- “Il change è coerente con il modello GitOps?”
- “La validazione YAML è passata?”
- “La sync Argo CD è completata?”
- “Qual è il rollback consigliato?”

### Informazioni richieste

- change summary;
- diff;
- validazione Tekton;
- policy check;
- link GitLab MR;
- stato Argo CD;
- rollback hint.

---

## 3.5 Persona P5 - Auditor / Change Manager

### Descrizione

L’**Auditor** o **Change Manager** non lavora necessariamente sui manifest, ma ha bisogno di ricostruire lo storico dei change.

### Obiettivi

L’Auditor vuole:

- sapere chi ha richiesto un change;
- sapere quando è stato eseguito;
- vedere commit e merge request;
- vedere esito della validazione;
- vedere esito della sincronizzazione;
- consultare evidenze prodotte automaticamente.

### Esempi di bisogni

- “Chi ha cambiato le repliche?”
- “Quando è stato introdotto `APP_VERSION=v3-green`?”
- “Quale PipelineRun ha validato il change?”
- “Il change è stato completato con successo?”
- “Quali evidenze sono state raccolte?”

### Informazioni richieste

- Change ID;
- richiedente;
- timestamp;
- stato finale;
- commit/MR;
- PipelineRun;
- revisione Argo CD;
- evidenze runtime.

---

## 4. Use Cases MVP

## 4.1 UC-001 - Visualizzare lista applicazioni Argo CD

### Persona primaria

DevOps Operator

### Obiettivo

Visualizzare le Application Argo CD disponibili e il loro stato principale.

### Precondizioni

- DevOps Control Plane è configurato con endpoint Argo CD API.
- Il token Argo CD è valido.
- Argo CD contiene almeno una Application visibile.

### Flusso principale

1. L’utente apre la lista applicazioni o chiama l’API `GET /api/applications`.
2. Il backend interroga Argo CD API.
3. Il backend normalizza le informazioni principali.
4. Il sistema mostra nome, progetto, namespace, sync status, health status e revision.

### Output atteso

```text
Application          Project          Sync      Health    Revision
demo-go-color-app    devops-ci-demo   Synced    Healthy   b8e1d6b
```

### Errori gestiti

- token Argo CD non valido;
- Argo CD API non raggiungibile;
- nessuna Application disponibile;
- errore autorizzativo.

### Acceptance criteria

- La lista mostra almeno nome, progetto, sync, health e revision.
- Gli errori sono leggibili e registrati.
- La richiesta è tracciata nei log applicativi senza esporre token.

---

## 4.2 UC-002 - Visualizzare dettaglio applicazione

### Persona primaria

DevOps Operator

### Persona secondaria

Platform Engineer

### Obiettivo

Visualizzare stato dettagliato di una Application Argo CD.

### Precondizioni

- L’Application esiste.
- Argo CD API è raggiungibile.

### Flusso principale

1. L’utente seleziona `demo-go-color-app`.
2. Il backend legge dettaglio Application da Argo CD.
3. Il backend legge resources gestite.
4. Il backend legge orphaned resources.
5. Il backend legge history Argo CD.
6. Il backend arricchisce il dato con ultimo commit Git, se disponibile.

### Output atteso

- sync status;
- health status;
- current revision;
- resources gestite;
- orphaned resources;
- history;
- repo/path.

### Acceptance criteria

- Il dettaglio mostra chiaramente se l’app è `Synced`, `OutOfSync`, `Healthy` o `Degraded`.
- Le orphaned resources sono distinte dalle risorse gestite.
- Il sistema non interpreta `OrphanedResourceWarning` come errore applicativo se l’app è `Healthy`.

---

## 4.3 UC-003 - Creare ChangeRequest per modifica repliche

### Persona primaria

DevOps Operator

### Obiettivo

Creare un change GitOps per modificare `spec.replicas` di un Deployment.

### Precondizioni

- L’applicazione è nota al sistema.
- Il repository GitLab è configurato.
- Il path GitOps contiene un Deployment supportato.

### Input

```yaml
application: demo-go-color-app
changeType: update-replicas
replicas: 3
description: "Scale applicazione per test capacità"
```

### Flusso principale

1. L’utente crea una ChangeRequest.
2. Il sistema salva la ChangeRequest in stato `Created`.
3. Il sistema legge metadata Application.
4. Il sistema legge file GitOps da GitLab.
5. Il sistema crea un branch GitLab.
6. Il sistema modifica `replicas`.
7. Il sistema produce diff.
8. Il sistema crea commit o merge request.
9. Il sistema avvia validazione Tekton.
10. Dopo merge o approvazione, il sistema avvia sync Argo CD.
11. Il sistema raccoglie evidenze.
12. La ChangeRequest passa a `Completed`.

### Acceptance criteria

- Il Deployment runtime non viene modificato direttamente.
- Il change è rappresentato da commit o MR.
- Lo storico interno contiene branch, commit, eventuale MR e stato finale.
- Il Deployment raggiunge il numero desiderato dopo sync.

---

## 4.4 UC-004 - Creare ChangeRequest per APP_VERSION

### Persona primaria

DevOps Operator

### Persona secondaria

Newbie

### Obiettivo

Aggiornare `APP_VERSION` tramite un change GitOps guidato.

### Precondizioni

- Il sistema conosce dove è definita `APP_VERSION`.
- `APP_VERSION` può essere definita nel Deployment o in una ConfigMap.

### Input

```yaml
application: demo-go-color-app
changeType: update-app-version
APP_VERSION: v3-green
description: "Aggiornamento versione demo"
```

### Flusso principale

1. Il sistema determina se `APP_VERSION` è inline o in ConfigMap.
2. Il sistema legge il file corretto da GitLab.
3. Il sistema crea branch.
4. Il sistema aggiorna il valore.
5. Il sistema crea diff.
6. Il sistema valida con Tekton.
7. Il sistema crea MR o commit.
8. Dopo merge, il sistema sincronizza Argo CD.
9. Il sistema verifica che il runtime esponga il valore aggiornato.

### Acceptance criteria

- Il valore viene aggiornato solo in Git.
- Il sistema mostra chiaramente il file modificato.
- Il sistema salva evidenza del valore runtime finale.

---

## 4.5 UC-005 - Creare ChangeRequest per PAGE_COLOR

### Persona primaria

DevOps Operator

### Obiettivo

Aggiornare `PAGE_COLOR` tramite GitOps.

### Input

```yaml
application: demo-go-color-app
changeType: update-page-color
PAGE_COLOR: "#1E90FF"
description: "Cambio colore pagina demo"
```

### Validazioni

- Il colore deve rispettare formato esadecimale.
- Il valore non deve essere vuoto.
- Il file target deve essere supportato.

### Acceptance criteria

- Valori non validi vengono rifiutati.
- Il change produce diff Git.
- La validazione Tekton passa.
- Argo CD sincronizza correttamente.

---

## 4.6 UC-006 - Modificare valori ConfigMap

### Persona primaria

DevOps Operator

### Persona secondaria

Platform Engineer

### Obiettivo

Modificare una ConfigMap GitOps gestita dall’applicazione.

### Precondizioni

- La ConfigMap esiste nel repository.
- La ConfigMap è inclusa in `kustomization.yaml`, se applicabile.
- L’AppProject consente `ConfigMap`.

### Flusso principale

1. L’utente seleziona una ConfigMap.
2. Il sistema mostra chiavi e valori correnti.
3. L’utente modifica uno o più valori.
4. Il sistema aggiorna il manifest GitOps.
5. Il sistema valida il change.
6. Il sistema crea commit/MR.
7. Dopo merge, Argo CD sincronizza.
8. Il sistema raccoglie evidenza.

### Errori gestiti

- ConfigMap non inclusa in Kustomize.
- ConfigMap non consentita da AppProject.
- Deployment referenzia chiavi non esistenti.
- YAML non valido.

### Acceptance criteria

- Il sistema impedisce o segnala prerequisiti mancanti.
- Il sistema produce messaggi chiari, per esempio:

```text
ConfigMap is not permitted in AppProject devops-ci-demo.
```

---

## 4.7 UC-007 - Validare change con Tekton

### Persona primaria

DevOps Operator

### Persona secondaria

Reviewer

### Obiettivo

Validare automaticamente un branch GitOps prima della promozione.

### Precondizioni

- Il branch GitLab esiste.
- Il template PipelineRun è disponibile.
- DevOps Control Plane ha permessi Kubernetes per creare PipelineRun.

### Flusso principale

1. Il sistema crea una PipelineRun di validazione.
2. Tekton clona il branch.
3. Tekton valida YAML/Kustomize.
4. Tekton esegue dry-run server-side.
5. Tekton esegue anti-secret check.
6. Il sistema osserva lo stato della PipelineRun.
7. Il sistema raccoglie TaskRun e log principali.
8. La ChangeRequest viene aggiornata.

### Acceptance criteria

- Lo stato `ValidationSucceeded` viene assegnato solo se Tekton termina con successo.
- In caso di errore, la ChangeRequest passa a `ValidationFailed`.
- I log principali sono associati alla ChangeRequest.

---

## 4.8 UC-008 - Sincronizzare Argo CD dopo merge

### Persona primaria

DevOps Operator

### Obiettivo

Sincronizzare la Application dopo che il change è stato integrato nel branch target.

### Precondizioni

- Il commit è presente nel branch target.
- Argo CD vede il repository.
- La Application è configurata correttamente.

### Flusso principale

1. Il sistema richiede sync tramite Argo CD API.
2. Il sistema attende completamento operazione.
3. Il sistema controlla sync status.
4. Il sistema controlla health status.
5. Il sistema salva revision applicata.
6. Il sistema raccoglie evidenze runtime.

### Errori gestiti

- sync failed;
- operation already in progress;
- AppProject non consente una risorsa;
- manifest Kubernetes non valido;
- Application resta OutOfSync;
- Application diventa Degraded.

### Acceptance criteria

- Il sistema distingue errore di sync, errore di health e timeout.
- Il sistema salva il messaggio errore Argo CD.
- Il sistema propone azione suggerita o troubleshooting.

---

## 4.9 UC-009 - Raccogliere evidenze change

### Persona primaria

Auditor / Change Manager

### Persona secondaria

DevOps Operator

### Obiettivo

Raccogliere e conservare evidenze tecniche di un change.

### Evidenze minime

- ChangeRequest summary;
- branch/MR/commit;
- Tekton PipelineRun;
- Argo CD sync status;
- Argo CD health status;
- Deployment status;
- Pod status;
- ConfigMap, se coinvolta;
- Route/healthz, se applicabile;
- errore finale, se il change fallisce.

### Acceptance criteria

- Ogni ChangeRequest completata ha almeno un record evidence.
- Le evidenze sono accessibili via API.
- Le evidenze non contengono token o secret.

---

## 4.10 UC-010 - Consultare storico change

### Persona primaria

Reviewer / Auditor

### Obiettivo

Consultare lo storico funzionale dei change.

### Flusso principale

1. L’utente apre lista change.
2. Il sistema mostra Change ID, Application, tipo, stato, timestamp.
3. L’utente apre dettaglio change.
4. Il sistema mostra eventi, commit, MR, PipelineRun, sync e evidenze.

### Acceptance criteria

- È possibile ricostruire il ciclo di vita di un change.
- Gli eventi sono ordinati temporalmente.
- Lo stato finale è chiaro.
- È presente un suggerimento di rollback quando applicabile.

---

## 5. Use Cases fuori scope per l’MVP

Questi use cases sono esplicitamente fuori scope iniziale:

- gestione completa utenti e gruppi enterprise;
- approval multi-step;
- integrazione ServiceNow/Jira;
- gestione Secret applicativi;
- promozione multi-ambiente dev/test/prod;
- gestione Helm chart avanzata;
- editor YAML visuale;
- rollback automatico multi-step;
- policy engine completo;
- generazione automatica di AppProject;
- full developer portal stile Backstage.

---

## 6. Mappa Persona -> Use Cases

| Persona | Use cases principali |
|---|---|
| DevOps Operator | UC-001, UC-002, UC-003, UC-004, UC-005, UC-006, UC-007, UC-008 |
| Platform Engineer | UC-002, UC-006, UC-008 |
| Newbie / Junior DevOps | UC-001, UC-002, UC-003, UC-004, UC-005 |
| Reviewer / Approver | UC-007, UC-010 |
| Auditor / Change Manager | UC-009, UC-010 |

---

## 7. Informazioni da mostrare in modo didattico

Per ogni workflow, il sistema dovrebbe mostrare:

- cosa sta facendo;
- perché lo sta facendo;
- quale strumento sta usando;
- quale file viene modificato;
- quale commit o branch viene creato;
- cosa sta validando Tekton;
- cosa sta sincronizzando Argo CD;
- quale evidenza viene raccolta.

Esempio di messaggio didattico:

```text
Il valore APP_VERSION è gestito da una ConfigMap.
Il sistema aggiornerà configmap.yaml nel repository GitLab.
Dopo il merge, Argo CD applicherà la modifica al cluster.
```

---

## 8. Errori da rendere comprensibili

DevOps Control Plane deve tradurre errori tecnici complessi in messaggi operativi chiari.

### Esempio 1 - ConfigMap non permessa

Errore tecnico:

```text
resource :ConfigMap is not permitted in project devops-ci-demo
```

Messaggio suggerito:

```text
La ConfigMap è presente nel repository, ma l’AppProject Argo CD non consente la gestione delle ConfigMap.
Aggiungere group: "" kind: ConfigMap alla namespaceResourceWhitelist dell’AppProject.
```

---

### Esempio 2 - valueFrom non valido

Errore tecnico:

```text
valueFrom: Invalid value: must specify configMapKeyRef, secretKeyRef, fieldRef or resourceFieldRef
```

Messaggio suggerito:

```text
Il Deployment contiene un blocco valueFrom incompleto o indentato male.
Verificare che configMapKeyRef sia indentato sotto valueFrom.
```

---

### Esempio 3 - Argo CD operation già in corso

Errore tecnico:

```text
another operation is already in progress
```

Messaggio suggerito:

```text
Argo CD ha già una sync o rollback in corso sulla Application.
Attendere la fine dell’operazione o terminarla se è bloccata.
```

---

### Esempio 4 - Repository Git non trovato

Errore tecnico:

```text
fatal: not a git repository
```

Messaggio suggerito:

```text
Il comando Git è stato eseguito fuori dalla directory del repository.
Entrare nella directory corretta prima di eseguire operazioni Git.
```

---

## 9. Journey end-to-end esempio: change repliche

```text
1. DevOps Operator apre DevOps Control Plane.
2. Seleziona demo-go-color-app.
3. Visualizza Synced/Healthy e revision corrente.
4. Seleziona change type: replicas.
5. Inserisce nuovo valore: 3.
6. Il sistema crea ChangeRequest CHG-2026-0001.
7. Il sistema crea branch GitLab change/CHG-2026-0001.
8. Il sistema modifica deployment.yaml.
9. Il sistema mostra diff.
10. Il sistema avvia validazione Tekton.
11. Tekton termina Succeeded.
12. Il sistema apre MR o crea commit.
13. Dopo merge, il sistema avvia sync Argo CD.
14. Argo CD raggiunge Synced/Healthy.
15. Il sistema verifica Deployment e Pod.
16. Il sistema raccoglie evidenze.
17. Il change passa a Completed.
```

---

## 10. Journey end-to-end esempio: change ConfigMap

```text
1. DevOps Operator seleziona demo-go-color-app.
2. Il sistema rileva che PAGE_COLOR e APP_VERSION sono gestiti da ConfigMap.
3. L’utente modifica APP_VERSION.
4. Il sistema verifica che configmap.yaml esista.
5. Il sistema verifica che configmap.yaml sia incluso in kustomization.yaml.
6. Il sistema verifica che AppProject consenta ConfigMap.
7. Il sistema crea branch GitLab.
8. Il sistema aggiorna configmap.yaml.
9. Il sistema valida con Tekton.
10. Il sistema produce commit o MR.
11. Dopo merge, il sistema sincronizza Argo CD.
12. Il sistema verifica ConfigMap runtime.
13. Il sistema verifica Deployment/Pod.
14. Il sistema raccoglie evidenze.
```

---

## 11. Dati minimi richiesti per configurare una Application

Per gestire una Application, il sistema deve conoscere o ricavare:

```yaml
applicationName: demo-go-color-app
argocdNamespace: openshift-gitops
argocdProject: devops-ci-demo
targetNamespace: devops-ci-demo
repoProvider: gitlab
repoProjectId: "<gitlab-project-id>"
repoUrl: "https://gitlab.example.local/group/demo-app-gitops.git"
targetRevision: main
path: apps/demo-go-color-app
defaultBranch: main
```

---

## 12. Relazione con i documenti successivi

Questo documento alimenta:

- `docs/03-functional-requirements.md`, per dettagliare i requisiti funzionali;
- `docs/05-architecture.md`, per disegnare componenti e adapter;
- `docs/10-data-model.md`, per derivare entità e relazioni;
- `docs/11-change-workflows.md`, per formalizzare i workflow operativi;
- `docs/13-api-design.md`, per definire gli endpoint.

---

## 13. Messaggio chiave

DevOps Control Plane deve essere progettato partendo dai workflow reali degli utenti.

Le personas aiutano a evitare due errori:

1. costruire un tool troppo tecnico che non aiuta i newbie;
2. costruire un’interfaccia troppo semplificata che nasconde informazioni necessarie a DevOps e Platform Engineer.

Il sistema deve quindi essere guidato, ma trasparente:

```text
Mostra cosa fa.
Mostra perché lo fa.
Mostra quale strumento usa.
Mostra quale evidenza produce.
```
