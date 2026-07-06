package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type uiData struct {
	Title               string
	Subtitle            string
	Active              string
	Mode                string
	Changes             []map[string]any
	Applications        []map[string]any
	SelectedApplication map[string]any
	Resources           []map[string]any
	History             []map[string]any
	Runtime             map[string]any
	SelectedChange      map[string]any
	Events              []map[string]any
	Evidence            []map[string]any
	Stats               map[string]any
	Flash               string
	ActionError         string
	Error               string
}

func (h *Handler) uiDashboard(w http.ResponseWriter, r *http.Request) {
	changes, err := h.uiChangesData(r)
	if err != nil {
		h.renderUI(w, http.StatusInternalServerError, uiData{Title: "Dashboard", Subtitle: "Applications and change overview", Active: "dashboard", Error: err.Error()})
		return
	}
	apps := h.uiApplicationsData(r)
	selected := preferredChange(changes, "CHG-2026-0005")
	events, evidence := h.uiChangeDetailsData(r, selected)
	h.renderUI(w, http.StatusOK, uiData{Title: "Dashboard", Subtitle: "Applications and change overview", Active: "dashboard", Changes: changes, Applications: apps, SelectedChange: selected, Events: events, Evidence: evidence, Stats: buildUIStats(changes, apps, evidence)})
}

func (h *Handler) uiChanges(w http.ResponseWriter, r *http.Request) {
	changes, err := h.uiChangesData(r)
	if err != nil {
		h.renderUI(w, http.StatusInternalServerError, uiData{Title: "Change Requests", Subtitle: "Change requests managed by the Control Plane", Active: "changes", Error: err.Error()})
		return
	}
	h.renderUI(w, http.StatusOK, uiData{Title: "Change Requests", Subtitle: "Change requests managed by the Control Plane", Active: "changes", Changes: changes, Stats: buildUIStats(changes, nil, nil)})
}

func (h *Handler) uiApplications(w http.ResponseWriter, r *http.Request) {
	apps := h.uiApplicationsData(r)
	changes, _ := h.uiChangesData(r)
	h.renderUI(w, http.StatusOK, uiData{Title: "Applications", Subtitle: "Applications managed by the Control Plane", Active: "applications", Applications: apps, Changes: changes, Stats: buildUIStats(changes, apps, nil)})
}

func (h *Handler) uiChangesAPIPage(w http.ResponseWriter, r *http.Request) {
	changes, err := h.uiChangesData(r)
	if err != nil {
		h.renderUI(w, http.StatusInternalServerError, uiData{Title: "Changes API", Subtitle: "API data preview", Active: "changes", Mode: "changesAPI", Error: err.Error()})
		return
	}
	h.renderUI(w, http.StatusOK, uiData{Title: "Changes API", Subtitle: "API data preview with navigation", Active: "changes", Mode: "changesAPI", Changes: changes, Stats: buildUIStats(changes, nil, nil)})
}

func (h *Handler) uiSettings(w http.ResponseWriter, r *http.Request) {
	h.renderUI(w, http.StatusOK, uiData{Title: "Settings", Subtitle: "Runtime readiness, environment and access placeholders", Active: "settings"})
}

func (h *Handler) uiApplicationDetail(w http.ResponseWriter, r *http.Request) {
	name := strings.Trim(strings.TrimPrefix(r.URL.Path, "/ui/applications/"), "/")
	if name == "" {
		h.uiApplications(w, r)
		return
	}
	apps := h.uiApplicationsData(r)
	selected, err := h.uiApplicationByName(r, name, apps)
	if err != nil {
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Application", Subtitle: name, Active: "applications", Applications: apps, Error: err.Error()})
		return
	}
	resources := toMapSlice(h.deps.Services.Applications.Resources(r.Context(), name))
	history := toMapSlice(h.deps.Services.Applications.History(r.Context(), name))
	runtime := toMap(h.deps.Services.Applications.Runtime(r.Context(), name))
	changes, _ := h.uiChangesData(r)
	h.renderUI(w, http.StatusOK, uiData{Title: fmt.Sprintf("Application: %s", str(get(selected, "name"))), Subtitle: "Application runtime and GitOps detail", Active: "applications", Applications: apps, SelectedApplication: selected, Resources: resources, History: history, Runtime: runtime, Changes: changes, Stats: buildUIStats(changes, apps, nil)})
}

func (h *Handler) uiChangeDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/ui/changes/"), "/")
	if path == "" {
		h.uiChanges(w, r)
		return
	}
	parts := strings.Split(path, "/")
	if len(parts) == 2 && parts[1] == "events" {
		h.uiChangeEvents(w, r, parts[0])
		return
	}
	if len(parts) == 2 && parts[1] == "evidence" {
		h.uiChangeEvidence(w, r, parts[0])
		return
	}
	if len(parts) != 1 {
		http.NotFound(w, r)
		return
	}
	id := parts[0]
	change, err := h.deps.Services.Changes.Get(r.Context(), id)
	if err != nil {
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Change Request", Subtitle: id, Active: "changes", Error: err.Error()})
		return
	}
	selected := toMap(change)
	events, evidence := h.uiChangeDetailsData(r, selected)
	if strings.TrimSpace(str(get(selected, "runtimeStatus"))) == "" {
		if status := latestRuntimeStatusFromEvents(events); status != "" {
			selected["runtimeStatus"] = status
		}
	}
	h.renderUI(w, http.StatusOK, uiData{Title: fmt.Sprintf("Change Request: %s", str(get(selected, "changeNumber"))), Subtitle: "Operational change detail", Active: "changes", SelectedChange: selected, Events: events, Evidence: evidence, Stats: map[string]any{"evidence": len(evidence)}, Flash: r.URL.Query().Get("flash"), ActionError: r.URL.Query().Get("error")})
}

