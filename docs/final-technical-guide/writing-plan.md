# DevOps Control Plane — Piano di scrittura della guida tecnica finale

Status: Draft  
Phase: 12.3 — Create detailed writing plan for final technical guide  
Owner: Vincenzo Marzario  
Language: Italian  
Last updated: 2026-07-09

## 1. Scopo

Questo documento definisce il piano di scrittura della guida tecnica finale del DevOps Control Plane.

La guida finale dovrà essere scritta in italiano, con comandi, nomi tecnici, risorse Kubernetes/OpenShift, riferimenti Git, API, nomi Tekton e nomi Argo CD mantenuti nel formato originale.

La guida finale non deve essere:

- un semplice stato di avanzamento;
- un elenco di titoli;
- una copia lineare della documentazione esistente;
- un documento troppo sintetico;
- un documento adatto solo a chi conosce già il progetto.

La guida finale deve essere:

- organica;
- narrativa;
- progressiva;
- tecnicamente accurata;
- utile per onboarding;
- utile per handover operativo;
- utile come riferimento tecnico;
- comprensibile anche a lettori non esperti.

## 2. Documenti di riferimento

La scrittura della guida finale partirà da:

- `docs/final-technical-guide/outline.md`;
- `docs/final-technical-guide/source-map.md`.

Il documento sorgente finale previsto è:

- `docs/final-technical-guide/final-technical-guide-it.md`.

L’output Word finale previsto è:

- `DevOps_Control_Plane_Guida_Tecnica_Finale.docx`.

## 3. Approccio di scrittura

La guida finale sarà scritta in modo incrementale.

Ogni parte dovrà essere completata, revisionata e committata prima di passare alla parte successiva.

Ordine di scrittura:

1. Parte 1 — Introduzione e visione
2. Parte 2 — Concetti fondamentali
3. Parte 3 — Architettura
4. Parte 4 — Workflow applicativi
5. Parte 5 — Evidence model
6. Parte 6 — UI e dashboard
7. Parte 7 — Environment Catalog e multi-cluster readiness
8. Parte 8 — Security e guardrail
9. Parte 9 — Operability
10. Parte 10 — Stato corrente e roadmap
11. Appendici
12. Revisione complessiva
13. Generazione Word

## 4. Regole generali di scrittura

Ogni capitolo dovrà seguire una struttura coerente:

1. spiegazione del concetto;
2. perché il concetto è importante nel progetto;
3. come il concetto è implementato nel DevOps Control Plane;
4. esempi concreti;
5. stato attuale;
6. limiti noti;
7. collegamento con altri capitoli.

Quando un concetto è complesso, la guida deve prima spiegarlo in modo semplice e poi passare al dettaglio tecnico.

Esempio:

- prima spiegare cosa significa `namespace`;
- poi spiegare cosa significa `namespace isolation`;
- poi spiegare come il progetto usa `devops-ci-demo`, `devops-ci-staging` e `devops-ci-production`.

## 5. Livello di dettaglio atteso

La guida finale deve essere significativamente più dettagliata di un documento di avanzamento.

Ogni capitolo principale deve contenere:

- testo narrativo;
- spiegazione dei concetti;
- esempi concreti;
- riferimenti allo stato reale del progetto;
- eventuali limiti noti;
- indicazioni operative dove utili.

La guida deve spiegare non solo cosa è stato fatto, ma anche perché è stato fatto.

## 6. Parte 1 — Introduzione e visione

Capitoli inclusi:

- Executive Summary;
- Visione del progetto;
- Scope del progetto.

Obiettivi:

- spiegare che cos’è il DevOps Control Plane;
- spiegare il problema che risolve;
- spiegare il valore dell’integrazione tra GitOps, Tekton, Argo CD e runtime evidence;
- presentare lo stato corrente del progetto;
- dichiarare chiaramente cosa è validato e cosa è rinviato.

Elementi obbligatori:

- baseline namespace-isolated;
- Fase 10 chiusa come baseline operativa avanzata;
- Fase 13 riallineata;
- Fase 15 chiusa come multi-cluster code-ready baseline;
- Fase 11 in standby;
- validazione fisica multi-cluster rinviata.

## 7. Parte 2 — Concetti fondamentali

Capitoli inclusi:

- Kubernetes e OpenShift in breve;
- Namespace isolation;
- GitOps;
- Kustomize base e overlays;
- Argo CD;
- Tekton.

Obiettivi:

- rendere il documento comprensibile anche a chi non parte da Kubernetes;
- introdurre gradualmente i concetti;
- collegare ogni concetto al progetto reale.

Esempi da includere:

- `apps/demo-go-color-app/overlays/staging`;
- `apps/demo-go-color-app/overlays/production`;
- `demo-go-color-app-staging`;
- `demo-go-color-app-production`;
- `devops-cp-validate-chg-2026-0049-nd7rm`;
- `devops-cp-validate-chg-2026-0050-8wqtv`.

## 8. Parte 3 — Architettura

Capitoli inclusi:

- Architettura generale;
- Backend Go;
- PostgreSQL e persistenza;
- Modello dati.

Elementi obbligatori:

- ChangeService;
- ChangeRequest;
- ChangeEvent;
- Evidence;
- repository layer;
- adapter per GitLab, Kubernetes, Tekton e Argo CD;
- runtime target resolver;
- provider registry.

## 9. Parte 4 — Workflow applicativi

Capitoli inclusi:

- ChangeRequest lifecycle;
- GitLab Merge Request workflow;
- Workflow runtime;
- Workflow dev, staging e production.

Elementi obbligatori:

- `create`;
- `collect-evidence`;
- `check-deployment`;
- `validate`;
- `check-validation`;
- audit log;
- runtime status;
- process status.

Esempi da includere:

- `CHG-2026-0049` per staging;
- `CHG-2026-0050` per production;
- validation path staging;
- validation path production;
- PipelineRun staging;
- PipelineRun production.

## 10. Parte 5 — Evidence model

Capitoli inclusi:

- Runtime evidence;
- Tekton validation evidence;
- Argo CD deployment evidence;
- Evidence sanitization.

Elementi obbligatori:

- runtime evidence card;
- Tekton validation card;
- failed task count;
- validation path;
- PipelineRun;
- reason;
- status;
- sanitized state;
- raw evidence diagnostics.

Da chiarire:

- l’evidence può contenere nomi di risorse;
- l’evidence non deve contenere Secret, token o kubeconfig;
- la UI deve mostrare solo informazioni sicure.

## 11. Parte 6 — UI e dashboard

Capitoli inclusi:

- Dashboard;
- ChangeRequest detail;
- UI environment awareness.

Elementi obbligatori:

- latest ChangeRequest;
- `Environments / Namespaces`;
- user box;
- Tekton validation card;
- runtime evidence card;
- audit log;
- sanitized raw evidence.

Da chiarire:

- la UI non deve più essere descritta solo come MVP;
- la UI è ora utile per validazione operativa;
- staging e production sono visibili come ambienti logici separati.

## 12. Parte 7 — Environment Catalog e multi-cluster readiness

Capitoli inclusi:

- Environment Catalog;
- Cluster Registry;
- Runtime target resolution;
- Multi-cluster code-ready baseline;
- Deferred real-cluster onboarding contract.

Elementi obbligatori:

- Environment Catalog;
- Cluster Registry;
- TechnicalRuntimeTarget;
- RuntimeClientProviderRegistry;
- simulated staging target;
- simulated production target;
- no fallback to `ocp-dev`;
- provider missing fail-closed;
- provider disabled fail-closed.

Formulazione obbligatoria:

Physical cross-cluster runtime validation is deferred by infrastructure availability.

Multi-cluster code readiness is completed, tested, documented and fail-closed.

## 13. Parte 8 — Security e guardrail

Capitoli inclusi:

- RBAC;
- Secret reference model;
- Runtime factories;
- AuthN/AuthZ e OAuth proxy.

