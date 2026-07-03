# DevOps Control Plane — Italian Documentation Inventory

## Document metadata

- **Project:** DevOps Control Plane
- **Phase:** 0.2 — Italian documentation inventory
- **Status:** Generated inventory baseline for documentation migration
- **Scope:** Markdown files under `docs/`
- **Policy reference:** `docs/documentation-language-policy.md`

---

## 1. Purpose

This inventory identifies documentation files that may contain Italian wording or mixed-language terminology.
The inventory is heuristic and must be reviewed by maintainers before migration work starts.

The official repository language is English. Existing Italian documentation should be migrated incrementally according to the repository language policy.

---

## 2. Summary

```text
Markdown files scanned: 44
Files with potential Italian content: 26
Potential Italian term matches: 850
```

---

## 3. Migration priority

Priority levels:

```text
P1 = Core documentation or ADR content; migrate first
P2 = Operational or index documentation; review after P1
P3 = Supporting or historical documentation; migrate or supersede later
```

| Priority | File | Matches | Recommendation |
|---|---|---:|---|
| P1 | `docs/00-vision.md` | 63 | Translate and align early vision content to English before final technical documentation. |
| P1 | `docs/01-scope-mvp.md` | 86 | Translate and normalize MVP scope terminology to English. |
| P1 | `docs/05-architecture.md` | 37 | Translate architecture narrative and align terminology with ADRs and runbooks. |
| P1 | `docs/adr/ADR-0001-git-source-of-truth.md` | 15 | ADR content and index must remain English; fix any residual mixed-language wording. |
| P1 | `docs/adr/ADR-0002-argocd-as-gitops-engine.md` | 5 | ADR content and index must remain English; fix any residual mixed-language wording. |
| P1 | `docs/adr/ADR-0003-tekton-validation-engine.md` | 8 | ADR content and index must remain English; fix any residual mixed-language wording. |
| P1 | `docs/adr/ADR-0004-postgresql-change-history.md` | 6 | ADR content and index must remain English; fix any residual mixed-language wording. |
| P1 | `docs/adr/ADR-0005-api-first-before-web-ui.md` | 3 | ADR content and index must remain English; fix any residual mixed-language wording. |
| P1 | `docs/adr/ADR-0006-adapter-based-architecture.md` | 3 | ADR content and index must remain English; fix any residual mixed-language wording. |
| P1 | `docs/adr/ADR-0007-gitlab-api-as-git-provider.md` | 5 | ADR content and index must remain English; fix any residual mixed-language wording. |
| P1 | `docs/adr/ADR-0008-kubernetes-api-for-tekton.md` | 1 | ADR content and index must remain English; fix any residual mixed-language wording. |
| P1 | `docs/adr/README.md` | 4 | ADR content and index must remain English; fix any residual mixed-language wording. |
| P3 | `docs/02-personas-use-cases.md` | 96 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/03-functional-requirements.md` | 43 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/04-non-functional-requirements.md` | 85 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/06-argocd-integration.md` | 66 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/07-gitlab-integration.md` | 51 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/08-tekton-integration.md` | 52 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/09-security-rbac.md` | 51 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/10-data-model.md` | 40 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/11-change-workflows.md` | 43 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/12-evidence-model.md` | 48 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/13-api-design.md` | 22 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/documentation-language-policy.md` | 10 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/phase-1-postgresql-change-repository.md` | 2 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |
| P3 | `docs/postgresql-integration-notes.md` | 5 | Review and translate if content is still relevant; otherwise mark as historical or superseded. |

---

## 4. Detailed findings

### `docs/00-vision.md`

- **Priority:** P1
- **Potential matches:** 63

| Line | Term | Snippet |
|---:|---|---|
| 13 | `ambienti` | Negli ambienti OpenShift moderni, un flusso DevOps/GitOps maturo coinvolge più componenti specializzati: |
| 15 | `stato` | - **Git** come sorgente dello stato desiderato; |
| 17 | `validazione` | - **Tekton / OpenShift Pipelines** come motore di build, validazione e automazione; |
| 24 | `modifica` | 2. modifica YAML; |
| 25 | `validazione` | 3. validazione locale; |
| 28 | `validazione` | 6. validazione Tekton; |
| 32 | `evidenze` | 10. raccolta evidenze; |
| 47 | `modifica` | - modifica manuale di file YAML; |
| 48 | `evidenze` | - raccolta manuale di evidenze. |
| 50 | `Questo` | Questo approccio funziona, ma presenta alcuni rischi: |
| 57 | `evidenze` | - evidenze raccolte in modo non uniforme; |
| 65 | `Obiettivo` | ## 3. Obiettivo del progetto |
| 67 | `obiettivo` | L’obiettivo del progetto è realizzare un’applicazione interna chiamata **DevOps Control Plane** che permetta di: |
| 70 | `stato` | - mostrare stato di sync, health, revisione corrente e orphaned resources; |
| 71 | `stato` | - leggere l’ultimo commit Git associato allo stato desiderato; |
| 77 | `stato` | - attendere lo stato `Synced` e `Healthy`; |
| 78 | `evidenze` | - raccogliere evidenze tecniche; |
| 85 | `stato` | ### 4.1 Git resta la sorgente dello stato desiderato |
| 87 | `stato` | Il DevOps Control Plane non deve sostituire Git come fonte dello stato desiderato. |
| 89 | `modifica` | Un change permanente deve essere rappresentato da una modifica Git: |
| 97 | `stato` | Il portale non deve applicare modifiche runtime come stato finale tramite comandi imperativi quali: |
| 114 | `responsabile` | Argo CD rimane responsabile di: |
| 116 | `stato` | - leggere lo stato desiderato da Git; |
| 117 | `stato` | - confrontarlo con lo stato runtime del cluster; |
| 119 | `stato` | - mostrare stato `Synced` / `OutOfSync`; |
| 120 | `stato` | - mostrare stato `Healthy` / `Degraded`; |
| 121 | `operativo` | - gestire history, sync e rollback operativo. |
| 127 | `evidenze` | ### 4.3 Tekton valida e produce evidenze |
| 129 | `validazione` | Tekton sarà usato come motore di validazione e automazione. |
| 139 | `evidenze` | - raccolta log ed evidenze. |
| 150 | `stato` | - perché è stato fatto? |
| 151 | `stato` | - quale workflow è stato eseguito? |
| 152 | `validazione` | - quale validazione Tekton è stata associata? |
| 153 | `evidenza` | - quale evidenza è stata raccolta? |
| 156 | `questo` | Per questo il DevOps Control Plane userà **PostgreSQL** come database interno per Change Request, eventi, evidenze e stato workflow. |
| 164 | `operativa` | Il supporto a **HTML templates + Bootstrap** sarà previsto fin dall’inizio, ma la Web UI operativa arriverà dopo, quando i workflow saranno chiari e validati. |
| 166 | `Questo` | Questo evita di costruire schermate sopra processi ancora instabili. |
| 208 | `modifica` | - modifica numero repliche; |
| 209 | `modifica` | - modifica `APP_VERSION`; |
| 210 | `modifica` | - modifica `PAGE_COLOR`; |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 63. |

### `docs/01-scope-mvp.md`

- **Priority:** P1
- **Potential matches:** 86