func (h *Handler) uiChangeEvents(w http.ResponseWriter, r *http.Request, id string) {
	change, err := h.deps.Services.Changes.Get(r.Context(), id)
	if err != nil {
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Audit Events", Subtitle: id, Active: "changes", Error: err.Error()})
		return
	}
	events, err := h.deps.Services.Changes.Events(r.Context(), id)
	if err != nil {
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Audit Events", Subtitle: id, Active: "changes", Error: err.Error()})
		return
	}
	selected := toMap(change)
	h.renderUI(w, http.StatusOK, uiData{Title: fmt.Sprintf("Audit events: %s", changeNumberOrID(selected)), Subtitle: "Change audit trail and technical workflow events", Active: "changes", Mode: "changeEvents", SelectedChange: selected, Events: toMapSlice(events), Stats: map[string]any{"events": len(events)}})
}

func (h *Handler) uiChangeEvidence(w http.ResponseWriter, r *http.Request, id string) {
	change, err := h.deps.Services.Changes.Get(r.Context(), id)
	if err != nil {
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Evidence", Subtitle: id, Active: "changes", Error: err.Error()})
		return
	}
	evidence, err := h.deps.Services.Evidence.List(r.Context(), id, "")
	if err != nil {
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Evidence", Subtitle: id, Active: "changes", Error: err.Error()})
		return
	}
	selected := toMap(change)
	h.renderUI(w, http.StatusOK, uiData{Title: fmt.Sprintf("Evidence: %s", changeNumberOrID(selected)), Subtitle: "Collected technical and runtime evidence", Active: "changes", Mode: "changeEvidence", SelectedChange: selected, Evidence: toMapSlice(evidence), Stats: map[string]any{"evidence": len(evidence)}})
}

func (h *Handler) uiChangeAction(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/ui/changes/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) != 3 || parts[1] != "actions" {
		http.NotFound(w, r)
		return
	}

	id := parts[0]
	action := parts[2]
	var err error

	switch action {
	case "validate":
		_, err = h.deps.Services.Changes.Validate(r.Context(), id)
	case "check-validation":
		_, err = h.deps.Services.Changes.CheckValidation(r.Context(), id)
	case "create-branch":
		_, err = h.deps.Services.Changes.CreateBranch(r.Context(), id)
	case "update-files":
		_, err = h.deps.Services.Changes.UpdateFiles(r.Context(), id)
	case "open-merge-request":
		_, err = h.deps.Services.Changes.OpenMergeRequest(r.Context(), id)
	case "merge-request":
		_, err = h.deps.Services.Changes.MergeRequest(r.Context(), id)
	case "check-deployment":
		_, err = h.deps.Services.Changes.CheckDeployment(r.Context(), id)
	case "collect-evidence":
		_, err = h.deps.Services.Changes.CollectEvidence(r.Context(), id)
	default:
		http.NotFound(w, r)
		return
	}

	redirectURL := "/ui/changes/" + url.PathEscape(id)
	if err != nil {
		redirectURL += "?error=" + url.QueryEscape(err.Error())
	} else {
		redirectURL += "?flash=" + url.QueryEscape("Action completed: "+action)
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *Handler) uiChangesData(r *http.Request) ([]map[string]any, error) {
	changes, err := h.deps.Services.Changes.List(r.Context())
	if err != nil {
		return nil, err
	}
	items := toMapSlice(changes)
	for _, item := range items {
		if strings.TrimSpace(str(get(item, "runtimeStatus"))) != "" {
			continue
		}
		id := changeNumberOrID(item)
		if id == "" {
			continue
		}
		events, err := h.deps.Services.Changes.Events(r.Context(), id)
		if err != nil {
			continue
		}
		if status := latestRuntimeStatusFromEvents(toMapSlice(events)); status != "" {
			item["runtimeStatus"] = status
		}
	}
	sort.SliceStable(items, func(i, j int) bool { return str(get(items[i], "changeNumber")) > str(get(items[j], "changeNumber")) })
	return items, nil
}

func (h *Handler) uiApplicationsData(r *http.Request) []map[string]any {
	apps, err := h.deps.Services.Applications.List(r.Context())
	if err != nil || len(apps) == 0 {
		return []map[string]any{{"name": "demo-go-color-app", "targetNamespace": "devops-ci-demo", "healthStatus": "Healthy", "syncStatus": "Synced"}}
	}
	return toMapSlice(apps)
}

func (h *Handler) uiApplicationByName(r *http.Request, name string, apps []map[string]any) (map[string]any, error) {
	if app, err := h.deps.Services.Applications.Get(r.Context(), name); err == nil {
		return toMap(app), nil
	}
	for _, app := range apps {
		if str(get(app, "name")) == name {
			return app, nil
		}
	}
	if name == "demo-go-color-app" {
		return map[string]any{"name": "demo-go-color-app", "targetNamespace": "devops-ci-demo", "healthStatus": "Healthy", "syncStatus": "Synced"}, nil
	}
	return nil, fmt.Errorf("application not found: %s", name)
}

func (h *Handler) uiChangeDetailsData(r *http.Request, change map[string]any) ([]map[string]any, []map[string]any) {
	if change == nil {
		return nil, nil
	}
	id := str(get(change, "changeNumber"))
	if id == "" {
		id = str(get(change, "id"))
	}
	eventsRaw, _ := h.deps.Services.Changes.Events(r.Context(), id)
	evidenceRaw, _ := h.deps.Services.Evidence.List(r.Context(), id, "deployment")
	return toMapSlice(eventsRaw), toMapSlice(evidenceRaw)
}

func (h *Handler) renderUI(w http.ResponseWriter, status int, data uiData) {
	funcs := template.FuncMap{"get": get, "str": str, "short": short, "badgeClass": badgeClass, "latestEvidence": latestEvidence, "kubeSummary": kubeSummary, "diagnosticsSummary": diagnosticsSummary, "eventStep": eventStep, "jsonPretty": jsonPretty, "changeNumberOrID": changeNumberOrID, "recommendedActions": recommendedActions, "advancedActions": advancedActions, "applicationName": applicationName, "recentChanges": recentChanges}
	tpl := template.Must(template.New("ui").Funcs(funcs).Parse(uiTemplate))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = tpl.Execute(w, data)
}

