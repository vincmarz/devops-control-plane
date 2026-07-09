# DevOps Control Plane — Mappa fonti per la guida tecnica finale

Status: Draft  
Phase: 12.2 — Inventory source documents for final technical guide  
Owner: Vincenzo Marzario  
Language: Italian  
Last updated: 2026-07-09

## 1. Scopo

Questo documento mappa i capitoli previsti nella guida tecnica finale del DevOps Control Plane con i documenti sorgente già presenti nel repository.

La guida finale dovrà essere scritta in italiano. I comandi, i nomi di file, le API, le risorse Kubernetes/OpenShift, i nomi Tekton, i nomi Argo CD e i riferimenti Git resteranno nel formato originale.

La mappa serve a evitare duplicazioni, lacune e incoerenze durante la scrittura del documento finale.

## 2. Documento guida di riferimento

Outline principale:

- `docs/final-technical-guide-outline.md`

Documento sorgente finale previsto:

- `docs/final-technical-guide.md`

Output Word previsto:

- `DevOps_Control_Plane_Guida_Tecnica_Finale.docx`

## 3. Fonti principali trasversali

Le seguenti fonti saranno usate in più parti del documento finale:

- `docs/00-vision.md`
- `docs/01-scope-mvp.md`
- `docs/03-functional-requirements.md`
- `docs/04-non-functional-requirements.md`
- `docs/05-architecture.md`
- `docs/11-change-workflows.md`
- `docs/12-evidence-model.md`
- `docs/13-api-design.md`
- `docs/environment-configuration-model.md`
- `docs/multi-cluster-environment-enablement-request.md`
- `docs/runtime-evidence-dashboard-maintenance-alignment.md`
- `docs/phase-10-operability-closure.md`

## 4. Parte 1 — Introduzione e visione

### Capitolo 1 — Executive Summary

Fonti principali:

- `docs/00-vision.md`
- `docs/01-scope-mvp.md`
- `docs/phase-10-operability-closure.md`
- `docs/multi-cluster-environment-enablement-request.md`
- `docs/runtime-evidence-dashboard-maintenance-alignment.md`

Uso previsto:

- spiegare che cos’è il DevOps Control Plane;
- spiegare lo stato corrente;
- chiarire cosa è validato;
- chiarire cosa è rinviato;
- introdurre baseline namespace-isolated e multi-cluster code-ready baseline.

### Capitolo 2 — Visione del progetto

Fonti principali:

- `docs/00-vision.md`
- `docs/05-architecture.md`
- `docs/03-functional-requirements.md`

Uso previsto:

- descrivere il problema da risolvere;
- descrivere la centralizzazione del ciclo ChangeRequest;
- introdurre GitOps, audit, evidence e workflow controllato.

### Capitolo 3 — Scope del progetto

Fonti principali:

- `docs/01-scope-mvp.md`
- `docs/03-functional-requirements.md`
- `docs/04-non-functional-requirements.md`

Uso previsto:

- distinguere MVP evoluto, baseline operativa e roadmap futura;
- chiarire cosa è incluso e cosa è escluso;
- esplicitare i limiti noti.

## 5. Parte 2 — Concetti fondamentali

### Capitolo 4 — Kubernetes e OpenShift in breve

Fonti principali:

- `docs/05-architecture.md`
- `docs/04-non-functional-requirements.md`
- `docs/runbooks/operability-health-check.md`

Uso previsto:

- spiegare cluster, namespace, pod, deployment, service, route, secret, configmap, RBAC;
- usare esempi dal progetto.

### Capitolo 5 — Namespace isolation

Fonti principali:

- `docs/environment-configuration-model.md`
- `docs/multi-environment-model.md`
- `docs/multi-cluster-environment-enablement-request.md`

Uso previsto:

- spiegare la topologia corrente:
  - `dev` -> `ocp-dev` / `devops-ci-demo`
  - `staging` -> `ocp-dev` / `devops-ci-staging`
  - `production` -> `ocp-dev` / `devops-ci-production`
