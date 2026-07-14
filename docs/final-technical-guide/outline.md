# DevOps Control Plane — Guida tecnica finale

Status: Draft  
Phase: 12.1 — Final technical document structure  
Owner: Vincenzo Marzario  
Language: Italian  
Last updated: 2026-07-09

## 1. Scopo del documento

Questo documento definisce la struttura della guida tecnica finale del progetto DevOps Control Plane.

La guida finale dovrà essere scritta in italiano e dovrà spiegare in modo organico, progressivo e completo:

- la visione del progetto;
- i concetti di base necessari per comprenderlo;
- l’architettura applicativa;
- il modello GitOps;
- l’integrazione con Argo CD;
- l’integrazione con Tekton;
- il modello ChangeRequest;
- il modello runtime evidence;
- la UI e la dashboard;
- il modello Environment Catalog;
- la baseline namespace-isolated;
- la readiness multi-cluster;
- i guardrail di sicurezza;
- l’operability;
- la manutenzione;
- il troubleshooting;
- lo stato corrente e la roadmap.

Il documento non deve essere un semplice stato di avanzamento.

Il documento deve essere utilizzabile come riferimento tecnico, materiale di onboarding e base per handover operativo.

## 2. Lingua e stile

Il documento finale sarà scritto in italiano.

I comandi, i nomi di file, i nomi di risorse Kubernetes/OpenShift, i nomi di API, i nomi di branch, i nomi di task Tekton, i nomi di commit e gli esempi tecnici resteranno in inglese o nel formato originale.

Esempi:

- `go test ./internal/api ./internal/app ./cmd/devops-control-plane`
- `oc get pod -n devops-control-plane`
- `ChangeRequest`
- `PipelineRun`
- `Application`
- `Environment Catalog`
- `RuntimeClientProviderRegistry`

Lo stile dovrà essere:

- narrativo;
- progressivo;
- adatto anche a lettori non esperti;
- tecnicamente accurato;
- ricco di esempi;
- orientato all’operatività;
- coerente con lo stato reale del progetto.

## 3. Audience

Il documento finale dovrà essere utile a più categorie di lettori.

### 3.1 Lettori introduttivi

Persone che conoscono poco o nulla di:

- Kubernetes;
- OpenShift;
- GitOps;
- Argo CD;
- Tekton;
- namespace isolation;
- runtime evidence.

Per questi lettori il documento dovrà spiegare i concetti prima di parlare dell’implementazione.

### 3.2 Sviluppatori

Sviluppatori che devono capire:

- come funziona il backend Go;
- come sono gestite le ChangeRequest;
- come viene persistito lo stato;
- come vengono raccolte le evidenze;
- dove intervenire in caso di nuove evoluzioni.

### 3.3 Platform engineer e DevOps engineer

Profili tecnici che devono capire:

- come il sistema gira su OpenShift;
- come interagisce con Argo CD;
- come interagisce con Tekton;
- come si validano dev, staging e production;
- quali sono i guardrail di sicurezza;
- come si prepara il futuro multi-cluster.

### 3.4 Operatori

Persone che devono eseguire:

- health check;
- manutenzione;
- troubleshooting;
- raccolta evidenze;
- validazione runtime;
- verifica Argo CD;
- verifica Tekton;
- verifica UI.

## 4. Stato corrente da rappresentare

Il documento finale dovrà rappresentare lo stato corrente del progetto.

### 4.1 Baseline runtime attuale

La baseline runtime validata è namespace-isolated sul cluster OpenShift disponibile `ocp-dev`.

Mappatura corrente:

- `dev` -> `ocp-dev` / `devops-ci-demo`
- `staging` -> `ocp-dev` / `devops-ci-staging`
- `production` -> `ocp-dev` / `devops-ci-production`

Questa baseline è validata e taggata.

Tag:

`namespace-isolated-baseline-20260709`

### 4.2 Stato multi-cluster