| Line | Term | Snippet |
|---:|---|---|
| 14 | `Questo` | Questo documento definisce il perimetro del primo MVP del progetto **DevOps Control Plane**. |
| 16 | `obiettivo` | L’obiettivo è stabilire in modo chiaro: |
| 25 | `operativo` | Il documento deve essere usato come riferimento operativo durante la progettazione e l’implementazione del backend, delle API, degli adapter e dei primi workflow GitOps. |
| 34 | `stato` | 2. consultare stato GitOps e runtime essenziale; |
| 37 | `modifica` | 5. modificare manifest GitOps in modo controllato; |
| 42 | `evidenze` | 10. raccogliere evidenze; |
| 59 | `responsabile` | Il backend sarà responsabile di: |
| 63 | `integrazione` | - integrazione con GitLab API; |
| 64 | `integrazione` | - integrazione con Argo CD API; |
| 65 | `integrazione` | - integrazione con Kubernetes API per Tekton; |
| 67 | `evidenze` | - raccolta evidenze; |
| 104 | `evidenze` | - evidenze; |
| 105 | `stato` | - stato finale del change. |
| 123 | `stato` | - leggere stato merge request. |
| 125 | `integrazione` | Nota: il repository sorgente del DevOps Control Plane può essere ospitato su GitHub nell’account `vincmarz`, ma l’applicazione implementerà l’integrazione target tramite GitLab ... |
| 146 | `stato` | - monitorare stato fino a `Synced` e `Healthy`. |
| 158 | `creazione` | - creazione `PipelineRun`; |
| 159 | `stato` | - osservazione stato `PipelineRun`; |
| 160 | `stato` | - osservazione stato `TaskRun`; |
| 162 | `validazione` | - raccolta esito validazione. |
| 172 | `ambiente` | Il sistema deve visualizzare le applicazioni Argo CD disponibili per il progetto/ambiente configurato. |
| 214 | `operativo` | Il sistema deve mostrare il dettaglio operativo di una Application Argo CD. |
| 238 | `stato` | - lo stato risulta coerente con quanto mostrato da Argo CD. |
| 246 | `modifica` | Il sistema deve consentire un change GitOps per modificare il numero di repliche di un Deployment. |
| 248 | `utente` | ### Input utente |
| 260 | `modifica` | 4. modificare `spec.replicas`; |
| 263 | `validazione` | 7. lanciare validazione Tekton; |
| 265 | `evidenze` | 9. raccogliere evidenze. |
| 271 | `modifica` | - il sistema produce una modifica Git tracciata; |
| 272 | `modifica` | - il sistema non modifica direttamente il Deployment runtime; |
| 274 | `stato` | - la ChangeRequest viene salvata in PostgreSQL con stato finale `Completed` o `Failed`. |
| 302 | `modifica` | - il sistema modifica il file corretto; |
| 332 | `validazione` | - il change passa la validazione Tekton; |
| 333 | `modifica` | - Argo CD applica la modifica; |
| 342 | `modifica` | Il sistema deve permettere di modificare chiavi/valori in una ConfigMap GitOps. |
| 361 | `modifica` | La modifica di una ConfigMap non sempre causa automaticamente rollout dei Pod. |
| 363 | `ambiente` | Se la ConfigMap è usata come variabile ambiente, i Pod devono essere ricreati per acquisire i nuovi valori. Nel futuro sarà necessario decidere una strategia, per esempio: |
| 370 | `Questa` | Questa decisione sarà formalizzata in un ADR dedicato. |
| 378 | `validazione` | Il sistema deve delegare a Tekton la validazione automatica del change prima della promozione. |
| 391 | `modifica` | - controllo presenza file modificati; |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 86. |

### `docs/05-architecture.md`

- **Priority:** P1
- **Potential matches:** 37

| Line | Term | Snippet |
|---:|---|---|
| 19 | `Questo` | Questo documento descrive l’architettura iniziale del progetto **DevOps Control Plane**. |
| 21 | `obiettivo` | L’obiettivo è definire: |
| 29 | `integrazione` | - modalità di integrazione con GitLab, Argo CD, Tekton, Kubernetes/OpenShift e PostgreSQL; |
| 30 | `sicurezza` | - principi di sicurezza e osservabilità architetturale; |
| 46 | `validazione` | Tekton       = motore di validazione workflow |
| 48 | `evidenze` | PostgreSQL   = storico ChangeRequest, eventi ed evidenze |
| 52 | `architettura` | Il principio architetturale fondamentale è: |
| 58 | `modifica` | DevOps Control Plane non deve diventare un pannello alternativo di modifica runtime del cluster. |
| 62 | `architettura` | ## 3. Stack architetturale scelto |
| 75 | `integrazione` | - integrazione adapter; |
| 77 | `evidenze` | - raccolta evidenze; |
| 120 | `operativo` | - audit trail operativo. |
| 138 | `stato` | - leggere stato merge request. |
| 140 | `integrazione` | Nota: il codice sorgente del progetto può essere ospitato su GitHub, ma l’integrazione funzionale MVP verso repository applicativi/GitOps target è GitLab API. |
| 158 | `stato` | - stato operazione; |
| 175 | `evidenze` | - salvare evidenze validazione. |
| 193 | `evidenze` | - raccogliere evidenze runtime; |
| 248 | `utenti` | - autenticare/autorizzare, quando implementato; |
| 325 | `stato` | - ogni fallimento produce stato esplicito; |
| 364 | `stato` | - aggiornare stato; |
| 369 | `evidenze` | - associare evidenze. |
| 382 | `evidenze` | Raccoglie e normalizza evidenze. |
| 384 | `evidenza` | Tipi di evidenza iniziali: |
| 462 | `stato` | - stato OutOfSync non sempre equivale a errore applicativo; |
| 487 | `evidenze` | - read-only per evidenze runtime, salvo creazione PipelineRun Tekton; |
| 488 | `modifica` | - non modificare workload applicativi come stato finale. |
| 513 | `stato` | - stato finale associato a ChangeRequest. |
| 602 | `architettura` | ## 8. Flussi architetturali principali |
| 651 | `operativo` | Esempio: rollback operativo Argo CD può rendere il cluster temporaneamente diverso da `main`. |
| 674 | `Regola` | ### Regola fondamentale |
| 879 | `evidenze` | - scrittura evidenze. |
| 912 | `evidenze` | - get/list/watch Pod/log per evidenze Tekton, dove necessario; |
| 1027 | `architettura` | ## 14. Decisioni architetturali da registrare in ADR |
| 1209 | `documentazione` | - il repository contiene documentazione base; |
| 1223 | `architettura` | L’architettura del DevOps Control Plane deve restare semplice, modulare e coerente con GitOps. |
| 1225 | `flusso` | Il valore non è costruire un nuovo orchestratore complesso, ma creare un livello ordinato che renda più sicuro e tracciabile il flusso: |
| 1235 | `evidenze` | Ogni componente deve avere una responsabilità chiara e ogni workflow deve produrre evidenze utili per operatori, reviewer, platform engineer e auditor. |

### `docs/adr/ADR-0001-git-source-of-truth.md`

- **Priority:** P1
- **Potential matches:** 15

| Line | Term | Snippet |
|---:|---|---|
| 18 | `modifica` | 1. modificare direttamente il runtime cluster; |
| 19 | `modifica` | 2. modificare lo stato desiderato in Git e lasciare che Argo CD riconcili il cluster. |
| 21 | `modifica` | Esempi di modifica runtime diretta: |
| 33 | `modifica` | - la modifica non è facilmente revisionabile; |
| 35 | `operativo` | - lo storico operativo è frammentato; |
| 37 | `stato` | - Argo CD può riportare il runtime allo stato Git precedente. |
| 39 | `questo` | Per questo progetto, il principio GitOps deve rimanere centrale. |
| 47 | `flusso` | Il flusso standard è: |
| 60 | `modifica` | DevOps Control Plane non deve modificare direttamente Deployment, ConfigMap, Service o Route come stato finale permanente. |
| 74 | `integrazione` | - Maggiore integrazione con Argo CD. |
| 75 | `evidenze` | - Possibilità di correlare change, commit, validation, sync ed evidenze. |
| 76 | `utenti` | - Maggiore valore formativo per utenti newbie. |
| 84 | `validazione` | - Richiede validazione manifest prima della sync. |
| 92 | `flusso` | Scartata come flusso standard perché rompe il modello GitOps e riduce l'auditabilità. |
| 96 | `stato` | Scartata per MVP perché rischia ambiguità tra stato desiderato e stato reale. |

