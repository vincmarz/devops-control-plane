# ADR-0003 - Tekton as validation engine

**Status:** Accepted  
**Date:** 2026-06-25  
**Project:** DevOps Control Plane  
**Owner iniziale:** Vincenzo Marzario

---

## Context

DevOps Control Plane deve validare i change GitOps prima della sync Argo CD.

La validazione deve essere:

- ripetibile;
- auditabile;
- eseguita su cluster;
- integrabile con OpenShift;
- associabile a una ChangeRequest;
- documentabile tramite evidenze.

Le validazioni candidate includono:

- clone del branch GitLab;
- validazione YAML;
- `kustomize build`;
- server-side dry-run;
- anti-secret check;
- AppProject policy check;
- generazione report.

Queste attività possono essere eseguite in vari modi:

- direttamente nel backend Go;
- con script shell;
- tramite un sistema CI esterno;
- tramite Tekton/OpenShift Pipelines.

---

## Decision

DevOps Control Plane userà Tekton/OpenShift Pipelines come motore di validazione dei change GitOps.

Ogni ChangeRequest che richiede validazione potrà generare una `PipelineRun` dedicata.

La pipeline MVP sarà concettualmente:

```text
validate-gitops-change
```

Task candidate:

```text
clone-repository
yaml-validate
kustomize-build
server-side-dry-run
anti-secret-check
appproject-policy-check
report
```

DevOps Control Plane creerà e monitorerà la PipelineRun, raccoglierà TaskRun e log sanitizzati, quindi salverà evidenze in PostgreSQL.

---

## Consequences

### Positive

- Validazione standardizzata.
- Evidenze tecniche associate a PipelineRun e TaskRun.
- Esecuzione Kubernetes-native.
- Separazione tra orchestrazione backend e task tecniche.
- Possibilità di evolvere le pipeline senza riscrivere tutto il backend.
- Buon valore didattico per colleghi newbie.

### Negative / Trade-off

- Richiede OpenShift Pipelines/Tekton installato.
- Richiede RBAC dedicato.
- Richiede gestione workspace, ServiceAccount e Secret Git.
- Richiede gestione timeout PipelineRun.
- Richiede raccolta log sanitizzata.

---

## Alternatives considered

### Validazione nel backend Go

Scartata come unica soluzione perché renderebbe il backend troppo accoppiato alle logiche tecniche di validazione.

### Script shell

Scartata come modello principale perché meno osservabile e meno Kubernetes-native.

### CI esterna

Possibile in futuro, ma non scelta per MVP perché Tekton è già coerente con OpenShift.

### Tekton validation engine

Scelta perché nativa Kubernetes/OpenShift e adatta a produrre evidenze.

---

## Related documents

- `docs/08-tekton-integration.md`
- `docs/09-security-rbac.md`
- `docs/11-change-workflows.md`
- `docs/12-evidence-model.md`
