# DevOps Control Plane - Security and RBAC

**Versione:** 0.1  
**Data:** 2026-06-25  
**Owner iniziale:** Vincenzo Marzario  
**Repository:** `https://github.com/vincmarz/devops-control-plane`  
**Documenti precedenti:**  
- `docs/00-vision.md`  
- `docs/01-scope-mvp.md`  
- `docs/02-personas-use-cases.md`  
- `docs/03-functional-requirements.md`  
- `docs/04-non-functional-requirements.md`  
- `docs/05-architecture.md`  
- `docs/06-argocd-integration.md`  
- `docs/07-gitlab-integration.md`  
- `docs/08-tekton-integration.md`  
**Stato:** Draft iniziale / Security and RBAC

---

## 1. Scopo del documento

Questo documento definisce il modello iniziale di **sicurezza**, **gestione credenziali** e **RBAC** del progetto **DevOps Control Plane**.

L’obiettivo è descrivere:

- principi di sicurezza del progetto;
- trust boundary principali;
- credenziali necessarie;
- gestione Secret e ConfigMap;
- permessi minimi per GitLab, Argo CD, Tekton e Kubernetes/OpenShift;
- ServiceAccount previsti;
- Role e RoleBinding indicativi;
- controlli anti-secret;
- logging sicuro;
- gestione evidenze senza dati sensibili;
- raccomandazioni per ambienti lab e futuri ambienti production-like.

Il documento è propedeutico alla creazione dei manifest in `manifests/` e alla definizione delle policy operative del DevOps Control Plane.

---

## 2. Principi di sicurezza

## 2.1 Principio del minimo privilegio

Ogni identità usata dal DevOps Control Plane deve avere solo i permessi strettamente necessari.

Regola:

```text
Nessun token o ServiceAccount deve essere cluster-admin per impostazione predefinita.
```

Permessi elevati possono essere usati solo in laboratorio controllato e devono essere rimossi prima di qualsiasi utilizzo stabile.

---

## 2.2 Separazione tra configurazione e segreti

Configurazioni non sensibili:

- endpoint Argo CD;
- endpoint GitLab;
- namespace Tekton;
- nome Pipeline;
- default branch;
- parametri timeout;
- log level.

Queste possono stare in `ConfigMap`.

Dati sensibili:

- token GitLab;
- token Argo CD;
- password PostgreSQL;
- stringa di connessione PostgreSQL completa se contiene password;
- token Kubernetes;
- chiavi private;
- credenziali Git per pipeline.

Questi devono stare in `Secret`.

---

## 2.3 Git è pubblico rispetto al modello secret

Il repository Git del progetto non deve mai contenere:

- token;
- password;
- kubeconfig reali;
- certificati privati;
- chiavi SSH private;
- dockerconfig reali;
- secret Kubernetes reali;
- `.env` con valori sensibili.

Sono ammessi solo template con placeholder.

Esempio ammesso:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: devops-control-plane-secrets
stringData:
  GITLAB_TOKEN: "<replace-me>"
  ARGOCD_AUTH_TOKEN: "<replace-me>"
```

---

## 2.4 Evidenze senza segreti

Le evidenze raccolte dal sistema devono essere utili per audit e troubleshooting, ma non devono contenere credenziali.

Regola:

```text
Evidence first, but sanitized evidence.
```

---

## 3. Trust boundaries

## 3.1 Boundary principali

```text
User / API Client
   -> DevOps Control Plane HTTP API
   -> PostgreSQL
   -> GitLab API
   -> Argo CD API
   -> Kubernetes/OpenShift API
   -> Tekton CRDs
