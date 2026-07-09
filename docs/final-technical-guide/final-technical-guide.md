# DevOps Control Plane — Guida tecnica finale

Status: Draft  
Phase: 12.4 — Start final technical guide with introduction and foundations  
Owner: Vincenzo Marzario  
Language: Italian  
Last updated: 2026-07-09

## 1. Executive Summary

Il DevOps Control Plane è una piattaforma applicativa pensata per governare, validare e osservare workflow DevOps basati su GitOps, Kubernetes/OpenShift, Tekton, Argo CD, GitLab e PostgreSQL.

L’obiettivo del progetto è centralizzare il ciclo di vita di una richiesta di cambiamento, denominata `ChangeRequest`, rendendo tracciabili le azioni tecniche, le decisioni operative, gli eventi di audit e le evidenze raccolte durante il processo.

In pratica, il DevOps Control Plane risponde a domande come:

- quale cambiamento è stato richiesto;
- da chi è stato richiesto;
- per quale applicazione;
- per quale ambiente;
- quale branch o revisione Git è coinvolta;
- quale validazione Tekton è stata eseguita;
- quale stato Argo CD è stato osservato;
- quale stato runtime Kubernetes/OpenShift è stato raccolto;
- quali evidenze sono state prodotte;
- se le evidenze sono sicure e sanificate;
- quale stato finale ha assunto la richiesta.

La piattaforma è stata sviluppata con un backend Go, con persistenza PostgreSQL e integrazioni verso GitLab, Tekton, Argo CD e Kubernetes/OpenShift.

La UI è stata evoluta da semplice MVP a superficie operativa avanzata. Oggi mostra:

- dashboard con selezione della ChangeRequest più recente;
- viste ChangeRequest;
- runtime evidence;
- Tekton validation evidence;
- mapping `Environments / Namespaces`;
- dettaglio delle evidenze associate alle richieste;
- informazioni utili per controllo e troubleshooting.

## 2. Stato corrente del progetto

Lo stato corrente del progetto è una baseline avanzata e funzionante.

Le principali componenti risultano completate o rafforzate:

- backend Go;
- persistenza PostgreSQL;
- ciclo di vita ChangeRequest;
- audit ChangeEvent;
- workflow GitLab;
- validazione Tekton;
- controllo Argo CD;
- runtime evidence Kubernetes/OpenShift;
- UI evidence-aware;
- Environment Catalog;
- Cluster Registry;
- runtime target resolution;
- multi-cluster code readiness;
- runbook operativi;
- troubleshooting;
- guardrail Secret, RBAC e factory.

La baseline runtime validata è namespace-isolated sul cluster OpenShift disponibile `ocp-dev`.

Mappatura corrente:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

Questa topologia non rappresenta ancora un multi-cluster fisico. Rappresenta una baseline multi-environment validata tramite isolamento per namespace sullo stesso cluster.

La validazione fisica cross-cluster resta rinviata perché non sono disponibili altri cluster OpenShift.

La readiness multi-cluster a livello codice è invece completata, testata e documentata.

Formulazione corretta dello stato:

Physical cross-cluster runtime validation is deferred by infrastructure availability.

Multi-cluster code readiness is completed, tested, documented and fail-closed.

## 3. Visione del progetto

La visione del DevOps Control Plane è costruire un punto centrale di controllo per orchestrare workflow applicativi e infrastrutturali in modo verificabile, auditabile e ripetibile.

In molti contesti DevOps, le informazioni sono distribuite tra strumenti diversi:

- GitLab contiene branch, commit e merge request;
- Argo CD contiene lo stato GitOps e il confronto tra stato desiderato e stato runtime;
- Tekton esegue pipeline tecniche;
- Kubernetes/OpenShift espone lo stato reale dei workload;
- PostgreSQL conserva dati applicativi;
- log e output tecnici sono spesso sparsi in più sistemi.

Il DevOps Control Plane nasce per unire questi elementi in un modello coerente.

La piattaforma non sostituisce GitLab, Argo CD, Tekton o OpenShift.

La piattaforma li coordina e li rende leggibili attraverso un lifecycle unico basato su `ChangeRequest`.

## 4. Scope del progetto

Lo scope attuale include:

- creazione e gestione delle ChangeRequest;
- persistenza su PostgreSQL;
- audit tramite ChangeEvent;
- workflow GitLab end-to-end;
- validazione tecnica con Tekton;
- controllo deployment con Argo CD;
- raccolta runtime evidence Kubernetes/OpenShift;
- UI dashboard e dettaglio ChangeRequest;
- Environment Catalog;
- namespace isolation;
- predisposizione multi-cluster a livello codice;
- Secret reference model;
- runtime factory guardrails;
- runbook operativi.

Lo scope attuale non include ancora:

- validazione fisica cross-cluster;
- onboarding reale di un cluster non-production separato;
- onboarding reale di un cluster production separato;
- CLI `devopsctl`, attualmente in standby;
- produzione enterprise definitiva con tutti gli hardening possibili;
- generazione automatica completa di report finali per tutti gli scenari.

Questi elementi restano parte della roadmap futura o di attività condizionate dalla disponibilità infrastrutturale.

## 5. Concetti fondamentali

Prima di descrivere l’implementazione, è utile chiarire alcuni concetti di base.

Il progetto combina strumenti e modelli che spesso vengono usati insieme nelle piattaforme cloud-native:

- Kubernetes/OpenShift;
- namespace;
- GitOps;
- Kustomize;
- Argo CD;
- Tekton;
- RBAC;
- Secret;
- evidence;
- workflow dichiarativi.

Questi concetti sono collegati tra loro.

Per esempio:

- GitOps descrive lo stato desiderato;
- Argo CD applica lo stato desiderato al cluster;
- Tekton valida tecnicamente il cambiamento;
- Kubernetes/OpenShift esegue i workload;
- il DevOps Control Plane raccoglie evidenze e mantiene il controllo del workflow.

## 6. Kubernetes e OpenShift in breve

Kubernetes è una piattaforma per eseguire applicazioni containerizzate.

OpenShift è una piattaforma enterprise basata su Kubernetes che aggiunge componenti, sicurezza, developer experience e integrazioni operative.

Nel progetto DevOps Control Plane, OpenShift è l’ambiente runtime principale.

I concetti fondamentali usati dal progetto sono:

- cluster;
- namespace;
- pod;
- deployment;
- service;
- route;
- secret;
- configmap;
- serviceaccount;
- role;
- rolebinding.

### 6.1 Cluster

Un cluster è l’insieme delle risorse Kubernetes/OpenShift che eseguono applicazioni.

Nel progetto attuale è disponibile il cluster:

`ocp-dev`

Su questo cluster sono stati simulati più ambienti logici tramite namespace separati.

### 6.2 Namespace

Un namespace è uno spazio logico dentro un cluster.

Permette di separare risorse, permessi, workload e configurazioni.

Nel progetto sono usati tre namespace principali per la baseline multi-environment:

- `devops-ci-demo`;
- `devops-ci-staging`;
- `devops-ci-production`.

### 6.3 Deployment

Un Deployment descrive come eseguire una o più repliche di un’applicazione.

Nel progetto, l’applicazione dimostrativa `demo-go-color-app` è distribuita nei namespace dev, staging e production simulati.

### 6.4 Service e Route

Un Service espone un’applicazione all’interno del cluster.

Una Route OpenShift espone l’applicazione verso l’esterno o verso utenti e sistemi che devono raggiungerla.

Nel runtime smoke test sono state validate le route `/healthz` per dev, staging e production.

### 6.5 Secret

Un Secret contiene dati sensibili, come token o credenziali.

Nel DevOps Control Plane i Secret non devono essere copiati in log, documentazione o evidence.

Il modello corretto usa Secret references, cioè riferimenti a Secret, non valori raw.

### 6.6 RBAC

RBAC significa Role-Based Access Control.

Serve a definire cosa può fare un utente o un ServiceAccount.

Nel progetto, RBAC è importante per:

- leggere deployment;
- osservare pod;
- creare PipelineRun Tekton;
- leggere TaskRun;
- accedere solo alle risorse strettamente necessarie.

Il principio operativo è il minimo privilegio.

## 7. Namespace isolation

Namespace isolation significa separare ambienti logici usando namespace diversi nello stesso cluster.

Nel progetto, la topologia validata è:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

Questa scelta è stata necessaria perché non sono disponibili altri cluster OpenShift.

Il vantaggio è che il progetto può validare workflow multi-environment anche senza multi-cluster fisico.

Il limite è che dev, staging e production condividono comunque lo stesso cluster fisico.

Per questo motivo il documento distingue sempre tra:

- ambiente logico;
- namespace;
- cluster fisico.

### 7.1 Ambiente logico

Un ambiente logico rappresenta uno scopo operativo.

Esempi:

- `dev` per sviluppo;
- `staging` per validazione pre-produzione;
- `production` per produzione simulata.

### 7.2 Namespace runtime

Un namespace runtime è il luogo Kubernetes/OpenShift in cui vengono eseguite le risorse dell’ambiente.

Esempi:

- `devops-ci-demo`;
- `devops-ci-staging`;
- `devops-ci-production`.

### 7.3 Cluster fisico

Un cluster fisico è il cluster OpenShift reale.

Al momento il cluster fisico disponibile è:

`ocp-dev`

La validazione fisica di cluster staging o production separati è rinviata.

### 7.4 Perché questa baseline è utile

La baseline namespace-isolated permette di validare:

- Environment Catalog;
- Argo CD Applications;
- Tekton validation path;
- runtime evidence;
- UI environment awareness;
- ChangeRequest workflow;
- RBAC namespace-specific;
- guardrail fail-closed.

Quando saranno disponibili cluster fisici separati, il progetto non dovrà riprogettare il modello.

Dovrà applicare il contratto di onboarding reale già documentato.

## 8. GitOps

GitOps è un modello operativo in cui Git rappresenta la fonte di verità dello stato desiderato.

In un modello GitOps, invece di modificare direttamente il cluster, si modifica un repository Git.

Uno strumento come Argo CD osserva il repository e porta il cluster verso lo stato dichiarato.

Nel progetto DevOps Control Plane, GitOps è centrale perché:

- le modifiche applicative sono versionate;
- le modifiche sono tracciabili;
- Argo CD può sincronizzare lo stato desiderato;
- Tekton può validare il contenuto GitOps;
- il DevOps Control Plane può collegare ChangeRequest, Git revision, validation path ed evidence.

Esempi di path GitOps validati:

- `apps/demo-go-color-app/overlays/staging`;
- `apps/demo-go-color-app/overlays/production`.

## 9. Kustomize base e overlays

Kustomize permette di definire una base comune e overlay specifici per ambiente.

La base contiene configurazioni condivise.

Gli overlay applicano differenze specifiche.

Nel progetto, gli overlay permettono di avere configurazioni distinte per:

- dev;
- staging;
- production.

Esempi:

- overlay staging: `apps/demo-go-color-app/overlays/staging`;
- overlay production: `apps/demo-go-color-app/overlays/production`.

Questa separazione è importante perché Tekton deve validare il path corretto per l’ambiente corretto.

Un errore di validation path potrebbe portare a validare staging mentre si pensa di validare production, o viceversa.

Per questo la correzione del validation path environment-specific è stata una parte importante della baseline.

## 10. Argo CD

Argo CD è il motore GitOps usato dal progetto.

Il compito di Argo CD è confrontare lo stato desiderato nel repository Git con lo stato effettivo del cluster.

I due stati principali sono:

- sync;
- health.

### 10.1 Sync

Lo stato `Synced` indica che il cluster è allineato con il repository Git.

### 10.2 Health

Lo stato `Healthy` indica che le risorse applicative risultano sane dal punto di vista di Argo CD.

### 10.3 Applications validate

Nel progetto sono state validate Argo CD Applications per:

- dev;
- staging;
- production.

Esempi:

- `demo-go-color-app`;
- `demo-go-color-app-staging`;
- `demo-go-color-app-production`.

La smoke matrix finale ha confermato stato:

- `Synced`;
- `Healthy`.

## 11. Tekton

Tekton è il motore di pipeline usato per validare tecnicamente il cambiamento.

I concetti principali sono:

- Task;
- Pipeline;
- PipelineRun;
- TaskRun.

### 11.1 Task

Un Task è un’unità di lavoro.

Esempi tipici:

- clonare un repository;
- validare manifest;
- costruire immagine;
- eseguire controlli tecnici.

### 11.2 Pipeline

Una Pipeline è una sequenza di Task.

Nel progetto, Tekton viene usato per validare contenuti GitOps.

### 11.3 PipelineRun

Una PipelineRun è un’esecuzione concreta di una Pipeline.

Esempi validati:

- `devops-cp-validate-chg-2026-0049-nd7rm`;
- `devops-cp-validate-chg-2026-0050-8wqtv`.

### 11.4 TaskRun

Una TaskRun è l’esecuzione concreta di un Task dentro una PipelineRun.

Se una validazione fallisce, le TaskRun aiutano a capire quale passaggio tecnico non è riuscito.

### 11.5 Tekton nel DevOps Control Plane

Il DevOps Control Plane usa Tekton per:

- avviare una validazione;
- controllare il risultato;
- raccogliere evidence;
- mostrare la validation evidence nella UI.

La validazione finale ha confermato:

- staging PipelineRun `Succeeded`;
- production PipelineRun `Succeeded`;
- failed task count pari a `0`;
- evidence sanitized pari a `true`.

## 12. Stato della Parte 1 e Parte 2

Questa prima versione del documento finale copre:

- introduzione;
- visione;
- scope;
- concetti Kubernetes/OpenShift;
- namespace isolation;
- GitOps;
- Kustomize;
- Argo CD;
- Tekton.

Le sezioni successive dovranno coprire:

- architettura;
- backend Go;
- PostgreSQL;
- modello dati;
- workflow;
- evidence model;
- UI;
- Environment Catalog;
- multi-cluster readiness;
- security;
- operability;
- roadmap;
- appendici.

## 12. Architettura generale

Il DevOps Control Plane è costruito come una piattaforma applicativa modulare. Il suo compito non è sostituire GitLab, Argo CD, Tekton o OpenShift, ma coordinarli in un workflow controllato, persistente e auditabile.

Il backend principale è scritto in Go e coordina più sistemi esterni:

- PostgreSQL per persistenza, audit ed evidence;
- GitLab per branch, commit e merge request;
- Argo CD per osservare lo stato GitOps e la salute delle applicazioni;
- Tekton per eseguire validazioni tecniche tramite PipelineRun;
- Kubernetes/OpenShift per osservare lo stato runtime reale;
- UI web per dashboard, ChangeRequest, audit log ed evidenze.

Vista semplificata:

```text
Utente / UI / API
       |
       v
DevOps Control Plane backend
       |
       +--> PostgreSQL
       +--> GitLab API
       +--> Argo CD API
       +--> Kubernetes/OpenShift API
       +--> Tekton API
```

Questa separazione permette di far evolvere il progetto in modo incrementale. Per esempio, la raccolta runtime evidence può essere migliorata senza riscrivere il lifecycle delle ChangeRequest, e il modello multi-cluster può essere predisposto nel codice anche se oggi è disponibile solo il cluster `ocp-dev`.

### 12.1 Principi architetturali

I principi architetturali principali sono:

- API-first;
- persistenza affidabile;
- auditabilità;
- adapter-based architecture;
- GitOps come fonte di verità;
- runtime evidence come prova operativa;
- fail-closed per configurazioni incomplete o non sicure;
- security-by-default;
- evoluzione incrementale verso il multi-cluster.

### 12.2 API-first

Il progetto è stato costruito dando priorità al backend e alle API.

Questa scelta è importante perché il valore principale del DevOps Control Plane non è solo visualizzare dati, ma orchestrare workflow tecnici affidabili.