- chiarire differenza tra ambiente logico e cluster fisico;
- spiegare perché questa baseline è validata.

### Capitolo 6 — GitOps

Fonti principali:

- `docs/00-vision.md`
- `docs/06-argocd-integration.md`
- `docs/11-change-workflows.md`
- `docs/adr/ADR-0001-git-source-of-truth.md`
- `docs/adr/ADR-0002-argocd-as-gitops-engine.md`

Uso previsto:

- spiegare Git come fonte di verità;
- spiegare il ruolo dei repository GitOps;
- collegare GitOps a ChangeRequest, Argo CD e audit.

### Capitolo 7 — Kustomize base e overlays

Fonti principali:

- `docs/06-argocd-integration.md`
- `docs/11-change-workflows.md`
- `docs/multi-cluster-environment-enablement-request.md`

Uso previsto:

- spiegare base comune e overlay per ambiente;
- spiegare i path:
  - `apps/demo-go-color-app/overlays/staging`
  - `apps/demo-go-color-app/overlays/production`
- collegare gli overlay alla validazione Tekton.

### Capitolo 8 — Argo CD

Fonti principali:

- `docs/06-argocd-integration.md`
- `docs/05-architecture.md`
- `docs/runbooks/operability-health-check.md`
- `docs/runbooks/maintenance-operations.md`

Uso previsto:

- spiegare Application, sync, health e target namespace;
- descrivere Applications dev, staging e production;
- spiegare stato `Synced` e `Healthy`.

### Capitolo 9 — Tekton

Fonti principali:

- `docs/08-tekton-integration.md`
- `docs/11-change-workflows.md`
- `docs/12-evidence-model.md`
- `docs/runbooks/maintenance-operations.md`

Uso previsto:

- spiegare Task, Pipeline, PipelineRun e TaskRun;
- spiegare validazione GitOps;
- descrivere `validate` e `check-validation`.

## 6. Parte 3 — Architettura

### Capitolo 10 — Architettura generale

Fonti principali:

- `docs/05-architecture.md`
- `docs/adr/ADR-0006-adapter-based-architecture.md`
- `docs/13-api-design.md`

Uso previsto:

- descrivere backend Go, PostgreSQL, GitLab API, Argo CD API, Kubernetes API, Tekton API e UI;
- spiegare adapter pattern;
- spiegare separazione tra applicazione, domain e adapter.

### Capitolo 11 — Backend Go

Fonti principali:

- `docs/05-architecture.md`
- `docs/13-api-design.md`
- `docs/11-change-workflows.md`

Uso previsto:

- descrivere ChangeService;
- descrivere repository layer;
- descrivere service options;
- descrivere runtime target resolver e provider registry.

### Capitolo 12 — PostgreSQL e persistenza

Fonti principali:

- `docs/postgresql-integration-notes.md`
- `docs/phase-1-postgresql-change-repository.md`
- `docs/10-data-model.md`
- `docs/runbooks/postgresql-backup-restore.md`

Uso previsto:

- spiegare ChangeRequest, ChangeEvent ed Evidence;
- spiegare audit trail;
- collegare persistenza a backup e restore.

### Capitolo 13 — Modello dati

Fonti principali:

- `docs/10-data-model.md`
- `docs/12-evidence-model.md`
- `docs/11-change-workflows.md`

Uso previsto:

- descrivere entità principali;
- descrivere stati lifecycle e runtime;
- descrivere relazione tra ChangeRequest, Eventi ed Evidence.

## 7. Parte 4 — Workflow applicativi

### Capitolo 14 — ChangeRequest lifecycle

Fonti principali:

- `docs/11-change-workflows.md`
- `docs/10-data-model.md`
- `docs/13-api-design.md`

Uso previsto:

- spiegare creazione, validazione, eventi, transizioni e audit.

### Capitolo 15 — GitLab Merge Request workflow

Fonti principali:

- `docs/07-gitlab-integration.md`
- `docs/11-change-workflows.md`
- `docs/adr/ADR-0007-gitlab-api-as-git-provider.md`