func toMap(v any) map[string]any {
	raw, _ := json.Marshal(v)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	return out
}
func toMapSlice(v any) []map[string]any {
	raw, _ := json.Marshal(v)
	var out []map[string]any
	_ = json.Unmarshal(raw, &out)
	if out == nil {
		return []map[string]any{}
	}
	return out
}
func get(m map[string]any, key string) any {
	if m == nil {
		return ""
	}
	return m[key]
}
func str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
func short(v any) string {
	s := str(v)
	if len(s) <= 12 {
		return s
	}
	return s[:12] + "…"
}

func applicationName(m map[string]any) string {
	return str(get(m, "name"))
}

func changeNumberOrID(m map[string]any) string {
	if v := str(get(m, "changeNumber")); v != "" {
		return v
	}
	return str(get(m, "id"))
}

func badgeClass(v any) string {
	s := strings.ToLower(str(v))
	s = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "_", ""), "-", ""), " ", "")
	switch {
	case strings.Contains(s, "healthy"), strings.Contains(s, "succeeded"), strings.Contains(s, "evidencecollected"), strings.Contains(s, "closed"), strings.Contains(s, "true"):
		return "badge-ok"
	case strings.Contains(s, "failed"), strings.Contains(s, "degraded"), strings.Contains(s, "error"):
		return "badge-bad"
	case strings.Contains(s, "running"), strings.Contains(s, "progressing"), strings.Contains(s, "merge"):
		return "badge-warn"
	case strings.Contains(s, "draft"):
		return "badge-info"
	default:
		return "badge-muted"
	}
}

func preferredChange(changes []map[string]any, preferred string) map[string]any {
	for _, ch := range changes {
		if str(get(ch, "changeNumber")) == preferred {
			return ch
		}
	}
	if len(changes) > 0 {
		return changes[0]
	}
	return nil
}