Elementi obbligatori:

- namespace-scoped RBAC;
- Secret reference;
- allow-list;
- runtime Secret loader disabled by default;
- Kubernetes runtime client factory;
- Tekton runtime client factory;
- Argo CD runtime client factory;
- unsupported kubeconfig;
- unsupported raw CA.

Da chiarire:

- un errore fail-closed è spesso il comportamento corretto;
- non bisogna aggirare provider disabled o missing provider;
- non bisogna forzare fallback a `ocp-dev`.

## 14. Parte 9 — Operability

Capitoli inclusi:

- Health check;
- Maintenance operations;
- Troubleshooting;
- Backup, restore e disaster recovery.

Elementi obbligatori:

- `/readyz`;
- dashboard HTTP;
- Argo CD matrix;
- deployment readiness matrix;
- route health matrix;
- Tekton validation matrix;
- evidence directory;
- incident triage;
- stop conditions;
- PostgreSQL backup e restore;
- DR limits.

## 15. Parte 10 — Stato corrente e roadmap

Capitoli inclusi:

- Stato delle fasi;
- Stato finale corrente;
- Roadmap futura.

Stato fasi da rappresentare:

- Fase 0: parziale e continua;
- Fasi 1-9: completate;
- Fase 10: completata come baseline operativa avanzata riallineata post-Fase 15;
- Fase 11: standby;
- Fase 12: in corso;
- Fase 13: completata e riallineata post-Fase 15;
- Fase 14: completata;
- Fase 15: completata come multi-cluster code-ready baseline.

## 16. Appendici

Appendici previste:

- Glossario;
- Comandi operativi principali;
- Commit e tag rilevanti;
- Limitazioni note.

Il glossario dovrà spiegare i termini tecnici in italiano.

I comandi dovranno restare in inglese o nel formato originale.

Commit e tag da includere almeno:

- `namespace-isolated-baseline-20260709`;
- `af6ddb3`;
- `052c446`;
- `9b72931`;
- `215a790`;
- `e1e81d1`;
- `b6c7c61`;
- `bdae466`;
- `24ca185`.

## 17. Regole per esempi e comandi

I comandi devono essere riportati solo quando servono davvero.

I comandi devono essere:

- copiabili;
- chiari;
- non distruttivi, salvo esplicita indicazione;
- accompagnati da output atteso;
- privi di Secret o token;
- coerenti con i runbook.

Esempi di comandi ammessi:

`git status --short`

`go test ./internal/api ./internal/app ./cmd/devops-control-plane`

`oc get pod -n devops-control-plane`

`oc get application demo-go-color-app-staging -n openshift-gitops`

## 18. Regole per diagrammi testuali

La guida finale dovrà includere diagrammi testuali semplici.

Esempi da includere:

- architettura generale;
- workflow ChangeRequest;
- GitOps flow;
- Tekton validation flow;
- evidence flow;
- namespace-isolated topology;
- future multi-cluster topology.

I diagrammi devono essere leggibili anche in Word.

## 19. Regole per tabelle

La guida finale dovrà includere tabelle per:

- mapping ambienti;
- mapping namespace;
- mapping fonti;
- stato fasi;
- evidence fields;
- runbook checks;
- guardrail;
- limitazioni note.

Le tabelle devono essere brevi e leggibili.

## 20. Criteri di revisione

Ogni parte dovrà essere revisionata verificando:

- accuratezza tecnica;
- coerenza con repository;
- coerenza con stato Fase 10, Fase 13 e Fase 15;
- chiarezza per lettori non esperti;
- assenza di Secret;
- assenza di claim fisici multi-cluster non validati;
- presenza di limiti noti;
- presenza di esempi concreti.

## 21. Strategia di commit

Il documento finale sarà scritto in modo incrementale.

Commit suggeriti:

- `Start final technical guide with introduction and concepts`
- `Add architecture chapters to final technical guide`
- `Add workflow and evidence chapters to final technical guide`
- `Add UI and multi-cluster readiness chapters to final technical guide`
- `Add security and operability chapters to final technical guide`
- `Complete final technical guide appendices`
- `Generate final technical guide Word document`

## 22. Definition of Done per 12.3

Questa fase è completata quando:

- il piano di scrittura esiste;
- il piano definisce ordine di scrittura;
- il piano definisce livello di dettaglio;
- il piano definisce esempi obbligatori;
- il piano definisce regole per comandi e diagrammi;
- il piano definisce criteri di revisione;
- il piano è committato nel repository.

<!-- CI_GUIDE_WRITING_PLAN_START -->
## Integrazione della baseline CI nella guida finale

### Obiettivo

Integrare nella guida tecnica italiana la baseline Continuous Integration e automated testing usando esclusivamente il documento autorevole:

```text
docs/continuous-integration-and-automated-testing.md
```

La guida non deve essere aggiornata direttamente dal log Git, dai file Go o dal workflow YAML senza passare dalla fonte Markdown approvata.

### Ordine di lavoro

1. aggiornare `docs/README.md` con il collegamento alla fonte CI;
2. aggiornare `docs/final-technical-guide/source-map.md`;
3. aggiornare `docs/final-technical-guide/outline.md`;
4. aggiornare `docs/final-technical-guide/writing-plan.md`;
5. aggiungere il capitolo italiano in `docs/final-technical-guide/final-technical-guide-it.md`;
6. riallineare la numerazione dei capitoli successivi;
7. aggiornare le sezioni Stato delle fasi, Stato finale corrente e Roadmap futura solo dove giustificato dalla fonte CI;
8. aggiornare Appendice C con PR e commit CI rilevanti;
9. eseguire i controlli strutturali e di secret scanning;
10. rigenerare il documento Word derivato;
11. aggiornare report di generazione e README finale.

### Contenuti da derivare dalla fonte CI

Il capitolo italiano deve coprire:

- scopo e perimetro della CI;
- workflow GitHub Actions;
- modello di trigger;
- formatting gate;
- static analysis con `go vet`;
- unit test, race detector e coverage;
- integration test PostgreSQL;
- test HTTP end-to-end;
- invariante di concorrenza lifecycle;
- invarianti TLS secure-by-default;
- modello di validazione tramite pull request;
- validazione locale degli sviluppatori;
- considerazioni di sicurezza;
- limitazioni correnti;
- evidenze e Definition of Done.

### Regole di scrittura

- scrivere il capitolo finale in italiano;
- mantenere comandi, nomi file, variabili, funzioni di test e identificativi tecnici nel formato originale;
- non presentare le credenziali disposable PostgreSQL della CI come credenziali runtime;
- non dichiarare che una pipeline verde valida il runtime OpenShift;
- non dichiarare test browser/UI se non esistono;
- non dichiarare pubblicazione automatica di immagini container;
- non dichiarare una coverage threshold minima finché non è configurata;
- non presentare la CI come validazione fisica multi-cluster;
- mantenere il Markdown italiano come source of truth del Word.

### Criteri di qualità

Prima della rigenerazione Word verificare:

- capitolo CI presente e coerente con la fonte autorevole;
- source map e numerazione allineate;
- riferimenti a `.github/workflows/ci.yml` corretti;
- comandi CI identici alla fonte;
- limitazioni CI esplicitate;
- nessun token, Secret, kubeconfig o credenziale reale;
- controlli strutturali PASS;
- working tree pulito prima della generazione finale.

### Definition of Done

L'integrazione CI nella guida finale è completata quando:

- outline e writing plan sono aggiornati;
- source map include la fonte CI;
- il capitolo italiano è presente;
- la numerazione è sequenziale;
- Stato corrente e appendici sono coerenti;
- i controlli strutturali non producono errori o warning;
- il Word è rigenerato dalla nuova sorgente italiana;
- README e report finali indicano il nuovo output.
<!-- CI_GUIDE_WRITING_PLAN_END -->