Uso previsto:

- spiegare branch, commit, MR, merge e relazione con GitOps.

### Capitolo 16 — Workflow runtime

Fonti principali:

- `docs/11-change-workflows.md`
- `docs/12-evidence-model.md`
- `docs/13-api-design.md`

Uso previsto:

- spiegare `create`;
- spiegare `collect-evidence`;
- spiegare `check-deployment`;
- spiegare `validate`;
- spiegare `check-validation`;
- collegare workflow e UI.

### Capitolo 17 — Workflow dev, staging e production

Fonti principali:

- `docs/multi-cluster-environment-enablement-request.md`
- `docs/11-change-workflows.md`
- `docs/runtime-evidence-dashboard-maintenance-alignment.md`

Uso previsto:

- descrivere workflow namespace-isolated;
- descrivere `CHG-2026-0049`;
- descrivere `CHG-2026-0050`;
- spiegare differenza tra baseline validata e multi-cluster fisico deferito.

## 8. Parte 5 — Evidence model

### Capitolo 18 — Runtime evidence

Fonti principali:

- `docs/12-evidence-model.md`
- `docs/runbooks/operability-health-check.md`
- `docs/05-architecture.md`

Uso previsto:

- descrivere stato osservato del runtime;
- descrivere deployment, pod, route, namespace e Argo CD.

### Capitolo 19 — Tekton validation evidence

Fonti principali:

- `docs/12-evidence-model.md`
- `docs/11-change-workflows.md`
- `docs/08-tekton-integration.md`

Uso previsto:

- descrivere PipelineRun, validation path, failed task count e sanitized evidence.

### Capitolo 20 — Argo CD deployment evidence

Fonti principali:

- `docs/06-argocd-integration.md`
- `docs/12-evidence-model.md`
- `docs/runbooks/operability-health-check.md`

Uso previsto:

- spiegare sync, health, revision, namespace e overlay.

### Capitolo 21 — Evidence sanitization

Fonti principali:

- `docs/12-evidence-model.md`
- `docs/09-security-rbac.md`
- `docs/runbooks/operability-health-check.md`

Uso previsto:

- spiegare dati ammessi e dati vietati;
- ribadire no raw Secret, no token, no kubeconfig.

## 9. Parte 6 — UI e dashboard

### Capitolo 22 — Dashboard

Fonti principali:

- `docs/05-architecture.md`
- `docs/13-api-design.md`
- `docs/runtime-evidence-dashboard-maintenance-alignment.md`

Uso previsto:

- spiegare latest ChangeRequest;
- spiegare Environments / Namespaces;
- spiegare KPI e cards.

### Capitolo 23 — ChangeRequest detail

Fonti principali:

- `docs/12-evidence-model.md`
- `docs/11-change-workflows.md`
- `docs/13-api-design.md`

Uso previsto:

- spiegare runtime evidence card;
- spiegare Tekton validation card;
- spiegare audit log e raw sanitized evidence.

### Capitolo 24 — UI environment awareness

Fonti principali:

- `docs/05-architecture.md`
- `docs/environment-configuration-model.md`
- `docs/runtime-evidence-dashboard-maintenance-alignment.md`

Uso previsto:

- spiegare visibilità di dev, staging e production;
- spiegare perché la UI non deve nascondere staging e production dietro dev.

## 10. Parte 7 — Environment Catalog e multi-cluster readiness

### Capitolo 25 — Environment Catalog

Fonti principali:

- `docs/environment-configuration-model.md`
- `docs/multi-environment-model.md`
- `docs/environment-catalog-ui-action-results.md`

Uso previsto:

- descrivere ambiente logico, namespace, Tekton namespace, Argo CD Application e validation path.

### Capitolo 26 — Cluster Registry

Fonti principali:

- `docs/cluster-registry-baseline-results.md`
- `docs/environment-cluster-resolver-results.md`
- `docs/multi-cluster-environment-enablement-request.md`

Uso previsto:

- descrivere cluster definition, enabled flag, allowed namespaces e Secret references.

### Capitolo 27 — Runtime target resolution

Fonti principali:

- `docs/technical-runtime-target-preparation-results.md`
- `docs/kubernetes-runtime-provider-wiring-results.md`
- `docs/tekton-runtime-provider-wiring-results.md`
- `docs/argocd-runtime-provider-wiring-results.md`

Uso previsto:

- spiegare EnvironmentClusterResolver;
- spiegare TechnicalRuntimeTarget;
- spiegare provider selection.

### Capitolo 28 — Multi-cluster code-ready baseline

Fonti principali:

- `docs/multi-cluster-environment-enablement-request.md`
- `docs/factory-aware-runtime-provider-main-wiring-results.md`
- `docs/controlled-enablement-plumbing-validation-results.md`

Uso previsto:

- spiegare code-ready;
- spiegare physical validation deferred;
- spiegare staging e production simulated;
- spiegare no fallback to ocp-dev.

### Capitolo 29 — Deferred real-cluster onboarding contract

Fonti principali:

- `docs/multi-cluster-environment-enablement-request.md`
- `docs/single-non-production-multi-cluster-enablement-plan.md`

Uso previsto:

- spiegare condizioni per onboarding futuro;
- spiegare input richiesti;
- spiegare RBAC, Secret references, readiness gates e rollback.

## 11. Parte 8 — Security e guardrail

### Capitolo 30 — RBAC

Fonti principali:

- `docs/09-security-rbac.md`
- `docs/runbooks/operability-health-check.md`
- `docs/runbooks/maintenance-operations.md`

Uso previsto:

- spiegare principio minimo privilegio;
- spiegare namespace-scoped RBAC;
- spiegare cosa evitare.

### Capitolo 31 — Secret reference model

Fonti principali:

- `docs/runtime-client-secret-config-model.md`
- `docs/runtime-client-secret-reference-loading.md`
- `docs/controlled-runtime-secret-loader-and-factory-enablement-design.md`

Uso previsto:

- spiegare Secret reference;
- spiegare allow-list;
- spiegare loader disabled-by-default;
- spiegare no raw Secret.

### Capitolo 32 — Runtime factories

Fonti principali:

- `docs/real-runtime-client-factory-implementation-design.md`
- `docs/real-runtime-client-factory-implementation-readiness.md`
- `docs/runtime-non-regression-factories-disabled-results.md`
- `docs/controlled-enablement-plumbing-validation-results.md`

Uso previsto:

- spiegare Kubernetes, Tekton e Argo CD factory;
- spiegare disabled-by-default;
- spiegare fail-closed;
- spiegare unsupported kubeconfig e raw CA.

### Capitolo 33 — AuthN/AuthZ e OAuth proxy

Fonti principali:

- `docs/adr/ADR-0009-authn-authz-strategy.md`
- `docs/adr/ADR-0010-oauth-proxy-deployment-design.md`
- `docs/runbooks/authn-authz.md`
- `docs/runbooks/oauth-proxy.md`

Uso previsto:

- spiegare proxy;
- spiegare forwarded headers;
- spiegare gruppi e autorizzazioni;
- spiegare fail-closed.

## 12. Parte 9 — Operability

### Capitolo 34 — Health check

Fonti principali:

- `docs/runbooks/operability-health-check.md`
- `docs/phase-10-operability-closure.md`

Uso previsto:

- spiegare health-check matrix post-Fase 15.

### Capitolo 35 — Maintenance operations

Fonti principali:

- `docs/runbooks/maintenance-operations.md`

Uso previsto:

- spiegare pre-maintenance, post-maintenance, smoke matrix e guardrail.

### Capitolo 36 — Troubleshooting

Fonti principali:

- `docs/runbooks/operability-health-check.md`
- `docs/runbooks/maintenance-operations.md`

Uso previsto:

- spiegare incident triage;
- spiegare evidence package;
- spiegare stop conditions.

