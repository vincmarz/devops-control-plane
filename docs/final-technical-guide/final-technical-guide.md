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

## 19. Workflow runtime

Il workflow runtime e l'insieme delle azioni tecniche che il DevOps Control Plane esegue o coordina dopo la creazione di una `ChangeRequest`.

Lo scopo del workflow runtime e verificare lo stato reale della modifica richiesta: non basta sapere che una richiesta esiste, o che un branch Git e stato creato. Il sistema deve anche osservare il runtime, validare il contenuto GitOps, raccogliere evidenze e rendere il risultato leggibile nella UI.

Nel progetto, il workflow runtime e composto principalmente da queste azioni:

- `collect-evidence`;
- `check-deployment`;
- `validate`;
- `check-validation`.

Vista semplificata:

```text
ChangeRequest
      |
      +--> resolve runtime target
      |
      +--> collect-evidence
      |
      +--> check-deployment
      |
      +--> validate
      |
      +--> check-validation
      |
      +--> persist evidence
      |
      +--> render evidence in UI
```

### 19.1 Obiettivo del workflow runtime

Il workflow runtime risponde a domande operative concrete:

- quale ambiente e stato scelto;
- quale namespace e stato controllato;
- quale Argo CD Application e stata osservata;
- quale deployment e stato verificato;
- quale validazione Tekton e stata eseguita;
- quale validation path e stato usato;
- quale PipelineRun ha prodotto il risultato;
- se il risultato e stato sanificato;
- se la UI mostra correttamente le informazioni raccolte.

Questo rende il DevOps Control Plane una piattaforma di controllo e non solo un frontend verso strumenti esistenti.

### 19.2 Risoluzione del runtime target

Prima di eseguire un'azione tecnica, il backend deve risolvere il runtime target.

Il runtime target deriva dal `targetEnvironment` della ChangeRequest.

Esempio staging:

```text
targetEnvironment = staging
clusterName = ocp-dev
kubernetesNamespace = devops-ci-staging
tektonNamespace = devops-ci-staging
argocdApplicationName = demo-go-color-app-staging
validationPath = apps/demo-go-color-app/overlays/staging
```

Esempio production:

```text
targetEnvironment = production
clusterName = ocp-dev
kubernetesNamespace = devops-ci-production
tektonNamespace = devops-ci-production
argocdApplicationName = demo-go-color-app-production
validationPath = apps/demo-go-color-app/overlays/production
```

Questo passaggio e fondamentale per evitare che un'azione venga eseguita nel namespace sbagliato.

### 19.3 collect-evidence

`collect-evidence` raccoglie informazioni tecniche dal runtime.

Nel baseline corrente, l'azione osserva risorse Kubernetes/OpenShift e stato Argo CD associati al target environment.

Esempi di dati raccolti:

- namespace target;
- deployment name;
- ready replicas;
- available replicas;
- updated replicas;
- pod status;
- service status;
- route status;
- Argo CD sync status;
- Argo CD health status.

L'evidence deve essere collegata alla ChangeRequest corretta.

L'evidence deve anche indicare l'ambiente e il namespace, perche dev, staging e production condividono oggi lo stesso cluster fisico `ocp-dev`.

### 19.4 check-deployment

`check-deployment` verifica se l'applicazione target risulta pronta nel namespace dell'ambiente selezionato.

Nel progetto, l'applicazione dimostrativa e `demo-go-color-app`.

La verifica deve essere environment-specific:

- dev viene verificato in `devops-ci-demo`;
- staging viene verificato in `devops-ci-staging`;
- production viene verificato in `devops-ci-production`.

Un risultato positivo in un namespace non dimostra lo stato degli altri namespace.

Per esempio, un deployment pronto in `devops-ci-demo` non prova che il deployment sia pronto anche in `devops-ci-production`.

### 19.5 validate

`validate` avvia una validazione Tekton.

La validazione deve usare il namespace Tekton corretto e il validation path corretto per l'ambiente.

Esempi validati:

```text
staging validationPath = apps/demo-go-color-app/overlays/staging
production validationPath = apps/demo-go-color-app/overlays/production
```

La validazione Tekton produce una PipelineRun.

Esempi reali:

```text
staging PipelineRun = devops-cp-validate-chg-2026-0049-nd7rm
production PipelineRun = devops-cp-validate-chg-2026-0050-8wqtv
```

### 19.6 check-validation

`check-validation` legge lo stato della PipelineRun Tekton e produce validation evidence.

Questa azione e importante perche trasforma un risultato tecnico Tekton in una evidence persistita e leggibile dal DevOps Control Plane.

Campi importanti della validation evidence:

- target environment;
- Tekton namespace;
- pipeline name;
- PipelineRun name;
- Git revision o branch;
- validation path;
- status;
- reason;
- failed task count;
- sanitization state.

Un risultato positivo atteso e:

```text
status=True
reason=Succeeded
failedTaskCount=0
evidence sanitized=true
```

### 19.7 Persistenza delle evidence

Le evidence prodotte dal workflow runtime vengono persistite in PostgreSQL.

Questo consente di consultarle successivamente tramite UI o API.

La persistenza e fondamentale perche i risultati runtime e Tekton possono cambiare nel tempo. Salvare una evidence significa conservare una fotografia del risultato osservato durante una certa fase del workflow.

### 19.8 Rendering nella UI

La UI rende visibili le evidence raccolte.

Nel dettaglio ChangeRequest, l'operatore deve poter vedere:

- runtime evidence;
- deployment evidence;
- Tekton validation evidence;
- failed task count;
- validation path;
- PipelineRun;
- Tekton namespace;
- stato sanitized.

La UI aiuta l'operatore a capire se il workflow e davvero riuscito dal punto di vista tecnico.

### 19.9 Relazione con Argo CD

Argo CD fornisce informazioni su sync e health.

Nel workflow runtime, Argo CD aiuta a rispondere a domande come:

- l'applicazione e allineata al repository GitOps;
- l'applicazione e healthy;
- la revisione osservata e quella attesa;
- l'overlay corretto e stato applicato.

Esempi di stato atteso:

```text
sync=Synced
health=Healthy
```

### 19.10 Relazione con Tekton

Tekton fornisce la validazione tecnica.

Nel workflow runtime, Tekton dimostra che il contenuto GitOps associato alla ChangeRequest e stato validato.

La PipelineRun e il risultato tecnico principale.

Le TaskRun aiutano nel troubleshooting quando una validazione fallisce.

### 19.11 Relazione con Kubernetes/OpenShift

Kubernetes/OpenShift rappresenta lo stato runtime reale.

Il workflow runtime osserva:

- deployment;
- pod;
- service;
- route;
- namespace;
- readiness.

Queste informazioni servono a verificare che l'applicazione non sia solo dichiarata in Git, ma anche eseguita correttamente nel cluster.

### 19.12 Fail-closed nel workflow runtime

Il workflow runtime deve rispettare i guardrail fail-closed.

Esempi:

- runtime target non risolto;
- provider mancante;
- provider disabled;
- Secret reference non allow-listed;
- factory disabled;
- API URL mancante;
- token mancante.

In questi casi il workflow deve fermarsi con errore esplicito.

Non deve eseguire fallback silenzioso verso `ocp-dev`.

### 19.13 Relazione con multi-cluster readiness

Il workflow runtime e attualmente validato sulla baseline namespace-isolated.

Tuttavia il codice e stato rafforzato con target simulati:

```text
staging -> ocp-staging-simulated
production -> ocp-production-simulated
```

I test confermano:

- nessun fallback silenzioso verso `ocp-dev`;
- provider mancante fail-closed;
- provider disabled fail-closed.

Questo dimostra che il workflow runtime puo essere esteso a cluster fisici futuri senza riprogettare il modello.

### 19.14 Esempio staging

Esempio riassuntivo staging:

```text
ChangeRequest: CHG-2026-0049
targetEnvironment: staging
kubernetesNamespace: devops-ci-staging
tektonNamespace: devops-ci-staging
validationPath: apps/demo-go-color-app/overlays/staging
PipelineRun: devops-cp-validate-chg-2026-0049-nd7rm
result: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

### 19.15 Esempio production

Esempio riassuntivo production:

```text
ChangeRequest: CHG-2026-0050
targetEnvironment: production
kubernetesNamespace: devops-ci-production
tektonNamespace: devops-ci-production
validationPath: apps/demo-go-color-app/overlays/production
PipelineRun: devops-cp-validate-chg-2026-0050-8wqtv
result: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

### 19.16 Sintesi

Il workflow runtime e la parte del DevOps Control Plane che collega la richiesta di cambiamento allo stato reale dei sistemi tecnici.

Attraverso `collect-evidence`, `check-deployment`, `validate` e `check-validation`, il sistema produce una vista coerente e persistente del risultato tecnico.

Questo workflow rende possibile osservare, validare e spiegare lo stato di dev, staging e production nella baseline namespace-isolated e prepara il progetto al futuro multi-cluster reale.

## 20. Workflow dev, staging e production

Il DevOps Control Plane supporta un modello multi-environment basato su tre ambienti logici:

- `dev`;
- `staging`;
- `production`.

Nel baseline attuale questi ambienti non sono ancora distribuiti su cluster fisici separati. Sono invece validati con isolamento per namespace sul cluster OpenShift disponibile `ocp-dev`.

La topologia validata e:

```text
dev        -> ocp-dev / devops-ci-demo
staging    -> ocp-dev / devops-ci-staging
production -> ocp-dev / devops-ci-production
```

Questa scelta permette di validare workflow, evidenze, Argo CD, Tekton, UI e operability multi-environment anche in assenza di cluster fisici aggiuntivi.

### 20.1 Ambiente dev

L'ambiente `dev` rappresenta l'ambiente di sviluppo attivo.

Nel baseline corrente:

```text
environment: dev
cluster: ocp-dev
namespace: devops-ci-demo
application: demo-go-color-app
```

L'ambiente dev e il primo ambiente validato nel progetto. Viene usato come baseline operativa iniziale per verificare che il DevOps Control Plane sia in grado di:

- creare ChangeRequest;
- interagire con GitLab;
- interrogare Argo CD;
- osservare lo stato runtime Kubernetes/OpenShift;
- raccogliere evidence;
- mostrare dati in UI.

### 20.2 Ambiente staging

L'ambiente `staging` rappresenta un ambiente logico pre-produzione.

Nel baseline corrente:

```text
environment: staging
cluster: ocp-dev
namespace: devops-ci-staging
Argo CD Application: demo-go-color-app-staging
validationPath: apps/demo-go-color-app/overlays/staging
```

Staging e stato validato come namespace separato sul cluster `ocp-dev`.

La validazione importante associata a staging e:

```text
ChangeRequest: CHG-2026-0049
PipelineRun: devops-cp-validate-chg-2026-0049-nd7rm
result: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

Questo dimostra che il workflow staging non e solo documentato, ma e stato eseguito e verificato.

### 20.3 Ambiente production

L'ambiente `production` rappresenta la produzione logica del modello.

Nel baseline corrente production non e un cluster fisico separato. E un namespace isolato sul cluster `ocp-dev`.

Mappatura corrente:

```text
environment: production
cluster: ocp-dev
namespace: devops-ci-production
Argo CD Application: demo-go-color-app-production
validationPath: apps/demo-go-color-app/overlays/production
```

La validazione importante associata a production e:

```text
ChangeRequest: CHG-2026-0050
PipelineRun: devops-cp-validate-chg-2026-0050-8wqtv
result: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

Questa validazione dimostra che il workflow production logico e stato testato senza dichiarare una produzione fisica reale.

### 20.4 Differenza tra ambiente logico e cluster fisico

E importante distinguere due concetti:

- ambiente logico;
- cluster fisico.