### `docs/adr/ADR-0002-argocd-as-gitops-engine.md`

- **Priority:** P1
- **Potential matches:** 5

| Line | Term | Snippet |
|---:|---|---|
| 14 | `architettura` | Nel lab e nel modello architetturale esiste già Argo CD / OpenShift GitOps come componente responsabile della riconciliazione tra repository Git e stato runtime del cluster. |
| 52 | `evidenze` | - salvare evidenze Argo CD; |
| 63 | `stato` | - Mantiene Argo CD come fonte autorevole dello stato GitOps. |
| 66 | `integrazione` | - Permette integrazione naturale con rollback e drift detection. |
| 74 | `modifica` | - Alcuni rollback Argo CD non modificano Git e possono produrre `OutOfSync` rispetto a `main`. |

### `docs/adr/ADR-0003-tekton-validation-engine.md`

- **Priority:** P1
- **Potential matches:** 8

| Line | Term | Snippet |
|---:|---|---|
| 14 | `validazione` | La validazione deve essere: |
| 21 | `evidenze` | - documentabile tramite evidenze. |
| 26 | `validazione` | - validazione YAML; |
| 44 | `validazione` | DevOps Control Plane userà Tekton/OpenShift Pipelines come motore di validazione dei change GitOps. |
| 46 | `validazione` | Ogni ChangeRequest che richiede validazione potrà generare una `PipelineRun` dedicata. |
| 66 | `evidenze` | DevOps Control Plane creerà e monitorerà la PipelineRun, raccoglierà TaskRun e log sanitizzati, quindi salverà evidenze in PostgreSQL. |
| 95 | `validazione` | Scartata come unica soluzione perché renderebbe il backend troppo accoppiato alle logiche tecniche di validazione. |
| 107 | `evidenze` | Scelta perché nativa Kubernetes/OpenShift e adatta a produrre evidenze. |

### `docs/adr/ADR-0004-postgresql-change-history.md`

- **Priority:** P1
- **Potential matches:** 6

| Line | Term | Snippet |
|---:|---|---|
| 16 | `stato` | Tekton conserva stato PipelineRun/TaskRun finché le risorse sono presenti o tramite sistemi come Tekton Results. |
| 44 | `stato` | - stato runtime sintetico; |
| 45 | `operativo` | - audit trail operativo. |
| 60 | `evidenze` | - Supporto naturale a JSONB per evidenze flessibili. |
| 78 | `stato` | Insufficiente perché Git non contiene stato Tekton, Argo CD e runtime. |
| 82 | `validazione` | Insufficiente perché Argo CD non contiene motivazione completa del change, requestedBy, validazione Tekton e diff summary applicativo. |

### `docs/adr/ADR-0005-api-first-before-web-ui.md`

- **Priority:** P1
- **Potential matches:** 3

| Line | Term | Snippet |
|---:|---|---|
| 17 | `creazione` | - creazione ChangeRequest; |
| 57 | `utente` | - Esperienza utente iniziale meno ricca. |
| 59 | `documentazione` | - Serve documentazione API più accurata. |

### `docs/adr/ADR-0006-adapter-based-architecture.md`

- **Priority:** P1
- **Potential matches:** 3

| Line | Term | Snippet |
|---:|---|---|
| 32 | `architettura` | Il progetto userà una architettura basata su adapter. |
| 89 | `manutenzione` | Scartata perché complica test e manutenzione. |
| 93 | `integrazione` | Scelta perché adatta a un sistema di orchestrazione multi-integrazione. |

### `docs/adr/ADR-0007-gitlab-api-as-git-provider.md`

- **Priority:** P1
- **Potential matches:** 5

| Line | Term | Snippet |
|---:|---|---|
| 20 | `stato` | - leggere commit e stato MR. |
| 29 | `stato` | Per MVP è stato deciso che il target funzionale sarà GitLab API. |
| 37 | `integrazione` | DevOps Control Plane userà GitLab API come integrazione Git primaria per l’MVP. |
| 61 | `integrazione` | - Migliore integrazione con ChangeRequest e evidence model. |
| 90 | `integrazione` | Scelta perché soddisfa le operazioni MVP e mantiene l’integrazione controllata. |

### `docs/adr/ADR-0008-kubernetes-api-for-tekton.md`

- **Priority:** P1
- **Potential matches:** 1

| Line | Term | Snippet |
|---:|---|---|
| 23 | `evidenze` | - salvare evidenze. |

### `docs/adr/README.md`

- **Priority:** P1
- **Potential matches:** 4

| Line | Term | Snippet |
|---:|---|---|
| 3 | `Questo` | Questo folder contiene gli ADR del progetto DevOps Control Plane. |
| 21 | `Regola` | ## Regola |
| 23 | `architettura` | Ogni decisione architetturale significativa deve essere documentata con un ADR versionato in Git. |
| 30 | `modifica` | - Gli ADR accettati non devono essere modificati retroattivamente per cambiarne il significato; eventuali nuove decisioni devono essere documentate con un nuovo ADR o con una re... |

### `docs/02-personas-use-cases.md`

- **Priority:** P3
- **Potential matches:** 96

| Line | Term | Snippet |
|---:|---|---|
| 14 | `Questo` | Questo documento descrive le **personas** e gli **use cases** principali del progetto **DevOps Control Plane**. |
| 21 | `utente` | - quali informazioni devono essere visibili a ogni tipologia di utente; |
| 28 | `utenti` | ## 2. Principi di progettazione orientati agli utenti |
| 42 | `evidenze` | - orientato alla raccolta evidenze. |
| 52 | `utente` | Il **DevOps Operator** è l’utente operativo principale del sistema. |
| 60 | `stato` | - vedere velocemente lo stato delle applicazioni; |
| 66 | `evidenze` | - raccogliere evidenze finali. |
| 72 | `modifica` | - “Voglio cambiare `PAGE_COLOR` senza modificare manualmente YAML.” |
| 78 | `stato` | - stato sync/health; |
| 82 | `stato` | - stato validazione Tekton; |
| 83 | `stato` | - stato runtime del Deployment; |
| 84 | `evidenze` | - evidenze post-change. |
| 110 | `modifica` | - “Il change sta tentando di modificare una risorsa non ammessa?” |
| 111 | `modifica` | - “Il workflow ha usato Git o ha modificato direttamente il cluster?” |
| 120 | `evidenze` | - evidenze di validazione; |
| 138 | `modifica` | - vedere quali file vengono modificati; |
| 145 | `modifica` | - “Perché devo modificare Git invece del Deployment direttamente?” |
| 149 | `questo` | - “Quale commit ha prodotto questo stato?” |
| 155 | `stato` | - stato workflow; |
| 158 | `evidenze` | - link a evidenze. |
| 166 | `promozione` | Il **Reviewer** verifica che un change sia corretto prima del merge o della promozione. |
| 176 | `validazione` | - vedere il risultato della validazione Tekton; |
| 178 | `operativo` | - valutare il rischio operativo; |
| 183 | `modifica` | - “Quali file sono stati modificati?” |
| 185 | `validazione` | - “La validazione YAML è passata?” |
| 193 | `validazione` | - validazione Tekton; |
| 196 | `stato` | - stato Argo CD; |
| 212 | `stato` | - sapere quando è stato eseguito; |
| 214 | `validazione` | - vedere esito della validazione; |
| 216 | `evidenze` | - consultare evidenze prodotte automaticamente. |
| 221 | `stato` | - “Quando è stato introdotto `APP_VERSION=v3-green`?” |
| 223 | `stato` | - “Il change è stato completato con successo?” |
| 224 | `evidenze` | - “Quali evidenze sono state raccolte?” |
| 231 | `stato` | - stato finale; |
| 235 | `evidenze` | - evidenze runtime. |
| 247 | `Obiettivo` | ### Obiettivo |
| 249 | `stato` | Visualizzare le Application Argo CD disponibili e il loro stato principale. |
| 259 | `utente` | 1. L’utente apre la lista applicazioni o chiama l’API `GET /api/applications`. |
| 282 | `richiesta` | - La richiesta è tracciata nei log applicativi senza esporre token. |
| 296 | `Obiettivo` | ### Obiettivo |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 96. |