La UI è stata evoluta dopo la stabilizzazione dei workflow principali. Oggi la UI è molto più avanzata rispetto al primo MVP, ma resta costruita sopra comportamenti backend già validati.

### 12.3 Adapter-based architecture

Il progetto usa adapter per isolare le integrazioni esterne.

Esempi:

- GitLab adapter;
- Kubernetes/OpenShift adapter;
- Tekton adapter;
- Argo CD adapter.

Questo consente al core applicativo di ragionare in termini di dominio e workflow, senza dipendere direttamente dai dettagli implementativi di ogni API esterna.

### 12.4 Stato runtime attuale

La baseline runtime attuale è:

- `dev` -> `ocp-dev` / `devops-ci-demo`;
- `staging` -> `ocp-dev` / `devops-ci-staging`;
- `production` -> `ocp-dev` / `devops-ci-production`.

Questa architettura rappresenta una topologia multi-environment con namespace isolation.

La validazione fisica multi-cluster resta rinviata perché non sono disponibili altri cluster OpenShift. Tuttavia il codice è stato predisposto per runtime target multi-cluster futuri.

### 12.5 Implicazione per il futuro multi-cluster

Il modello architetturale non è legato rigidamente a un solo cluster.

Il codice include concetti come:

- Environment Catalog;
- Cluster Registry;
- Environment Cluster Resolver;
- Technical Runtime Target;
- Runtime Client Provider Registry;
- Secret reference model;
- runtime client factories.

Questo significa che, quando un cluster reale aggiuntivo sarà disponibile, il progetto dovrà eseguire onboarding e validazioni, non riprogettare l’architettura.

La regola da mantenere è:

```text
Physical cross-cluster runtime validation is deferred by infrastructure availability.
Multi-cluster code readiness is completed, tested, documented and fail-closed.
```

## 13. Backend Go

Il backend Go è il cuore applicativo del DevOps Control Plane.

Il backend espone API, coordina i workflow tecnici, conserva lo stato applicativo, produce eventi di audit e raccoglie evidenze dai sistemi integrati. La UI e gli operatori non devono orchestrare direttamente GitLab, Tekton, Argo CD o Kubernetes/OpenShift: questa responsabilità viene centralizzata nel backend.

In termini pratici, il backend riceve una richiesta, la interpreta nel contesto di una `ChangeRequest`, determina l’ambiente target, esegue o controlla le integrazioni tecniche necessarie e persiste il risultato.

Vista semplificata:

```text
API / UI request
      |
      v
Go backend
      |
      +--> ChangeService
      +--> Repositories
      +--> Runtime target resolution
      +--> Provider selection
      +--> External adapters
      +--> Evidence persistence
```

### 13.1 Ruolo del backend

Il backend svolge diverse responsabilità fondamentali:

- validare input applicativi;
- creare e aggiornare `ChangeRequest`;
- registrare `ChangeEvent` di audit;
- coordinare workflow GitLab;
- avviare o controllare workflow Tekton;
- leggere stato Argo CD;
- raccogliere runtime evidence Kubernetes/OpenShift;
- applicare regole di environment awareness;
- applicare guardrail fail-closed;
- fornire dati alla UI.

Questa centralizzazione evita che ogni operatore o script debba conoscere direttamente tutti i dettagli tecnici dei sistemi integrati.

### 13.2 ChangeService

`ChangeService` è uno dei componenti principali del backend.

Il suo ruolo è orchestrare le operazioni legate alle ChangeRequest.

Esempi di operazioni coordinate da `ChangeService`:

- creazione di una nuova ChangeRequest;
- aggiornamento dello stato lifecycle;
- aggiornamento dello stato runtime;
- salvataggio degli eventi di audit;
- raccolta evidence;
- controllo deployment;
- avvio validazione Tekton;
- controllo risultato validazione;
- preparazione dei dati che saranno poi mostrati in UI.

Vista concettuale:

```text
ChangeService
      |
      +--> ChangeRepository
      +--> ChangeEventRepository
      +--> EvidenceRepository
      +--> GitLab workflow adapter
      +--> Kubernetes runtime adapter
      +--> Tekton runtime adapter
      +--> Argo CD runtime adapter
```

La logica applicativa rimane quindi concentrata in un servizio coerente, mentre i dettagli tecnici restano delegati ad adapter e interfacce.

### 13.3 Repository layer

Il repository layer astrae la persistenza.

Questo significa che il resto del codice non deve conoscere direttamente tutti i dettagli SQL o la struttura fisica delle tabelle.

Il repository layer gestisce principalmente:

- `ChangeRequest`;
- `ChangeEvent`;
- `Evidence`.

Questa separazione rende il codice più leggibile, testabile e manutenibile.

Per esempio, quando il workflow produce una nuova evidenza, il servizio applicativo non deve sapere come viene scritta esattamente nel database. Deve solo passare i dati al repository corretto.

### 13.4 Service options e dependency injection

Il backend usa opzioni di configurazione per collegare resolver, registry e adapter.

Esempi di opzioni già presenti nel codice:

- `WithTechnicalRuntimeTargetResolver`;
- `WithRuntimeClientProviderRegistry`;
- `WithRuntimeClientSecretRefsRegistry`;
- `WithKubernetesRuntimeClientProviderRegistry`.

Questo approccio è utile perché permette di iniettare implementazioni diverse in base al contesto.

Durante i test, per esempio, è possibile usare resolver o provider simulati. In runtime, invece, vengono collegati provider e adapter reali o conservativi.

Il vantaggio principale è che il codice può essere validato anche senza avere a disposizione un cluster fisico aggiuntivo.

### 13.5 Runtime target resolution

La runtime target resolution è il processo con cui il backend traduce un ambiente logico in un target tecnico.

Esempio logico:

```text
targetEnvironment = staging
```

Target tecnico risultante nella baseline corrente:

```text
clusterName = ocp-dev
kubernetesNamespace = devops-ci-staging
tektonNamespace = devops-ci-staging
argocdApplicationName = demo-go-color-app-staging
validationPath = apps/demo-go-color-app/overlays/staging
```

Questo modello è fondamentale perché evita hardcoding e rende esplicito dove deve essere eseguita ogni azione tecnica.

La stessa logica vale per production:

```text
targetEnvironment = production
clusterName = ocp-dev
kubernetesNamespace = devops-ci-production
tektonNamespace = devops-ci-production
argocdApplicationName = demo-go-color-app-production
validationPath = apps/demo-go-color-app/overlays/production
```

### 13.6 EnvironmentClusterResolver

`EnvironmentClusterResolver` collega l’Environment Catalog al Cluster Registry.

Il suo compito è rispondere alla domanda:

```text
Dato un ambiente, quale cluster e quale configurazione runtime devo usare?
```

Nel modello corrente:

- `dev` risolve verso `ocp-dev` e `devops-ci-demo`;
- `staging` risolve verso `ocp-dev` e `devops-ci-staging`;
- `production` risolve verso `ocp-dev` e `devops-ci-production`.

Nei test di readiness multi-cluster, staging e production sono stati anche simulati come target esterni distinti:

- `staging` -> `ocp-staging-simulated`;
- `production` -> `ocp-production-simulated`.

Questa simulazione dimostra che il codice non è vincolato rigidamente a `ocp-dev`.

### 13.7 TechnicalRuntimeTarget

`TechnicalRuntimeTarget` è il risultato della risoluzione del target runtime.

Contiene le informazioni tecniche che servono alle operazioni successive, tra cui:

- target environment;
- environment name;
- cluster name;
- cluster display name;
- cluster enabled flag;
- Kubernetes namespace;
- Tekton namespace;
- Tekton pipeline name;
- Argo CD Application name;
- Git target branch;
- validation path.

Questo oggetto è importante perché rende esplicita la destinazione tecnica di ogni workflow.

Un’operazione come `check-deployment` o `validate` non deve decidere autonomamente dove andare. Deve usare il target tecnico risolto.

### 13.8 RuntimeClientProviderRegistry