Un ambiente logico rappresenta lo scopo operativo di un workflow: sviluppo, staging o production.

Un cluster fisico e invece un cluster OpenShift reale.

Nel progetto attuale:

```text
ambienti logici: dev, staging, production
cluster fisico disponibile: ocp-dev
```

Quindi staging e production sono ambienti logici validati tramite namespace isolation, non cluster fisici separati.

### 20.5 Perche usare namespace isolation

Namespace isolation consente di validare molte caratteristiche senza attendere cluster aggiuntivi.

Il progetto ha potuto validare:

- mapping environment-to-namespace;
- Argo CD Applications separate;
- overlay GitOps environment-specific;
- Tekton namespace separati;
- validation path staging e production;
- route health per ambiente;
- runtime evidence environment-aware;
- UI Environments / Namespaces;
- ChangeRequest detail con evidence specifica.

Questa strategia ha permesso di avanzare senza bloccare il progetto per indisponibilita infrastrutturale.

### 20.6 Workflow comune ai tre ambienti

Il workflow logico e simile per dev, staging e production.

Passaggi principali:

1. creare o selezionare una ChangeRequest;
2. risolvere il target environment;
3. risolvere cluster e namespace;
4. controllare Argo CD;
5. controllare deployment e route;
6. avviare o verificare Tekton validation;
7. raccogliere evidence;
8. mostrare evidence in UI.

La differenza principale e nei metadati environment-specific.

### 20.7 Mapping environment-specific

Tabella concettuale:

```text
Environment   Cluster   Namespace                 Argo CD Application
----------    -------   -----------------------   -------------------------------
dev           ocp-dev   devops-ci-demo            demo-go-color-app
staging       ocp-dev   devops-ci-staging         demo-go-color-app-staging
production    ocp-dev   devops-ci-production      demo-go-color-app-production
```

Validation path:

```text
staging     apps/demo-go-color-app/overlays/staging
production  apps/demo-go-color-app/overlays/production
```

Questi valori devono restare espliciti nelle evidence e nei workflow.

### 20.8 Argo CD per ambiente

Argo CD osserva Applications distinte per ambiente.

Le verifiche finali hanno confermato:

```text
dev        sync=Synced health=Healthy
staging    sync=Synced health=Healthy
production sync=Synced health=Healthy
```

Questo risultato conferma che le Applications GitOps sono coerenti con la baseline namespace-isolated.

### 20.9 Deployment readiness per ambiente

Il deployment `demo-go-color-app` e stato verificato nei tre namespace.

La smoke matrix finale ha confermato readiness nei namespace:

- `devops-ci-demo`;
- `devops-ci-staging`;
- `devops-ci-production`.

Questo e importante perche un deployment healthy in dev non prova automaticamente lo stato di staging o production.

### 20.10 Route health per ambiente

Ogni ambiente espone una route applicativa.

La validazione finale ha verificato `/healthz` per tutti e tre gli ambienti:

```text
dev_healthz_http=200
staging_healthz_http=200
production_healthz_http=200
```

Route e health check sono parte delle evidenze operative di runtime.

### 20.11 Tekton per staging e production

Tekton e stato usato per validare staging e production.

Staging:

```text
PipelineRun: devops-cp-validate-chg-2026-0049-nd7rm
status: True
reason: Succeeded
```

Production:

```text
PipelineRun: devops-cp-validate-chg-2026-0050-8wqtv
status: True
reason: Succeeded
```

Questi risultati confermano che la validazione tecnica e stata completata sugli ambienti logici corretti.

### 20.12 UI per dev, staging e production

La UI e stata aggiornata per mostrare il modello multi-environment.

Elementi importanti:

- dashboard latest ChangeRequest;
- sezione `Environments / Namespaces`;
- ChangeRequest detail;
- runtime evidence card;
- Tekton validation evidence card;
- evidence sanitized.

La UI deve rendere chiaro che staging e production sono ambienti logici separati, anche se condividono il cluster fisico `ocp-dev`.

### 20.13 Simulazione cluster separati

Dopo la chiusura della Fase 15, il codice e stato rafforzato con test di target cluster simulati:

```text
staging -> ocp-staging-simulated
production -> ocp-production-simulated
```

Questi test non rappresentano validazione fisica cross-cluster.

Rappresentano una validazione del modello codice.

I test dimostrano che:

- staging puo risolvere un cluster diverso da `ocp-dev`;
- production puo risolvere un cluster diverso da `ocp-dev`;
- non avviene fallback silenzioso verso `ocp-dev`;
- provider mancante fallisce fail-closed;
- provider disabled fallisce fail-closed.

### 20.14 Cosa resta deferred

Restano deferred per indisponibilita infrastrutturale:

- validazione fisica su cluster staging reale;
- validazione fisica su cluster production reale;
- Secret loading reale cross-cluster;
- RBAC reale cross-cluster;
- rollback reale da onboarding fisico fallito;
- smoke test cross-cluster fisico.

Questi elementi sono rinviati per mancanza di cluster, non per mancanza del modello codice.

### 20.15 Regola operativa

La regola da mantenere e:

```text
Physical cross-cluster runtime validation is deferred by infrastructure availability.
Multi-cluster code readiness is completed, tested, documented and fail-closed.
```

Questo significa che il progetto e pronto a livello modello e codice, ma non dichiara una validazione fisica che non e stata possibile eseguire.

### 20.16 Sintesi

Il workflow dev, staging e production dimostra che il DevOps Control Plane e capace di controllare piu ambienti logici in modo coerente.

La baseline corrente e namespace-isolated su `ocp-dev`, ma include tutti gli elementi necessari per spiegare e validare il futuro percorso multi-cluster:

- environment mapping;
- namespace mapping;
- Argo CD Applications;
- Tekton validation;
- runtime evidence;
- UI environment awareness;
- fail-closed guardrails;
- simulazione staging e production cluster.

Questo capitolo chiude la parte dedicata ai workflow applicativi e prepara la guida ai capitoli sull'evidence model.

## 21. Runtime evidence

La runtime evidence e l'insieme delle prove tecniche raccolte osservando lo stato reale dei sistemi runtime.

Nel DevOps Control Plane, una ChangeRequest non e considerata completa solo perche esiste nel database o perche una pipeline e stata avviata. Il sistema deve anche poter dimostrare cosa e stato osservato nell'ambiente target.

La runtime evidence risponde alla domanda:

```text
Cosa risultava effettivamente in esecuzione nel cluster al momento del controllo?
```

Per questo motivo la runtime evidence e centrale per operability, troubleshooting, audit tecnico e UI evidence rendering.

### 21.1 Perche serve la runtime evidence

In un workflow DevOps distribuito, lo stato puo essere frammentato tra strumenti diversi.

Per esempio:

- GitLab conosce il codice e la merge request;
- Argo CD conosce sync e health dell'applicazione;
- Tekton conosce l'esito della pipeline;
- Kubernetes/OpenShift conosce pod, deployment, service e route;
- il database del DevOps Control Plane conosce ChangeRequest, eventi ed evidence.

La runtime evidence collega la ChangeRequest allo stato osservato nel cluster. Questo permette di dimostrare che una richiesta non e solo stata processata, ma che il runtime e stato controllato e documentato.

### 21.2 Cosa puo contenere la runtime evidence

La runtime evidence puo includere informazioni come:

- target environment;
- cluster name;
- Kubernetes namespace;
- deployment name;
- ready replicas;
- desired replicas;
- available replicas;
- updated replicas;
- pod status;
- service status;
- route host;
- route health;
- Argo CD sync status;
- Argo CD health status;
- timestamp del controllo;
- stato di sanitizzazione.

Queste informazioni aiutano un operatore a capire se l'ambiente target e coerente con lo stato atteso.

### 21.3 Namespace e ambiente target

La runtime evidence deve sempre conservare il contesto dell'ambiente target.

Nel baseline attuale gli ambienti sono namespace-isolated sul cluster `ocp-dev`:

```text
dev        -> ocp-dev / devops-ci-demo
staging    -> ocp-dev / devops-ci-staging
production -> ocp-dev / devops-ci-production
```

Per questo motivo non basta dire che un deployment e pronto. Bisogna dire in quale namespace e stato osservato.

Un deployment pronto in `devops-ci-demo` non prova che staging o production siano pronti.

### 21.4 Deployment evidence

Una parte importante della runtime evidence e la deployment evidence.

La deployment evidence descrive lo stato osservato del Deployment applicativo.

Esempio:

```text
environment = production
namespace = devops-ci-production
deployment = demo-go-color-app
readyReplicas = 2
desiredReplicas = 2
availableReplicas = 2
updatedReplicas = 2
```

Questa evidence permette di distinguere tra deployment esistente ma non pronto, deployment pronto, deployment assente, numero di repliche non coerente o rollout non completato.

### 21.5 Pod evidence

La pod evidence aiuta a capire lo stato dei pod associati a un deployment.

Informazioni utili:

- nome pod;
- fase pod;
- container ready;
- restart count;
- eventi rilevanti.

La pod evidence e utile quando il deployment non e pronto, per esempio per image pull error, crash del container, probe fallita o configurazione errata.

### 21.6 Service e route evidence

La runtime evidence puo includere informazioni su Service e Route.

Nel progetto OpenShift, la Route e importante per verificare la raggiungibilita dell'applicazione.

La smoke matrix finale ha validato `/healthz` per dev, staging e production.

Risultato atteso:

```text
dev_healthz_http=200
staging_healthz_http=200
production_healthz_http=200
```

Questi controlli dimostrano che il workload non e solo presente nel cluster, ma risponde anche attraverso il percorso esposto.

### 21.7 Argo CD runtime evidence

Argo CD fornisce una vista GitOps dello stato dell'applicazione.

Le informazioni piu importanti sono:

- sync status;
- health status;
- revision;
- Application name;
- target namespace;
- GitOps path.

Stato atteso nella baseline validata:

```text
sync=Synced
health=Healthy
```

La Argo CD evidence e utile per confrontare lo stato desiderato GitOps con lo stato osservato nel cluster.

### 21.8 Runtime evidence e Tekton validation evidence

Runtime evidence e Tekton validation evidence sono collegate ma non identiche.

La Tekton validation evidence risponde alla domanda:

```text
La pipeline di validazione tecnica e riuscita?
```

La runtime evidence risponde alla domanda:

```text
Il runtime osservato risulta coerente e sano?
```

Entrambe sono necessarie. Una pipeline Tekton puo riuscire, ma un deployment puo non essere pronto. Oppure un deployment puo essere pronto, ma la validazione GitOps puo fallire.

### 21.9 Persistenza della runtime evidence

La runtime evidence viene persistita in PostgreSQL come evidence associata a una ChangeRequest.

Questo consente di:

- consultare lo storico;
- mostrare la evidence nella UI;
- analizzare incidenti;
- confrontare esecuzioni diverse;
- preservare una fotografia dello stato osservato.

La persistenza e importante perche lo stato runtime puo cambiare dopo il controllo.

### 21.10 Runtime evidence nella UI

La UI mostra la runtime evidence nelle pagine ChangeRequest.

La UI deve aiutare l'operatore a capire:

- quale ambiente era coinvolto;
- quale namespace e stato osservato;
- quale deployment e stato controllato;
- se il deployment era pronto;
- se la route rispondeva;
- se Argo CD era `Synced` e `Healthy`.

La UI puo offrire raw sanitized evidence come dettaglio diagnostico, ma non deve esporre dati sensibili.

### 21.11 Sanitizzazione della runtime evidence

La runtime evidence deve essere sanificata.

Dati ammessi:

- namespace;
- nomi risorse;
- stato deployment;
- stato route;
- status Argo CD;
- reason;
- timestamp;
- validation path;
- PipelineRun name, quando collegato.

Dati vietati:

- token;
- password;
- kubeconfig raw;
- private key;
- Secret decodificati;
- credenziali applicative;
- bearer token.

La regola e:

```text
show operational metadata, never expose credentials
```

### 21.12 Runtime evidence e troubleshooting

La runtime evidence e uno strumento di troubleshooting.

Quando qualcosa non funziona, l'operatore puo usare le evidence per capire:

- se il problema e nel deployment;
- se il problema e nella route;
- se il problema e in Argo CD;
- se il problema e nel namespace sbagliato;
- se il problema e nella validazione Tekton;
- se il problema e nella UI o nella persistenza.

Senza runtime evidence, il troubleshooting richiederebbe di interrogare manualmente molti sistemi diversi.

### 21.13 Runtime evidence e operability

I runbook operativi usano la runtime evidence come base per health check e manutenzione.

Esempi di controlli:

- `/readyz` del DevOps Control Plane;
- dashboard HTTP;
- Argo CD Application matrix;
- deployment readiness matrix;
- route health matrix;
- Tekton validation matrix;
- UI ChangeRequest detail.

Questi controlli producono evidence directory e summary utili per incidenti e manutenzione.

### 21.14 Dev evidence

Per l'ambiente dev, la runtime evidence riguarda:

```text
environment = dev
cluster = ocp-dev
namespace = devops-ci-demo
application = demo-go-color-app
```

Dev rappresenta la baseline iniziale del progetto.

### 21.15 Staging evidence

Per staging, la runtime evidence riguarda:

```text
environment = staging
cluster = ocp-dev
namespace = devops-ci-staging
application = demo-go-color-app
Argo CD Application = demo-go-color-app-staging
```

La staging evidence e collegata anche alla validation evidence della ChangeRequest `CHG-2026-0049`.

### 21.16 Production evidence

Per production, la runtime evidence riguarda:

```text
environment = production
cluster = ocp-dev
namespace = devops-ci-production
application = demo-go-color-app
Argo CD Application = demo-go-color-app-production
```

La production evidence e collegata anche alla validation evidence della ChangeRequest `CHG-2026-0050`.

### 21.17 Relazione con multi-cluster readiness

La runtime evidence deve essere compatibile con il futuro multi-cluster.

Oggi staging e production sono namespace-isolated su `ocp-dev`. Domani potrebbero puntare a cluster fisici diversi.

Per questo la evidence deve preservare sempre:

- target environment;
- cluster name;
- namespace;
- provider selection, quando disponibile;
- Argo CD Application;
- Tekton namespace;
- validation path.

Quando arriveranno cluster reali, la evidence dovra dimostrare che il workflow non e ricaduto per errore su `ocp-dev`.

### 21.18 Cosa la runtime evidence non deve essere

La runtime evidence non deve diventare:

- un dump completo e non sanificato del cluster;
- un contenitore di Secret;
- una copia dei log grezzi senza controllo;
- una fonte di credenziali;
- una sostituzione dei runbook;
- una dichiarazione automatica di successo production.

La runtime evidence deve essere una prova tecnica utile, sicura e contestualizzata.

### 21.19 Sintesi

La runtime evidence e una delle funzioni piu importanti del DevOps Control Plane.

Permette di collegare una ChangeRequest allo stato reale osservato in OpenShift, Argo CD e Tekton.

Grazie alla runtime evidence, il sistema puo spiegare non solo che una richiesta e stata elaborata, ma anche cosa e stato osservato nel runtime e quali prove sono disponibili per verificarlo.

## 22. Tekton validation evidence

La Tekton validation evidence e l'evidenza che descrive il risultato di una validazione tecnica eseguita tramite Tekton.

Nel DevOps Control Plane, Tekton non e usato come elemento isolato. Tekton e parte del workflow della `ChangeRequest`: una richiesta di cambiamento puo generare una validazione, la validazione produce una `PipelineRun`, la `PipelineRun` produce uno stato, e il DevOps Control Plane trasforma questo stato in evidence persistita e visualizzabile nella UI.

La Tekton validation evidence risponde alla domanda:

```text
La validazione tecnica associata alla ChangeRequest e riuscita?
```

Questa evidence e diversa dalla runtime evidence. La runtime evidence osserva il runtime. La Tekton validation evidence osserva il risultato della pipeline tecnica.

### 22.1 Perche serve la Tekton validation evidence

La validazione Tekton consente di verificare il cambiamento prima o durante il percorso di promozione tecnica.

Nel progetto, Tekton viene usato per validare contenuti GitOps e per produrre un risultato tecnico collegato alla ChangeRequest.

Senza Tekton validation evidence, un operatore potrebbe vedere che una PipelineRun e stata creata, ma non avrebbe una rappresentazione persistente e leggibile del suo esito nel control plane.

Con la validation evidence, invece, il DevOps Control Plane puo mostrare:

- quale PipelineRun e stata eseguita;
- in quale namespace Tekton;
- con quale validation path;
- quale stato finale ha avuto;
- quale reason e stata restituita;
- se ci sono task fallite;
- se l'evidence e stata sanificata.

### 22.2 Pipeline, PipelineRun e TaskRun

Tekton organizza il lavoro tecnico attraverso concetti specifici.

Una `Pipeline` descrive una sequenza di passaggi.

Una `PipelineRun` e un'esecuzione concreta di una Pipeline.

Una `TaskRun` e l'esecuzione concreta di un singolo task all'interno di una PipelineRun.

Nel troubleshooting, la `PipelineRun` dice se la validazione complessiva e riuscita. Le `TaskRun` aiutano a capire dove si e verificato un errore.

Vista semplificata:

```text
Tekton Pipeline
      |
      v
PipelineRun
      |
      +--> TaskRun clone-repository
      +--> TaskRun validate-gitops
      +--> TaskRun collect-result
```

### 22.3 Campi principali della validation evidence

Una Tekton validation evidence deve contenere informazioni sufficienti per spiegare il risultato senza obbligare l'operatore a interrogare manualmente Tekton.

Campi importanti:

- target environment;
- Tekton namespace;
- pipeline name;
- PipelineRun name;
- Git revision o branch;
- validation path;
- status;
- reason;
- failed task count;
- failed tasks, se disponibili;
- sanitization state.

Questi campi aiutano a rispondere a domande operative precise.

Esempio:

```text
Quale overlay GitOps e stato validato?
Quale PipelineRun ha prodotto il risultato?
La validazione e terminata con Succeeded?
Ci sono TaskRun fallite?
L'evidence e sicura da mostrare in UI?
```

### 22.4 Validation path

Il validation path e uno dei campi piu importanti.

Indica quale porzione del repository GitOps e stata validata.

Nel baseline corrente i path validati sono environment-specific:

```text
staging     apps/demo-go-color-app/overlays/staging
production  apps/demo-go-color-app/overlays/production
```

Questa distinzione evita errori pericolosi.

Una validazione production non deve validare per errore l'overlay staging.

Una validazione staging non deve validare per errore l'overlay production.

### 22.5 Stato e reason

Tekton espone condizioni che indicano lo stato della PipelineRun.

Per una validazione riuscita, il risultato atteso e:

```text
status=True
reason=Succeeded
```

Se la PipelineRun fallisce, lo stato e la reason aiutano a capire la natura del problema.

La validation evidence deve preservare questi valori per renderli visibili nella UI e disponibili per troubleshooting.

### 22.6 Failed task count

`failedTaskCount` indica quante TaskRun sono fallite.

Nel baseline validato, staging e production hanno avuto:

```text
failedTaskCount=0
```

Questo valore e importante perche una PipelineRun puo avere molte attivita interne. Sapere che il numero di task fallite e zero rende immediata la lettura operativa del risultato.

Quando il numero e maggiore di zero, l'operatore deve passare ad analizzare TaskRun e log sanificati.

### 22.7 Evidence sanitization

La Tekton validation evidence deve essere sanificata.

Dati ammessi:

- PipelineRun name;
- TaskRun name;
- namespace;
- validation path;
- status;
- reason;
- failed task count;
- Git revision o branch;
- timestamp.

Dati vietati:

- token Git;
- password;
- kubeconfig raw;
- private key;
- Secret decodificati;
- credenziali contenute nei log;
- bearer token.

La UI puo mostrare lo stato sanitized, per esempio:

```text
evidence sanitized=true
```

### 22.8 Esempio staging

Esempio validato per staging:

```text
ChangeRequest: CHG-2026-0049
environment: staging
Tekton namespace: devops-ci-staging
PipelineRun: devops-cp-validate-chg-2026-0049-nd7rm
validationPath: apps/demo-go-color-app/overlays/staging
status: True
reason: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

Questa evidence dimostra che la validazione associata a staging e stata completata correttamente.

### 22.9 Esempio production

Esempio validato per production:

```text
ChangeRequest: CHG-2026-0050
environment: production
Tekton namespace: devops-ci-production
PipelineRun: devops-cp-validate-chg-2026-0050-8wqtv
validationPath: apps/demo-go-color-app/overlays/production
status: True
reason: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

Questa evidence dimostra che la validazione associata a production logica e stata completata correttamente nella baseline namespace-isolated.

### 22.10 Relazione con ChangeRequest

La validation evidence deve essere collegata alla ChangeRequest corretta.

Il motivo e semplice: senza questo collegamento, un operatore potrebbe vedere una PipelineRun riuscita ma non sapere quale richiesta di cambiamento l'ha generata.

Il collegamento permette di navigare dalla ChangeRequest alla validation evidence e viceversa.

Vista concettuale:

```text
ChangeRequest CHG-2026-0050
      |
      +--> validate
      |
      +--> PipelineRun devops-cp-validate-chg-2026-0050-8wqtv
      |
      +--> Tekton validation evidence
```

### 22.11 Relazione con runtime evidence

Tekton validation evidence e runtime evidence si completano a vicenda.

Tekton validation evidence dimostra che la validazione tecnica e riuscita.

Runtime evidence dimostra cosa e stato osservato nel cluster.

Un workflow completo deve poter mostrare entrambe.

Esempio:

```text
Tekton validation: Succeeded
Runtime deployment: Ready
Argo CD: Synced / Healthy
Route health: HTTP 200
```

Questa combinazione fornisce una visione operativa molto piu solida di un singolo controllo isolato.

### 22.12 Relazione con UI

La UI espone la Tekton validation evidence nella pagina di dettaglio della ChangeRequest.

La card di validazione dovrebbe rendere visibili:

- PipelineRun;
- Tekton namespace;
- pipeline;
- validation path;
- status;
- reason;
- failed task count;
- sanitized state.

Questo evita che l'operatore debba usare immediatamente `oc` o la console Tekton per capire l'esito della validazione.

La UI non sostituisce Tekton, ma rende il risultato Tekton leggibile nel contesto della ChangeRequest.

### 22.13 Relazione con troubleshooting

Quando una validazione fallisce, la validation evidence e il primo punto di analisi.

L'operatore deve verificare:

- PipelineRun status;
- reason;
- failed task count;
- TaskRun fallite;
- validation path;
- namespace Tekton;
- Git revision o branch;
- eventuale errore di accesso a Git;
- eventuale problema RBAC;
- eventuale problema di workspace o parametri.

La evidence deve guidare il troubleshooting senza esporre credenziali.

### 22.14 Relazione con operability

I runbook operativi includono controlli sulle PipelineRun staging e production.

Esempi:

```text
staging PipelineRun status=True reason=Succeeded
production PipelineRun status=True reason=Succeeded
```

Questi controlli fanno parte della smoke matrix post-Fase 15.

### 22.15 Relazione con multi-cluster readiness

Oggi la validazione Tekton e namespace-isolated su `ocp-dev`.

Domani staging e production potrebbero puntare a cluster fisici diversi.

La Tekton validation evidence deve quindi preservare:

- target environment;
- Tekton namespace;
- cluster name quando disponibile;
- provider selection quando disponibile;
- validation path.

I test simulati post-Fase 15 hanno dimostrato che staging e production possono risolvere target cluster simulati senza fallback a `ocp-dev`.

### 22.16 Errori tipici

Errori possibili nella validazione Tekton:

- Pipeline non trovata;
- Task non trovata;
- PipelineRun fallita;
- Git clone fallito;
- validation path errato;
- RBAC insufficiente;
- ServiceAccount non autorizzato;
- workspace mancante;
- Secret reference non valida;
- timeout della pipeline.

Questi errori devono essere riportati come evidence sanificata e collegati alla ChangeRequest.

### 22.17 Cosa non deve fare la validation evidence

La validation evidence non deve:

- copiare log completi non sanificati;
- esporre token Git;
- esporre Secret;
- nascondere il namespace;
- nascondere il validation path;
- dichiarare production fisica quando si tratta di production logica namespace-isolated.

### 22.18 Sintesi

La Tekton validation evidence rende verificabile il risultato della validazione tecnica.

Essa collega ChangeRequest, PipelineRun, validation path, stato, reason, failed task count e UI.

Insieme alla runtime evidence, permette al DevOps Control Plane di fornire una vista completa e auditabile del cambiamento.

## 23. Argo CD deployment evidence

La Argo CD deployment evidence descrive lo stato GitOps osservato da Argo CD per una applicazione gestita.

Nel DevOps Control Plane, Argo CD e il componente che confronta lo stato desiderato nel repository GitOps con lo stato effettivo applicato nel cluster OpenShift. La deployment evidence proveniente da Argo CD permette di capire se l'applicazione e allineata al repository e se le risorse risultano sane.

La domanda principale a cui risponde questa evidence e:

```text
Lo stato GitOps dell'applicazione e sincronizzato e sano per l'ambiente target?
```

Questa evidence completa la runtime evidence e la Tekton validation evidence.

### 23.1 Perche serve la Argo CD deployment evidence

In un modello GitOps, Git rappresenta lo stato desiderato.

Argo CD osserva il repository Git e applica o confronta quello stato con il cluster.

Senza Argo CD deployment evidence, un operatore potrebbe sapere che una pipeline Tekton e riuscita, ma non sapere se Argo CD vede l'applicazione come `Synced` e `Healthy`.

Questa evidence e quindi fondamentale per rispondere a domande come:

- l'applicazione e allineata con Git?
- Argo CD vede differenze tra Git e cluster?
- l'applicazione e sana?
- quale revisione Git e osservata?
- quale namespace e target dell'Application?
- quale overlay GitOps e collegato all'ambiente?

### 23.2 Application Argo CD

Il concetto centrale di Argo CD e `Application`.

Una Application descrive:

- repository Git;
- path GitOps;
- cluster o destinazione;
- namespace target;
- stato di sync;
- stato di health;
- revisione osservata.

Nel DevOps Control Plane, le Application rappresentano la vista GitOps degli ambienti logici.

Esempi validati:

- `demo-go-color-app`;
- `demo-go-color-app-staging`;
- `demo-go-color-app-production`.

### 23.3 Sync status

Il sync status indica se lo stato del cluster e allineato allo stato desiderato nel repository Git.

Stato atteso nella baseline validata:

```text
sync=Synced
```

Se lo stato e `OutOfSync`, Argo CD vede una differenza tra Git e cluster.

Questa condizione richiede analisi prima di dichiarare completato un workflow.

### 23.4 Health status

Il health status indica se le risorse gestite da Argo CD appaiono sane.

Stato atteso nella baseline validata:

```text
health=Healthy
```

Se lo stato e `Degraded`, `Progressing` o un altro stato non atteso, l'applicazione potrebbe non essere pronta anche se il repository Git contiene la configurazione desiderata.

### 23.5 Revision

La revision indica quale revisione Git Argo CD sta osservando.

Questa informazione e importante per collegare:

- repository Git;
- commit o branch;
- Application Argo CD;
- ChangeRequest;
- evidence.

In un workflow auditabile, e utile sapere quale revisione era associata allo stato osservato.

### 23.6 Target namespace

La Argo CD deployment evidence deve preservare il namespace target.

Nel baseline corrente:

```text
dev        -> devops-ci-demo
staging    -> devops-ci-staging
production -> devops-ci-production
```

Il namespace target e necessario per evitare ambiguita.

Una Application `Healthy` per staging non dimostra automaticamente che production sia corretta.

### 23.7 Application dev

L'Application dev e associata all'ambiente di sviluppo.

Mappatura concettuale:

```text
environment = dev
Argo CD Application = demo-go-color-app
namespace = devops-ci-demo
```

Dev rappresenta la baseline iniziale da cui il progetto e stato esteso verso staging e production.

### 23.8 Application staging

L'Application staging e associata all'ambiente logico staging.

Mappatura concettuale:

```text
environment = staging
Argo CD Application = demo-go-color-app-staging
namespace = devops-ci-staging
validationPath = apps/demo-go-color-app/overlays/staging
```

Lo stato atteso e:

```text
sync=Synced
health=Healthy
```

Questa Application dimostra che staging non e solo un'etichetta logica, ma ha un proprio namespace, un proprio overlay e una propria vista GitOps.

### 23.9 Application production

L'Application production e associata all'ambiente logico production.

Mappatura concettuale:

```text
environment = production
Argo CD Application = demo-go-color-app-production
namespace = devops-ci-production
validationPath = apps/demo-go-color-app/overlays/production
```

Lo stato atteso e:

```text
sync=Synced
health=Healthy
```

Questa Application rappresenta la produzione logica nella baseline namespace-isolated.

Non deve essere descritta come produzione fisica separata, perche il cluster fisico resta `ocp-dev`.

### 23.10 Argo CD evidence e runtime evidence

Argo CD deployment evidence e runtime evidence sono collegate.

Argo CD deployment evidence dice se l'applicazione e allineata e sana dal punto di vista GitOps.

Runtime evidence dice cosa e stato osservato nel cluster.

Esempio di vista completa:

```text
Argo CD sync: Synced
Argo CD health: Healthy
Deployment ready: true
Route health: HTTP 200
```

La combinazione dei due punti di vista rende il controllo piu affidabile.

### 23.11 Argo CD evidence e Tekton validation evidence

Argo CD e Tekton svolgono ruoli diversi.

Tekton valida tecnicamente il contenuto o il path GitOps.

Argo CD osserva e riconcilia lo stato GitOps sul cluster.

Insieme, le evidenze rispondono a due domande complementari:

```text
Tekton: il contenuto e stato validato?
Argo CD: lo stato desiderato e sincronizzato e sano?
```

Un workflow completo deve considerare entrambe.

### 23.12 Argo CD evidence nella UI

La UI deve rendere disponibili le informazioni Argo CD in modo utile per l'operatore.

Informazioni da mostrare o rendere disponibili:

- Application name;
- environment;
- namespace;
- sync status;
- health status;
- revision;
- eventuale stato non atteso.

La UI non deve nascondere staging e production dietro una rappresentazione dev-only.

### 23.13 Argo CD evidence e troubleshooting

Quando Argo CD non e `Synced` o non e `Healthy`, l'operatore deve analizzare:

- Application status;
- events;
- Git repository revision;
- path GitOps;
- target namespace;
- risorse Kubernetes associate;
- eventuali errori di sync;
- drift manuale nel cluster.

Esempi di problemi:

- overlay errato;
- namespace mancante;
- manifest non valido;
- risorsa non applicabile;
- permessi insufficienti;
- drift rispetto a Git;
- risorsa degradata.

Questi casi devono essere registrati come evidence sanificata quando collegati a una ChangeRequest.

### 23.14 OutOfSync

`OutOfSync` indica che il cluster non corrisponde allo stato desiderato in Git.

Possibili cause:

- modifica non ancora sincronizzata;
- errore di sync;
- modifica manuale nel cluster;
- repository aggiornato ma non ancora riconciliato;
- differenza tra overlay atteso e overlay configurato.

Un workflow non dovrebbe essere considerato completamente sano se l'Application target e `OutOfSync`, salvo eccezioni operative esplicitamente documentate.

### 23.15 Degraded

`Degraded` indica che Argo CD considera non sane una o piu risorse gestite.

Possibili cause:

- pod non ready;
- deployment non disponibile;
- errore applicativo;
- probe fallita;
- risorsa Kubernetes non valida;
- dipendenza non disponibile.

In caso di `Degraded`, bisogna correlare Argo CD evidence con runtime evidence Kubernetes/OpenShift.

### 23.16 Evidence sanitization

La Argo CD deployment evidence deve essere sanificata.

Dati ammessi:

- Application name;
- namespace;
- sync status;
- health status;
- revision;
- GitOps path;
- resource status;
- reason o message non sensibili.

Dati vietati:

- token Argo CD;
- credenziali Git;
- Secret;
- bearer token;
- dettagli sensibili non necessari;
- payload non sanificati.

### 23.17 Relazione con multi-cluster readiness

In futuro, Argo CD dovra essere valutato anche rispetto a cluster fisici separati.

Oggi la baseline e namespace-isolated su `ocp-dev`.

Domani staging o production potrebbero avere cluster fisici dedicati.

La Argo CD evidence dovra allora preservare anche il cluster target effettivo, per dimostrare che non c'e stato fallback verso `ocp-dev`.

### 23.18 Stato corrente validato

La smoke matrix finale ha confermato Applications `Synced` e `Healthy` per:

```text
dev
staging
production
```

Questa validazione sostiene la baseline namespace-isolated e fornisce un riferimento operativo per health check e manutenzione.

### 23.19 Sintesi

La Argo CD deployment evidence dimostra lo stato GitOps dell'applicazione.

Essa collega repository Git, Application Argo CD, namespace target, stato di sync, stato di health e ChangeRequest.

Insieme a runtime evidence e Tekton validation evidence, permette al DevOps Control Plane di fornire una vista completa del cambiamento: codice validato, GitOps sincronizzato e runtime osservato.

## 24. Evidence sanitization

La evidence sanitization e il processo con cui il DevOps Control Plane conserva e mostra solo informazioni tecniche sicure, evitando di esporre credenziali, token, Secret o altri dati sensibili.

Nel progetto, le evidenze sono fondamentali per audit, troubleshooting, validazione tecnica e UI. Tuttavia, una evidence utile non deve diventare un canale di esposizione di informazioni riservate.

La regola principale e:

```text
collect useful operational evidence, never expose raw credentials
```

La sanitizzazione permette quindi di mantenere il valore operativo delle evidenze senza compromettere la sicurezza.

### 24.1 Perche la sanitizzazione e necessaria

Il DevOps Control Plane integra sistemi che gestiscono informazioni potenzialmente sensibili:

- GitLab;
- Argo CD;
- Tekton;
- Kubernetes/OpenShift;
- Secret;
- token di accesso;
- configurazioni runtime;
- log tecnici.

Durante una validazione o un controllo runtime, il sistema puo attraversare dati provenienti da questi strumenti.

Senza una regola esplicita di sanitizzazione, il rischio e salvare o mostrare informazioni che non dovrebbero mai uscire dal sistema di origine.

### 24.2 Cosa puo essere salvato

Una evidence puo contenere metadati operativi sicuri.

Esempi di dati ammessi:

- numero ChangeRequest;
- target environment;
- cluster name;
- namespace;
- nome Deployment;
- nome Pod;
- nome Service;
- nome Route;
- Argo CD Application name;
- sync status;
- health status;
- Git revision o branch;
- validation path;
- PipelineRun name;
- TaskRun name;
- status;
- reason;
- failed task count;
- timestamp;
- stato `evidence sanitized=true`.