La validazione fisica cross-cluster è rinviata perché non sono disponibili altri cluster OpenShift.

La readiness multi-cluster a livello codice è completata, testata e documentata.

Il documento dovrà usare una formulazione chiara:

Physical cross-cluster runtime validation is deferred by infrastructure availability.

Multi-cluster code readiness is completed, tested, documented and fail-closed.

### 4.3 Simulazione staging e production cluster

Il codice è stato rafforzato con test che simulano target esterni distinti:

- `staging` -> `ocp-staging-simulated`
- `production` -> `ocp-production-simulated`

I test validano:

- nessun fallback silenzioso verso `ocp-dev`;
- provider mancante fail-closed;
- provider disabled fail-closed;
- metadata runtime environment-specific.

## 5. Struttura proposta del documento finale

## Parte 1 — Introduzione e visione

### Capitolo 1 — Executive Summary

Contenuti:

- che cos’è il DevOps Control Plane;
- quale problema risolve;
- quali strumenti integra;
- qual è lo stato attuale;
- cosa è validato;
- cosa è rinviato;
- perché il progetto è utile.

### Capitolo 2 — Visione del progetto

Contenuti:

- obiettivo generale;
- centralizzazione del ciclo di vita ChangeRequest;
- audit;
- GitOps;
- validazione tecnica;
- runtime evidence;
- readiness multi-cluster;
- ruolo della UI.

### Capitolo 3 — Scope del progetto

Contenuti:

- cosa è incluso;
- cosa è escluso;
- cosa è MVP evoluto;
- cosa è baseline operativa;
- cosa resta roadmap futura.

## Parte 2 — Concetti fondamentali

### Capitolo 4 — Kubernetes e OpenShift in breve

Contenuti:

- cluster;
- namespace;
- pod;
- deployment;
- service;
- route;
- secret;
- configmap;
- serviceaccount;
- RBAC.

### Capitolo 5 — Namespace isolation

Contenuti:

- cosa significa isolamento per namespace;
- perché viene usato nel progetto;
- differenza tra ambiente logico e cluster fisico;
- dev, staging e production sullo stesso cluster;
- vantaggi;
- limiti;
- relazione con il futuro multi-cluster.

### Capitolo 6 — GitOps

Contenuti:

- Git come fonte di verità;
- manifest dichiarativi;
- pull-based reconciliation;
- differenza tra cambio manuale e cambio GitOps;
- relazione con audit e rollback;
- distinzione tra GitLab SCM e repository GitHub consumato dal runtime;
- sincronizzazione GitLab-GitHub non dimostrata.

### Capitolo 7 — Kustomize base e overlays

Contenuti:

- base comune;
- overlay dev;
- overlay staging;
- overlay production;
- validation path environment-specific;
- esempi logici.

### Capitolo 8 — Argo CD

Contenuti:

- ruolo di Argo CD;
- Application;
- sync;
- health;
- target namespace;
- overlay GitOps;
- stato Synced e Healthy;
- applicazioni staging e production;
- una applicazione logica con tre istanze ambientali;
- `demo-app` come Argo CD Application storica standalone.

### Capitolo 9 — Tekton

Contenuti:

- Task;
- Pipeline;
- PipelineRun;
- TaskRun;
- validazione GitOps;
- check-validation;
- evidence Tekton.

## Parte 3 — Architettura del DevOps Control Plane

### Capitolo 10 — Architettura generale

Contenuti:

- backend Go;
- PostgreSQL;
- GitLab API;
- Argo CD API;
- Kubernetes API;
- Tekton API;
- UI;
- evidence store;
- adapter pattern.

### Capitolo 11 — Backend Go

Contenuti:

- ChangeService;
- repository layer;
- adapter;
- configurazione;
- service options;
- runtime target resolver;
- provider registry.

### Capitolo 12 — PostgreSQL e persistenza

Contenuti:

- ChangeRequest;
- ChangeEvent;
- Evidence;
- audit trail;
- retention futura;
- backup e restore.