func recentChanges(items []map[string]any) []map[string]any {
	const limit = 5
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func buildUIStats(changes []map[string]any, apps []map[string]any, evidence []map[string]any) map[string]any {
	stats := map[string]any{"applications": len(apps), "changes": len(changes), "completed": 0, "running": 0, "failed": 0, "evidence": len(evidence)}
	for _, ch := range changes {
		status := strings.ToLower(str(get(ch, "status")) + " " + str(get(ch, "runtimeStatus")))
		if strings.Contains(status, "closed") || strings.Contains(status, "succeeded") || strings.Contains(status, "evidencecollected") {
			stats["completed"] = stats["completed"].(int) + 1
		}
		if strings.Contains(status, "running") || strings.Contains(status, "executing") {
			stats["running"] = stats["running"].(int) + 1
		}
		if strings.Contains(status, "failed") || strings.Contains(status, "degraded") {
			stats["failed"] = stats["failed"].(int) + 1
		}
	}
	return stats
}
func latestEvidence(items []map[string]any) map[string]any {
	for _, item := range items {
		if kube := kubeSummary(item); len(kube) > 0 {
			return item
		}
	}
	if len(items) == 0 {
		return nil
	}
	return items[0]
}

func latestRuntimeStatusFromEvents(events []map[string]any) string {
	for i := len(events) - 1; i >= 0; i-- {
		if step := eventStep(events[i]); step != "" {
			return step
		}
	}
	return ""
}
func kubeSummary(ev map[string]any) map[string]any {
	payload, _ := get(ev, "payload").(map[string]any)
	kube, _ := payload["kubernetes"].(map[string]any)
	if kube == nil {
		return map[string]any{}
	}
	return kube
}
func diagnosticsSummary(ev map[string]any) map[string]any {
	payload, _ := get(ev, "payload").(map[string]any)
	diagnostics, _ := payload["diagnostics"].(map[string]any)
	if diagnostics == nil {
		return map[string]any{}
	}
	return diagnostics
}
func eventStep(ev map[string]any) string {
	payload, _ := get(ev, "payload").(map[string]any)
	if payload != nil {
		if step := str(payload["step"]); step != "" {
			return step
		}
	}
	return str(get(ev, "eventType"))
}

func uiAction(name string, label string, description string, primary bool) map[string]any {
	return map[string]any{"name": name, "label": label, "description": description, "primary": primary}
}

func allUIActions() []map[string]any {
	return []map[string]any{
		uiAction("validate", "Validate", "Start a Tekton validation PipelineRun.", true),
		uiAction("check-validation", "Check Validation", "Poll latest Tekton PipelineRun result.", false),
		uiAction("check-deployment", "Check Deployment", "Check Argo CD sync and health state.", false),
		uiAction("collect-evidence", "Collect Evidence", "Collect post-deployment Kubernetes/OpenShift evidence.", false),
		uiAction("create-branch", "Create Branch", "Create the Git change branch.", false),
		uiAction("update-files", "Update Files", "Commit generated GitOps files on the change branch.", false),
		uiAction("open-merge-request", "Open Merge Request", "Open the GitLab merge request.", false),
		uiAction("merge-request", "Merge Request", "Merge the approved GitLab merge request.", false),
	}
}

func recommendedActions(change map[string]any) []map[string]any {
	runtime := strings.ToLower(strings.TrimSpace(str(get(change, "runtimeStatus"))))
	lifecycle := strings.ToLower(strings.TrimSpace(str(get(change, "status"))))
	switch runtime {
	case "":
		if lifecycle == "draft" || lifecycle == "" {
			return []map[string]any{uiAction("validate", "Validate", "Start validation for this draft change.", true), uiAction("create-branch", "Create Branch", "Prepare the Git branch for the change.", false)}
		}
	case "validationrunning":
		return []map[string]any{uiAction("check-validation", "Check Validation", "Read the latest Tekton validation result.", true)}
	case "validationsucceeded":
		return []map[string]any{uiAction("create-branch", "Create Branch", "Create or verify the Git change branch.", true), uiAction("update-files", "Update Files", "Generate and commit GitOps files.", false)}
	case "validationfailed":
		return []map[string]any{uiAction("validate", "Validate", "Retry Tekton validation after remediation.", true)}
	case "branchcreated":
		return []map[string]any{uiAction("update-files", "Update Files", "Commit generated GitOps files on the change branch.", true)}
	case "commitcreated":
		return []map[string]any{uiAction("open-merge-request", "Open Merge Request", "Open the merge request for review.", true)}
	case "mergerequestopened":
		return []map[string]any{uiAction("merge-request", "Merge Request", "Merge the GitLab MR when governance allows it.", true)}
	case "mergerequestmerged":
		return []map[string]any{uiAction("check-deployment", "Check Deployment", "Verify Argo CD sync and application health.", true)}
	case "deploymentprogressing", "deploymentoutofsync", "deploymentdegraded", "deploymentunknown":
		return []map[string]any{uiAction("check-deployment", "Check Deployment", "Re-check Argo CD deployment status.", true)}
	case "deploymentsyncedhealthy":
		return []map[string]any{uiAction("collect-evidence", "Collect Evidence", "Collect post-deployment runtime evidence.", true)}
	case "evidencecollected":
		return []map[string]any{uiAction("check-deployment", "Check Deployment", "Re-check the deployed application if needed.", false)}
	}
	return []map[string]any{}
}

func advancedActions(change map[string]any) []map[string]any {
	recommended := recommendedActions(change)
	recommendedNames := map[string]bool{}
	for _, action := range recommended {
		recommendedNames[str(action["name"])] = true
	}
	advanced := make([]map[string]any, 0)
	for _, action := range allUIActions() {
		if !recommendedNames[str(action["name"])] {
			advanced = append(advanced, action)
		}
	}
	return advanced
}

func jsonPretty(v any) string { raw, _ := json.MarshalIndent(v, "", "  "); return string(raw) }

const uiTemplate = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>{{.Title}} - DevOps Control Plane</title><style>
:root{--nav:#071b33;--nav2:#0b2646;--blue:#2563eb;--text:#0f172a;--muted:#64748b;--line:#e2e8f0;--bg:#f8fafc;--card:#fff}*{box-sizing:border-box}body{margin:0;font-family:Inter,Segoe UI,Roboto,Arial,sans-serif;background:var(--bg);color:var(--text)}a{color:#075eea;text-decoration:none}.app{display:flex;min-height:100vh}.sidebar{width:280px;background:linear-gradient(180deg,var(--nav),var(--nav2));color:white;padding:24px 16px;display:flex;flex-direction:column;gap:22px;position:fixed;inset:0 auto 0 0}.brand{display:flex;align-items:center;gap:12px;font-weight:800;font-size:18px}.brand-icon{width:36px;height:36px;border:2px solid #3b82f6;border-radius:10px;display:grid;place-items:center;color:#60a5fa}.nav-group{border-top:1px solid rgba(255,255,255,.12);padding-top:14px}.nav a{display:flex;align-items:center;gap:12px;color:#dbeafe;padding:12px 14px;border-radius:8px;margin:4px 0;font-weight:600}.nav a.active,.nav a:hover{background:#1d4ed8;color:#fff}.sys{margin-top:auto;border:1px solid rgba(255,255,255,.16);border-radius:10px;padding:14px;background:rgba(255,255,255,.03)}.sys h4{margin:0 0 12px}.sys-row{display:flex;justify-content:space-between;font-size:13px;margin:10px 0}.dot{display:inline-block;width:9px;height:9px;border-radius:50%;background:#22c55e;margin-right:8px}.version{font-size:12px;color:#cbd5e1}.main{margin-left:280px;width:calc(100% - 280px)}.topbar{height:88px;padding:20px 28px;border-bottom:1px solid var(--line);background:white;display:flex;align-items:center;justify-content:space-between}.title h1{margin:0;font-size:25px}.title p{margin:6px 0 0;color:var(--muted)}.user{display:flex;align-items:center;gap:18px}.select,.avatar{border:1px solid #cbd5e1;border-radius:8px;background:white;padding:10px 16px}.avatar{width:42px;height:42px;display:grid;place-items:center;border-radius:50%;font-weight:700;color:#475569}.content{padding:24px 28px}.cards{display:grid;grid-template-columns:repeat(5,1fr);gap:16px;margin-bottom:20px}.card{background:var(--card);border:1px solid var(--line);border-radius:10px;box-shadow:0 8px 20px rgba(15,23,42,.04)}.metric{padding:18px;display:flex;gap:16px;align-items:center}.metric .icon{width:46px;height:46px;border-radius:50%;display:grid;place-items:center;font-weight:800}.kpi-title-line{display:flex;align-items:center;gap:.45rem;white-space:nowrap}.kpi-title{color:var(--text);font-size:16px}.kpi-counter{display:inline-flex;align-items:center;justify-content:center;min-width:1.55rem;height:1.75rem;padding:0 .35rem;border:1px solid #334155;border-radius:.25rem;background:#f8fafc;color:#0f172a;font-size:1.35rem;font-weight:700;line-height:1}.metric>div>span{display:block;color:var(--muted);font-size:13px;margin-top:6px}.grid{display:grid;grid-template-columns:1.05fr 1.55fr 1.35fr;gap:18px}.panel{padding:0}.panel h3{font-size:16px;margin:0;padding:18px;border-bottom:1px solid var(--line)}.list{padding:0;margin:0;list-style:none}.list li{display:flex;align-items:center;justify-content:space-between;padding:13px 18px;border-bottom:1px solid var(--line)}.small{font-size:13px;color:var(--muted)}.badge{border-radius:7px;padding:4px 9px;font-size:12px;font-weight:700}.badge-ok{background:#dcfce7;color:#15803d;border:1px solid #86efac}.badge-bad{background:#fee2e2;color:#b91c1c;border:1px solid #fecaca}.badge-warn{background:#fef3c7;color:#b45309;border:1px solid #fde68a}.badge-info{background:#dbeafe;color:#1d4ed8;border:1px solid #bfdbfe}.badge-muted{background:#f1f5f9;color:#475569}.detail{padding:18px}.detail-head{display:flex;justify-content:space-between;align-items:center;margin-bottom:18px}.kv{display:grid;grid-template-columns:1fr 1fr;gap:12px 24px}.kv .label{color:var(--muted);font-size:13px;margin-bottom:4px}.section{border-top:1px solid var(--line);margin-top:18px;padding-top:18px}.actions{display:flex;flex-wrap:wrap;gap:10px}.action-groups{display:grid;gap:14px}.action-card{border:1px solid var(--line);border-radius:10px;padding:10px;background:#f8fafc;max-width:260px}.action-card form{margin:0 0 6px}.action-desc{line-height:1.35}.btn{border:1px solid #2563eb;color:#1d4ed8;background:white;padding:10px 14px;border-radius:8px;font-weight:700;cursor:pointer}.btn.primary{background:#2563eb;color:white}.timeline{padding:18px}.step{display:flex;gap:12px;margin:0 0 17px}.circle{width:18px;height:18px;border:2px solid #94a3b8;border-radius:50%;margin-top:2px}.circle.done{background:#16a34a;border-color:#16a34a}.evidence{padding:18px}.ev-row{display:flex;justify-content:space-between;gap:12px;padding:13px 0;border-bottom:1px solid var(--line)}.table{width:100%;border-collapse:collapse;background:white}.table th,.table td{border-bottom:1px solid var(--line);padding:12px;text-align:left}.table th{font-size:12px;color:#475569;text-transform:uppercase}.full{grid-column:1 / -1}.json{white-space:pre-wrap;background:#0f172a;color:#dbeafe;border-radius:8px;padding:14px;max-height:360px;overflow:auto}.evidence-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:12px;margin-top:12px}.evidence-card{border:1px solid var(--line);border-radius:10px;padding:12px;background:#f8fafc}.evidence-card h4{margin:0 0 8px;font-size:14px}.evidence-kv{display:flex;justify-content:space-between;gap:12px;padding:4px 0;color:#334155;font-size:13px}.pod-list{margin:0;padding-left:18px}.pod-list li{margin:5px 0;color:#334155;font-size:13px}details{margin-top:12px}summary{cursor:pointer;color:#1d4ed8;font-weight:700;font-size:13px}.alert{padding:12px 14px;border-radius:8px;margin-bottom:16px;font-weight:700}.alert-ok{background:#dcfce7;color:#15803d;border:1px solid #86efac}.alert-error{background:#fee2e2;color:#b91c1c;border:1px solid #fecaca}.footer{text-align:center;color:var(--muted);font-size:12px;margin-top:28px}@media(max-width:1300px){.cards{grid-template-columns:repeat(2,1fr)}.grid{grid-template-columns:1fr}.sidebar{position:static;width:100%}.main{margin-left:0;width:100%}.app{display:block}}
</style></head><body><div class="app"><aside class="sidebar"><div class="brand"><div class="brand-icon">☁</div><div>DevOps Control Plane</div></div><nav class="nav"><a class="{{if eq .Active "dashboard"}}active{{end}}" href="/">▣ Dashboard</a><a class="{{if eq .Active "applications"}}active{{end}}" href="/ui/applications">▧ Applications</a><div class="nav-group"><a class="{{if eq .Active "changes"}}active{{end}}" href="/ui/changes">◌ Change Requests</a><a href="/ui/changes">All changes</a><a href="/ui/changes-api">Changes API</a></div><div class="nav-group"><a href="/ui/changes/CHG-2026-0005/evidence">▤ Evidence</a><a href="/ui/changes/CHG-2026-0005/events">☷ Audit log</a></div><div class="nav-group"><a class="{{if eq .Active "settings"}}active{{end}}" href="/ui/settings">⚙ Settings</a></div></nav><div class="sys"><h4>System status</h4><div class="sys-row"><span><i class="dot"></i>API</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>Database</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>GitLab</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>Tekton</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>Argo CD</span><b>OK</b></div></div><div class="version">DevOps Control Plane<br>v0.1.0</div></aside><main class="main"><header class="topbar"><div class="title"><h1>{{.Title}}</h1><p>{{.Subtitle}}</p></div><div class="user"><div class="select">Environment<br><b>dev</b></div><div class="avatar">A</div><b>admin</b></div></header><section class="content">{{if .Flash}}<div class="alert alert-ok">{{.Flash}}</div>{{end}}{{if .ActionError}}<div class="alert alert-error">{{.ActionError}}</div>{{end}}{{if .Error}}<div class="card detail"><b>Error:</b> {{.Error}}</div>{{else}}{{if eq .Mode "changeEvents"}}{{template "changeEventsPage" .}}{{else if eq .Mode "changeEvidence"}}{{template "changeEvidencePage" .}}{{else if eq .Active "settings"}}{{template "settingsPage" .}}{{else if eq .Mode "changesAPI"}}{{template "changesAPIPage" .}}{{else if eq .Active "changes"}}{{if .SelectedChange}}{{template "changeDetail" .}}{{else}}{{template "changesList" .}}{{end}}{{else if eq .Active "applications"}}{{if .SelectedApplication}}{{template "applicationDetail" .}}{{else}}{{template "applicationsList" .}}{{end}}{{else}}{{template "dashboard" .}}{{end}}{{end}}<div class="footer">© 2026 DevOps Control Plane <span style="float:right">v0.1.0</span></div></section></main></div></body></html>
{{define "dashboard"}}<div class="cards"><div class="card metric"><div class="icon" style="background:#dbeafe;color:#2563eb">□</div><div><div class="kpi-title-line"><span class="kpi-title">Applications</span><b class="kpi-counter">{{get .Stats "applications"}}</b></div><span>Managed total</span></div></div><div class="card metric"><div class="icon" style="background:#dcfce7;color:#16a34a">✓</div><div><div class="kpi-title-line"><span class="kpi-title">Completed changes</span><b class="kpi-counter">{{get .Stats "completed"}}</b></div><span>Last 30 days</span></div></div><div class="card metric"><div class="icon" style="background:#fef3c7;color:#d97706">◷</div><div><div class="kpi-title-line"><span class="kpi-title">Running changes</span><b class="kpi-counter">{{get .Stats "running"}}</b></div><span>Currently running</span></div></div><div class="card metric"><div class="icon" style="background:#fee2e2;color:#dc2626">!</div><div><div class="kpi-title-line"><span class="kpi-title">Failed changes</span><b class="kpi-counter">{{get .Stats "failed"}}</b></div><span>Last 30 days</span></div></div><div class="card metric"><div class="icon" style="background:#ede9fe;color:#7c3aed">▤</div><div><div class="kpi-title-line"><span class="kpi-title">Collected evidence</span><b class="kpi-counter">{{get .Stats "evidence"}}</b></div><span>For selected change</span></div></div></div><div class="grid"><div><div class="card panel"><h3>Applications <a style="float:right;font-size:13px" href="/ui/applications">View all</a></h3><ul class="list">{{range .Applications}}<li><div><a href="/ui/applications/{{get . "name"}}"><b>{{get . "name"}}</b></a><div class="small">{{get . "targetNamespace"}}</div></div><span class="badge {{badgeClass (get . "healthStatus")}}">{{get . "healthStatus"}}</span></li>{{end}}</ul></div><div class="card panel" style="margin-top:18px"><h3>Recent changes <a style="float:right;font-size:13px" href="/ui/changes">View all</a></h3><ul class="list">{{range recentChanges .Changes}}<li><div><a href="/ui/changes/{{changeNumberOrID .}}"><b>{{changeNumberOrID .}}</b></a><div class="small">{{get . "applicationName"}} · Environment: {{get . "targetEnvironment"}} · Requested by: {{get . "requestedBy"}}</div></div><span class="badge {{badgeClass (get . "runtimeStatus")}}">{{get . "runtimeStatus"}}</span></li>{{end}}</ul></div></div><div>{{template "changeCard" .}}</div><div><div class="card timeline"><h3>Workflow Change</h3>{{range .Events}}<div class="step"><span class="circle done"></span><div><b>{{eventStep .}}</b><div class="small">{{get . "createdAt"}}</div></div></div>{{else}}<div class="small">No events available</div>{{end}}</div><div class="card evidence" style="margin-top:18px"><h3>Available evidence</h3>{{range .Evidence}}<div class="ev-row"><div><b>{{get . "name"}}</b><div class="small">{{get . "summary"}}</div></div><a href="/ui/changes/{{get . "changeNumber"}}/evidence">View</a></div>{{else}}<div class="small">No evidence available</div>{{end}}</div></div></div>{{end}}
{{define "changeCard"}}<div class="card detail">{{if .SelectedChange}}<div class="detail-head"><h3>Change Request: {{changeNumberOrID .SelectedChange}}</h3><span class="badge {{badgeClass (get .SelectedChange "runtimeStatus")}}">{{get .SelectedChange "runtimeStatus"}}</span></div><div class="kv"><div><div class="label">Application</div><b>{{get .SelectedChange "applicationName"}}</b></div><div><div class="label">Requester</div><b>{{get .SelectedChange "requestedBy"}}</b></div><div><div class="label">Environment</div><b>{{get .SelectedChange "targetEnvironment"}}</b></div><div><div class="label">Process lifecycle status</div><span class="badge {{badgeClass (get .SelectedChange "status")}}">{{get .SelectedChange "status"}}</span></div><div><div class="label">Technical runtime status</div><span class="badge {{badgeClass (get .SelectedChange "runtimeStatus")}}">{{get .SelectedChange "runtimeStatus"}}</span></div><div><div class="label">Created at</div><b>{{get .SelectedChange "createdAt"}}</b></div></div><div class="section"><h3>Status meaning</h3><p class="small">Process lifecycle status tracks the governance and approval state of the ChangeRequest. Technical runtime status tracks the latest automation, validation, deployment or evidence observation.</p></div><div class="section"><h3>Description</h3><p>{{get .SelectedChange "description"}}</p><div class="actions"><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}/evidence">View evidence</a><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}/events">View audit events</a></div></div><div class="section"><h3>Technical actions</h3><div class="action-groups"><div><div class="small" style="margin-bottom:8px">Recommended next actions</div><div class="actions">{{range recommendedActions .SelectedChange}}<div class="action-card"><form method="post" action="/ui/changes/{{changeNumberOrID $.SelectedChange}}/actions/{{get . "name"}}"><button class="btn {{if get . "primary"}}primary{{end}}">{{get . "label"}}</button></form><div class="small action-desc">{{get . "description"}}</div></div>{{else}}<div class="small">No recommended technical action for the current state.</div>{{end}}</div></div><details><summary>Advanced/manual actions</summary><div class="actions" style="margin-top:10px">{{range advancedActions .SelectedChange}}<div class="action-card"><form method="post" action="/ui/changes/{{changeNumberOrID $.SelectedChange}}/actions/{{get . "name"}}"><button class="btn">{{get . "label"}}</button></form><div class="small action-desc">{{get . "description"}}</div></div>{{end}}</div></details></div></div>{{with latestEvidence .Evidence}}<div class="section"><h3>Latest runtime evidence</h3><div class="small">{{get . "summary"}}</div>{{with diagnosticsSummary .}}<div class="evidence-card" style="margin-top:12px"><h4>Deployment diagnostics</h4><div class="evidence-kv"><span>Summary</span><b>{{get . "summary"}}</b></div><div class="evidence-kv"><span>Argo CD synced</span><b>{{get . "argocdSynced"}}</b></div><div class="evidence-kv"><span>Argo CD healthy</span><b>{{get . "argocdHealthy"}}</b></div><div class="evidence-kv"><span>Deployment ready</span><b>{{get . "deploymentReady"}}</b></div><div class="evidence-kv"><span>Replicas</span><b>{{get . "readyReplicas"}}</b></div><div class="evidence-kv"><span>Pods</span><b>{{get . "podsReady"}}</b></div><div class="evidence-kv"><span>Restarts</span><b>{{get . "totalRestarts"}}</b></div><div class="evidence-kv"><span>Service available</span><b>{{get . "serviceAvailable"}}</b></div><div class="evidence-kv"><span>Route available</span><b>{{get . "routeAvailable"}}</b></div>{{with get . "warnings"}}<div class="small" style="margin-top:8px"><b>Warnings</b><ul class="pod-list">{{range .}}<li>{{.}}</li>{{end}}</ul></div>{{end}}</div>{{end}}<div class="evidence-grid">{{with get (kubeSummary .) "deployment"}}<div class="evidence-card"><h4>Deployment</h4><div class="evidence-kv"><span>Name</span><b>{{get . "name"}}</b></div><div class="evidence-kv"><span>Namespace</span><b>{{get . "namespace"}}</b></div><div class="evidence-kv"><span>Ready</span><b>{{get . "readyReplicas"}}/{{get . "desiredReplicas"}}</b></div><div class="evidence-kv"><span>Available</span><b>{{get . "availableReplicas"}}</b></div><div class="evidence-kv"><span>Updated</span><b>{{get . "updatedReplicas"}}</b></div></div>{{end}}{{with get (kubeSummary .) "service"}}<div class="evidence-card"><h4>Service</h4><div class="evidence-kv"><span>Name</span><b>{{get . "name"}}</b></div><div class="evidence-kv"><span>Type</span><b>{{get . "type"}}</b></div><div class="evidence-kv"><span>Cluster IP</span><b>{{get . "clusterIP"}}</b></div></div>{{end}}{{with get (kubeSummary .) "route"}}<div class="evidence-card"><h4>Route</h4><div class="evidence-kv"><span>Host</span><b>{{get . "host"}}</b></div><div class="evidence-kv"><span>TLS</span><b>{{get . "tlsTermination"}}</b></div><div class="evidence-kv"><span>To</span><b>{{get . "to"}}</b></div></div>{{end}}{{with get (kubeSummary .) "pods"}}<div class="evidence-card"><h4>Pods</h4><ul class="pod-list">{{range .}}<li><b>{{get . "name"}}</b> - {{get . "phase"}}, ready={{get . "ready"}}, restarts={{get . "restartCount"}}, node={{get . "nodeName"}}</li>{{end}}</ul></div>{{end}}</div><details><summary>View raw deployment evidence</summary><pre class="json">{{jsonPretty .}}</pre></details></div>{{end}}{{else}}<p>No ChangeRequest available.</p>{{end}}</div>{{end}}
{{define "changesList"}}<div class="card panel full"><h3>Change Requests</h3><table class="table"><thead><tr><th>Change</th><th>Application</th><th>Requested by</th><th>Environment</th><th>Process lifecycle</th><th>Technical runtime</th><th>Action</th></tr></thead><tbody>{{range .Changes}}<tr><td><b>{{changeNumberOrID .}}</b></td><td>{{get . "applicationName"}}</td><td>{{get . "requestedBy"}}</td><td>{{get . "targetEnvironment"}}</td><td><span class="badge {{badgeClass (get . "status")}}">{{get . "status"}}</span></td><td><span class="badge {{badgeClass (get . "runtimeStatus")}}">{{get . "runtimeStatus"}}</span></td><td><a href="/ui/changes/{{changeNumberOrID .}}">Open</a></td></tr>{{end}}</tbody></table></div>{{end}}

{{define "settingsPage"}}<div class="grid"><div class="full"><div class="card detail"><div class="detail-head"><h3>Settings</h3><span class="badge badge-info">MVP</span></div><div class="kv"><div><div class="label">Readiness</div><b><a href="/readyz">/readyz</a></b><div class="small">Technical readiness endpoint for API and database checks.</div></div><div><div class="label">Environment</div><b>dev</b><div class="small">Static display only. Multi-environment selection is not implemented yet.</div></div><div><div class="label">User</div><b>admin</b><div class="small">Static placeholder only. Authentication and authorization are planned for Phase 9.3.</div></div><div><div class="label">Version</div><b>v0.1.0</b><div class="small">UI MVP for lab validation.</div></div></div><div class="section"><h3>Production readiness notes</h3><ul class="pod-list"><li>AuthN/AuthZ is intentionally deferred to Phase 9.3.</li><li>Environment selector is intentionally static until multi-environment support is implemented.</li><li>Readiness remains available as JSON at <a href="/readyz">/readyz</a>.</li><li>OpenShift deployment, TLS, secrets and RBAC will be handled in Phase 8 and Phase 9.</li></ul></div></div></div></div>{{end}}
{{define "applicationsList"}}<div class="card panel full"><h3>Applications</h3><table class="table"><thead><tr><th>Name</th><th>Namespace</th><th>Sync</th><th>Health</th><th>Revision</th><th>Action</th></tr></thead><tbody>{{range .Applications}}<tr><td><b>{{get . "name"}}</b></td><td>{{get . "targetNamespace"}}</td><td><span class="badge {{badgeClass (get . "syncStatus")}}">{{get . "syncStatus"}}</span></td><td><span class="badge {{badgeClass (get . "healthStatus")}}">{{get . "healthStatus"}}</span></td><td>{{short (get . "revision")}}</td><td><a href="/ui/applications/{{get . "name"}}">Open</a></td></tr>{{else}}<tr><td colspan="6" class="small">No applications available.</td></tr>{{end}}</tbody></table></div>{{end}}
{{define "applicationDetail"}}<div class="grid"><div class="full"><div class="card detail"><div class="detail-head"><h3>Application: {{get .SelectedApplication "name"}}</h3><span class="badge {{badgeClass (get .SelectedApplication "healthStatus")}}">{{get .SelectedApplication "healthStatus"}}</span></div><div class="kv"><div><div class="label">Namespace</div><b>{{get .SelectedApplication "targetNamespace"}}</b></div><div><div class="label">Sync status</div><span class="badge {{badgeClass (get .SelectedApplication "syncStatus")}}">{{get .SelectedApplication "syncStatus"}}</span></div><div><div class="label">Health status</div><span class="badge {{badgeClass (get .SelectedApplication "healthStatus")}}">{{get .SelectedApplication "healthStatus"}}</span></div><div><div class="label">Revision</div><b>{{short (get .SelectedApplication "revision")}}</b></div><div><div class="label">Repository</div><b>{{get .SelectedApplication "repoURL"}}</b></div><div><div class="label">Path</div><b>{{get .SelectedApplication "path"}}</b></div></div><div class="section"><h3>Runtime summary</h3><pre class="json">{{jsonPretty .Runtime}}</pre></div></div></div><div class="card panel full"><h3>Resources</h3><table class="table"><thead><tr><th>Kind</th><th>Name</th><th>Namespace</th><th>Status</th></tr></thead><tbody>{{range .Resources}}<tr><td>{{get . "kind"}}</td><td>{{get . "name"}}</td><td>{{get . "namespace"}}</td><td><span class="badge {{badgeClass (get . "status")}}">{{get . "status"}}</span></td></tr>{{else}}<tr><td colspan="4" class="small">No resource details available.</td></tr>{{end}}</tbody></table></div><div class="card panel full"><h3>Deployment history</h3><table class="table"><thead><tr><th>Revision</th><th>Status</th><th>Deployed at</th></tr></thead><tbody>{{range .History}}<tr><td>{{short (get . "revision")}}</td><td><span class="badge {{badgeClass (get . "status")}}">{{get . "status"}}</span></td><td>{{get . "deployedAt"}}</td></tr>{{else}}<tr><td colspan="3" class="small">No deployment history available.</td></tr>{{end}}</tbody></table></div></div>{{end}}

{{define "changeEvidencePage"}}<div class="grid"><div class="full"><div class="card detail"><div class="detail-head"><h3>Evidence for {{changeNumberOrID .SelectedChange}}</h3><span class="badge {{badgeClass (get .SelectedChange "runtimeStatus")}}">{{get .SelectedChange "runtimeStatus"}}</span></div><div class="actions"><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}">Back to change</a><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}/events">View audit events</a></div></div></div><div class="card panel full"><h3>Collected evidence</h3><table class="table"><thead><tr><th>Name</th><th>Type</th><th>Summary</th><th>Sanitized</th><th>Created</th></tr></thead><tbody>{{range .Evidence}}<tr><td><b>{{get . "name"}}</b></td><td>{{get . "evidenceType"}}</td><td>{{get . "summary"}}</td><td><span class="badge {{badgeClass (get . "sanitized")}}">{{get . "sanitized"}}</span></td><td>{{get . "createdAt"}}</td></tr>{{else}}<tr><td colspan="5" class="small">No evidence available.</td></tr>{{end}}</tbody></table></div>{{range .Evidence}}<div class="card panel full"><h3>{{get . "name"}}</h3><pre class="json">{{jsonPretty .}}</pre></div>{{end}}</div>{{end}}

