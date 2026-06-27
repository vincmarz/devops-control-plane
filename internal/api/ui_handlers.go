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
	Title          string
	Subtitle       string
	Active         string
	Changes        []map[string]any
	Applications   []map[string]any
	SelectedChange map[string]any
	Events         []map[string]any
	Evidence       []map[string]any
	Stats          map[string]any
	Flash          string
	ActionError    string
	Error          string
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

func (h *Handler) uiChangeDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/ui/changes/"), "/")
	if id == "" {
		h.uiChanges(w, r)
		return
	}
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
	funcs := template.FuncMap{"get": get, "str": str, "short": short, "badgeClass": badgeClass, "latestEvidence": latestEvidence, "kubeSummary": kubeSummary, "eventStep": eventStep, "jsonPretty": jsonPretty, "changeNumberOrID": changeNumberOrID}
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
func eventStep(ev map[string]any) string {
	payload, _ := get(ev, "payload").(map[string]any)
	if payload != nil {
		if step := str(payload["step"]); step != "" {
			return step
		}
	}
	return str(get(ev, "eventType"))
}
func jsonPretty(v any) string { raw, _ := json.MarshalIndent(v, "", "  "); return string(raw) }

const uiTemplate = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>{{.Title}} - DevOps Control Plane</title><style>
:root{--nav:#071b33;--nav2:#0b2646;--blue:#2563eb;--text:#0f172a;--muted:#64748b;--line:#e2e8f0;--bg:#f8fafc;--card:#fff}*{box-sizing:border-box}body{margin:0;font-family:Inter,Segoe UI,Roboto,Arial,sans-serif;background:var(--bg);color:var(--text)}a{color:#075eea;text-decoration:none}.app{display:flex;min-height:100vh}.sidebar{width:280px;background:linear-gradient(180deg,var(--nav),var(--nav2));color:white;padding:24px 16px;display:flex;flex-direction:column;gap:22px;position:fixed;inset:0 auto 0 0}.brand{display:flex;align-items:center;gap:12px;font-weight:800;font-size:18px}.brand-icon{width:36px;height:36px;border:2px solid #3b82f6;border-radius:10px;display:grid;place-items:center;color:#60a5fa}.nav-group{border-top:1px solid rgba(255,255,255,.12);padding-top:14px}.nav a{display:flex;align-items:center;gap:12px;color:#dbeafe;padding:12px 14px;border-radius:8px;margin:4px 0;font-weight:600}.nav a.active,.nav a:hover{background:#1d4ed8;color:#fff}.sys{margin-top:auto;border:1px solid rgba(255,255,255,.16);border-radius:10px;padding:14px;background:rgba(255,255,255,.03)}.sys h4{margin:0 0 12px}.sys-row{display:flex;justify-content:space-between;font-size:13px;margin:10px 0}.dot{display:inline-block;width:9px;height:9px;border-radius:50%;background:#22c55e;margin-right:8px}.version{font-size:12px;color:#cbd5e1}.main{margin-left:280px;width:calc(100% - 280px)}.topbar{height:88px;padding:20px 28px;border-bottom:1px solid var(--line);background:white;display:flex;align-items:center;justify-content:space-between}.title h1{margin:0;font-size:25px}.title p{margin:6px 0 0;color:var(--muted)}.user{display:flex;align-items:center;gap:18px}.select,.avatar{border:1px solid #cbd5e1;border-radius:8px;background:white;padding:10px 16px}.avatar{width:42px;height:42px;display:grid;place-items:center;border-radius:50%;font-weight:700;color:#475569}.content{padding:24px 28px}.cards{display:grid;grid-template-columns:repeat(5,1fr);gap:16px;margin-bottom:20px}.card{background:var(--card);border:1px solid var(--line);border-radius:10px;box-shadow:0 8px 20px rgba(15,23,42,.04)}.metric{padding:18px;display:flex;gap:16px;align-items:center}.metric .icon{width:46px;height:46px;border-radius:50%;display:grid;place-items:center;font-weight:800}.metric b{font-size:28px}.metric span{display:block;color:var(--muted);font-size:13px;margin-top:6px}.grid{display:grid;grid-template-columns:1.05fr 1.55fr 1.35fr;gap:18px}.panel{padding:0}.panel h3{font-size:16px;margin:0;padding:18px;border-bottom:1px solid var(--line)}.list{padding:0;margin:0;list-style:none}.list li{display:flex;align-items:center;justify-content:space-between;padding:13px 18px;border-bottom:1px solid var(--line)}.small{font-size:13px;color:var(--muted)}.badge{border-radius:7px;padding:4px 9px;font-size:12px;font-weight:700}.badge-ok{background:#dcfce7;color:#15803d;border:1px solid #86efac}.badge-bad{background:#fee2e2;color:#b91c1c;border:1px solid #fecaca}.badge-warn{background:#fef3c7;color:#b45309;border:1px solid #fde68a}.badge-info{background:#dbeafe;color:#1d4ed8;border:1px solid #bfdbfe}.badge-muted{background:#f1f5f9;color:#475569}.detail{padding:18px}.detail-head{display:flex;justify-content:space-between;align-items:center;margin-bottom:18px}.kv{display:grid;grid-template-columns:1fr 1fr;gap:12px 24px}.kv .label{color:var(--muted);font-size:13px;margin-bottom:4px}.section{border-top:1px solid var(--line);margin-top:18px;padding-top:18px}.actions{display:flex;flex-wrap:wrap;gap:10px}.btn{border:1px solid #2563eb;color:#1d4ed8;background:white;padding:10px 14px;border-radius:8px;font-weight:700;cursor:pointer}.btn.primary{background:#2563eb;color:white}.timeline{padding:18px}.step{display:flex;gap:12px;margin:0 0 17px}.circle{width:18px;height:18px;border:2px solid #94a3b8;border-radius:50%;margin-top:2px}.circle.done{background:#16a34a;border-color:#16a34a}.evidence{padding:18px}.ev-row{display:flex;justify-content:space-between;gap:12px;padding:13px 0;border-bottom:1px solid var(--line)}.table{width:100%;border-collapse:collapse;background:white}.table th,.table td{border-bottom:1px solid var(--line);padding:12px;text-align:left}.table th{font-size:12px;color:#475569;text-transform:uppercase}.full{grid-column:1 / -1}.json{white-space:pre-wrap;background:#0f172a;color:#dbeafe;border-radius:8px;padding:14px;max-height:360px;overflow:auto}.alert{padding:12px 14px;border-radius:8px;margin-bottom:16px;font-weight:700}.alert-ok{background:#dcfce7;color:#15803d;border:1px solid #86efac}.alert-error{background:#fee2e2;color:#b91c1c;border:1px solid #fecaca}.footer{text-align:center;color:var(--muted);font-size:12px;margin-top:28px}@media(max-width:1300px){.cards{grid-template-columns:repeat(2,1fr)}.grid{grid-template-columns:1fr}.sidebar{position:static;width:100%}.main{margin-left:0;width:100%}.app{display:block}}
</style></head><body><div class="app"><aside class="sidebar"><div class="brand"><div class="brand-icon">☁</div><div>DevOps Control Plane</div></div><nav class="nav"><a class="{{if eq .Active "dashboard"}}active{{end}}" href="/">▣ Dashboard</a><a href="/api/v1/applications">▧ Applications</a><div class="nav-group"><a class="{{if eq .Active "changes"}}active{{end}}" href="/ui/changes">◌ Change Requests</a><a href="/ui/changes">All changes</a><a href="/api/v1/changes">New change</a></div><div class="nav-group"><a href="/ui/changes/CHG-2026-0005">▤ Evidence</a><a href="/ui/changes/CHG-2026-0005">☷ Audit Log</a></div><div class="nav-group"><a href="/readyz">⚙ Settings</a></div></nav><div class="sys"><h4>System status</h4><div class="sys-row"><span><i class="dot"></i>API</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>Database</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>GitLab</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>Tekton</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>Argo CD</span><b>OK</b></div></div><div class="version">DevOps Control Plane<br>v0.1.0</div></aside><main class="main"><header class="topbar"><div class="title"><h1>{{.Title}}</h1><p>{{.Subtitle}}</p></div><div class="user"><div class="select">Environment<br><b>dev</b></div><div class="avatar">A</div><b>admin</b></div></header><section class="content">{{if .Flash}}<div class="alert alert-ok">{{.Flash}}</div>{{end}}{{if .ActionError}}<div class="alert alert-error">{{.ActionError}}</div>{{end}}{{if .Error}}<div class="card detail"><b>Error:</b> {{.Error}}</div>{{else}}{{if eq .Active "changes"}}{{if .SelectedChange}}{{template "changeDetail" .}}{{else}}{{template "changesList" .}}{{end}}{{else}}{{template "dashboard" .}}{{end}}{{end}}<div class="footer">© 2026 DevOps Control Plane <span style="float:right">v0.1.0</span></div></section></main></div></body></html>
{{define "dashboard"}}<div class="cards"><div class="card metric"><div class="icon" style="background:#dbeafe;color:#2563eb">□</div><div>Applications<b>{{get .Stats "applications"}}</b><span>Managed total</span></div></div><div class="card metric"><div class="icon" style="background:#dcfce7;color:#16a34a">✓</div><div>Completed changes<b>{{get .Stats "completed"}}</b><span>Last 30 days</span></div></div><div class="card metric"><div class="icon" style="background:#fef3c7;color:#d97706">◷</div><div>Running changes<b>{{get .Stats "running"}}</b><span>Currently running</span></div></div><div class="card metric"><div class="icon" style="background:#fee2e2;color:#dc2626">!</div><div>Failed changes<b>{{get .Stats "failed"}}</b><span>Last 30 days</span></div></div><div class="card metric"><div class="icon" style="background:#ede9fe;color:#7c3aed">▤</div><div>Collected evidence<b>{{get .Stats "evidence"}}</b><span>For selected change</span></div></div></div><div class="grid"><div><div class="card panel"><h3>Applications <a style="float:right;font-size:13px" href="/api/v1/applications">View all</a></h3><ul class="list">{{range .Applications}}<li><div><a href="/api/v1/applications/{{get . "name"}}"><b>{{get . "name"}}</b></a><div class="small">{{get . "targetNamespace"}}</div></div><span class="badge {{badgeClass (get . "healthStatus")}}">{{get . "healthStatus"}}</span></li>{{end}}</ul></div><div class="card panel" style="margin-top:18px"><h3>Recent changes <a style="float:right;font-size:13px" href="/ui/changes">View all</a></h3><ul class="list">{{range .Changes}}<li><div><a href="/ui/changes/{{changeNumberOrID .}}"><b>{{changeNumberOrID .}}</b></a><div class="small">{{get . "applicationName"}}</div></div><span class="badge {{badgeClass (get . "runtimeStatus")}}">{{get . "runtimeStatus"}}</span></li>{{end}}</ul></div></div><div>{{template "changeCard" .}}</div><div><div class="card timeline"><h3>Workflow Change</h3>{{range .Events}}<div class="step"><span class="circle done"></span><div><b>{{eventStep .}}</b><div class="small">{{get . "createdAt"}}</div></div></div>{{else}}<div class="small">No events available</div>{{end}}</div><div class="card evidence" style="margin-top:18px"><h3>Available evidence</h3>{{range .Evidence}}<div class="ev-row"><div><b>{{get . "name"}}</b><div class="small">{{get . "summary"}}</div></div><a href="/api/v1/changes/{{get . "changeNumber"}}/evidence/{{get . "evidenceType"}}">View</a></div>{{else}}<div class="small">No evidence available</div>{{end}}</div></div></div>{{end}}
{{define "changeCard"}}<div class="card detail">{{if .SelectedChange}}<div class="detail-head"><h3>Change Request: {{changeNumberOrID .SelectedChange}}</h3><span class="badge {{badgeClass (get .SelectedChange "runtimeStatus")}}">{{get .SelectedChange "runtimeStatus"}}</span></div><div class="kv"><div><div class="label">Application</div><b>{{get .SelectedChange "applicationName"}}</b></div><div><div class="label">Requester</div><b>{{get .SelectedChange "requestedBy"}}</b></div><div><div class="label">Environment</div><b>{{get .SelectedChange "targetEnvironment"}}</b></div><div><div class="label">Lifecycle status</div><span class="badge {{badgeClass (get .SelectedChange "status")}}">{{get .SelectedChange "status"}}</span></div><div><div class="label">Runtime Status</div><span class="badge {{badgeClass (get .SelectedChange "runtimeStatus")}}">{{get .SelectedChange "runtimeStatus"}}</span></div><div><div class="label">Created at</div><b>{{get .SelectedChange "createdAt"}}</b></div></div><div class="section"><h3>Description</h3><p>{{get .SelectedChange "description"}}</p></div><div class="section"><h3>Technical actions</h3><div class="actions"><form method="post" action="/ui/changes/{{changeNumberOrID .SelectedChange}}/actions/validate"><button class="btn primary">Validate</button></form><form method="post" action="/ui/changes/{{changeNumberOrID .SelectedChange}}/actions/check-validation"><button class="btn">Check Validation</button></form><form method="post" action="/ui/changes/{{changeNumberOrID .SelectedChange}}/actions/check-deployment"><button class="btn">Check Deployment</button></form><form method="post" action="/ui/changes/{{changeNumberOrID .SelectedChange}}/actions/collect-evidence"><button class="btn">Collect Evidence</button></form><form method="post" action="/ui/changes/{{changeNumberOrID .SelectedChange}}/actions/create-branch"><button class="btn">Create Branch</button></form><form method="post" action="/ui/changes/{{changeNumberOrID .SelectedChange}}/actions/update-files"><button class="btn">Update Files</button></form><form method="post" action="/ui/changes/{{changeNumberOrID .SelectedChange}}/actions/open-merge-request"><button class="btn">Open Merge Request</button></form><form method="post" action="/ui/changes/{{changeNumberOrID .SelectedChange}}/actions/merge-request"><button class="btn">Merge Request</button></form></div></div>{{with latestEvidence .Evidence}}<div class="section"><h3>Latest runtime evidence</h3><div class="small">{{get . "summary"}}</div><pre class="json">{{jsonPretty (kubeSummary .)}}</pre></div>{{end}}{{else}}<p>No ChangeRequest available.</p>{{end}}</div>{{end}}
{{define "changesList"}}<div class="card panel full"><h3>Change Requests</h3><table class="table"><thead><tr><th>Change</th><th>Application</th><th>Environment</th><th>Lifecycle</th><th>Runtime</th><th>Action</th></tr></thead><tbody>{{range .Changes}}<tr><td><b>{{changeNumberOrID .}}</b></td><td>{{get . "applicationName"}}</td><td>{{get . "targetEnvironment"}}</td><td><span class="badge {{badgeClass (get . "status")}}">{{get . "status"}}</span></td><td><span class="badge {{badgeClass (get . "runtimeStatus")}}">{{get . "runtimeStatus"}}</span></td><td><a href="/ui/changes/{{changeNumberOrID .}}">Open</a></td></tr>{{end}}</tbody></table></div>{{end}}
{{define "changeDetail"}}<div class="grid"><div class="full">{{template "changeCard" .}}</div><div class="card panel full"><h3>Audit events</h3><table class="table"><thead><tr><th>Event</th><th>Step</th><th>Created</th></tr></thead><tbody>{{range .Events}}<tr><td>{{get . "eventType"}}</td><td>{{eventStep .}}</td><td>{{get . "createdAt"}}</td></tr>{{end}}</tbody></table></div></div>{{end}}`