Questi dati aiutano operatori e sviluppatori a capire cosa e stato osservato, senza rivelare materiale sensibile.

### 24.3 Cosa non deve mai essere salvato

Una evidence non deve contenere credenziali o contenuti raw sensibili.

Dati vietati:

- password;
- bearer token;
- token GitLab;
- token Argo CD;
- token Kubernetes;
- kubeconfig raw;
- private key;
- contenuto Secret decodificato;
- certificate private material;
- variabili di ambiente contenenti segreti;
- log completi non revisionati;
- payload che possono includere credenziali.

Se un dato non e necessario per spiegare il risultato operativo, non deve essere salvato nella evidence.

### 24.4 Secret reference invece di Secret value

Il progetto usa un modello basato su Secret reference.

Una Secret reference descrive dove si trova un Secret e quali chiavi sono attese, ma non contiene il valore del Secret.

Esempio di informazione accettabile:

```text
secretRefName = dcp-cluster-ocp-dev-token
secretNamespace = devops-control-plane
key = token
```

Esempio di informazione vietata:

```text
token = actual-token-value
```

La documentazione, la UI, i log e le evidence possono mostrare riferimenti, ma non valori raw.

### 24.5 Sanitizzazione e UI

La UI deve presentare solo evidence sicure.

La UI puo mostrare:

- stato di validazione;
- PipelineRun;
- namespace;
- validation path;
- failed task count;
- Argo CD Application;
- sync e health;
- deployment readiness;
- route health;
- stato sanitized.

La UI non deve mostrare:

- token;
- Secret raw;
- kubeconfig;
- password;
- private key;
- contenuto decodificato di Secret.

La UI deve essere una superficie operativa, non un contenitore di credenziali.

### 24.6 Sanitizzazione e raw evidence

In alcuni casi puo essere utile conservare una forma di raw evidence diagnostica.

Questa raw evidence deve comunque essere sanificata.

Il termine raw, in questo contesto, non significa non filtrato. Significa piu dettagliato rispetto alla card sintetica mostrata nella UI.

Una raw sanitized evidence puo contenere:

- campi tecnici normalizzati;
- status dettagliati;
- reason;
- nomi risorse;
- messaggi di errore non sensibili.

Non deve contenere valori riservati.

### 24.7 Sanitizzazione e Tekton

Tekton puo produrre log e risultati tecnici molto dettagliati.

La validation evidence deve estrarre solo le informazioni utili:

- PipelineRun;
- TaskRun fallite;
- status;
- reason;
- failed task count;
- validation path;
- namespace;
- Git revision o branch.

I log Tekton completi devono essere trattati con cautela, perche potrebbero contenere dati non adatti alla persistenza o alla UI.

### 24.8 Sanitizzazione e Argo CD

Argo CD puo esporre informazioni su Application, repository, revision e stato delle risorse.

La evidence puo conservare:

- Application name;
- sync status;
- health status;
- revision;
- namespace target;
- GitOps path;
- messaggi non sensibili.

La evidence non deve conservare token Argo CD, credenziali Git o altri dati riservati.

### 24.9 Sanitizzazione e Kubernetes/OpenShift

Kubernetes/OpenShift espone molte informazioni operative.

La evidence puo conservare:

- namespace;
- deployment;
- pod;
- service;
- route;
- readiness;
- replica count;
- eventi non sensibili.

La evidence non deve conservare contenuto Secret, token di ServiceAccount, kubeconfig o altri dati sensibili.

### 24.10 Sanitizzazione e troubleshooting

Durante troubleshooting, la tentazione puo essere copiare output completi per velocizzare l'analisi.

Questa pratica e rischiosa.

Gli operatori devono preservare solo evidenze utili e sicure.

Regole operative:

- non decodificare Secret in terminale se l'output viene salvato;
- non copiare token in ticket;
- non allegare kubeconfig;
- non salvare log completi senza revisione;
- preferire summary sanificati;
- indicare i nomi delle risorse invece dei valori segreti.

### 24.11 Evidence sanitized flag

Quando una evidence e stata filtrata correttamente, il sistema puo indicare uno stato come:

```text
evidence sanitized=true
```

Questo campo aiuta l'operatore a capire che l'evidence e stata preparata per essere mostrata o conservata.

Tuttavia, il flag non deve sostituire la responsabilita tecnica. Se un output contiene dati sospetti, deve essere rivisto anche se dichiarato sanificato.

### 24.12 Relazione con audit e compliance

La sanitizzazione e importante anche per audit e compliance.

Un audit trail utile deve spiegare cosa e successo, ma non deve esporre credenziali.

Una buona evidence auditabile contiene:

- chiara associazione alla ChangeRequest;
- ambiente target;
- namespace;
- strumento coinvolto;
- risultato;
- timestamp;
- stato di sanitizzazione.

Non deve contenere materiale che aumenti il rischio di sicurezza.

### 24.13 Relazione con multi-cluster readiness

La readiness multi-cluster richiede ancora piu attenzione alla sanitizzazione.

Quando verranno aggiunti cluster reali, il sistema potra gestire Secret reference, token e endpoint separati per cluster diversi.

Le evidence dovranno dimostrare:

- ambiente target;
- cluster target;
- namespace target;
- provider selection;
- risultato tecnico;
- assenza di fallback non voluto.

Ma non dovranno mai mostrare credenziali dei cluster.

### 24.14 Errori da evitare

Errori comuni da evitare:

- copiare output di `oc describe secret` con dati sensibili;
- decodificare Secret e salvare il risultato;
- inserire token in documentazione;
- allegare kubeconfig a evidenze;
- salvare log completi senza revisione;
- mostrare in UI payload non filtrati;
- confondere Secret reference con Secret value;
- bypassare allow-list per velocizzare un test.

### 24.15 Esempio di evidence corretta

Esempio sicuro:

```text
ChangeRequest: CHG-2026-0050
environment: production
namespace: devops-ci-production
PipelineRun: devops-cp-validate-chg-2026-0050-8wqtv
validationPath: apps/demo-go-color-app/overlays/production
status: True
reason: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

Questa evidence e utile e non contiene credenziali.

### 24.16 Esempio di evidence non corretta

Esempio non accettabile:

```text
bearerToken: actual-token-value
kubeconfig: raw-kubeconfig-content
password: actual-password
```

Questi valori non devono essere salvati, mostrati o committati.

### 24.17 Sintesi

La evidence sanitization e un guardrail fondamentale del DevOps Control Plane.

Permette di conservare prove tecniche utili senza trasformare il sistema in un archivio di segreti.

La regola finale e semplice:

```text
le evidenze devono spiegare cosa e successo, non rivelare credenziali
```

## 25. Dashboard

La dashboard del DevOps Control Plane e la superficie operativa principale per avere una vista sintetica dello stato della piattaforma.

Nelle prime fasi del progetto la UI era un MVP pensato soprattutto per visualizzare dati di base. Dopo le evoluzioni successive, la dashboard e diventata una vista operativa environment-aware ed evidence-aware.

La dashboard oggi non deve essere vista come una semplice pagina grafica. Deve essere considerata un punto di ingresso per comprendere:

- quale ChangeRequest e piu recente;
- quali ambienti logici sono configurati;
- quali namespace sono associati agli ambienti;
- quali evidenze runtime sono disponibili;
- quali evidenze Tekton sono disponibili;
- se il modello dev, staging e production e visibile agli operatori.

### 25.1 Scopo della dashboard

Lo scopo della dashboard e fornire una sintesi rapida dello stato del DevOps Control Plane.

Un operatore deve poter aprire la dashboard e capire rapidamente:

- se il backend risponde;
- se la UI e aggiornata;
- qual e la ChangeRequest piu recente;
- quali ambienti sono rappresentati;
- quali evidenze sono disponibili;
- se la piattaforma sta lavorando sul namespace corretto.

La dashboard non sostituisce i runbook, ma aiuta a orientare l'analisi.

### 25.2 Latest ChangeRequest

La dashboard seleziona la ChangeRequest piu recente disponibile nel backend.

Questo comportamento e importante perche evita una dashboard statica o legata a una richiesta storica hardcoded.

Il comportamento atteso e:

- mostrare la ChangeRequest piu recente;
- mantenere le ChangeRequest precedenti accessibili nella lista;
- permettere all'operatore di aprire il dettaglio della richiesta;
- mostrare evidenze collegate quando disponibili.

In questo modo la dashboard riflette l'attivita corrente del sistema.

### 25.3 Recent changes

La dashboard puo mostrare un elenco compatto di ChangeRequest recenti.

Questa lista aiuta a vedere rapidamente il contesto operativo recente senza trasformare la dashboard in una cronologia completa.

La cronologia dettagliata resta disponibile nelle viste dedicate e tramite il modello persistente su PostgreSQL.

### 25.4 KPI e contatori

La dashboard puo includere contatori o indicatori sintetici.

Esempi:

- numero di ChangeRequest;
- richieste recenti;
- stati principali;
- evidenze disponibili;
- stato operativo sintetico.

Questi indicatori devono essere interpretati come supporto alla navigazione, non come sostituti delle evidenze tecniche.

### 25.5 Environments / Namespaces

Una funzionalita importante della dashboard post-Fase 15 e la sezione `Environments / Namespaces`.

Questa sezione rende visibile la mappatura corrente:

```text
dev        -> devops-ci-demo
staging    -> devops-ci-staging
production -> devops-ci-production
```

La visibilita di questa mappatura e fondamentale perche il progetto usa namespace isolation sul cluster `ocp-dev`.

L'operatore deve poter vedere che staging e production sono ambienti logici distinti, anche se condividono lo stesso cluster fisico.

### 25.6 User box

La dashboard include una rappresentazione compatta del contesto utente.

La user box aiuta a mostrare chi sta consultando la UI o quale identita e stata propagata tramite il layer di autenticazione.

Questa informazione non deve oscurare la sezione degli ambienti. La priorita operativa resta rendere visibili environment, namespace ed evidenze.

### 25.7 Runtime evidence summary

La dashboard puo mostrare una sintesi delle runtime evidence disponibili.

La runtime evidence aiuta a capire se lo stato osservato nel cluster e coerente con il risultato atteso.

Esempi di informazioni utili:

- ambiente target;
- namespace;
- deployment readiness;
- Argo CD sync e health;
- route health;
- timestamp di raccolta.

La dashboard deve sintetizzare queste informazioni, mentre la pagina di dettaglio della ChangeRequest puo mostrare informazioni piu ricche.

### 25.8 Tekton validation summary

La dashboard puo anche mostrare o collegare informazioni sulla validazione Tekton.

Una sintesi utile include:

- PipelineRun;
- status;
- reason;
- failed task count;
- validation path;
- stato sanitized.

Per analisi piu dettagliate, l'operatore deve aprire la pagina di dettaglio della ChangeRequest.

### 25.9 Dashboard e ChangeRequest detail

La dashboard e un punto di ingresso.

Il dettaglio della ChangeRequest e il punto in cui l'operatore trova le informazioni complete.

Flusso previsto:

```text
Dashboard
      |
      v
Latest ChangeRequest
      |
      v
ChangeRequest detail
      |
      +--> audit log
      +--> runtime evidence
      +--> Tekton validation evidence
      +--> raw sanitized evidence