`RuntimeClientProviderRegistry` collega un cluster risolto a un provider runtime.

Un provider runtime descrive come costruire o selezionare i client necessari per operare su un cluster.

Nel baseline corrente, `ocp-dev` è il provider runtime principale.

Il comportamento corretto per i cluster non configurati è fail-closed.

Questo significa:

- se il provider manca, l’operazione fallisce;
- se il provider è disabled, l’operazione fallisce;
- il sistema non deve ricadere silenziosamente su `ocp-dev`.

Questa regola è fondamentale per il futuro multi-cluster.

Se domani `staging` dovesse puntare a un cluster fisico dedicato, un errore di configurazione non deve causare un’esecuzione involontaria su `ocp-dev`.

### 13.9 Secret reference registry

Il backend supporta anche un modello di Secret references.

Una Secret reference non contiene il valore del Secret. Contiene solo riferimenti a dove il Secret si trova e quali chiavi devono essere lette.

Questo approccio consente di predisporre il runtime per cluster futuri senza esporre credenziali nel codice, nei log, nella documentazione o nelle evidence.

La regola operativa è semplice:

```text
reference sì, valore raw no
```

Il backend può quindi associare a un provider runtime informazioni come:

- cluster name;
- namespace del Secret;
- nome del Secret;
- chiave token;
- eventuali chiavi tecniche richieste.

Il caricamento dei valori reali resta protetto da loader, allow-list e flag disabled-by-default.

### 13.10 Fail-closed come principio applicativo

Nel DevOps Control Plane, fail-closed significa che una configurazione incompleta o non sicura deve bloccare l’operazione.

Esempi:

- ambiente sconosciuto;
- cluster sconosciuto;
- cluster disabled;
- provider mancante;
- provider disabled;
- Secret reference non allow-listed;
- factory disabled;
- token mancante;
- API URL mancante;
- raw CA non supportata;
- kubeconfig non supportato.

Questo comportamento non è un difetto. È un guardrail di sicurezza.

Un errore esplicito è preferibile a un’azione eseguita nel namespace o nel cluster sbagliato.

### 13.11 Relazione con i test

Il backend è stato validato con test unitari e test di non regressione.

Esempi di aspetti coperti:

- Environment Catalog;
- Cluster Registry;
- EnvironmentClusterResolver;
- TechnicalRuntimeTarget;
- RuntimeClientProviderRegistry;
- provider missing fail-closed;
- provider disabled fail-closed;
- simulazione staging e production cluster;
- Secret reference model;
- factory disabled-by-default.

I test più recenti rafforzano la readiness multi-cluster simulando:

```text
staging -> ocp-staging-simulated
production -> ocp-production-simulated
```

e verificando che non ci sia fallback verso `ocp-dev`.

### 13.12 Sintesi

Il backend Go è la parte che trasforma il DevOps Control Plane da semplice dashboard a vero control plane applicativo.

Le sue responsabilità principali sono:

- orchestrare workflow;
- persistere stato;
- produrre audit;
- risolvere target runtime;
- selezionare provider;
- raccogliere evidence;
- applicare guardrail;
- alimentare la UI.

Grazie a questa architettura, il progetto può oggi operare sulla baseline namespace-isolated e, allo stesso tempo, essere pronto a supportare il futuro multi-cluster reale quando l’infrastruttura sarà disponibile.

## 14. PostgreSQL e persistenza

PostgreSQL è il database relazionale usato dal DevOps Control Plane per conservare lo stato applicativo della piattaforma.

La persistenza è una parte fondamentale del progetto perché il DevOps Control Plane non deve limitarsi a invocare API esterne. Il sistema deve ricordare cosa è stato richiesto, cosa è stato eseguito, quali eventi sono avvenuti e quali evidenze sono state raccolte.

Senza persistenza, una ChangeRequest sarebbe solo un’operazione temporanea. Con PostgreSQL, invece, una ChangeRequest diventa un oggetto tracciabile, consultabile e auditabile nel tempo.

Vista semplificata:

```text
ChangeRequest
      |
      +--> ChangeEvent
      |
      +--> Evidence
      |
      +--> PostgreSQL
```

PostgreSQL rappresenta quindi la memoria applicativa del control plane.

### 14.1 Perché PostgreSQL

PostgreSQL è stato scelto perché offre caratteristiche adatte a una piattaforma di controllo e audit:

- modello relazionale;
- transazioni;
- consistenza dei dati;
- interrogazioni affidabili;
- supporto a dati strutturati;
- maturità operativa;
- strumenti consolidati per backup e restore.

Nel progetto, PostgreSQL conserva lo stato delle ChangeRequest e delle informazioni correlate. Questo permette alla UI, alle API e ai runbook operativi di lavorare su dati persistenti e non su informazioni volatili.

### 14.2 Cosa viene persistito

Le entità principali persistite sono:

- `ChangeRequest`;
- `ChangeEvent`;
- `Evidence`.

Queste tre entità rispondono a tre domande diverse.

`ChangeRequest` risponde alla domanda:

```text
Quale cambiamento è stato richiesto e qual è il suo stato corrente?
```

`ChangeEvent` risponde alla domanda:

```text
Quali eventi sono avvenuti durante il ciclo di vita della richiesta?
```

`Evidence` risponde alla domanda:

```text
Quali prove tecniche sono state raccolte per dimostrare lo stato osservato?
```

### 14.3 ChangeRequest persistence

Una `ChangeRequest` rappresenta una richiesta di cambiamento applicativo o tecnico.

Esempi di informazioni associate a una ChangeRequest:

- numero della richiesta;
- titolo;
- descrizione;
- applicazione target;
- ambiente target;
- requester;
- stato del processo;
- stato runtime;
- riferimenti Git;
- timestamp di creazione e aggiornamento.

La persistenza della ChangeRequest permette di ricostruire lo stato corrente del workflow anche dopo riavvii applicativi, rollout del pod o nuove sessioni utente.

Esempi reali usati durante la validazione:

- `CHG-2026-0049` per staging;
- `CHG-2026-0050` per production.

Questi record sono stati usati dalla UI per mostrare runtime evidence e Tekton validation evidence.

### 14.4 ChangeEvent persistence

Un `ChangeEvent` rappresenta un evento di audit collegato a una ChangeRequest.

Esempi di eventi:

- ChangeRequest creata;
- workflow GitLab avviato;
- branch creato;
- merge request creata;
- validazione Tekton avviata;
- validazione Tekton completata;
- evidence raccolta;
- deployment controllato;
- errore registrato.

La persistenza degli eventi permette di ricostruire la storia della richiesta.

Questo è importante perché un operatore non deve vedere solo lo stato finale. Deve anche poter capire quali passaggi hanno portato a quello stato.

Gli eventi sono quindi la base dell’audit trail applicativo.

### 14.5 Evidence persistence

`Evidence` rappresenta una prova tecnica raccolta dal sistema.

Le evidenze possono provenire da più fonti:

- Kubernetes/OpenShift;
- Argo CD;
- Tekton;
- GitLab;
- workflow interni del DevOps Control Plane.

Esempi di evidenze:

- stato di una Argo CD Application;
- stato di un Deployment;
- replica readiness;
- route health;
- PipelineRun Tekton;
- TaskRun fallite;
- validation path;
- failed task count;
- stato sanitized.

Le evidenze devono essere associate alla ChangeRequest corretta.

Questo collegamento è fondamentale perché consente alla UI di mostrare, nel dettaglio di una ChangeRequest, esattamente le prove tecniche relative a quella richiesta.

### 14.6 Sanitizzazione delle evidenze

Le evidenze devono essere sanificate prima di essere considerate sicure per persistenza, visualizzazione o condivisione.

Le evidenze possono contenere metadati tecnici sicuri, come:

- namespace;
- nomi di risorse;
- nomi di PipelineRun;
- nomi di Argo CD Application;
- validation path;
- status;
- reason;
- failed task count.

Le evidenze non devono contenere:

- token;
- password;
- kubeconfig raw;
- private key;
- contenuto Secret decodificato;
- credenziali applicative;
- bearer token;
- materiale sensibile non necessario.

La regola operativa è:

```text
persist readable evidence, never persist raw credentials
```

### 14.7 Relazione tra PostgreSQL e UI

La UI non deve ricostruire lo stato operativo interrogando direttamente tutti i sistemi esterni.

La UI legge lo stato applicativo dal backend, che a sua volta usa PostgreSQL per recuperare dati persistiti.

Questo vale per:

- lista ChangeRequest;
- dettaglio ChangeRequest;
- audit events;
- runtime evidence;
- Tekton validation evidence;
- dashboard latest ChangeRequest.

La UI può quindi presentare una vista coerente anche quando le evidenze sono state raccolte in momenti diversi.

### 14.8 Relazione tra PostgreSQL e runtime evidence

La runtime evidence osserva il mondo esterno, ma viene conservata nel mondo applicativo.

Esempio:

```text
OpenShift Deployment status
       |
       v
runtime evidence
       |
       v
PostgreSQL
       |
       v
ChangeRequest detail UI
```

Questo flusso permette di conservare una fotografia dello stato osservato durante una specifica fase del workflow.

La stessa logica vale per Tekton validation evidence:

```text
Tekton PipelineRun
       |
       v
check-validation
       |
       v
validation evidence
       |
       v
PostgreSQL
       |
       v
UI evidence card
```

### 14.9 Backup e restore

PostgreSQL è incluso nella baseline operativa del progetto.

Sono disponibili runbook dedicati per:

- backup;
- restore;
- restore isolato;
- disaster recovery;
- validazione post-restore;
- limiti e raccomandazioni per produzione.

Il backup di PostgreSQL è importante perché il database contiene:

- stato delle richieste;
- eventi di audit;
- evidenze raccolte;
- storico utile per troubleshooting e compliance.

La perdita del database non cancella necessariamente lo stato GitOps presente nel repository o nel cluster, ma compromette la memoria applicativa del DevOps Control Plane.

Per questo motivo il database deve essere trattato come componente critico.

### 14.10 Restore isolato

Il restore isolato è preferibile durante test e verifiche.

Un restore isolato permette di validare un backup senza sovrascrivere l’istanza runtime attiva.

Questo approccio riduce il rischio operativo.

La regola generale è:

```text
validate restore safely before considering active replacement
```

Un restore in ambiente attivo deve essere eseguito solo con procedura approvata, evidenze raccolte e chiara responsabilità operativa.

### 14.11 Relazione con operability

PostgreSQL è parte della Fase 10 di operability.

Gli operatori devono verificare:

- pod PostgreSQL running;
- connettività dal backend;
- readiness del DevOps Control Plane;
- assenza di errori di connessione;
- validità dei backup;
- possibilità di restore isolato.

La readiness applicativa `/readyz` dipende anche dalla possibilità del backend di usare correttamente le proprie dipendenze.

Se PostgreSQL non è disponibile, il DevOps Control Plane può perdere la capacità di leggere o aggiornare ChangeRequest, eventi ed evidenze.

### 14.12 Relazione con audit e compliance

La persistenza su PostgreSQL supporta audit e compliance applicativa.

Il sistema può dimostrare:

- chi ha richiesto un cambiamento;
- quale ambiente è stato scelto;
- quali eventi sono avvenuti;
- quali evidenze sono state raccolte;
- quale validazione Tekton è stata eseguita;
- quale stato Argo CD è stato osservato;
- quale stato runtime è stato rilevato.

Questo rende il DevOps Control Plane più di un semplice orchestratore tecnico.

Il sistema diventa anche una fonte di tracciabilità.

### 14.13 Limiti attuali

La baseline attuale include PostgreSQL funzionante e runbook operativi, ma non rappresenta ancora necessariamente una configurazione enterprise definitiva per ogni contesto produttivo.

Restano possibili evoluzioni future:

- alta disponibilità PostgreSQL;
- retention policy più articolate;
- archiviazione storica delle evidenze;
- metriche specifiche database;
- alerting dedicato;
- restore periodici misurati con RPO e RTO formali.

Questi aspetti fanno parte del percorso di production hardening e non invalidano la baseline attuale.

### 14.14 Sintesi

PostgreSQL è la memoria persistente del DevOps Control Plane.

Conserva:

- richieste;
- eventi;
- evidenze;
- storia operativa.

Grazie a PostgreSQL, il progetto può offrire:

- auditabilità;
- consultazione storica;
- UI coerente;
- troubleshooting;
- operability;
- backup e restore;
- base per future esigenze di compliance.

La persistenza è quindi uno dei pilastri che trasformano il DevOps Control Plane da automazione temporanea a piattaforma di controllo strutturata.

## 15. Modello dati

Il modello dati del DevOps Control Plane descrive le informazioni che la piattaforma deve conservare per governare un cambiamento in modo tracciabile, auditabile e verificabile.

Il modello non e pensato solo per salvare dati applicativi. Il modello rappresenta il ciclo completo di una richiesta di cambiamento: richiesta, stato, eventi, workflow tecnici ed evidenze.

Le entita principali sono:

- `ChangeRequest`;
- `ChangeEvent`;
- `Evidence`.

Vista semplificata:

```text
ChangeRequest
      |
      +--> ChangeEvent
      |
      +--> Evidence
```

Questa struttura permette di rispondere a tre domande fondamentali:

```text
Che cosa e stato richiesto?
Che cosa e successo durante il workflow?
Quali prove tecniche dimostrano lo stato osservato?
```

### 15.1 ChangeRequest

`ChangeRequest` e l'entita principale del dominio.

Una ChangeRequest rappresenta una richiesta di cambiamento controllata dal DevOps Control Plane. La ChangeRequest collega il mondo funzionale con il mondo tecnico: da un lato descrive cosa si vuole cambiare, dall'altro permette al sistema di eseguire workflow tecnici come GitLab, Argo CD, Tekton e runtime evidence.

Informazioni concettuali tipiche:

- numero della richiesta;
- titolo;
- descrizione;
- applicazione target;
- ambiente target;
- requester;
- stato del processo;
- stato runtime;
- branch Git o riferimento Git;
- timestamp di creazione;
- timestamp di aggiornamento.

Esempi reali usati nella baseline:

- `CHG-2026-0049` per staging;
- `CHG-2026-0050` per production.

Queste ChangeRequest sono importanti perche hanno validato il workflow namespace-isolated per staging e production, includendo Tekton validation evidence e UI rendering.

### 15.2 Numero ChangeRequest

Il numero della ChangeRequest e l'identificativo leggibile usato da operatori e UI.

Esempi:

```text
CHG-2026-0049
CHG-2026-0050
```

Il numero deve essere stabile e riconoscibile, perche compare in dashboard, pagine di dettaglio, audit, evidence, troubleshooting e runbook operativi.

### 15.3 Target environment

`targetEnvironment` indica l'ambiente logico richiesto per la ChangeRequest.

Ambienti correnti:

- `dev`;
- `staging`;
- `production`.

Il target environment e fondamentale perche determina il runtime target tecnico.

Nel baseline corrente:

```text
dev        -> ocp-dev / devops-ci-demo
staging    -> ocp-dev / devops-ci-staging
production -> ocp-dev / devops-ci-production
```

Il target environment deve essere persistito perche tutte le operazioni successive devono sapere quale ambiente era stato richiesto. Senza questo campo, non sarebbe possibile distinguere correttamente una validazione staging da una validazione production.

### 15.4 Stato del processo

Lo stato del processo descrive l'avanzamento logico della ChangeRequest.

Esempi concettuali:

- creata;
- in lavorazione;
- in validazione;
- completata;
- fallita;
- in attesa di azione.

Questo stato non deve essere confuso con lo stato runtime. Una richiesta puo essere stata processata correttamente dal backend, ma il deployment puo comunque non essere pronto.

### 15.5 Stato runtime