### `docs/03-functional-requirements.md`

- **Priority:** P3
- **Potential matches:** 43

| Line | Term | Snippet |
|---:|---|---|
| 17 | `Questo` | Questo documento definisce i **requisiti funzionali** del primo MVP di **DevOps Control Plane**. |
| 25 | `validazione` | - regole di validazione; |
| 65 | `operativo` | - messaggio operativo leggibile; |
| 71 | `Regola` | ## 2.3 Regola GitOps fondamentale |
| 73 | `modifica` | Ogni change permanente su workload applicativi deve essere rappresentato da una modifica Git. |
| 75 | `stato` | Il sistema **non deve** usare modifiche runtime dirette come stato finale. |
| 201 | `operativo` | Il sistema deve mostrare il dettaglio operativo di una Application. |
| 228 | `stato` | - Lo stato `OrphanedResourceWarning` viene riportato ma non trattato automaticamente come errore applicativo se `healthStatus=Healthy`. |
| 394 | `modifica` | ## FR-GIT-004 - Commit file modificati |
| 447 | `stato` | - stato MR. |
| 518 | `stato` | ## FR-CHG-002 - Aggiornare stato ChangeRequest |
| 526 | `stato` | Il sistema deve aggiornare lo stato della ChangeRequest durante il workflow. |
| 553 | `stato` | - Ogni cambio stato crea un evento in `change_events`. |
| 577 | `stato` | - Lista change ordinabile per data e stato. |
| 578 | `evidenze` | - Dettaglio change include eventi, Git, Tekton, Argo CD ed evidenze. |
| 637 | `modifica` | - Il valore viene modificato nel file corretto. |
| 638 | `validazione` | - La validazione Tekton passa. |
| 662 | `validazione` | - Il change passa validazione. |
| 676 | `modifica` | Il sistema deve permettere modifica di chiavi in una ConfigMap GitOps. |
| 683 | `modifica` | - Il sistema deve segnalare se la modifica ConfigMap richiede rollout Pod. |
| 687 | `modifica` | - Il sistema modifica la ConfigMap corretta. |
| 689 | `evidenza` | - Il sistema raccoglie evidenza runtime. |
| 695 | `validazione` | ## FR-TKN-001 - Creare PipelineRun di validazione |
| 729 | `stato` | Il sistema deve monitorare stato PipelineRun fino a successo, fallimento o timeout. |
| 761 | `evidenza` | - I log salvati sono consultabili come evidenza. |
| 799 | `stato` | Il sistema deve attendere che l’Application raggiunga stato desiderato. |
| 809 | `stato` | - Il sistema salva stato finale. |
| 810 | `operativo` | - Il sistema produce messaggio operativo leggibile. |
| 830 | `operativo` | ### Messaggio operativo |
| 853 | `evidenze` | Il sistema deve creare un riepilogo evidenze per ogni ChangeRequest. |
| 876 | `evidenze` | ## FR-EVD-002 - Raccogliere evidenze runtime |
| 884 | `stato` | Il sistema dovrebbe raccogliere stato runtime da Kubernetes/OpenShift. |
| 897 | `evidenze` | - Le evidenze runtime non contengono secret. |
| 927 | `operativo` | - Ogni errore ha messaggio tecnico e messaggio operativo. |
| 940 | `stato` | Un fallimento non deve lasciare il sistema in stato ambiguo. |
| 947 | `stato` | - Ogni stato deve avere messaggio e timestamp. |
| 961 | `ambiente` | Il sistema deve leggere token e credenziali da variabili ambiente o Secret montate. |
| 967 | `evidenze` | - Nessun token nelle evidenze. |
| 980 | `modifica` | Il sistema dovrebbe integrare un controllo anti-secret sui file modificati prima del commit/MR. |
| 1028 | `evidenze` | \| UC-009 Raccolta evidenze \| FR-EVD-001, FR-EVD-002 \| |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 43. |

### `docs/04-non-functional-requirements.md`

- **Priority:** P3
- **Potential matches:** 85

| Line | Term | Snippet |
|---:|---|---|
| 18 | `Questo` | Questo documento definisce i **requisiti non funzionali** del progetto **DevOps Control Plane**. |
| 20 | `questo` | I requisiti non funzionali descrivono le qualità attese del sistema, indipendentemente dalle singole funzionalità applicative. In particolare, questo documento copre: |
| 22 | `sicurezza` | - sicurezza; |
| 39 | `architettura` | Il documento completa `docs/03-functional-requirements.md` e sarà usato come riferimento per architettura, data model, API design, deployment e implementazione. |
| 49 | `sicurezza` | \| MUST \| Obbligatorio per MVP o per sicurezza base \| |
| 60 | `questo` | Per questo motivo, anche nel primo MVP, il sistema deve essere progettato con attenzione a: |
| 66 | `stato` | - separazione tra stato desiderato Git e stato runtime cluster; |
| 93 | `evidenze` | - Nessun token deve essere incluso nelle evidenze. |
| 94 | `ambiente` | - I token devono essere forniti tramite Secret Kubernetes/OpenShift o variabili ambiente sicure. |
| 100 | `utenti` | - Gli errori di autenticazione non stampano token. |
| 101 | `evidenze` | - Le evidenze non contengono valori sensibili. |
| 134 | `stato` | - leggere stato TaskRun e log rilevanti; |
| 135 | `evidenze` | - leggere risorse runtime necessarie alle evidenze; |
| 142 | `produzione` | - I token personali non sono usati come soluzione stabile di produzione. |
| 177 | `modifica` | Il sistema dovrebbe eseguire un controllo anti-secret sui file modificati prima di commit/MR. |
| 196 | `stato` | - Se il controllo trova pattern sospetti, il workflow passa in stato `ValidationFailed` o richiede review. |
| 226 | `stato` | - stato finale; |
| 231 | `stato` | - Ogni stato produce un evento in `change_events`. |
| 245 | `evidenze` | Le evidenze principali devono essere persistite in PostgreSQL o in storage referenziato dal database. |
| 247 | `evidenza` | ### Tipi evidenza |
| 252 | `stato` | - stato Argo CD; |
| 253 | `stato` | - stato Deployment; |
| 254 | `stato` | - stato Pod; |
| 255 | `stato` | - stato ConfigMap, se coinvolta; |
| 256 | `stato` | - stato Route/health check, se disponibile. |
| 260 | `evidenze` | - Ogni ChangeRequest completata o fallita ha evidenze associate. |
| 261 | `evidenze` | - Le evidenze non contengono token o secret. |
| 262 | `evidenze` | - Le evidenze restano disponibili dopo riavvio applicazione. |
| 280 | `Regola` | ### Regola didattica |
| 283 | `stato` | Argo CD descrive cosa è stato sincronizzato sul cluster. |
| 284 | `stato` | DevOps Control Plane descrive perché il change è stato fatto, da chi, con quale workflow e con quali evidenze. |
| 324 | `richiesta` | - Ogni richiesta API genera log start/end o equivalente. |
| 377 | `validazione` | - durata validazione Tekton. |
| 399 | `stato` | - Se fallisce GitLab branch creation, lo stato deve indicare il fallimento. |
| 400 | `stato` | - Se fallisce commit, lo stato deve indicare il fallimento. |
| 401 | `stato` | - Se fallisce Tekton, lo stato deve essere `ValidationFailed`. |
| 402 | `stato` | - Se fallisce Argo CD sync, lo stato deve essere `SyncFailed`. |
| 407 | `stato` | - Nessuna ChangeRequest resta indefinitamente in stato intermedio senza timeout. |
| 427 | `evidenze` | - raccolta evidenze; |
| 450 | `integrazione` | Ogni integrazione esterna deve avere timeout. |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 85. |