```

Ogni attraversamento boundary deve essere controllato tramite:

- autenticazione;
- autorizzazione;
- timeout;
- logging sicuro;
- gestione errori;
- masking dei dati sensibili.

---

## 3.2 Boundary HTTP ingresso

Nel primo MVP, l’autenticazione utente applicativa può essere semplificata o demandata a meccanismi esterni.

Tuttavia, l’architettura deve prevedere in futuro:

- autenticazione via OpenShift OAuth / reverse proxy;
- ruoli applicativi;
- autorizzazione per change type;
- audit utente reale.

Per MVP/lab, il campo `requestedBy` può essere passato dal client o impostato staticamente, ma deve essere considerato non affidabile finché non esiste autenticazione reale.

---

## 3.3 Boundary verso API esterne

Ogni adapter esterno deve:

- usare token da Secret;
- non stampare token nei log;
- usare timeout;
- normalizzare errori;
- non ritornare payload sensibili al client;
- non salvare credenziali in PostgreSQL.

---

## 4. Credenziali richieste

## 4.1 GitLab

Variabile sensibile:

```text
GITLAB_TOKEN
```

Uso:

- leggere repository;
- leggere file;
- creare branch;
- creare commit;
- aprire merge request;
- leggere stato merge request.

Raccomandazione:

- per lab può essere usato token tecnico controllato;
- per uso stabile preferire Project Access Token, Group Access Token o service account/bot dedicato;
- evitare Personal Access Token nominali in uso stabile;
- usare scope e ruolo minimi compatibili.

---

## 4.2 Argo CD

Variabile sensibile:

```text
ARGOCD_AUTH_TOKEN
```

Uso:

- listare Application;
- leggere dettaglio Application;
- leggere resources/history;
- lanciare sync;
- leggere stato operation.

Raccomandazione:

- usare account/token dedicato;
- limitare accesso ad Application/progetti richiesti, se possibile;
- evitare token admin globali come configurazione stabile.

---

## 4.3 PostgreSQL

Variabili sensibili possibili:

```text
DATABASE_URL
DATABASE_PASSWORD
```

Regola:

Se `DATABASE_URL` contiene user/password, deve essere considerata sensibile e stare in Secret.

---

## 4.4 Kubernetes/OpenShift

Quando DevOps Control Plane gira dentro OpenShift, l’accesso Kubernetes avviene tramite ServiceAccount token montato automaticamente nel Pod.

Quando gira fuori cluster, per sviluppo/lab:

```text
KUBECONFIG
```

Regola:

- non salvare kubeconfig nel repository;
- non includere kubeconfig nelle evidenze;
- preferire in-cluster ServiceAccount nel deployment target.

---

## 5. ConfigMap e Secret previsti

## 5.1 ConfigMap applicativa

Esempio concettuale:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: devops-control-plane-config
  namespace: devops-control-plane
data:
  HTTP_ADDR: ":8080"
  LOG_LEVEL: "info"
  ARGOCD_BASE_URL: "https://openshift-gitops-server-openshift-gitops.apps.example.local"
  GITLAB_BASE_URL: "https://gitlab.example.local"
  TEKTON_NAMESPACE: "devops-control-plane"
  TEKTON_PIPELINE_NAME: "validate-gitops-change"
  TEKTON_SERVICE_ACCOUNT: "pipeline"
  TEKTON_TIMEOUT_SECONDS: "900"
  TEKTON_POLL_INTERVAL_SECONDS: "5"
```

---

## 5.2 Secret applicativo

Esempio template con placeholder:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: devops-control-plane-secrets
  namespace: devops-control-plane
type: Opaque
stringData:
  GITLAB_TOKEN: "<replace-me>"
  ARGOCD_AUTH_TOKEN: "<replace-me>"
  DATABASE_URL: "postgres://devops_cp:<replace-me>@postgresql:5432/devops_control_plane?sslmode=disable"