Lo stato runtime descrive cosa e stato osservato nei sistemi tecnici.

Puo derivare da:

- Kubernetes/OpenShift;
- Argo CD;
- Tekton;
- controlli interni;
- evidence raccolte.

Questa distinzione evita di dichiarare riuscito un cambiamento solo perche il processo applicativo e avanzato. Il DevOps Control Plane deve distinguere successo del processo, successo tecnico del runtime, fallimento della validazione, fallimento del deployment, evidence mancante o evidence incompleta.

### 15.6 ChangeEvent

`ChangeEvent` rappresenta un evento di audit collegato a una ChangeRequest.

Esempi di eventi:

- ChangeRequest creata;
- branch Git creato;
- merge request aperta;
- merge request completata;
- runtime evidence raccolta;
- check deployment eseguito;
- validazione Tekton avviata;
- check-validation completato;
- errore registrato.

Gli eventi permettono di ricostruire la storia della richiesta.

Vista concettuale:

```text
ChangeRequest CHG-2026-0050
      |
      +--> event: created
      +--> event: validate requested
      +--> event: PipelineRun created
      +--> event: check-validation succeeded
      +--> event: evidence stored
```

### 15.7 Audit trail

L'audit trail e la sequenza degli eventi associati a una ChangeRequest.

L'audit trail e utile per:

- troubleshooting;
- revisione operativa;
- verifica dei passaggi eseguiti;
- ricostruzione storica;
- responsabilita e governance.

Il valore dell'audit trail non e solo tecnico. L'audit trail aiuta anche a spiegare perche una richiesta si trova in un certo stato.

### 15.8 Evidence

`Evidence` rappresenta una prova tecnica associata a una ChangeRequest.

Una evidence risponde alla domanda:

```text
Quale stato tecnico e stato osservato?
```

Esempi di evidence:

- stato Argo CD Application;
- stato Deployment;
- replica readiness;
- route health;
- PipelineRun Tekton;
- TaskRun Tekton;
- failed task count;
- validation path;
- stato sanitized;
- diagnostica runtime.

La evidence deve sempre essere collegata al contesto corretto: ChangeRequest, target environment, namespace, timestamp e tipo di evidence.

### 15.9 Tipi di evidence

Nel progetto si possono distinguere varie famiglie di evidence:

- runtime evidence;
- deployment evidence;
- Tekton validation evidence;
- Argo CD deployment evidence;
- GitLab workflow evidence;
- diagnostic evidence;
- raw sanitized evidence.

Queste famiglie non devono essere confuse. Una Tekton validation evidence dimostra l'esito di una PipelineRun. Una runtime evidence dimostra invece lo stato osservato nel cluster.

### 15.10 Evidence sanitization

La evidence deve essere sanificata.

Dati ammessi:

- nomi di namespace;
- nomi di risorse;
- nomi di PipelineRun;
- nomi di Argo CD Application;
- validation path;
- status;
- reason;
- failed task count;
- timestamp;
- stato `evidence sanitized=true`.

Dati vietati:

- token;
- password;
- kubeconfig raw;
- private key;
- bearer token;
- contenuto Secret decodificato;
- raw CA non gestita in modo sicuro.

La sanitizzazione e un requisito di sicurezza, non un dettaglio secondario.

### 15.11 Relazione tra ChangeRequest ed Evidence

Una ChangeRequest puo avere piu evidence nel tempo.

Esempio:

```text
CHG-2026-0050
      |
      +--> Argo CD evidence
      +--> deployment evidence
      +--> Tekton validation evidence
      +--> runtime diagnostics evidence
```

Questo e importante perche il workflow puo essere eseguito in piu passaggi. La UI deve mostrare le evidence piu rilevanti per l'operatore, in particolare la latest validation evidence quando disponibile.

### 15.12 Latest validation evidence

La latest validation evidence e l'evidenza di validazione piu recente associata a una ChangeRequest.

La UI deve mostrare in modo chiaro:

- PipelineRun;
- Tekton namespace;
- Pipeline;
- Git revision o branch;
- validation path;
- status;
- reason;
- failed task count;
- stato sanitized.

Esempio production validato:

```text
ChangeRequest: CHG-2026-0050
Tekton namespace: devops-ci-production
PipelineRun: devops-cp-validate-chg-2026-0050-8wqtv
validationPath: apps/demo-go-color-app/overlays/production
failedTaskCount: 0
evidence sanitized: true
result: Succeeded
```

### 15.13 Relazione con Environment Catalog

Il modello dati e collegato all'Environment Catalog perche la ChangeRequest contiene il target environment.

Il target environment viene risolto in:

- cluster name;
- Kubernetes namespace;
- Tekton namespace;
- Argo CD Application;
- validation path.

Queste informazioni possono poi comparire nelle evidence. Questo collegamento e essenziale per evitare ambiguita tra ambienti.

### 15.14 Relazione con Cluster Registry

Il modello dati deve essere compatibile con il Cluster Registry.

Oggi dev, staging e production sono namespace-isolated su `ocp-dev`. Domani staging o production potranno puntare a cluster fisici diversi.

Per questo motivo le evidence e i runtime target devono preservare informazioni come:

- target environment;
- cluster name;
- namespace;
- provider selection.

Quando arrivera un cluster reale, sara importante dimostrare che il workflow non e ricaduto per errore su `ocp-dev`.

### 15.15 Relazione con la UI

La UI e una vista del modello dati e delle evidenze.

La UI usa questi dati per mostrare:

- dashboard;
- lista ChangeRequest;
- dettaglio ChangeRequest;
- audit log;
- runtime evidence card;
- Tekton validation card;
- raw sanitized evidence.

La UI non deve inventare uno stato. La UI deve rappresentare lo stato persistito e le evidence raccolte.

### 15.16 Relazione con operability

Il modello dati sostiene anche l'operability.

Durante troubleshooting o manutenzione, un operatore puo usare ChangeRequest, eventi ed evidence per capire:

- quale ambiente era coinvolto;
- quale namespace era coinvolto;
- quale workflow e stato eseguito;
- quale PipelineRun e stata creata;
- quale stato Argo CD e stato osservato;
- quale stato runtime e stato raccolto;
- dove si e verificato un errore.

Senza modello dati persistente, l'operatore dovrebbe ricostruire tutto da log e sistemi esterni.

### 15.17 Relazione con security

Il modello dati deve rispettare i guardrail di sicurezza.

In particolare:

- non deve persistere Secret raw;
- non deve persistere token;
- non deve persistere kubeconfig raw;
- non deve rendere disponibili credenziali in UI;
- deve conservare solo informazioni operative sicure;
- deve indicare se la evidence e sanificata.

La sicurezza del modello dati e parte della sicurezza della piattaforma.

### 15.18 Relazione con multi-cluster readiness

Il modello dati supporta la readiness multi-cluster perche conserva il target environment e le informazioni runtime correlate.

I test post-Fase 15 hanno validato target simulati:

```text
staging -> ocp-staging-simulated
production -> ocp-production-simulated
```

Il comportamento atteso e:

- nessun fallback silenzioso verso `ocp-dev`;
- provider mancante fail-closed;
- provider disabled fail-closed.

Il modello dati deve continuare a rendere visibile quale ambiente e quale cluster erano attesi.

### 15.19 Esempio completo

Esempio concettuale basato sulla validazione production:

```text
ChangeRequest
  number: CHG-2026-0050
  targetEnvironment: production

Runtime target
  clusterName: ocp-dev
  kubernetesNamespace: devops-ci-production
  tektonNamespace: devops-ci-production
  argocdApplicationName: demo-go-color-app-production
  validationPath: apps/demo-go-color-app/overlays/production

Evidence
  type: validation
  pipelineRun: devops-cp-validate-chg-2026-0050-8wqtv
  status: True
  reason: Succeeded
  failedTaskCount: 0
  sanitized: true
```

Questo esempio mostra come dominio, runtime target ed evidence siano collegati.

### 15.20 Sintesi