### `docs/06-argocd-integration.md`

- **Priority:** P3
- **Potential matches:** 66

| Line | Term | Snippet |
|---:|---|---|
| 20 | `Questo` | Questo documento descrive come **DevOps Control Plane** deve integrarsi con **Argo CD / OpenShift GitOps**. |
| 22 | `obiettivo` | L’obiettivo è definire: |
| 24 | `integrazione` | - responsabilità dell’integrazione Argo CD; |
| 26 | `utenti` | - modello di autenticazione; |
| 33 | `sicurezza` | - requisiti di sicurezza; |
| 36 | `responsabile` | DevOps Control Plane non sostituisce Argo CD. Argo CD resta il motore GitOps responsabile della riconciliazione tra stato desiderato Git e stato runtime nel cluster. |
| 40 | `architettura` | ## 2. Ruolo di Argo CD nell’architettura |
| 47 | `stato` | - confrontare stato Git con stato runtime cluster; |
| 59 | `stato` | - mostrare stato e history; |
| 61 | `stato` | - attendere stato finale; |
| 63 | `evidenze` | - salvare evidenze in PostgreSQL; |
| 64 | `stato` | - correlare stato Argo CD con GitLab, Tekton e runtime OpenShift. |
| 68 | `integrazione` | ## 3. Principi di integrazione |
| 74 | `flusso` | Il flusso corretto è: |
| 87 | `integrazione` | L’integrazione target deve usare **Argo CD API**, non parsing fragile dell’output CLI. |
| 101 | `Regola` | Regola didattica: |
| 105 | `stato` | Argo CD dice cosa è stato sincronizzato. |
| 106 | `evidenze` | DevOps Control Plane dice perché, da chi e con quali evidenze. |
| 113 | `stato` | `OutOfSync` significa che lo stato runtime non coincide con lo stato desiderato Git. |
| 120 | `operativo` | - rollback operativo temporaneo; |
| 121 | `modifica` | - risorsa runtime modificata manualmente; |
| 132 | `Obiettivo` | ### Obiettivo |
| 171 | `Obiettivo` | ### Obiettivo |
| 173 | `operativo` | Mostrare il dettaglio operativo di una Application. |
| 209 | `Obiettivo` | ### Obiettivo |
| 256 | `questo` | Per questo motivo, nel campo `group`, il valore può essere vuoto. |
| 268 | `Obiettivo` | ### Obiettivo |
| 294 | `modifica` | La history Argo CD non modifica Git. Un rollback operativo Argo CD può riportare temporaneamente il cluster a una revision precedente, ma Git resta sulla branch configurata, ad ... |
| 300 | `Obiettivo` | ### Obiettivo |
| 316 | `evidenza` | - il sistema deve produrre evidenza. |
| 329 | `Obiettivo` | ### Obiettivo |
| 331 | `stato` | Attendere che la Application raggiunga stato finale desiderato. |
| 416 | `utenti` | - autenticazione; |
| 545 | `utenti` | ## 7. Autenticazione verso Argo CD |
| 549 | `utenti` | Il DevOps Control Plane deve autenticarsi ad Argo CD tramite token. |
| 571 | `ambienti` | In ambienti lab può essere necessario gestire certificati non trusted o route reencrypt. |
| 580 | `Regola` | Regola: |
| 582 | `produzione` | - in produzione preferire CA/certificati trusted; |
| 592 | `operativo` | \| Stato Argo CD \| Significato operativo \| |
| 594 | `stato` | \| Synced \| Runtime allineato allo stato desiderato Git \| |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 66. |

### `docs/07-gitlab-integration.md`

- **Priority:** P3
- **Potential matches:** 51

| Line | Term | Snippet |
|---:|---|---|
| 21 | `Questo` | Questo documento descrive come **DevOps Control Plane** deve integrarsi con **GitLab**. |
| 23 | `obiettivo` | L’obiettivo è definire: |
| 25 | `integrazione` | - responsabilità dell’integrazione GitLab; |
| 27 | `utenti` | - modello di autenticazione; |
| 33 | `sicurezza` | - requisiti di sicurezza; |
| 34 | `evidenze` | - evidenze GitLab da salvare; |
| 41 | `architettura` | ## 2. Ruolo di GitLab nell’architettura |
| 49 | `modifica` | - permettere lettura e modifica file tramite API; |
| 58 | `modifica` | - modificare file YAML in modo controllato; |
| 61 | `stato` | - leggere stato MR; |
| 68 | `integrazione` | ## 3. Principi di integrazione |
| 90 | `integrazione` | L’integrazione target deve usare **GitLab REST API**. |
| 111 | `stato` | - stato MR; |
| 135 | `validazione` | - validazione Tekton prima del merge; |
| 147 | `Obiettivo` | ### Obiettivo |
| 173 | `Obiettivo` | ### Obiettivo |
| 195 | `evidenza` | - produrre evidenza. |
| 207 | `Obiettivo` | ### Obiettivo |
| 232 | `modifica` | ## 4.4 Commit file modificati |
| 234 | `Obiettivo` | ### Obiettivo |
| 271 | `Obiettivo` | ### Obiettivo |
| 312 | `stato` | - Lo stato MR deve essere interrogabile. |
| 318 | `Obiettivo` | ### Obiettivo |
| 336 | `stato` | ## 4.7 Lettura stato MR |
| 338 | `Obiettivo` | ### Obiettivo |
| 410 | `utenti` | - header autenticazione; |
| 427 | `creazione` | - creazione branch; |
| 444 | `stato` | - mapping stato MR. |
| 528 | `utenti` | ## 7. Autenticazione verso GitLab |
| 554 | `produzione` | Per MVP/lab può essere usato un token tecnico, ma in produzione è preferibile usare token con privilegi minimi e scadenza definita. |
| 563 | `Regola` | ### Regola di sicurezza |
| 579 | `ambienti` | In molti scenari GitLab, lo scope `api` abilita l’uso completo delle API richieste. Per ambienti più restrittivi, valutare token/project role con permessi minimi compatibili. |
| 664 | `modifica` | - MR descrizione include file modificati; |
| 665 | `stato` | - MR descrizione include stato validazione Tekton, se già disponibile; |
| 682 | `utente` | Il polling MR può essere implementato dopo il primo workflow end-to-end. Nel primissimo MVP può essere manuale: l’utente conferma che la MR è stata mergiata. |
| 726 | `modifica` | ## 10. File modification strategy |
| 728 | `modifica` | ## 10.1 YAML modifications |
| 730 | `modifica` | DevOps Control Plane dovrà modificare YAML in modo sicuro. |
| 755 | `evidenza` | - evidenza; |
| 759 | `Regola` | ### Regola |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 51. |