```

Questo flusso rende la UI utile sia per una vista rapida sia per l'analisi tecnica.

### 25.10 Dashboard e namespace isolation

La dashboard deve rappresentare correttamente la baseline namespace-isolated.

Non deve suggerire che staging e production siano gia cluster fisici separati.

La rappresentazione corretta e:

```text
cluster fisico disponibile: ocp-dev
ambienti logici: dev, staging, production
isolamento corrente: namespace separation
```

Questo evita claim errati e mantiene la documentazione coerente con la realta runtime.

### 25.11 Dashboard e multi-cluster readiness

La dashboard si appoggia a un backend multi-cluster code-ready.

Il backend sa modellare ambienti, cluster, runtime target e provider.

Tuttavia, la dashboard deve distinguere tra:

- baseline fisica validata;
- readiness multi-cluster a livello codice;
- simulazioni staging e production;
- validazione fisica futura.

La UI non deve dichiarare validazione fisica cross-cluster finche non saranno disponibili cluster reali.

### 25.12 Dashboard e sicurezza

La dashboard non deve mostrare materiale sensibile.

Dati ammessi:

- numero ChangeRequest;
- ambiente;
- namespace;
- stato;
- reason;
- nomi PipelineRun;
- nomi Argo CD Application;
- validation path;
- evidence sanitized state.

Dati vietati:

- Secret raw;
- token;
- password;
- kubeconfig;
- private key;
- bearer token;
- contenuto Secret decodificato.

La dashboard deve essere sicura da consultare e da usare durante troubleshooting e review operative.

### 25.13 Dashboard e operability

La dashboard e usata anche nelle procedure operative.

Nei runbook post-Fase 15, un controllo dashboard positivo include:

```text
dashboard_http=200
```

Un operatore deve verificare anche che la dashboard mostri elementi coerenti con la baseline attuale:

- latest ChangeRequest;
- `Environments / Namespaces`;
- user box;
- evidenze se disponibili;
- nessun dato sensibile.

### 25.14 Errori tipici della dashboard

Possibili problemi:

- dashboard HTTP non 200;
- latest ChangeRequest non aggiornata;
- sezione `Environments / Namespaces` assente;
- evidence non visualizzate;
- user box non coerente;
- pod con immagine non aggiornata;
- browser cache;
- problema OAuth proxy o forwarded headers.

In questi casi bisogna distinguere se il problema e nella UI, nel backend, nella persistenza o nel runtime.

### 25.15 Sintesi

La dashboard e oggi una superficie operativa del DevOps Control Plane.

Essa mostra lo stato recente, la visibilita degli ambienti, il contesto utente e l'accesso alle evidenze.

La dashboard non e piu solo una UI MVP iniziale. E una vista evidence-aware ed environment-aware, coerente con la baseline namespace-isolated e con la readiness multi-cluster a livello codice.

## 26. ChangeRequest detail

La pagina di dettaglio della `ChangeRequest` e la vista piu importante per analizzare una richiesta specifica.

La dashboard offre una vista sintetica, mentre il dettaglio ChangeRequest permette di entrare nel contesto operativo completo: dati della richiesta, stato del processo, stato runtime, audit log, evidence, validazioni Tekton, stato Argo CD e diagnostica sanificata.

In pratica, questa vista risponde alla domanda:

```text
Cosa e successo per questa ChangeRequest e quali prove tecniche lo dimostrano?
```

### 26.1 Scopo della pagina di dettaglio

La pagina di dettaglio deve aiutare operatori, sviluppatori e reviewer a capire lo stato reale di una singola richiesta.

La pagina deve mostrare:

- dati principali della ChangeRequest;
- target environment;
- stato del processo;
- stato runtime;
- audit log;
- runtime evidence;
- Tekton validation evidence;
- Argo CD deployment evidence;
- raw sanitized evidence quando utile;
- eventuali azioni tecniche disponibili.

La pagina di dettaglio non deve essere solo una vista anagrafica. Deve essere una vista operativa.

### 26.2 Dati principali della ChangeRequest

La sezione iniziale della pagina deve mostrare i dati principali della richiesta.

Esempi di informazioni utili:

- numero ChangeRequest;
- titolo;
- descrizione;
- applicazione target;
- ambiente target;
- requester;
- branch Git o riferimento Git;
- timestamp;
- stato corrente.

Esempi reali usati nella guida:

```text
CHG-2026-0049
CHG-2026-0050
```

Questi dati permettono all'operatore di capire immediatamente quale richiesta sta analizzando.

### 26.3 Target environment nel dettaglio

Il dettaglio deve mostrare chiaramente l'ambiente target.

Nel baseline attuale:

```text
dev        -> ocp-dev / devops-ci-demo
staging    -> ocp-dev / devops-ci-staging
production -> ocp-dev / devops-ci-production
```

Per una ChangeRequest staging, la UI deve rendere chiaro che il namespace e `devops-ci-staging`.

Per una ChangeRequest production, la UI deve rendere chiaro che il namespace e `devops-ci-production`.

Questa chiarezza evita ambiguita tra ambienti logici e cluster fisico.

### 26.4 Lifecycle status

Il lifecycle status descrive l'avanzamento logico della richiesta.

Esempi:

- richiesta creata;
- workflow GitLab eseguito;
- evidence raccolta;
- deployment controllato;
- validazione avviata;
- validazione completata;
- richiesta completata;
- richiesta fallita.

Il lifecycle status aiuta a capire in quale fase del processo si trova la richiesta.

### 26.5 Runtime status

Il runtime status descrive lo stato tecnico osservato.

Esempi:

- deployment ready;
- deployment not ready;
- Argo CD `Synced`;
- Argo CD `Healthy`;
- Tekton `Succeeded`;
- route health HTTP `200`;
- evidence non disponibile;
- validazione fallita.

Il runtime status deve essere interpretato insieme alle evidence.

### 26.6 Audit log

L'audit log mostra gli eventi associati alla ChangeRequest.

Esempi di eventi:

- ChangeRequest creata;
- workflow avviato;
- branch creato;
- Merge Request creata;
- runtime evidence raccolta;
- Tekton validation avviata;
- check-validation completato;
- errore registrato.

L'audit log permette di ricostruire la sequenza delle azioni.

### 26.7 Runtime evidence card

La runtime evidence card mostra una sintesi dello stato osservato nel runtime.

Informazioni utili:

- ambiente target;
- cluster;
- namespace;
- deployment;
- readiness;
- route health;
- Argo CD sync e health;
- timestamp.

Questa card evita che l'operatore debba cercare manualmente tutte le informazioni nei sistemi esterni.

### 26.8 Tekton validation evidence card

La Tekton validation evidence card mostra l'esito della validazione tecnica.

Informazioni utili:

- PipelineRun;
- Tekton namespace;
- pipeline;
- validation path;
- status;
- reason;
- failed task count;
- stato sanitized.

Esempio staging:

```text
ChangeRequest: CHG-2026-0049
PipelineRun: devops-cp-validate-chg-2026-0049-nd7rm
validationPath: apps/demo-go-color-app/overlays/staging
result: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

Esempio production:

```text
ChangeRequest: CHG-2026-0050
PipelineRun: devops-cp-validate-chg-2026-0050-8wqtv
validationPath: apps/demo-go-color-app/overlays/production
result: Succeeded
failedTaskCount: 0
evidence sanitized: true
```

### 26.9 Argo CD deployment evidence

La pagina di dettaglio puo mostrare anche evidenze Argo CD.

Informazioni utili:

- Application name;
- sync status;
- health status;
- revision;
- target namespace;
- GitOps path.

Stato atteso nella baseline validata:

```text
sync=Synced
health=Healthy
```

Queste informazioni aiutano a collegare la ChangeRequest allo stato GitOps osservato.

### 26.10 Raw sanitized evidence

La pagina di dettaglio puo offrire una vista piu tecnica della raw sanitized evidence.

Questa vista deve essere utile per troubleshooting, ma non deve contenere dati sensibili.

Puo includere:

- payload tecnici normalizzati;
- status;
- reason;
- nomi risorse;
- validation path;
- timestamp;
- informazioni diagnostiche non sensibili.

Non deve includere:

- token;
- password;
- kubeconfig;
- private key;
- Secret raw;
- bearer token.

### 26.11 Azioni tecniche

La pagina di dettaglio puo esporre azioni tecniche, in base allo stato della richiesta e ai permessi dell'utente.

Esempi di azioni:

- collect evidence;
- check deployment;
- validate;
- check validation;
- aprire dettagli o link correlati.

Le azioni devono rispettare i guardrail:

- target environment valido;
- provider disponibile;
- provider enabled;
- Secret reference valida;
- factory configurata quando richiesta;
- nessun fallback silenzioso verso namespace o cluster sbagliati.

### 26.12 Relazione con RBAC e ruoli

Non tutti gli utenti devono poter eseguire tutte le azioni.

La UI deve riflettere il modello di autorizzazione.

Un utente senza permessi adeguati non deve vedere o poter attivare azioni tecniche sensibili.

La visibilita delle azioni deve essere coerente con il backend, che resta la fonte autorevole per enforcement e fail-closed.

### 26.13 Relazione con troubleshooting

La pagina ChangeRequest detail e uno dei primi punti da consultare durante troubleshooting.

L'operatore puo verificare:

- target environment;
- namespace;
- audit log;
- evidence disponibili;
- validazione Tekton;
- stato Argo CD;
- stato runtime;
- eventuali failure.

Se la UI non mostra evidence attese, bisogna verificare:

- che la evidence sia stata raccolta;
- che sia stata persistita;
- che sia collegata alla ChangeRequest corretta;
- che la UI stia eseguendo una versione aggiornata;
- che non ci siano errori backend.

### 26.14 Relazione con operability

I runbook operativi includono controlli sulle pagine di dettaglio ChangeRequest.

Esempio di risultato atteso:

```text
chg0049_ui_http=200
chg0050_ui_http=200
```

Questi controlli dimostrano che la UI e in grado di rendere le informazioni principali della richiesta e delle evidence.

### 26.15 Sicurezza della pagina di dettaglio

La pagina di dettaglio non deve mai esporre materiale sensibile.

Dati vietati:

- Secret raw;
- token;
- password;
- kubeconfig;
- private key;
- bearer token;
- contenuto Secret decodificato.

La pagina puo mostrare metadati tecnici sicuri:

- namespace;
- application name;
- PipelineRun;
- Argo CD Application;
- validation path;
- status;
- reason;
- failed task count;
- sanitized state.

### 26.16 Relazione con multi-cluster readiness

La pagina ChangeRequest detail deve essere pronta a rappresentare target futuri multi-cluster.

Oggi staging e production sono namespace-isolated su `ocp-dev`.

Domani potranno essere cluster fisici distinti.

La UI dovra quindi continuare a mostrare:

- target environment;
- cluster name;
- namespace;
- provider selection, quando disponibile;
- evidence associate;
- eventuali errori fail-closed.

La UI non deve nascondere un fallback non voluto.

### 26.17 Sintesi

La pagina ChangeRequest detail e il punto in cui il DevOps Control Plane rende comprensibile una singola richiesta.

Essa collega dati di dominio, audit trail, runtime evidence, Tekton validation evidence, Argo CD evidence e azioni tecniche.

Questa vista e essenziale per trasformare il control plane in uno strumento operativo reale, non solo in un archivio di richieste.

## 27. UI environment awareness

La UI environment awareness e la capacita della UI del DevOps Control Plane di rappresentare chiaramente gli ambienti logici, i namespace e il contesto runtime associato a una ChangeRequest.

Questa funzionalita e importante perche il progetto non lavora piu con un'unica vista dev-only. La piattaforma gestisce oggi una baseline multi-environment namespace-isolated:

```text
dev        -> ocp-dev / devops-ci-demo
staging    -> ocp-dev / devops-ci-staging
production -> ocp-dev / devops-ci-production
```

La UI deve quindi aiutare l'operatore a capire in quale ambiente logico si trova, quale namespace e coinvolto e quali evidenze appartengono a quell'ambiente.

### 27.1 Perche serve environment awareness nella UI

Senza environment awareness, la UI rischia di nascondere informazioni operative fondamentali.

Per esempio, se la UI mostrasse solo un'etichetta `dev`, l'operatore potrebbe non capire se una ChangeRequest riguarda staging o production.