```

### Regola

Questo file, se commitato, deve essere un template. I valori reali devono essere applicati tramite procedura sicura fuori Git.

---

## 6. ServiceAccount previsti

## 6.1 ServiceAccount applicativo

Nome proposto:

```text
devops-control-plane
```

Namespace:

```text
devops-control-plane
```

Responsabilità:

- eseguire il backend Go;
- leggere risorse runtime autorizzate;
- creare e monitorare PipelineRun Tekton;
- raccogliere evidenze.

---

## 6.2 ServiceAccount Tekton Pipeline

Nome proposto:

```text
pipeline
```

oppure:

```text
devops-control-plane-pipeline
```

Responsabilità:

- eseguire le Task della pipeline `validate-gitops-change`;
- clonare repository;
- eseguire validazioni YAML/Kustomize;
- eseguire dry-run server-side;
- accedere solo alle risorse minime necessarie.

---

## 7. RBAC DevOps Control Plane - namespace applicativo

## 7.1 Permessi nel namespace `devops-control-plane`

Il backend deve poter gestire risorse Tekton nel namespace dedicato.

Permessi minimi indicativi:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: devops-control-plane-tekton-runner
  namespace: devops-control-plane
rules:
  - apiGroups:
      - tekton.dev
    resources:
      - pipelineruns
    verbs:
      - create
      - get
      - list
      - watch
  - apiGroups:
      - tekton.dev
    resources:
      - taskruns
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - pods
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - pods/log
    verbs:
      - get
```

RoleBinding:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: devops-control-plane-tekton-runner
  namespace: devops-control-plane
subjects:
  - kind: ServiceAccount
    name: devops-control-plane
    namespace: devops-control-plane
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: devops-control-plane-tekton-runner
```

---

## 8. RBAC runtime evidence sui namespace applicativi

## 8.1 Permessi read-only namespace target

Per raccogliere evidenze runtime, DevOps Control Plane deve leggere risorse in namespace applicativi autorizzati, ad esempio:

```text
devops-ci-demo
```

Permessi minimi indicativi:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: devops-control-plane-runtime-reader
  namespace: devops-ci-demo
rules:
  - apiGroups:
      - apps
    resources:
      - deployments
      - replicasets
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - ""
    resources:
      - pods
      - services
      - configmaps
      - events
    verbs:
      - get
      - list
      - watch
  - apiGroups:
      - route.openshift.io
    resources:
      - routes
    verbs:
      - get
      - list
      - watch
```

RoleBinding nel namespace target:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: devops-control-plane-runtime-reader
  namespace: devops-ci-demo
subjects:
  - kind: ServiceAccount
    name: devops-control-plane
    namespace: devops-control-plane
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: devops-control-plane-runtime-reader
```

### Nota

Il RoleBinding cross-namespace è ammesso perché il subject punta al ServiceAccount del namespace `devops-control-plane`, mentre il RoleBinding vive nel namespace target.

---

## 9. RBAC pipeline validation

## 9.1 Permessi della PipelineRun

La pipeline `validate-gitops-change` potrebbe dover eseguire:

```bash
oc apply --dry-run=server
```

Per farlo, il ServiceAccount della PipelineRun deve avere permessi sufficienti a validare le risorse target.

### Approccio consigliato MVP

Separare due livelli:

1. ServiceAccount backend: crea e monitora PipelineRun.
2. ServiceAccount pipeline: esegue le task di validazione.

---

## 9.2 Permessi dry-run nel namespace target

Permessi indicativi per la pipeline nel namespace applicativo:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: devops-control-plane-validation-dry-run
  namespace: devops-ci-demo
rules:
  - apiGroups:
      - ""
    resources:
      - services
      - configmaps
    verbs:
      - get
      - list
      - create
      - patch
      - update
  - apiGroups:
      - apps
    resources:
      - deployments
    verbs:
      - get
      - list
      - create
      - patch
      - update
  - apiGroups:
      - route.openshift.io
    resources:
      - routes
    verbs:
      - get
      - list
      - create
      - patch
      - update
```

### Nota di sicurezza

Anche se l’operazione è `--dry-run=server`, Kubernetes RBAC può richiedere permessi simili all’operazione reale. Per questo motivo questi permessi devono essere valutati con attenzione e limitati ai namespace autorizzati.

---

## 9.3 RoleBinding pipeline

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: devops-control-plane-validation-dry-run
  namespace: devops-ci-demo