### `docs/08-tekton-integration.md`

- **Priority:** P3
- **Potential matches:** 52

| Line | Term | Snippet |
|---:|---|---|
| 22 | `Questo` | Questo documento descrive come **DevOps Control Plane** deve integrarsi con **Tekton / OpenShift Pipelines**. |
| 24 | `obiettivo` | L’obiettivo è definire: |
| 26 | `integrazione` | - responsabilità dell’integrazione Tekton; |
| 31 | `validazione` | - pipeline di validazione MVP; |
| 33 | `stato` | - raccolta stato e log; |
| 34 | `evidenze` | - evidenze prodotte; |
| 39 | `validazione` | Tekton non sostituisce GitLab o Argo CD. Nel DevOps Control Plane, Tekton ha il ruolo di **motore di validazione e automazione tecnica** dei change GitOps. |
| 43 | `architettura` | ## 2. Ruolo di Tekton nell’architettura |
| 49 | `validazione` | - eseguire pipeline di validazione; |
| 56 | `stato` | - produrre stato `Succeeded` o `Failed`; |
| 63 | `stato` | - monitorare stato PipelineRun; |
| 66 | `evidenze` | - salvare evidenze in PostgreSQL; |
| 67 | `stato` | - aggiornare lo stato della ChangeRequest; |
| 68 | `operativo` | - interpretare errori Tekton in modo operativo. |
| 74 | `architettura` | La scelta architetturale iniziale è: |
| 84 | `integrazione` | - integrazione nativa con OpenShift/Kubernetes; |
| 91 | `Regola` | ### Regola |
| 106 | `validazione` | - validazione YAML; |
| 121 | `stato` | - stato task; |
| 144 | `validazione` | Nel DevOps Control Plane, ogni ChangeRequest che richiede validazione deve generare una PipelineRun dedicata. |
| 162 | `validazione` | ## 5.1 Creare PipelineRun di validazione |
| 164 | `Obiettivo` | ### Obiettivo |
| 191 | `Obiettivo` | ### Obiettivo |
| 193 | `stato` | DevOps Control Plane deve monitorare la PipelineRun fino allo stato finale. |
| 209 | `stato` | - stato finale salvato in PostgreSQL; |
| 217 | `Obiettivo` | ### Obiettivo |
| 236 | `Obiettivo` | ### Obiettivo |
| 252 | `Obiettivo` | ## 6.1 Obiettivo |
| 263 | `modifica` | - non ci sono secret evidenti nei file modificati; |
| 312 | `stato` | Nel lab è stato osservato che un errore di indentazione `valueFrom/configMapKeyRef` può causare failure durante la sync Argo CD. |
| 352 | `modifica` | - cercare pattern sospetti nei file modificati; |
| 353 | `promozione` | - bloccare commit/promozione se vengono trovati secret evidenti. |
| 387 | `questa` | Nota: questa task può essere implementata in modo incrementale. Nel primissimo MVP può essere un controllo statico o opzionale. |
| 522 | `stato` | - polling stato PipelineRun; |
| 720 | `validazione` | userMessage: "La validazione Tekton del change è fallita." |
| 735 | `operativo` | Messaggio operativo: |
| 766 | `sicurezza` | ## 15. RBAC e sicurezza |
| 796 | `evidenza` | - Token non devono essere salvati in evidenza. |
| 840 | `stato` | Il Workflow Engine deve poter essere testato senza cluster reale. |
| 864 | `stato` | 7. verificare stato finale Succeeded; |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 52. |

### `docs/09-security-rbac.md`

- **Priority:** P3
- **Potential matches:** 51

| Line | Term | Snippet |
|---:|---|---|
| 23 | `Questo` | Questo documento definisce il modello iniziale di **sicurezza**, **gestione credenziali** e **RBAC** del progetto **DevOps Control Plane**. |
| 25 | `obiettivo` | L’obiettivo è descrivere: |
| 27 | `sicurezza` | - principi di sicurezza del progetto; |
| 36 | `evidenze` | - gestione evidenze senza dati sensibili; |
| 37 | `ambienti` | - raccomandazioni per ambienti lab e futuri ambienti production-like. |
| 39 | `creazione` | Il documento è propedeutico alla creazione dei manifest in `manifests/` e alla definizione delle policy operative del DevOps Control Plane. |
| 43 | `sicurezza` | ## 2. Principi di sicurezza |
| 49 | `Regola` | Regola: |
| 59 | `segreti` | ## 2.2 Separazione tra configurazione e segreti |
| 116 | `segreti` | ## 2.4 Evidenze senza segreti |
| 118 | `evidenze` | Le evidenze raccolte dal sistema devono essere utili per audit e troubleshooting, ma non devono contenere credenziali. |
| 120 | `Regola` | Regola: |
| 144 | `utenti` | - autenticazione; |
| 155 | `utenti` | Nel primo MVP, l’autenticazione utente applicativa può essere semplificata o demandata a meccanismi esterni. |
| 157 | `architettura` | Tuttavia, l’architettura deve prevedere in futuro: |
| 159 | `utenti` | - autenticazione via OpenShift OAuth / reverse proxy; |
| 162 | `utente` | - audit utente reale. |
| 164 | `stato` | Per MVP/lab, il campo `requestedBy` può essere passato dal client o impostato staticamente, ma deve essere considerato non affidabile finché non esiste autenticazione reale. |
| 198 | `stato` | - leggere stato merge request. |
| 223 | `stato` | - leggere stato operation. |
| 242 | `Regola` | Regola: |
| 258 | `Regola` | Regola: |
| 261 | `evidenze` | - non includere kubeconfig nelle evidenze; |
| 309 | `Regola` | ### Regola |
| 311 | `Questo` | Questo file, se commitato, deve essere un template. I valori reali devono essere applicati tramite procedura sicura fuori Git. |
| 336 | `evidenze` | - raccogliere evidenze. |
| 436 | `evidenze` | Per raccogliere evidenze runtime, DevOps Control Plane deve leggere risorse in namespace applicativi autorizzati, ad esempio: |
| 522 | `validazione` | 2. ServiceAccount pipeline: esegue le task di validazione. |
| 570 | `sicurezza` | ### Nota di sicurezza |
| 572 | `questo` | Anche se l’operazione è `--dry-run=server`, Kubernetes RBAC può richiedere permessi simili all’operazione reale. Per questo motivo questi permessi devono essere valutati con att... |
| 607 | `Regola` | ### Regola |
| 653 | `stato` | - leggere stato merge request. |
| 672 | `utente` | Creare un utente dedicato: |
| 684 | `utente` | L’utente applicativo deve avere permessi solo sul database/schema del DevOps Control Plane. |
| 693 | `ambienti` | In ambienti più strutturati, separare owner schema e application user può essere valutato successivamente. |
| 730 | `evidenze` | ## 14.1 Dati ammessi nelle evidenze |
| 734 | `stato` | - stato Argo CD; |
| 746 | `evidenze` | ## 14.2 Dati vietati nelle evidenze |
| 787 | `validazione` | bloccare validazione o richiedere review esplicita |
| 834 | `questa` | Nel futuro, questa autorizzazione potrà essere modellata in PostgreSQL o ConfigMap. |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 51. |

### `docs/10-data-model.md`

- **Priority:** P3
- **Potential matches:** 40