### Capitolo 13 — Modello dati

Contenuti:

- entità principali;
- stati lifecycle;
- stati runtime;
- eventi;
- evidenze;
- relazioni.

## Parte 4 — Workflow applicativi

### Capitolo 14 — ChangeRequest lifecycle

Contenuti:

- creazione;
- validazione;
- eventi;
- transizioni;
- processo tecnico;
- audit.

### Capitolo 15 — GitLab Merge Request workflow

Contenuti:

- branch;
- commit;
- merge request;
- merge;
- stato GitLab;
- confine tra workflow SCM GitLab e repository GitOps runtime;
- assenza di sincronizzazione GitLab-GitHub dimostrata;
- relazione con GitOps tramite ChangeRequest, ambiente ed evidence.

### Capitolo 16 — Workflow runtime

Contenuti:

- `create`;
- `collect-evidence`;
- `check-deployment`;
- `validate`;
- `check-validation`;
- UI evidence rendering.

### Capitolo 17 — Workflow dev, staging e production

Contenuti:

- baseline namespace-isolated;
- differenze tra ambienti;
- mapping namespace;
- evidenze finali;
- esempi `CHG-2026-0049` e `CHG-2026-0050`.

## Parte 5 — Evidence model

### Capitolo 18 — Runtime evidence

Contenuti:

- cosa viene osservato;
- namespace;
- deployment;
- pods;
- services;
- routes;
- Argo CD;
- diagnostica.

### Capitolo 19 — Tekton validation evidence

Contenuti:

- PipelineRun;
- Tekton namespace;
- pipeline;
- validation path;
- failed task count;
- reason;
- status;
- sanitized evidence.

### Capitolo 20 — Argo CD deployment evidence

Contenuti:

- Application;
- sync;
- health;
- revision;
- target namespace;
- overlay path.

### Capitolo 21 — Evidence sanitization

Contenuti:

- cosa si può salvare;
- cosa non si deve salvare;
- Secret;
- token;
- kubeconfig;
- raw CA;
- safe summary.

## Parte 6 — UI e dashboard

### Capitolo 22 — Dashboard

Contenuti:

- latest ChangeRequest;
- KPI;
- recent changes;
- evidence summary;
- Environments / Namespaces;
- user box.

### Capitolo 23 — ChangeRequest detail

Contenuti:

- runtime evidence card;
- Tekton validation card;
- raw sanitized evidence;
- audit log;
- azioni tecniche.

### Capitolo 24 — UI environment awareness

Contenuti:

- dev;
- staging;
- production;
- namespace visibility;
- no static dev-only representation.

## Parte 7 — Environment Catalog e multi-cluster readiness

### Capitolo 25 — Environment Catalog

Contenuti:

- ambiente logico;
- namespace;
- Tekton namespace;
- Argo CD Application;
- validation path;
- technical actions.

### Capitolo 26 — Cluster Registry

Contenuti:

- cluster definition;
- enabled flag;
- API URL;
- Secret references;
- allowed namespaces.

### Capitolo 27 — Runtime target resolution

Contenuti:

- EnvironmentClusterResolver;
- TechnicalRuntimeTarget;
- target environment;
- cluster name;
- namespace;
- provider selection.

### Capitolo 28 — Multi-cluster code-ready baseline

Contenuti:

- cosa significa code-ready;
- cosa non significa;
- physical validation deferred;
- simulated staging e production targets;
- no fallback to ocp-dev;
- fail-closed behavior.

### Capitolo 29 — Deferred real-cluster onboarding contract

Contenuti:

- dati richiesti;
- Secret references;
- RBAC;
- Argo CD;
- Tekton;
- readiness gates;
- rollback.

## Parte 8 — Security e guardrail

### Capitolo 30 — RBAC

Contenuti:

- principio del minimo privilegio;
- namespace-scoped RBAC;
- Role e RoleBinding;
- cosa evitare.

### Capitolo 31 — Secret reference model