### Capitolo 37 — Backup, restore e disaster recovery

Fonti principali:

- `docs/runbooks/postgresql-backup-restore.md`
- `docs/runbooks/disaster-recovery.md`
- `docs/phase-10-operability-closure.md`

Uso previsto:

- spiegare backup PostgreSQL;
- spiegare restore isolato;
- spiegare RPO/RTO;
- spiegare limiti GitLab, Argo CD e Tekton.

## 13. Parte 10 — Stato corrente e roadmap

### Capitolo 38 — Stato delle fasi

Fonti principali:

- `docs/phase-10-operability-closure.md`
- `docs/runtime-evidence-dashboard-maintenance-alignment.md`
- `docs/multi-cluster-environment-enablement-request.md`
- `docs/final-technical-guide-outline.md`

Uso previsto:

- riassumere stato fasi;
- evidenziare Fase 10, Fase 13 e Fase 15;
- esplicitare Fase 11 standby e Fase 12 in corso.

### Capitolo 39 — Stato finale corrente

Fonti principali:

- `docs/multi-cluster-environment-enablement-request.md`
- `docs/runbooks/operability-health-check.md`
- `docs/12-evidence-model.md`

Uso previsto:

- sintetizzare cosa è completato;
- sintetizzare cosa è rinviato;
- sintetizzare cosa è pronto a livello codice.

### Capitolo 40 — Roadmap futura

Fonti principali:

- `docs/final-technical-guide-outline.md`
- `docs/multi-cluster-environment-enablement-request.md`
- `docs/production-readiness-checklist.md`

Uso previsto:

- descrivere documento finale;
- descrivere real-cluster onboarding futuro;
- descrivere eventuale CLI;
- descrivere observability e produzione reale.

## 14. Appendici

### Appendice A — Glossario

Fonti principali:

- tutti i documenti principali;
- in particolare `docs/05-architecture.md`, `docs/11-change-workflows.md`, `docs/12-evidence-model.md`.

### Appendice B — Comandi operativi principali

Fonti principali:

- `docs/runbooks/operability-health-check.md`
- `docs/runbooks/maintenance-operations.md`
- `docs/runbooks/postgresql-backup-restore.md`

### Appendice C — Commit e tag rilevanti

Fonti principali:

- cronologia Git;
- `docs/multi-cluster-environment-enablement-request.md`;
- `docs/runtime-evidence-dashboard-maintenance-alignment.md`.

Commit e tag da includere:

- `namespace-isolated-baseline-20260709`
- `af6ddb3`
- `052c446`
- `9b72931`
- `215a790`
- `e1e81d1`
- `b6c7c61`

### Appendice D — Limitazioni note

Fonti principali:

- `docs/multi-cluster-environment-enablement-request.md`
- `docs/production-readiness-checklist.md`
- `docs/phase-10-operability-closure.md`

Limitazioni da includere:

- physical multi-cluster validation deferred;
- no additional OpenShift cluster available;
- no real production cluster onboarding;
- runtime Secret loader disabled by default;
- runtime factories disabled by default;
- CLI in standby.

## 15. Priorità di scrittura

La guida finale sarà scritta in questo ordine:

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

## 16. Regole di riuso delle fonti

Durante la scrittura del documento finale:

- non copiare intere sezioni senza adattarle al lettore finale;
- trasformare risultati e runbook in spiegazioni narrative;
- mantenere comandi e nomi tecnici nel formato originale;
- evitare contraddizioni tra documenti;
- dichiarare sempre cosa è validato e cosa è deferred;
- usare esempi concreti tratti dalle validazioni completate;
- evitare di presentare la simulazione come validazione fisica reale;
- non dichiarare produzione reale se non esiste un cluster production fisico.

## 17. Definition of Done per 12.2

Questa fase è completata quando:

- la mappa fonti esiste;
- ogni parte della guida finale ha documenti sorgente associati;
- le fonti operative, architetturali e di evidence sono mappate;
- le limitazioni note sono mappate;
- la mappa è committata nel repository.
