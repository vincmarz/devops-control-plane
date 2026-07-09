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