| Line | Term | Snippet |
|---:|---|---|
| 24 | `Questo` | Questo documento definisce il **modello dati iniziale** del progetto **DevOps Control Plane**. |
| 26 | `obiettivo` | L’obiettivo è descrivere: |
| 36 | `sicurezza` | - regole di sicurezza sui dati; |
| 40 | `obiettivo` | Il modello dati deve supportare il primo obiettivo del prodotto: |
| 46 | `flusso` | Il flusso dati minimo da rappresentare è: |
| 71 | `evidenze` | - salvare evidenze tecniche; |
| 98 | `evidenze` | Le evidenze devono essere sanitizzate prima della persistenza. |
| 114 | `Regola` | Regola: |
| 125 | `stato` | Ogni cambio stato rilevante deve generare un record in `change_events`. |
| 127 | `Questo` | Questo non è un event sourcing completo, ma consente di ricostruire la timeline del workflow. |
| 162 | `stato` | - stato sync/health corrente; |
| 169 | `richiesta` | Rappresenta una richiesta di change GitOps. |
| 191 | `validazione` | - validazione Tekton fallita; |
| 193 | `evidenza` | - evidenza raccolta. |
| 199 | `evidenza` | Rappresenta una evidenza tecnica associata a una ChangeRequest. |
| 361 | `stato` | Snapshot ultimo dello stato Argo CD. |
| 380 | `richiesta` | Conserva la richiesta di change e lo stato corrente del workflow. |
| 661 | `evidenze` | Conserva evidenze tecniche associate alla ChangeRequest. |
| 796 | `utenti` | Quando sarà introdotta autenticazione reale: |
| 916 | `Regola` | Regola: |
| 919 | `stato` | Una ChangeRequest in stato non finale deve avere updated_at recente o un timeout gestito. |
| 942 | `Regola` | Regola: |
| 952 | `evidenza` | La colonna `sanitized` indica se l’evidenza è stata filtrata. |
| 1221 | `stato` | - stato Tekton; |
| 1222 | `stato` | - stato Argo CD; |
| 1223 | `evidenze` | - evidenze. |
| 1285 | `Regola` | Regola: |
| 1291 | `Questa` | Questa regola può essere applicata nel codice all’inizio e poi con constraint/check in futuro. |
| 1303 | `stato` | Se `argocd_application` è valorizzato, `argocd_sync_status` e `argocd_health_status` dovrebbero rappresentare l’ultimo stato osservato. |
| 1311 | `evidenze` | PostgreSQL conserva storico change ed evidenze. |
| 1323 | `stato` | - restore testato; |
| 1328 | `evidenze` | ## 19.2 Retention evidenze |
| 1330 | `evidenze` | Le evidenze possono crescere. |
| 1342 | `evidenze` | - compressione evidenze; |
| 1354 | `evidenze` | - le evidenze sono persistite; |
| 1365 | `Questo` | Questo documento alimenta: |
| 1387 | `modifica` | Che cosa voleva modificare? |
| 1388 | `stato` | Quale branch GitLab è stato creato? |
| 1392 | `stato` | Quale stato runtime è stato osservato? |
| 1393 | `evidenze` | Quali evidenze sono disponibili? |

### `docs/11-change-workflows.md`

- **Priority:** P3
- **Potential matches:** 43

| Line | Term | Snippet |
|---:|---|---|
| 25 | `Questo` | Questo documento descrive i **workflow operativi** del progetto **DevOps Control Plane**. |
| 27 | `obiettivo` | L’obiettivo è definire, in modo ripetibile e implementabile: |
| 34 | `evidenze` | - le evidenze da raccogliere; |
| 35 | `integrazione` | - i punti di integrazione con GitLab, Tekton, Argo CD e OpenShift; |
| 36 | `sicurezza` | - le policy di sicurezza e governance applicate durante il change; |
| 39 | `Questo` | Questo documento è pensato sia per guidare l’implementazione backend sia per diventare una guida didattica per operatori e colleghi meno esperti. |
| 45 | `modifica` | Il DevOps Control Plane non modifica direttamente il runtime come stato finale permanente. |
| 141 | `stato` | Una ChangeRequest in stato finale non deve più avanzare automaticamente. |
| 161 | `Regola` | Regola: |
| 164 | `stato` | Nessuna ChangeRequest deve restare indefinitamente in uno stato intermedio senza timeout o azione richiesta. |
| 197 | `flusso` | Nel lab iniziale può essere supportato un flusso semplificato: |
| 214 | `Questo` | Questo flusso è più veloce, ma meno adatto alla governance enterprise perché salta la review tramite MR. |
| 220 | `stato` | Qualunque stato operativo può transitare a `Failed` se l’errore è non recuperabile. |
| 231 | `Regola` | Regola: |
| 245 | `modifica` | Il workflow deve modificare `spec.replicas` nel file GitOps, non scalare direttamente il Deployment runtime. |
| 289 | `modifica` | 11. Anti-secret check sui file modificati |
| 358 | `validazione` | - validazione Tekton riuscita; |
| 363 | `evidenze` | - evidenze salvate. |
| 436 | `Regola` | ## 7.4 Regola ConfigMap e rollout |
| 457 | `evidenze` | - evidenze salvate. |
| 573 | `creazione` | - chiavi richieste devono essere presenti o policy deve consentire creazione chiavi. |
| 617 | `validazione` | Eseguire validazione Tekton su branch o commit senza avviare sync Argo CD. |
| 665 | `stato` | - si vuole riconciliare lo stato. |
| 672 | `stato` | - repository Git è già nello stato desiderato; |
| 708 | `evidenze` | Raccogliere evidenze di una Application senza modificare Git o lanciare sync. |
| 739 | `Regola` | Regola MVP: |
| 789 | `operativo` | ## 14.2 Rollback Argo CD operativo |
| 791 | `modifica` | Rollback Argo CD può riportare il cluster a una revision precedente, ma non modifica Git. |
| 819 | `creazione` | - creazione automatica MR di revert; |
| 850 | `operativo` | - messaggio operativo; |
| 852 | `stato` | - stato ultimo osservato; |
| 879 | `stato` | - non perdere lo stato precedente; |
| 903 | `stato` | - aggiornare stato; |
| 905 | `stato` | - salvare stato ultimo osservato; |
| 919 | `stato` | Crea ChangeRequest in stato `Created`. |
| 983 | `modifica` | - modifica Git e non runtime diretto; |
| 985 | `validazione` | - esegue validazione Tekton; |
| 986 | `validazione` | - blocca sync se validazione fallisce; |
| 989 | `evidenze` | - raccoglie evidenze; |
| 990 | `stato` | - salva stato finale; |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 43. |

### `docs/12-evidence-model.md`

- **Priority:** P3
- **Potential matches:** 48