Questo sarebbe pericoloso perche:

- staging e production hanno namespace diversi;
- staging e production hanno Argo CD Applications diverse;
- staging e production hanno validation path diversi;
- staging e production hanno PipelineRun diverse;
- le evidenze devono essere lette nel contesto dell'ambiente corretto.

La UI deve quindi rendere esplicito il modello multi-environment.

### 27.2 Ambiente logico, namespace e cluster fisico

La UI deve aiutare a distinguere tre concetti.

Ambiente logico:

```text
dev
staging
production
```

Namespace:

```text
devops-ci-demo
devops-ci-staging
devops-ci-production
```

Cluster fisico:

```text
ocp-dev
```

Nel baseline corrente tutti gli ambienti condividono il cluster fisico `ocp-dev`, ma restano separati come namespace e come contesto runtime.

### 27.3 Environments / Namespaces

La sezione `Environments / Namespaces` rende visibile la mappatura corrente degli ambienti.

La UI deve mostrare una mappatura simile a:

```text
dev        -> devops-ci-demo
staging    -> devops-ci-staging
production -> devops-ci-production
```

Questa sezione e utile per operatori, reviewer e sviluppatori perche chiarisce subito come il progetto sta rappresentando gli ambienti.

### 27.4 Evitare rappresentazioni dev-only

Una UI dev-only non e piu sufficiente.

Nelle fasi iniziali del progetto poteva essere accettabile mostrare solo `dev` come placeholder. Dopo l'introduzione di Environment Catalog, runtime target resolution, staging e production namespace-isolated, la UI deve rappresentare tutti gli ambienti rilevanti.

La UI deve evitare formulazioni che facciano pensare che:

- esista solo dev;
- staging non sia implementato;
- production non sia rappresentata;
- tutte le azioni vengano eseguite implicitamente su dev.

### 27.5 Relazione con Environment Catalog

La UI non dovrebbe inventare il modello degli ambienti.

La fonte concettuale del modello e l'Environment Catalog.

L'Environment Catalog contiene informazioni come:

- nome ambiente;
- namespace Kubernetes;
- namespace Tekton;
- Argo CD Application;
- validation path;
- stato enabled;
- technical actions abilitate.

La UI deve essere coerente con questo modello.

### 27.6 Relazione con runtime target resolution

Quando una ChangeRequest ha un `targetEnvironment`, il backend risolve un `TechnicalRuntimeTarget`.

La UI deve mostrare informazioni coerenti con quel target.

Esempio staging:

```text
targetEnvironment = staging
clusterName = ocp-dev
kubernetesNamespace = devops-ci-staging
tektonNamespace = devops-ci-staging
argocdApplicationName = demo-go-color-app-staging
validationPath = apps/demo-go-color-app/overlays/staging
```

Esempio production:

```text
targetEnvironment = production
clusterName = ocp-dev
kubernetesNamespace = devops-ci-production
tektonNamespace = devops-ci-production
argocdApplicationName = demo-go-color-app-production
validationPath = apps/demo-go-color-app/overlays/production
```

Queste informazioni aiutano l'operatore a verificare che il workflow stia agendo sul target corretto.

### 27.7 UI e ChangeRequest detail

La pagina ChangeRequest detail deve essere environment-aware.

Deve mostrare o rendere chiaro:

- target environment;
- namespace Kubernetes;
- namespace Tekton;
- Argo CD Application;
- validation path;
- runtime evidence collegata;
- Tekton validation evidence collegata;
- stato sanitized delle evidence.

Questo rende possibile analizzare una singola richiesta senza perdere il contesto dell'ambiente.

### 27.8 UI e dashboard

La dashboard deve fornire una vista sintetica degli ambienti.

Elementi importanti:

- latest ChangeRequest;
- `Environments / Namespaces`;
- user box;
- summary delle evidence;
- link al dettaglio della richiesta.

La dashboard deve aiutare a capire rapidamente se la piattaforma rappresenta correttamente dev, staging e production.

### 27.9 UI e evidence rendering

Le evidence visualizzate nella UI devono mantenere il contesto environment-aware.

Per esempio, una Tekton validation evidence deve indicare:

- environment;
- Tekton namespace;
- PipelineRun;
- validation path;
- result;
- failed task count.

Una runtime evidence deve indicare:

- environment;
- cluster;
- namespace;
- deployment;
- readiness;
- Argo CD status, se disponibile.

### 27.10 UI e multi-cluster readiness

La UI deve essere pronta a rappresentare in futuro cluster fisici separati.

Oggi lo stato validato e namespace-isolated su `ocp-dev`.

Domani staging e production potrebbero essere associati a cluster diversi.

La UI deve quindi evitare assunzioni rigide come:

```text
all environments always run on ocp-dev
```

La UI deve invece mostrare il target risolto e lasciare evidente la relazione tra ambiente, namespace e cluster.

### 27.11 Simulazione staging e production cluster

Il codice ha validato target simulati:

```text
staging -> ocp-staging-simulated
production -> ocp-production-simulated
```

Questa simulazione e una validazione del modello codice, non una validazione fisica runtime.

La UI deve quindi supportare il modello, ma non dichiarare validazione fisica cross-cluster finche non esistono cluster reali.

### 27.12 Cosa mostrare

La UI puo mostrare informazioni operative sicure:

- ambiente;
- cluster name;
- namespace;
- Argo CD Application;
- PipelineRun;
- validation path;
- status;
- reason;
- failed task count;
- sanitized state.

Queste informazioni sono utili e non espongono credenziali.

### 27.13 Cosa non mostrare

La UI non deve mostrare:

- Secret raw;
- token;
- password;
- kubeconfig;
- bearer token;
- private key;
- contenuto Secret decodificato;
- credenziali Git o Argo CD.

Environment awareness non deve compromettere la sicurezza.

### 27.14 Errori da evitare

Errori tipici da evitare nella UI:

- mostrare sempre `dev` anche per staging o production;
- non mostrare il namespace;
- non mostrare il validation path;
- nascondere il Tekton namespace;
- confondere production logica con cluster production fisico;
- non distinguere tra baseline validata e validazione fisica deferred;
- mostrare evidence senza stato sanitized.

### 27.15 Relazione con operability

I runbook operativi richiedono di verificare che la UI esponga gli elementi corretti.

Dopo manutenzione o rollout, l'operatore deve controllare:

- dashboard HTTP `200`;
- presenza di `Environments / Namespaces`;
- latest ChangeRequest;
- ChangeRequest detail;
- runtime evidence card;
- Tekton validation card;
- assenza di dati sensibili.

### 27.16 Sintesi

La UI environment awareness rende visibile il modello multi-environment del DevOps Control Plane.

Essa consente agli operatori di distinguere dev, staging e production, di vedere namespace e target runtime, e di interpretare le evidence nel contesto corretto.

Questa funzionalita e essenziale nella baseline namespace-isolated e sara ancora piu importante quando saranno disponibili cluster fisici separati.

## 28. Environment Catalog

L'Environment Catalog e il modello con cui il DevOps Control Plane descrive gli ambienti logici supportati dalla piattaforma.

Un ambiente logico non coincide necessariamente con un cluster fisico. Un ambiente logico rappresenta uno scopo operativo, come sviluppo, staging o produzione. Il catalogo serve a trasformare questo concetto logico in metadati tecnici utilizzabili dal backend, dai workflow, dalla UI e dai runbook operativi.

Nel baseline corrente, gli ambienti sono:

- `dev`;
- `staging`;
- `production`.

La mappatura validata e namespace-isolated sul cluster OpenShift disponibile `ocp-dev`:

```text
dev        -> ocp-dev / devops-ci-demo
staging    -> ocp-dev / devops-ci-staging
production -> ocp-dev / devops-ci-production
```

L'Environment Catalog e quindi il punto in cui il DevOps Control Plane inizia a distinguere tra ambiente logico, namespace e cluster.

### 28.1 Perche serve l'Environment Catalog

Senza un Environment Catalog, il sistema rischierebbe di avere configurazioni hardcoded o sparse nel codice.

Questo sarebbe fragile perche:

- renderebbe difficile aggiungere nuovi ambienti;
- renderebbe difficile distinguere staging e production;
- aumenterebbe il rischio di azioni nel namespace sbagliato;
- ridurrebbe la possibilita di supportare cluster futuri;
- renderebbe meno chiara la UI.

L'Environment Catalog centralizza la descrizione degli ambienti e consente al backend di risolvere target tecnici in modo coerente.

### 28.2 Ambiente logico

Un ambiente logico rappresenta il contesto operativo della ChangeRequest.

Esempi:

```text
dev
staging
production
```

Quando una ChangeRequest indica `targetEnvironment = staging`, il DevOps Control Plane deve sapere quali namespace, Application Argo CD, path GitOps e pipeline usare per staging.

L'Environment Catalog rende questa associazione esplicita.

### 28.3 Metadati dell'ambiente

Un ambiente nel catalogo puo includere metadati come:

- nome ambiente;
- display name;
- flag enabled;
- cluster name;
- Kubernetes namespace;
- Tekton namespace;
- Argo CD Application name;
- Git target branch;
- validation path;
- flag per azioni tecniche abilitate.

Questi campi non sono puramente descrittivi. Sono usati per costruire il runtime target tecnico.

### 28.4 Namespace Kubernetes

Il Kubernetes namespace indica dove il workload applicativo viene osservato o gestito.

Nel baseline corrente:

```text
dev        -> devops-ci-demo
staging    -> devops-ci-staging
production -> devops-ci-production
```

Questa informazione e usata da azioni come:

- `collect-evidence`;
- `check-deployment`;
- runtime evidence;
- route health check;
- UI environment visibility.

### 28.5 Namespace Tekton

Il Tekton namespace indica dove vengono create o controllate le PipelineRun di validazione.

Nel baseline corrente, staging e production usano namespace Tekton coerenti con il proprio namespace applicativo:

```text
staging    -> devops-ci-staging
production -> devops-ci-production
```

Questa informazione e essenziale per evitare che una validazione venga eseguita nel namespace sbagliato.

### 28.6 Argo CD Application

L'Environment Catalog deve associare ogni ambiente alla corretta Argo CD Application.

Esempi:

```text
dev        -> demo-go-color-app
staging    -> demo-go-color-app-staging
production -> demo-go-color-app-production
```

Questa associazione permette al DevOps Control Plane di raccogliere Argo CD deployment evidence nel contesto corretto.

### 28.7 Validation path

Il validation path indica quale parte del repository GitOps deve essere validata da Tekton.

Esempi validati:

```text
staging     apps/demo-go-color-app/overlays/staging
production  apps/demo-go-color-app/overlays/production
```

Il validation path environment-specific e fondamentale. Se il path e sbagliato, una validazione potrebbe controllare l'overlay errato.

### 28.8 Technical actions

L'Environment Catalog puo indicare se un ambiente consente azioni tecniche.

Esempi di azioni:

- collect evidence;
- check deployment;
- validate;
- check validation.

Questo consente alla UI e al backend di applicare guardrail environment-aware.

Un ambiente disabled o non abilitato alle azioni tecniche deve essere trattato in modo conservativo.

### 28.9 Relazione con la UI

La UI usa il modello environment-aware per mostrare informazioni corrette agli operatori.

La dashboard e la pagina ChangeRequest detail devono rendere visibile:

- ambiente target;
- namespace;
- Argo CD Application;
- validation path;
- evidence associate.

La sezione `Environments / Namespaces` della dashboard e una rappresentazione diretta del fatto che il modello multi-environment e esplicito.

### 28.10 Relazione con Runtime Target Resolution

L'Environment Catalog e uno degli input della runtime target resolution.

Il processo concettuale e:

```text
ChangeRequest targetEnvironment
      |
      v
Environment Catalog
      |
      v
Cluster Registry
      |
      v
TechnicalRuntimeTarget
```

Il risultato e un target tecnico completo, usato da workflow runtime, evidence collection, Tekton validation e UI.

### 28.11 Relazione con Cluster Registry

L'Environment Catalog indica a quale cluster logico e associato un ambiente.

Il Cluster Registry descrive invece il cluster.

Quindi:

- Environment Catalog risponde alla domanda: quale cluster usa questo ambiente?
- Cluster Registry risponde alla domanda: che cosa sappiamo di quel cluster?

Questa separazione e importante per il futuro multi-cluster.

### 28.12 Baseline corrente

Baseline corrente:

```text
Environment   Cluster   Kubernetes namespace     Tekton namespace       Argo CD Application
----------    -------   ---------------------    ------------------     -------------------------------
dev           ocp-dev   devops-ci-demo           devops-ci-demo         demo-go-color-app
staging       ocp-dev   devops-ci-staging        devops-ci-staging      demo-go-color-app-staging
production    ocp-dev   devops-ci-production     devops-ci-production   demo-go-color-app-production
```

Questa tabella descrive la baseline validata e non deve essere confusa con una topologia multi-cluster fisica.

### 28.13 Relazione con multi-cluster readiness

L'Environment Catalog e uno dei componenti che rende il codice multi-cluster-ready.

Oggi staging e production puntano al cluster fisico disponibile `ocp-dev`.

Nei test post-Fase 15, staging e production sono stati anche modellati come target simulati:

```text
staging -> ocp-staging-simulated
production -> ocp-production-simulated
```

Questi test dimostrano che il modello puo rappresentare cluster diversi.

La validazione fisica resta deferred per indisponibilita di cluster aggiuntivi.

### 28.14 Fail-closed

L'Environment Catalog deve partecipare al comportamento fail-closed.

Esempi di condizioni da bloccare:

- ambiente sconosciuto;
- ambiente disabled;
- mapping incompleto;
- namespace mancante;
- cluster reference non valida;
- validation path mancante per workflow che lo richiede.

In questi casi il sistema deve fermarsi con errore esplicito, non procedere con default non sicuri.

### 28.15 Errori da evitare

Errori da evitare:

- hardcodare staging o production nel codice;
- usare sempre `devops-ci-demo` come default;
- nascondere l'ambiente nella UI;
- validare production con il path staging;
- trattare production logica come cluster fisico production;
- fare fallback silenzioso verso `ocp-dev` quando un ambiente dovrebbe puntare altrove.

### 28.16 Sintesi

L'Environment Catalog e il punto di controllo per descrivere gli ambienti logici del DevOps Control Plane.

Collega ChangeRequest, namespace, Argo CD, Tekton, validation path, UI e runtime target resolution.

Grazie a questo modello, il progetto puo operare oggi con namespace isolation e prepararsi domani a un vero multi-cluster senza riprogettare il workflow.

## 29. Cluster Registry

Il Cluster Registry e il modello con cui il DevOps Control Plane descrive i cluster disponibili o previsti.

Se l'Environment Catalog risponde alla domanda:

```text
Quale ambiente logico sto usando?
```

il Cluster Registry risponde alla domanda:

```text
Che cosa so del cluster associato a quell'ambiente?
```

Questa distinzione e importante perche un ambiente logico non deve essere confuso con un cluster fisico. Oggi `dev`, `staging` e `production` sono ambienti logici validati sullo stesso cluster `ocp-dev`, usando namespace separati. In futuro, staging e production potrebbero essere associati a cluster fisici distinti.

### 29.1 Perche serve il Cluster Registry

Senza Cluster Registry, le informazioni sui cluster rischierebbero di essere hardcoded o distribuite in punti diversi del codice.

Questo sarebbe fragile perche:

- renderebbe difficile aggiungere nuovi cluster;
- renderebbe difficile disabilitare un cluster in modo controllato;
- aumenterebbe il rischio di fallback verso il cluster sbagliato;
- renderebbe poco chiara la separazione tra ambiente logico e runtime fisico;
- complicherebbe l'onboarding futuro di cluster reali.

Il Cluster Registry centralizza le informazioni essenziali sui cluster e permette al backend di applicare guardrail coerenti.

### 29.2 Differenza tra Environment Catalog e Cluster Registry

Environment Catalog e Cluster Registry hanno responsabilita diverse.

Environment Catalog:

- descrive ambienti logici;
- associa un ambiente a namespace, Tekton namespace, Argo CD Application e validation path;
- indica quale cluster logico usare per l'ambiente.

Cluster Registry:

- descrive i cluster;
- indica se un cluster e abilitato;
- conserva metadati tecnici del cluster;
- definisce namespace consentiti;
- prepara l'integrazione con Secret references e provider runtime.

Vista concettuale:

```text
ChangeRequest targetEnvironment
      |
      v
Environment Catalog
      |
      v
clusterName
      |
      v
Cluster Registry
      |
      v
TechnicalRuntimeTarget
```

### 29.3 Cluster name

Il `clusterName` e l'identificativo logico del cluster.

Esempio corrente:

```text
ocp-dev
```

Esempi simulati usati nei test post-Fase 15:

```text
ocp-staging-simulated
ocp-production-simulated
```

Il cluster name deve essere esplicito nelle risoluzioni runtime e nelle evidence quando rilevante.

### 29.4 Enabled flag

Un cluster deve poter essere abilitato o disabilitato.

Il flag enabled e un guardrail operativo.

Un cluster disabled non deve essere usato per azioni runtime.

Comportamento atteso:

```text
cluster disabled -> fail closed
```

Questo evita che un cluster venga usato prima che siano completate readiness, RBAC, Secret references, provider e smoke test.

### 29.5 API URL

Il Cluster Registry puo includere l'API URL del cluster.

Esempio concettuale:

```text
apiURL = https://api.ocp-dev.example:6443
```

L'API URL e necessaria quando il runtime deve costruire client reali verso un cluster.

Se un client factory richiede API URL e quel valore manca, il comportamento corretto e fail-closed.

### 29.6 Default namespace e allowed namespaces

Il Cluster Registry puo descrivere namespace predefiniti e namespace consentiti.

Esempio:

```text
cluster = ocp-dev
defaultNamespace = devops-ci-demo
allowedNamespaces = devops-ci-demo, devops-ci-staging, devops-ci-production
```

Questi dati aiutano a evitare accessi non intenzionali a namespace non previsti.

Il principio operativo e:

```text
allow only what is required
```

### 29.7 Secret references

Il Cluster Registry e collegato al modello Secret reference.

Un cluster reale puo richiedere credenziali, token o CA reference per costruire client runtime.

Il modello corretto non salva valori raw nel registry.

Il registry o i modelli collegati devono fare riferimento a Secret references.

Esempio accettabile:

```text
clusterName = ocp-nonprod
secretRefName = dcp-ocp-nonprod-runtime-token
secretNamespace = devops-control-plane
```

Esempio non accettabile:

```text
token = actual-token-value
```

### 29.8 Provider selection

Il Cluster Registry contribuisce alla provider selection.

Dopo aver risolto il cluster target, il backend deve selezionare un provider runtime adatto.

Se il provider manca, l'azione deve fallire.

Se il provider e disabled, l'azione deve fallire.

Il sistema non deve ricadere automaticamente su `ocp-dev`.

Questa regola e fondamentale per evitare azioni nel cluster sbagliato.

### 29.9 Cluster corrente ocp-dev

Nel baseline validato, il cluster fisico disponibile e:

```text
ocp-dev
```

Su questo cluster sono stati validati tre ambienti logici tramite namespace isolation:

```text
dev        -> devops-ci-demo
staging    -> devops-ci-staging
production -> devops-ci-production
```

Il Cluster Registry deve quindi rappresentare `ocp-dev` come cluster corrente della baseline operativa.

### 29.10 Cluster simulati staging e production

Dopo la chiusura della Fase 15, il codice e stato rafforzato con test per cluster simulati distinti:

```text
staging -> ocp-staging-simulated
production -> ocp-production-simulated
```

Questi test dimostrano che il modello puo risolvere cluster diversi da `ocp-dev`.

I test confermano:

- nessun fallback silenzioso verso `ocp-dev`;
- provider mancante fail-closed;
- provider disabled fail-closed;
- metadati runtime environment-specific preservati.

Questa e una validazione del modello codice, non una validazione fisica runtime.

### 29.11 Fail-closed behavior

Il Cluster Registry deve supportare comportamenti fail-closed.

Esempi:

- cluster sconosciuto;
- cluster disabled;
- namespace non consentito;
- API URL mancante quando richiesta;
- Secret reference mancante;
- provider mancante;
- provider disabled.

In tutti questi casi l'azione deve fermarsi con errore chiaro.

Un errore esplicito e preferibile a un'azione eseguita sul cluster sbagliato.

### 29.12 Relazione con Environment Catalog

L'Environment Catalog indica quale cluster usare.

Il Cluster Registry descrive quel cluster.

Esempio:

```text
Environment Catalog
  staging -> clusterName ocp-dev

Cluster Registry
  ocp-dev -> enabled, namespaces allowed, metadata
```

In futuro:

```text
Environment Catalog
  staging -> clusterName ocp-nonprod

Cluster Registry
  ocp-nonprod -> enabled, API URL, allowed namespaces, Secret references
```

Questo modello permette di spostare un ambiente verso un cluster reale senza cambiare il significato della ChangeRequest.

### 29.13 Relazione con runtime target resolution

La runtime target resolution combina Environment Catalog e Cluster Registry.

Il risultato e un `TechnicalRuntimeTarget`.

Questo target contiene informazioni come:

- target environment;
- cluster name;
- Kubernetes namespace;
- Tekton namespace;
- Argo CD Application;
- validation path;
- cluster enabled flag;
- eventuali metadati cluster.

Il workflow runtime deve usare questo target e non inventare destinazioni alternative.

### 29.14 Relazione con runtime factories

Quando si usano client runtime reali, le runtime factories possono richiedere dati provenienti dal Cluster Registry o dalla configurazione collegata.

Esempi:

- Kubernetes API URL;
- Argo CD base URL;
- token reference;
- CA settings;
- namespace consentiti.

Se una factory non ha i dati necessari, deve fallire in modo esplicito.

Esempi di fail-closed:

- API URL mancante;
- token mancante;
- raw CA non supportata;
- kubeconfig non supportato;
- factory disabled.

### 29.15 Real-cluster onboarding futuro

Quando sara disponibile un cluster reale aggiuntivo, il Cluster Registry dovra essere aggiornato con dati controllati.

Informazioni richieste:

- cluster name;
- display name;
- enabled flag inizialmente conservativo;
- API URL;
- allowed namespaces;
- Secret references;
- RBAC previsto;
- provider runtime;
- readiness gates;
- rollback plan.

Il cluster non deve essere abilitato fino al completamento dei controlli di readiness.

### 29.16 Errori da evitare

Errori da evitare:

- usare `ocp-dev` come fallback implicito;
- abilitare un cluster senza readiness;
- usare namespace non allow-listed;
- salvare token raw nel registry;
- confondere cluster simulato con cluster fisico reale;
- dichiarare validazione fisica multi-cluster senza cluster reale;
- abilitare provider o factory per bypassare un errore.

### 29.17 Sintesi

Il Cluster Registry e il componente che rende esplicita la conoscenza dei cluster nel DevOps Control Plane.

Insieme all'Environment Catalog, permette di trasformare un ambiente logico in un target runtime tecnico.

Oggi rappresenta la baseline `ocp-dev` namespace-isolated. Domani permettera l'onboarding controllato di cluster fisici aggiuntivi, mantenendo fail-closed, Secret references e guardrail operativi.