Il modello dati del DevOps Control Plane permette di trasformare workflow tecnici distribuiti in una storia coerente e persistente.

Le entita principali sono:

- `ChangeRequest`, che rappresenta la richiesta;
- `ChangeEvent`, che rappresenta la storia;
- `Evidence`, che rappresenta la prova tecnica.

Il modello dati collega governance, automazione, audit, runtime evidence, UI e operability.

Questo e uno dei motivi per cui il DevOps Control Plane puo essere considerato una piattaforma di controllo e non una semplice raccolta di script.

## 17. ChangeRequest lifecycle

Una `ChangeRequest` rappresenta il punto centrale del DevOps Control Plane.

Tutte le azioni principali partono da una richiesta di cambiamento e vengono collegate a quella richiesta: eventi, workflow GitLab, validazioni Tekton, controlli Argo CD, runtime evidence, audit trail e visualizzazione nella UI.

La ChangeRequest permette di trasformare un insieme di operazioni tecniche distribuite in un workflow coerente e tracciabile.

Vista semplificata:

```text
ChangeRequest
      |
      +--> lifecycle status
      +--> runtime status
      +--> ChangeEvent audit trail
      +--> GitLab workflow
      +--> Argo CD checks
      +--> Tekton validation
      +--> runtime evidence
      +--> UI detail view
```

### 17.1 Perche serve una ChangeRequest

Senza una ChangeRequest, le operazioni DevOps restano sparse tra strumenti diversi.

Per esempio:

- GitLab conosce branch, commit e merge request;
- Argo CD conosce sync e health delle applicazioni;
- Tekton conosce PipelineRun e TaskRun;
- Kubernetes/OpenShift conosce deployment, pod, service e route;
- PostgreSQL conserva dati applicativi.

La ChangeRequest collega questi elementi e fornisce un riferimento unico.

Questo riferimento unico consente di rispondere a domande operative come:

- quale cambiamento era richiesto;
- quale ambiente era coinvolto;
- quale namespace e stato usato;
- quale validazione e stata lanciata;
- quale evidence e stata raccolta;
- quale stato finale ha assunto il workflow.

### 17.2 Informazioni principali della ChangeRequest

Una ChangeRequest contiene informazioni funzionali e tecniche.

Informazioni funzionali:

- numero della richiesta;
- titolo;
- descrizione;
- requester;
- applicazione coinvolta.

Informazioni tecniche:

- target environment;
- branch o revisione Git;
- stato del processo;
- stato runtime;
- riferimenti a eventi;
- riferimenti a evidenze.

Il campo `targetEnvironment` e particolarmente importante perche guida la risoluzione del runtime target.

### 17.3 Target environment

Gli ambienti logici attualmente validati sono:

- `dev`;
- `staging`;
- `production`.

La baseline corrente usa namespace isolation sul cluster `ocp-dev`:

```text
dev        -> ocp-dev / devops-ci-demo
staging    -> ocp-dev / devops-ci-staging
production -> ocp-dev / devops-ci-production
```

Il lifecycle di una ChangeRequest deve preservare l'ambiente richiesto.

Una richiesta per `staging` non deve essere trattata come `dev`.

Una richiesta per `production` non deve essere trattata come `dev`.

Questa regola e ancora piu importante in ottica multi-cluster futura.

### 17.4 Creazione della ChangeRequest

Il workflow inizia con la creazione della ChangeRequest.

Durante la creazione, il backend deve:

- validare i dati in ingresso;
- verificare il target environment;
- creare il record persistente;
- produrre un primo evento di audit;
- inizializzare gli stati del processo e del runtime.

Il risultato atteso e che la richiesta sia disponibile tramite API e UI.

La dashboard puo poi mostrare la richiesta piu recente come riferimento operativo.

### 17.5 Lifecycle status

Il lifecycle status rappresenta lo stato logico del processo.

Esempi concettuali:

- richiesta creata;
- workflow in corso;
- validazione richiesta;
- validazione completata;
- errore;
- completata.

Questo stato descrive il percorso applicativo della richiesta.

Il lifecycle status non deve essere confuso con il runtime status.

### 17.6 Runtime status

Il runtime status rappresenta lo stato tecnico osservato.

Esempi:

- deployment pronto;
- deployment non pronto;
- Argo CD `Synced`;
- Argo CD `Healthy`;
- PipelineRun `Succeeded`;
- PipelineRun `Failed`;
- evidence mancante;
- check non ancora eseguito.

Il runtime status deriva dalle integrazioni tecniche e dalle evidence raccolte.

### 17.7 ChangeEvent durante il lifecycle

Ogni passaggio importante del lifecycle puo generare un `ChangeEvent`.

Esempi:

```text
ChangeRequest created
Git branch prepared
Merge request created
Runtime evidence collected
Deployment checked
Tekton validation started
Tekton validation checked
Evidence stored
Workflow failed
Workflow completed
```

Questi eventi formano l'audit trail della richiesta.

### 17.8 Workflow tecnico collegato

Una ChangeRequest puo attivare o coordinare diversi workflow tecnici:

- workflow GitLab;
- collect-evidence;
- check-deployment;
- validate;
- check-validation;
- UI evidence rendering.

Questi workflow non sono indipendenti dalla ChangeRequest: devono essere associati alla richiesta corretta.

### 17.9 collect-evidence

`collect-evidence` raccoglie informazioni runtime dal target environment.

Nel baseline corrente, questo significa osservare il namespace corretto:

- `devops-ci-demo` per dev;
- `devops-ci-staging` per staging;
- `devops-ci-production` per production.

L'evidence deve indicare chiaramente ambiente, namespace e risorse osservate.

### 17.10 check-deployment

`check-deployment` verifica se l'applicazione target e pronta nel namespace corretto.

Il check deve restare environment-specific.

Un deployment pronto in `devops-ci-demo` non dimostra che staging o production siano pronti.

### 17.11 validate

`validate` avvia la validazione Tekton.

La validazione deve usare il path GitOps corretto per l'ambiente.

Esempi validati:

- staging: `apps/demo-go-color-app/overlays/staging`;
- production: `apps/demo-go-color-app/overlays/production`.

Questa distinzione evita di validare un overlay sbagliato.

### 17.12 check-validation

`check-validation` legge l'esito della PipelineRun Tekton e produce validation evidence.

La validation evidence deve includere:

- ChangeRequest;
- target environment;
- Tekton namespace;
- PipelineRun;
- validation path;
- status;
- reason;
- failed task count;
- sanitization state.

Esempio staging:

```text
ChangeRequest: CHG-2026-0049
Tekton namespace: devops-ci-staging
PipelineRun: devops-cp-validate-chg-2026-0049-nd7rm
validationPath: apps/demo-go-color-app/overlays/staging
result: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

Esempio production:

```text
ChangeRequest: CHG-2026-0050
Tekton namespace: devops-ci-production
PipelineRun: devops-cp-validate-chg-2026-0050-8wqtv
validationPath: apps/demo-go-color-app/overlays/production
result: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

### 17.13 UI e lifecycle

La UI rende visibile il lifecycle della ChangeRequest.

La dashboard mostra la richiesta piu recente.

La pagina di dettaglio mostra:

- dati principali della richiesta;
- audit log;
- runtime evidence;
- Tekton validation evidence;
- raw sanitized evidence quando utile;
- stato e contesto operativo.

La UI non deve mostrare solo un elenco statico di dati. Deve aiutare l'operatore a capire cosa e successo e quale stato e stato osservato.

### 17.14 Fail-closed nel lifecycle

Il lifecycle deve rispettare i guardrail fail-closed.

Esempi:

- target environment sconosciuto;
- cluster reference non valida;
- provider runtime mancante;
- provider runtime disabled;
- Secret reference non allow-listed;
- factory non configurata;
- factory disabled.

In questi casi l'operazione deve fallire in modo esplicito.

Non deve essere eseguito un fallback silenzioso verso `ocp-dev`.

### 17.15 Relazione con multi-cluster readiness

Il lifecycle e oggi validato sulla baseline namespace-isolated.