subjects:
  - kind: ServiceAccount
    name: pipeline
    namespace: devops-control-plane
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: devops-control-plane-validation-dry-run
```

---

## 10. Argo CD authorization model

## 10.1 Token Argo CD

Il token Argo CD deve consentire:

- list Application autorizzate;
- get Application;
- get resources/history;
- sync Application autorizzate.

### Regola

Non usare token admin globale in modo stabile.

---

## 10.2 AppProject governance

DevOps Control Plane deve rispettare gli AppProject.

Esempio:

```yaml
namespaceResourceWhitelist:
  - group: ""
    kind: Service
  - group: ""
    kind: ConfigMap
  - group: apps
    kind: Deployment
  - group: route.openshift.io
    kind: Route
```

Se una risorsa non è autorizzata, il sistema deve segnalare:

```text
ARGO_RESOURCE_NOT_PERMITTED
```

E suggerire la correzione del manifest AppProject.

---

## 11. GitLab authorization model

## 11.1 Permessi GitLab minimi

Il token GitLab deve potere:

- leggere repository;
- leggere file;
- leggere commit;
- creare branch;
- creare commit;
- aprire merge request;
- leggere stato merge request.

---

## 11.2 Raccomandazione token

Ordine preferibile:

1. Project Access Token per singolo progetto;
2. Group Access Token per gruppo di repository;
3. service account/bot user controllato;
4. Personal Access Token solo per sviluppo/lab.

---

## 12. PostgreSQL security

## 12.1 Utente database dedicato

Creare un utente dedicato:

```text
devops_cp
```

Evita usare superuser PostgreSQL.

---

## 12.2 Permessi DB minimi

L’utente applicativo deve avere permessi solo sul database/schema del DevOps Control Plane.

Esempio concettuale:

```sql
CREATE USER devops_cp WITH PASSWORD '<replace-me>';
CREATE DATABASE devops_control_plane OWNER devops_cp;
```

In ambienti più strutturati, separare owner schema e application user può essere valutato successivamente.

---

## 13. Logging sicuro

## 13.1 Dati vietati nei log

Non loggare:

- token GitLab;
- token Argo CD;
- password DB;
- kubeconfig;
- Secret Kubernetes;
- header Authorization;
- contenuto completo di `.dockerconfigjson`;
- chiavi private.

---

## 13.2 Masking

Se serve loggare configurazioni, mascherare valori sensibili.

Esempio:

```text
GITLAB_TOKEN=****
ARGOCD_AUTH_TOKEN=****
DATABASE_URL=postgres://devops_cp:****@postgresql:5432/devops_control_plane
```

---

## 14. Evidence sanitization

## 14.1 Dati ammessi nelle evidenze

- nome Application;
- namespace;
- stato Argo CD;
- commit SHA;
- MR URL;
- PipelineRun name;
- TaskRun status;
- Deployment status;
- Pod status;
- ConfigMap keys non sensibili;
- errori tecnici sanitizzati.

---

## 14.2 Dati vietati nelle evidenze

- token;
- password;
- secret values;
- kubeconfig;
- private key;
- raw Secret YAML;
- docker auth JSON;
- header Authorization.

---

## 15. Anti-secret check

## 15.1 Pattern minimi

Il controllo anti-secret deve cercare almeno:

```text
token
password
secret
auth
PRIVATE KEY
BEGIN RSA
ghp_
github_pat_
AKIA
ASIA
.dockerconfigjson
config.json
```

---

## 15.2 Azione in caso di match

Per MVP:

```text
bloccare validazione o richiedere review esplicita
```

Stato suggerito:

```text
ValidationFailed
```

Codice errore:

```text
VALIDATION_SECRET_DETECTED
```

---

## 16. Namespace e blast radius

## 16.1 Namespace dedicato

Usare namespace dedicato:

```text
devops-control-plane
```

Vantaggi:

- separazione logica;
- RBAC più chiaro;
- Secret isolati;
- PipelineRun isolate;
- PVC/workspace dedicati.

---

## 16.2 Namespace applicativi autorizzati

DevOps Control Plane deve poter operare solo sui namespace esplicitamente autorizzati.

Esempio MVP:

```text
devops-ci-demo
```

Nel futuro, questa autorizzazione potrà essere modellata in PostgreSQL o ConfigMap.

---

## 17. File da versionare e file da non versionare

## 17.1 Versionabili

- manifest template;
- ConfigMap non sensibili;
- Secret template con placeholder;
- Role/RoleBinding;
- ServiceAccount;
- Pipeline e Task Tekton;
- documentazione;
- codice Go;
- migrations SQL.

---

## 17.2 Non versionabili

- `.env` reale;
- token;
- password;
- private key;
- kubeconfig;
- Secret reali;
- certificati privati;
- output evidence contenente dati sensibili.

---

## 18. Threat model MVP

## 18.1 Minacce principali

| Minaccia | Impatto | Mitigazione |
|---|---|---|
| Token GitLab esposto in Git | Accesso repository | Secret, `.gitignore`, anti-secret check |
| Token Argo CD esposto nei log | Deploy non autorizzati | masking log, no token logging |
| ServiceAccount troppo privilegiata | Impatto cluster ampio | RBAC minimo per namespace |
| Evidence con Secret | Data leakage | sanitizzazione evidence |
| Pipeline con permessi eccessivi | Modifica runtime non controllata | ServiceAccount dedicato e limitato |
| Drift GitOps | Stato non auditabile | blocco modifiche runtime permanenti |
| Branch/commit non tracciato | Audit incompleto | Change ID in branch e commit |

---

## 19. Checklist sicurezza MVP

Prima di considerare pronta una milestone:

- `.env` è in `.gitignore`;
- nessun token è presente in Git;
- Secret reali non sono commitati;
- token non compaiono nei log;
- evidence non contiene secret;
- ServiceAccount non è cluster-admin;
- Role/RoleBinding sono namespace-scoped dove possibile;
- GitLab token ha privilegi minimi;
- Argo CD token non è admin globale, se possibile;
- PipelineRun usa ServiceAccount dedicato;
- anti-secret check è previsto nel workflow;
- i timeout sono configurati;
- errori 401/403 sono gestiti senza leak.

---

## 20. Roadmap sicurezza post-MVP

Possibili evoluzioni:

- autenticazione utenti tramite OpenShift OAuth;
- RBAC applicativo interno;
- mapping utenti reali -> ChangeRequest;
- approvazioni per change ad alto impatto;
- integrazione con Vault o secret manager;
- rotazione token automatizzata;
- policy engine con OPA/Kyverno;
- audit export;
- firma commit;
- verifica firma commit;
- scanning secret più avanzato;
- integrazione SIEM.

---

## 21. Relazione con altri documenti

Questo documento alimenta:

- `manifests/serviceaccount.yaml`;
- `manifests/role.yaml`;
- `manifests/rolebinding.yaml`;
- `manifests/configmap.yaml`;
- `manifests/secret-template.yaml`;
- `pipelines/validate-gitops-change.yaml`;
- `docs/10-data-model.md`, per campi sensibili da evitare;
- `docs/11-change-workflows.md`, per gli step di validazione sicurezza;
- `docs/12-evidence-model.md`, per sanitizzazione evidence;
- ADR sulle scelte security più rilevanti.

---

## 22. Messaggio chiave

La sicurezza del DevOps Control Plane deve essere progettata fin dal primo MVP.

Il sistema orchestra GitLab, Argo CD, Tekton e OpenShift: quindi un errore nella gestione dei permessi può avere impatto significativo.

Regola fondamentale:

```text
Least privilege, no secrets in Git, no secrets in logs, no secrets in evidence.
```

Il DevOps Control Plane deve rendere i change più sicuri e tracciabili, non introdurre un nuovo punto di rischio non governato.