{{define "changesAPIPage"}}<div class="grid"><div class="full"><div class="card detail"><div class="detail-head"><h3>Changes API</h3><span class="badge badge-info">UI wrapper</span></div><p>This page shows the Change Requests API data inside the web UI, so users do not land directly on a raw JSON-only browser page.</p><div class="actions"><a class="btn primary" href="/">Back to dashboard</a><a class="btn" href="/ui/changes">Back to changes</a><a class="btn" href="/api/v1/changes">Open raw JSON API</a></div></div></div><div class="card panel full"><h3>Change Requests API preview</h3><pre class="json">{{jsonPretty .Changes}}</pre></div></div>{{end}}
{{define "changeEventsPage"}}<div class="grid"><div class="full"><div class="card detail"><div class="detail-head"><h3>Audit events for {{changeNumberOrID .SelectedChange}}</h3><span class="badge {{badgeClass (get .SelectedChange "status")}}">{{get .SelectedChange "status"}}</span></div><div class="actions"><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}">Back to change</a><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}/evidence">View evidence</a></div></div></div><div class="card panel full"><h3>Audit trail</h3><table class="table"><thead><tr><th>Event</th><th>Step</th><th>Previous</th><th>New</th><th>Source</th><th>Created</th></tr></thead><tbody>{{range .Events}}<tr><td>{{get . "eventType"}}</td><td>{{eventStep .}}</td><td>{{get . "previousStatus"}}</td><td>{{get . "newStatus"}}</td><td>{{get . "source"}}</td><td>{{get . "createdAt"}}</td></tr>{{else}}<tr><td colspan="6" class="small">No audit events available.</td></tr>{{end}}</tbody></table></div>{{range .Events}}<div class="card panel full"><h3>{{get . "eventType"}} - {{eventStep .}}</h3><pre class="json">{{jsonPretty .}}</pre></div>{{end}}</div>{{end}}
{{define "changeDetail"}}<div class="grid"><div class="full">{{template "changeCard" .}}</div><div class="card panel full"><h3>Audit events</h3><table class="table"><thead><tr><th>Event</th><th>Step</th><th>Created</th></tr></thead><tbody>{{range .Events}}<tr><td>{{get . "eventType"}}</td><td>{{eventStep .}}</td><td>{{get . "createdAt"}}</td></tr>{{end}}</tbody></table></div></div>{{end}}`