| Line | Term | Snippet |
|---:|---|---|
| 26 | `Questo` | Questo documento definisce il **modello delle evidenze** del progetto **DevOps Control Plane**. |
| 28 | `obiettivo` | L’obiettivo è descrivere: |
| 30 | `evidenza` | - cosa si intende per evidenza; |
| 31 | `evidenze` | - quali evidenze devono essere raccolte; |
| 36 | `evidenze` | - come correlare evidenze GitLab, Tekton, Argo CD e Kubernetes/OpenShift; |
| 37 | `evidenze` | - come rendere le evidenze utili per audit, troubleshooting e formazione dei newbie. |
| 39 | `evidenze` | Il modello delle evidenze è centrale per il valore del DevOps Control Plane: non basta eseguire un change, bisogna poter ricostruire cosa è successo, quando, con quali strumenti... |
| 43 | `evidenza` | ## 2. Definizione di evidenza |
| 45 | `evidenza` | Nel DevOps Control Plane, una **evidenza** è un’informazione tecnica, normalizzata e sanitizzata, associata a una ChangeRequest. |
| 47 | `evidenza` | Una evidenza deve aiutare a rispondere a domande come: |
| 50 | `stato` | Quale branch GitLab è stato creato? |
| 66 | `evidenze` | Ogni workflow deve produrre evidenze. |
| 68 | `Regola` | Regola: |
| 71 | `modifica` | Se un workflow modifica Git, valida con Tekton, sincronizza con Argo CD o legge runtime OpenShift, deve produrre evidenza. |
| 78 | `evidenze` | Le evidenze devono essere utili ma sicure. |
| 96 | `evidenza` | Ogni evidenza deve avere: |
| 109 | `evidenza` | Ogni evidenza deve essere associata a: |
| 127 | `evidenza` | ## 4. Tipi di evidenza MVP |
| 172 | `modifica` | Dimostrare quale modifica Git ha rappresentato il change. |
| 178 | `creazione` | - dopo creazione branch; |
| 220 | `utente` | Mostrare cosa è cambiato nei manifest GitOps senza costringere l’utente a leggere il diff completo. |
| 226 | `modifica` | - dopo modifica file; |
| 228 | `validazione` | - prima della validazione Tekton. |
| 281 | `sicurezza` | ### Nota sicurezza |
| 283 | `modifica` | Se il campo modificato sembra contenere secret, il valore non deve essere salvato integralmente. |
| 302 | `stato` | Dimostrare che il change è stato validato tecnicamente prima della sync Argo CD. |
| 308 | `creazione` | - alla creazione della PipelineRun; |
| 394 | `stato` | Dimostrare quale stato Argo CD è stato osservato e quale revision è stata sincronizzata. |
| 401 | `richiesta` | - dopo richiesta sync; |
| 473 | `stato` | Dimostrare lo stato runtime osservato dopo sync o durante un controllo evidence-only. |
| 548 | `sicurezza` | ### Nota sicurezza |
| 598 | `sicurezza` | Dimostrare che sono stati eseguiti controlli di sicurezza minimi sul change. |
| 708 | `validazione` | "userMessage": "La validazione Tekton del change è fallita.", |
| 716 | `evidenze` | ## 15. Naming evidenze |
| 718 | `evidenza` | ## 15.1 Nome evidenza |
| 811 | `evidenza` | Una evidenza viene creata da: |
| 824 | `evidenze` | Per MVP, preferire evidenze append-only. |
| 826 | `Regola` | Regola: |
| 829 | `evidenze` | Non sovrascrivere evidenze storiche, aggiungere una nuova evidenza. |
| 842 | `evidenze` | - pruning evidenze non critiche; |
| ... | ... | Output truncated after 40 matches for this file. Total matches: 48. |

### `docs/13-api-design.md`

- **Priority:** P3
- **Potential matches:** 22

| Line | Term | Snippet |
|---:|---|---|
| 27 | `Questo` | Questo documento definisce il **design iniziale delle API HTTP/REST** del progetto **DevOps Control Plane**. |
| 29 | `obiettivo` | L’obiettivo è descrivere: |
| 44 | `sicurezza` | - requisiti di sicurezza e sanitizzazione. |
| 46 | `Questo` | Questo documento guiderà l’implementazione del backend Go e deve rimanere allineato a: |
| 60 | `architettura` | La UI HTML con Bootstrap potrà usare queste stesse API, ma non deve guidare prematuramente le scelte architetturali. |
| 309 | `operativo` | Restituisce dettaglio operativo di una Application. |
| 533 | `evidenze` | Restituisce dettaglio ChangeRequest con timeline ed evidenze sintetiche. |
| 701 | `validazione` | Crea e monitora una PipelineRun Tekton di validazione. |
| 800 | `evidenze` | Raccoglie evidenze runtime e workflow. |
| 823 | `evidenze` | Restituisce tutte le evidenze associate a una ChangeRequest. |
| 852 | `evidenze` | Restituisce evidenze filtrate per tipo. |
| 880 | `stato` | Restituisce stato validazione Tekton della ChangeRequest. |
| 905 | `validazione` | Restituisce TaskRun collegate alla PipelineRun di validazione. |
| 979 | `Questo` | Questo endpoint va usato con cautela: per change completi è preferibile usare `/changes/{id}/sync`, in modo da mantenere correlazione con ChangeRequest. |
| 1006 | `stato` | Restituisce stato runtime OpenShift/Kubernetes normalizzato. |
| 1046 | `utenti` | \| 401 \| Non autenticato \| |
| 1119 | `utenti` | - autenticazione OpenShift OAuth; |
| 1121 | `utente` | - mapping utente reale; |
| 1174 | `richiesta` | Ogni richiesta dovrebbe avere un `requestId`. |
| 1239 | `creazione` | - creazione ChangeRequest funziona; |
| 1242 | `evidenze` | - evidenze sono consultabili; |
| 1252 | `Questo` | Questo documento alimenta: |

### `docs/documentation-language-policy.md`

- **Priority:** P3
- **Potential matches:** 10

| Line | Term | Snippet |
|---:|---|---|
| 158 | `collaudo` | collaudo |
| 159 | `produzione` | produzione |
| 165 | `collaudo` | collaudo   -> staging |
| 166 | `produzione` | produzione -> production |
| 193 | `promozione` | promozione ChangeRequest |
| 194 | `ambiente` | ambiente production |
| 195 | `evidenza` | evidenza deployment |
| 196 | `manutenzione` | manutenzione operations |
| 341 | `Questo` | grep -R -n "Questo\\|Questa\\|Obiettivo\\|Regola\\|ambiente\\|produzione\\|collaudo\\|manutenzione\\|ripristino\\|evidenza" docs |
| 347 | `Questo` | grep -R -n "Questo\\|Questa\\|Obiettivo\\|Regola\\|ambiente\\|produzione\\|collaudo\\|manutenzione\\|ripristino\\|evidenza" docs \ |

### `docs/phase-1-postgresql-change-repository.md`

- **Priority:** P3
- **Potential matches:** 2

| Line | Term | Snippet |
|---:|---|---|
| 3 | `Obiettivo` | ## Obiettivo |
| 22 | `questo` | L'ID restituito da questo step è l'UUID PostgreSQL della tabella `change_requests`. |

### `docs/postgresql-integration-notes.md`

- **Priority:** P3
- **Potential matches:** 5

| Line | Term | Snippet |
|---:|---|---|
| 3 | `questo` | ## Stato di questo step |
| 5 | `Questo` | Questo step sostituisce il database placeholder con una connessione PostgreSQL reale basata su `pgxpool`. |
| 7 | `modifica` | ## File modificati o aggiunti |
| 25 | `validazione` | ## Comandi di validazione |
| 48 | `questo` | Da questo step in poi `go run ./cmd/devops-control-plane` fallisce se PostgreSQL non è raggiungibile o se `DATABASE_URL` non è configurata. |

---

## 5. Recommended migration plan

Recommended order:

```text
1. Complete and commit the repository language policy if not already committed.
2. Translate and align docs/00-vision.md.
3. Translate and align docs/01-scope-mvp.md.
4. Translate and align docs/05-architecture.md.
5. Review ADR index and ADR files for residual mixed-language content.
6. Review remaining documents by priority.
7. Keep all new documentation in English.
```

---

## 6. Acceptance criteria for Phase 0.2

Phase 0.2 is complete when:

```text
Italian documentation inventory has been generated.
Potential migration targets are listed.
Migration priority is documented.
Repository language policy is referenced.
No automatic translation has been applied without review.
```

---

## 7. Revision history

| Date | Phase | Description |
|---|---:|---|
| 2026-07-03 | 0.2 | Initial generated Italian documentation inventory baseline. |