Tuttavia il codice e stato anche validato con target simulati:

```text
staging -> ocp-staging-simulated
production -> ocp-production-simulated
```

I test confermano che:

- staging non ricade su `ocp-dev`;
- production non ricade su `ocp-dev`;
- provider mancante fallisce fail-closed;
- provider disabled fallisce fail-closed.

Questo dimostra che il lifecycle e predisposto al futuro multi-cluster reale.

### 17.16 Sintesi

Il lifecycle della ChangeRequest e il filo conduttore del DevOps Control Plane.

La ChangeRequest collega:

- governance;
- workflow GitLab;
- runtime target resolution;
- Argo CD;
- Tekton;
- Kubernetes/OpenShift;
- evidence;
- audit;
- UI;
- operability.

Questo modello consente di trasformare operazioni tecniche distribuite in un processo unico, persistente, verificabile e comprensibile.

## 18. GitLab Merge Request workflow

Il workflow GitLab collega una `ChangeRequest` al ciclo di vita del codice e della configurazione GitOps.

Nel DevOps Control Plane, GitLab non viene usato solo come repository remoto. GitLab rappresenta il punto in cui una modifica viene preparata, versionata, proposta e infine resa disponibile al modello GitOps.

Il workflow GitLab permette di collegare:

- richiesta di cambiamento;
- branch Git;
- commit;
- merge request;
- merge;
- revisione usata da Argo CD e Tekton;
- audit ed evidence.

Vista semplificata:

```text
ChangeRequest
      |
      +--> branch Git
      +--> commit
      +--> Merge Request
      +--> merge
      +--> GitOps revision
      +--> Argo CD / Tekton validation
```

### 18.1 Perche GitLab e importante

GitLab e importante perche conserva la storia delle modifiche.

In un flusso GitOps, una modifica al runtime non dovrebbe essere un'azione manuale e non tracciata sul cluster. La modifica dovrebbe passare dal repository, essere revisionabile e avere un collegamento con una richiesta di cambiamento.

Il DevOps Control Plane usa GitLab per rendere il cambiamento:

- tracciabile;
- revisionabile;
- collegato a una ChangeRequest;
- auditabile;
- compatibile con GitOps.

### 18.2 Branch workflow

Il workflow puo creare o usare un branch dedicato alla ChangeRequest.

Il branch rappresenta l'area di lavoro tecnica del cambiamento.

Esempio concettuale:

```text
ChangeRequest: CHG-2026-0050
branch: change/CHG-2026-0050
```

Il branch permette di separare la modifica dal branch principale finche non e stata validata e approvata.

### 18.3 Commit workflow

Il commit rappresenta la modifica effettiva salvata in Git.

Nel contesto GitOps, un commit puo modificare manifest, overlay, configurazioni o altri file dichiarativi.

Il commit deve essere collegabile alla ChangeRequest.

Questo collegamento permette di sapere quale richiesta ha introdotto una modifica nel repository.

### 18.4 Merge Request

La Merge Request e il punto di revisione.

Una MR permette di confrontare la modifica proposta con il branch di destinazione, analizzare il diff e decidere se procedere.

Nel DevOps Control Plane, la MR e parte del workflow controllato.

La piattaforma puo conservare riferimenti come:

- branch sorgente;
- branch target;
- URL o identificativo della MR;
- stato della MR;
- risultato del merge.

### 18.5 Merge

Il merge porta la modifica nel branch principale o nel branch GitOps target.

Dopo il merge, gli strumenti GitOps possono osservare la revisione aggiornata.

Argo CD puo quindi confrontare il repository con lo stato del cluster.

Tekton puo validare path e contenuti GitOps in modo coerente con l'ambiente target.

### 18.6 Relazione con GitOps

Il workflow GitLab e collegato direttamente al modello GitOps.

GitLab conserva la modifica.

Argo CD legge la modifica.

Tekton valida la modifica.

Il DevOps Control Plane coordina il processo e conserva stato, eventi ed evidence.

Vista semplificata:

```text
GitLab repository
      |
      v
GitOps manifests
      |
      +--> Argo CD sync and health
      +--> Tekton validation
      +--> DevOps Control Plane evidence
```

### 18.7 Relazione con Kustomize overlays

Per staging e production, il workflow deve considerare i path GitOps corretti.

Esempi validati:

```text
apps/demo-go-color-app/overlays/staging
apps/demo-go-color-app/overlays/production
```

Questi path sono importanti perche una validazione staging non deve validare per errore l'overlay production, e una validazione production non deve validare per errore l'overlay staging.

Il validation path environment-specific e una parte essenziale del workflow.

### 18.8 Relazione con ChangeRequest

Il workflow GitLab deve restare collegato alla ChangeRequest.

La ChangeRequest conserva il contesto funzionale e operativo.

GitLab conserva il contesto tecnico della modifica.

Il collegamento tra i due permette di rispondere a domande come:

- quale ChangeRequest ha generato questo branch;
- quale MR e collegata alla richiesta;
- quale commit e stato validato;
- quale revisione e stata osservata da Argo CD;
- quale path e stato validato da Tekton.

### 18.9 Relazione con audit

Ogni passaggio importante del workflow GitLab puo produrre un ChangeEvent.

Esempi:

- branch creato;
- commit creato;
- merge request aperta;
- merge request aggiornata;
- merge completato;
- errore GitLab;
- workflow GitLab completato.

Questi eventi diventano parte dell'audit trail della ChangeRequest.

### 18.10 Relazione con evidence

Il workflow GitLab puo contribuire indirettamente alle evidence.

Per esempio, una Tekton validation evidence puo contenere o riferire:

- Git revision;
- branch;
- validation path;
- repository path.

Questi dati aiutano a capire quale contenuto e stato validato.

### 18.11 Relazione con Argo CD

Dopo che una modifica GitOps e disponibile nel repository, Argo CD puo osservarla.

Argo CD produce stato come:

- `Synced`;
- `OutOfSync`;
- `Healthy`;
- `Degraded`.

Il DevOps Control Plane raccoglie o rappresenta queste informazioni come deployment evidence o runtime evidence.

### 18.12 Relazione con Tekton

Tekton valida tecnicamente il contenuto GitOps.

Nel progetto, le validazioni importanti includono i path environment-specific:

- staging: `apps/demo-go-color-app/overlays/staging`;
- production: `apps/demo-go-color-app/overlays/production`.

Il risultato Tekton viene raccolto tramite `check-validation` e collegato alla ChangeRequest.

### 18.13 Errori tipici del workflow GitLab

Possibili errori:

- repository non raggiungibile;
- branch gia esistente;
- branch target errato;
- commit fallito;
- merge request non creata;
- merge request non mergeabile;
- problema di permessi;
- token GitLab non valido;
- errore TLS o trust bundle.

In questi casi il DevOps Control Plane deve registrare l'errore e preservare il contesto nella ChangeRequest.

### 18.14 Sicurezza del workflow GitLab

Il workflow GitLab puo richiedere token o credenziali.

Questi valori non devono essere stampati nei log, nelle evidence o nella documentazione.

La documentazione puo citare nomi di Secret o riferimenti, ma non deve contenere valori raw.

La regola operativa resta:

```text
reference yes, raw secret no
```

### 18.15 Stato corrente

Il workflow GitLab e stato completato end-to-end nelle fasi precedenti del progetto.

Oggi il suo ruolo nella guida finale e spiegare come la parte Git si integra con:

- ChangeRequest lifecycle;
- GitOps;
- Argo CD;
- Tekton;
- evidence;
- audit trail.

### 18.16 Sintesi

Il GitLab Merge Request workflow fornisce la dimensione Git del DevOps Control Plane.

Il workflow collega una richiesta di cambiamento a una modifica versionata, revisionabile e compatibile con GitOps.

Grazie a questo collegamento, il DevOps Control Plane puo dimostrare non solo che una validazione e stata eseguita, ma anche quale contenuto Git e stato validato e da quale ChangeRequest e nato il cambiamento.