Contenuti:

- reference invece di valori;
- allow-list;
- loader disabled-by-default;
- no raw Secret in logs;
- no raw Secret in evidence.

### Capitolo 32 — Runtime factories

Contenuti:

- Kubernetes factory;
- Tekton factory;
- Argo CD factory;
- disabled-by-default;
- fail-closed;
- token-based readiness;
- kubeconfig unsupported;
- raw CA unsupported.

### Capitolo 33 — AuthN/AuthZ e OAuth proxy

Contenuti:

- proxy;
- forwarded headers;
- gruppi;
- admin;
- fail-closed authorization.

## Parte 9 — Operability

### Capitolo 34 — Health check

Contenuti:

- pod;
- readyz;
- dashboard;
- Argo CD matrix;
- deployment matrix;
- route health;
- Tekton matrix;
- evidence directory.

### Capitolo 35 — Maintenance operations

Contenuti:

- pre-maintenance;
- post-maintenance;
- Argo CD checks;
- Tekton checks;
- UI checks;
- Secret/RBAC/factory checks.

### Capitolo 36 — Troubleshooting

Contenuti:

- pod not ready;
- readyz failure;
- Argo CD degraded;
- Tekton failed;
- UI missing evidence;
- provider missing;
- provider disabled;
- Secret loading failure.

### Capitolo 37 — Backup, restore e disaster recovery

Contenuti:

- PostgreSQL backup;
- restore isolato;
- DR;
- limiti GitLab, Argo CD e Tekton;
- RPO e RTO.

## Parte 10 — Stato corrente e roadmap

### Capitolo 38 — Stato delle fasi

Contenuti:

- elenco fasi;
- stato corrente;
- Fase 10 chiusa;
- Fase 13 allineata;
- Fase 15 chiusa;
- Fase 11 standby;
- Fase 12 in corso.

### Capitolo 39 — Stato finale corrente

Contenuti:

- cosa è completato;
- cosa è rafforzato;
- cosa è deferred;
- cosa resta da fare.

### Capitolo 40 — Roadmap futura

Contenuti:

- documento finale;
- eventuale CLI;
- real-cluster onboarding;
- produzione reale;
- observability avanzata;
- metriche;
- alerting.

## Appendici

### Appendice A — Glossario

Termini:

- ChangeRequest;
- ChangeEvent;
- Evidence;
- GitOps;
- Argo CD;
- Tekton;
- PipelineRun;
- TaskRun;
- Namespace;
- Environment Catalog;
- Cluster Registry;
- RuntimeTarget;
- Provider Registry;
- Secret reference;
- fail-closed.

### Appendice B — Comandi operativi principali

Contenuti:

- health check;
- smoke matrix;
- Argo CD checks;
- Tekton checks;
- route checks;
- git hygiene.

### Appendice C — Commit e tag rilevanti

Contenuti:

- baseline namespace-isolated;
- chiusura Fase 15;
- simulazione staging e production;
- chiusura Fase 10;
- allineamento Fase 13.

### Appendice D — Limitazioni note

Contenuti:

- no physical multi-cluster validation;
- no additional clusters available;
- no real production cluster onboarding;
- secret loader disabled by default;
- factories disabled by default;
- CLI in standby.

## 6. Fonti documentali di riferimento

La guida finale dovrà usare come fonti principali:

- `docs/00-vision.md`
- `docs/01-scope-mvp.md`
- `docs/03-functional-requirements.md`
- `docs/04-non-functional-requirements.md`
- `docs/05-architecture.md`
- `docs/06-argocd-integration.md`
- `docs/08-tekton-integration.md`
- `docs/09-security-rbac.md`
- `docs/10-data-model.md`
- `docs/11-change-workflows.md`
- `docs/12-evidence-model.md`
- `docs/13-api-design.md`
- `docs/environment-configuration-model.md`
- `docs/multi-cluster-environment-enablement-request.md`
- `docs/runtime-evidence-dashboard-maintenance-alignment.md`
- `docs/phase-10-operability-closure.md`
- `docs/runbooks/operability-health-check.md`
- `docs/runbooks/maintenance-operations.md`

## 7. Criteri di qualità

Il documento finale sarà considerato accettabile solo se:

- è scritto in italiano;
- spiega i concetti prima dell’implementazione;
- non è solo una raccolta di titoli;
- contiene esempi concreti;
- contiene workflow passo-passo;
- distingue chiaramente stato validato e stato rinviato;
- spiega namespace isolation;
- spiega GitOps;
- spiega Argo CD;
- spiega Tekton;
- spiega RBAC;
- spiega Secret guardrails;
- spiega runtime evidence;
- spiega UI e dashboard;
- spiega multi-cluster code readiness;
- contiene una sezione operativa;
- contiene troubleshooting;
- contiene glossario;
- è coerente con i commit e i documenti aggiornati.

## 8. Definition of Done

La Fase 12 sarà completata quando:

1. la guida sorgente Markdown sarà completa;
2. la guida sarà revisionata;
3. il documento Word sarà generato dalla guida sorgente;
4. il documento Word sarà leggibile e strutturato;
5. le sezioni tecniche saranno coerenti con il repository;
6. le limitazioni saranno esplicitate;
7. la guida sarà adatta a onboarding e handover;
8. il repository sarà pulito;
9. la versione finale sarà committata o allegata secondo la strategia scelta.

## 9. Strategia di generazione Word

La guida finale dovrà essere scritta prima come documento Markdown versionabile.

Documento sorgente previsto:

`docs/final-technical-guide/final-technical-guide-it.md`

Output Word previsto:

`DevOps_Control_Plane_Guida_Tecnica_Finale.docx`

Il Word sarà generato solo dopo il completamento e la revisione del contenuto Markdown.

Il Word non deve essere la sorgente primaria.

<!-- CI_GUIDE_OUTLINE_START -->
### Continuous Integration e test automatizzati

Scopo del capitolo:

- descrivere la baseline CI implementata con GitHub Actions;
- spiegare i quality gate applicati alle pull request e ai push su `main`;
- distinguere test unitari, test HTTP end-to-end e test PostgreSQL con build tag `integration`;
- descrivere race detector, coverage atomica e artifact di coverage;
- spiegare la protezione della concorrenza lifecycle con `SELECT ... FOR UPDATE`;
- documentare gli invarianti TLS secure-by-default per configurazione, Argo CD, GitLab, Kubernetes e Tekton;
- chiarire cosa dimostra una pipeline CI verde e cosa resta fuori dal suo perimetro.

Fonte autorevole:

- `docs/continuous-integration-and-automated-testing.md`

Elementi tecnici da includere:

- workflow `.github/workflows/ci.yml`;
- trigger `pull_request` e push su `main`;
- Go 1.22 e cache delle dipendenze;
- controllo `gofmt`;
- `go vet ./...`;
- `go test ./... -race -covermode=atomic -coverprofile=coverage.out`;
- PostgreSQL 16 disposable service container;
- `go test -tags=integration ./internal/database/... -run Integration -v`;
- test HTTP per health, readiness, ChangeRequest API e lifecycle routes;
- test `TestConcurrentApproveOnlyOneWins`;
- test TLS `TestNewVerifiesTLSByDefault` e opt-in esplicito per la modalità insecure;
- required status check `test` e branch protection su `main`;
- limitazioni correnti della CI.

Posizionamento proposto:

- Parte 3 - Architettura del DevOps Control Plane;
- dopo il capitolo Backend Go e prima di PostgreSQL e persistenza.

Impatto sulla numerazione:

- l'inserimento richiede il riallineamento controllato dei capitoli successivi nella guida italiana e nella source map;
- il documento Word deve essere rigenerato solo dopo l'aggiornamento del Markdown e i controlli strutturali.
<!-- CI_GUIDE_OUTLINE_END -->
