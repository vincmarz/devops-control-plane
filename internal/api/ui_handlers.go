package api

import (
	"context"
	"encoding/json"
	"fmt"
	appsvc "github.com/vincmarz/devops-control-plane/internal/app"
	"github.com/vincmarz/devops-control-plane/internal/domain"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type uiData struct {
	Title                  string
	Subtitle               string
	Active                 string
	ActiveNav              string
	Mode                   string
	Changes                []map[string]any
	Applications           []map[string]any
	LogicalApplications    []map[string]any
	StandaloneApplications []map[string]any
	SelectedApplication    map[string]any
	Resources              []map[string]any
	History                []map[string]any
	Runtime                map[string]any
	SelectedChange         map[string]any
	ChangeRuntimeState     map[string]any
	Events                 []map[string]any
	Evidence               []map[string]any
	Stats                  map[string]any
	Flash                  string
	ActionError            string
	Error                  string
}

func (h *Handler) uiDashboard(w http.ResponseWriter, r *http.Request) {
	changes, err := h.uiChangesData(r)
	if err != nil {
		h.renderUI(w, http.StatusInternalServerError, uiData{Title: "Dashboard", Subtitle: "Applications and change overview", Active: "dashboard", ActiveNav: "dashboard", Error: err.Error()})
		return
	}
	apps := h.uiApplicationsData(r)
	grouping := h.deps.Services.Applications.GroupByEnvironment(toApplicationSlice(apps), appsvc.DefaultEnvironmentCatalog())
	logicalApplications := toMapSlice(grouping.LogicalApplications)
	standaloneApplications := toMapSlice(grouping.StandaloneApplications)
	selected := preferredChange(changes, "")
	events, evidence := h.uiChangeDetailsData(r, selected)
	stats := buildUIStats(changes, apps, evidence)
	addApplicationGroupingStats(stats, logicalApplications, standaloneApplications)
	h.renderUI(w, http.StatusOK, uiData{Title: "Dashboard", Subtitle: "Applications and change overview", Active: "dashboard", ActiveNav: "dashboard", Changes: changes, Applications: apps, LogicalApplications: logicalApplications, StandaloneApplications: standaloneApplications, SelectedChange: selected, Events: events, Evidence: evidence, Stats: stats})
}

func (h *Handler) uiChanges(w http.ResponseWriter, r *http.Request) {
	changes, err := h.uiChangesData(r)
	if err != nil {
		h.renderUI(w, http.StatusInternalServerError, uiData{Title: "Change Requests", Subtitle: "Change requests managed by the Control Plane", Active: "changes", ActiveNav: "all-changes", Error: err.Error()})
		return
	}
	h.renderUI(w, http.StatusOK, uiData{Title: "Change Requests", Subtitle: "Change requests managed by the Control Plane", Active: "changes", ActiveNav: "all-changes", Changes: changes, Stats: buildUIStats(changes, nil, nil)})
}

func (h *Handler) uiApplications(w http.ResponseWriter, r *http.Request) {
	apps := h.uiApplicationsData(r)
	grouping := h.deps.Services.Applications.GroupByEnvironment(toApplicationSlice(apps), appsvc.DefaultEnvironmentCatalog())
	logicalApplications := toMapSlice(grouping.LogicalApplications)
	standaloneApplications := toMapSlice(grouping.StandaloneApplications)
	changes, _ := h.uiChangesData(r)
	stats := buildUIStats(changes, apps, nil)
	addApplicationGroupingStats(stats, logicalApplications, standaloneApplications)
	h.renderUI(w, http.StatusOK, uiData{Title: "Applications", Subtitle: "Logical applications grouped by environment and standalone Argo CD applications", Active: "applications", ActiveNav: "applications", Applications: apps, LogicalApplications: logicalApplications, StandaloneApplications: standaloneApplications, Changes: changes, Stats: stats})
}

func (h *Handler) uiChangesAPIPage(w http.ResponseWriter, r *http.Request) {
	changes, err := h.uiChangesData(r)
	if err != nil {
		h.renderUI(w, http.StatusInternalServerError, uiData{Title: "Changes API", Subtitle: "API data preview", Active: "changes", ActiveNav: "changes-api", Mode: "changesAPI", Error: err.Error()})
		return
	}
	h.renderUI(w, http.StatusOK, uiData{Title: "Changes API", Subtitle: "API data preview with navigation", Active: "changes", ActiveNav: "changes-api", Mode: "changesAPI", Changes: changes, Stats: buildUIStats(changes, nil, nil)})
}

func (h *Handler) uiSettings(w http.ResponseWriter, r *http.Request) {
	h.renderUI(w, http.StatusOK, uiData{Title: "Settings", Subtitle: "Runtime readiness, environment and access placeholders", Active: "settings", ActiveNav: "settings"})
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
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Application", Subtitle: name, Active: "applications", ActiveNav: "applications", Applications: apps, Error: err.Error()})
		return
	}
	resources := toMapSlice(h.deps.Services.Applications.Resources(r.Context(), name))
	history := toMapSlice(h.deps.Services.Applications.History(r.Context(), name))
	runtime := toMap(h.deps.Services.Applications.Runtime(r.Context(), name))
	changes, _ := h.uiChangesData(r)
	h.renderUI(w, http.StatusOK, uiData{Title: fmt.Sprintf("Application: %s", str(get(selected, "name"))), Subtitle: "Application runtime and GitOps detail", Active: "applications", ActiveNav: "applications", Applications: apps, SelectedApplication: selected, Resources: resources, History: history, Runtime: runtime, Changes: changes, Stats: buildUIStats(changes, apps, nil)})
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
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Change Request", Subtitle: id, Active: "changes", ActiveNav: "change-requests", Error: err.Error()})
		return
	}
	selected := withUIActionVisibility(r.Context(), toMap(change))
	runtimeState, err := h.deps.Services.Changes.GetRuntimeState(r.Context(), id)
	if err != nil {
		h.renderUI(w, http.StatusInternalServerError, uiData{Title: "Change Request", Subtitle: id, Active: "changes", ActiveNav: "change-requests", Error: err.Error()})
		return
	}
	events, evidence := h.uiChangeDetailsData(r, selected)
	if strings.TrimSpace(str(get(selected, "runtimeStatus"))) == "" {
		if status := latestRuntimeStatusFromEvents(events); status != "" {
			selected["runtimeStatus"] = status
		}
	}
	h.renderUI(w, http.StatusOK, uiData{Title: fmt.Sprintf("Change Request: %s", str(get(selected, "changeNumber"))), Subtitle: "Operational change detail", Active: "changes", ActiveNav: "change-requests", SelectedChange: selected, ChangeRuntimeState: toMap(runtimeState), Events: events, Evidence: evidence, Stats: map[string]any{"evidence": len(evidence)}, Flash: r.URL.Query().Get("flash"), ActionError: r.URL.Query().Get("error")})
}

func (h *Handler) uiChangeEvents(w http.ResponseWriter, r *http.Request, id string) {
	change, err := h.deps.Services.Changes.Get(r.Context(), id)
	if err != nil {
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Audit Events", Subtitle: id, Active: "changes", ActiveNav: "audit-log", Error: err.Error()})
		return
	}
	events, err := h.deps.Services.Changes.Events(r.Context(), id)
	if err != nil {
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Audit Events", Subtitle: id, Active: "changes", ActiveNav: "audit-log", Error: err.Error()})
		return
	}
	selected := withUIActionVisibility(r.Context(), toMap(change))
	h.renderUI(w, http.StatusOK, uiData{Title: fmt.Sprintf("Audit events: %s", changeNumberOrID(selected)), Subtitle: "Change audit trail and technical workflow events", Active: "changes", ActiveNav: "audit-log", Mode: "changeEvents", SelectedChange: selected, Events: toMapSlice(events), Stats: map[string]any{"events": len(events)}})
}

func (h *Handler) uiChangeEvidence(w http.ResponseWriter, r *http.Request, id string) {
	change, err := h.deps.Services.Changes.Get(r.Context(), id)
	if err != nil {
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Evidence", Subtitle: id, Active: "changes", ActiveNav: "evidence", Error: err.Error()})
		return
	}
	evidence, err := h.deps.Services.Evidence.List(r.Context(), id, "")
	if err != nil {
		h.renderUI(w, http.StatusNotFound, uiData{Title: "Evidence", Subtitle: id, Active: "changes", ActiveNav: "evidence", Error: err.Error()})
		return
	}
	selected := withUIActionVisibility(r.Context(), toMap(change))
	h.renderUI(w, http.StatusOK, uiData{Title: fmt.Sprintf("Evidence: %s", changeNumberOrID(selected)), Subtitle: "Collected technical and runtime evidence", Active: "changes", ActiveNav: "evidence", Mode: "changeEvidence", SelectedChange: selected, Evidence: toMapSlice(evidence), Stats: map[string]any{"evidence": len(evidence)}})
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
	case "start-build":
		_, err = h.deps.Services.Changes.StartBuild(r.Context(), id)
	case "check-build":
		_, err = h.deps.Services.Changes.CheckBuild(r.Context(), id)
	case "update-gitops":
		_, err = h.deps.Services.Changes.UpdateGitOps(r.Context(), id)
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

func addApplicationGroupingStats(stats map[string]any, logical, standalone []map[string]any) {
	instances := 0
	for _, application := range logical {
		if environments, ok := application["environments"].([]map[string]any); ok {
			instances += len(environments)
		}
	}
	stats["logicalApplications"] = len(logical)
	stats["environmentInstances"] = instances
	stats["standaloneApplications"] = len(standalone)
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
	evidenceRaw, _ := h.deps.Services.Evidence.List(r.Context(), id, "")
	return toMapSlice(eventsRaw), toMapSlice(evidenceRaw)
}

func (h *Handler) renderUI(w http.ResponseWriter, status int, data uiData) {
	funcs := template.FuncMap{"get": get, "str": str, "short": short, "badgeClass": badgeClass, "latestEvidence": latestEvidence, "kubeSummary": kubeSummary, "diagnosticsSummary": diagnosticsSummary, "eventStep": eventStep, "jsonPretty": jsonPretty, "changeNumberOrID": changeNumberOrID, "userCanSeeTechnicalActions": userCanSeeTechnicalActions, "recommendedActions": recommendedActions, "advancedActions": advancedActions,
		"environmentAllowsTechnicalActions": environmentAllowsTechnicalActions,
		"environmentActionWarning":          environmentActionWarning, "applicationName": applicationName, "recentChanges": recentChanges,
		"latestValidationEvidence": latestValidationEvidence, "validationField": validationField, "environmentSummaries": environmentSummaries}
	tpl := template.Must(template.New("ui").Funcs(funcs).Parse(uiTemplate))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = tpl.Execute(w, data)
}

func toApplicationSlice(v any) []domain.Application {
	raw, _ := json.Marshal(v)
	var out []domain.Application
	_ = json.Unmarshal(raw, &out)
	if out == nil {
		return []domain.Application{}
	}
	return out
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

func environmentSummaries() []map[string]any {
	catalog := appsvc.DefaultEnvironmentCatalog()
	names := []string{"dev", "staging", "production"}
	out := make([]map[string]any, 0, len(names))
	for _, name := range names {
		definition, ok := catalog.Resolve(name)
		if !ok {
			continue
		}
		kubernetesNamespace := strings.TrimSpace(definition.KubernetesNamespace)
		if kubernetesNamespace == "" {
			kubernetesNamespace = "-"
		}
		tektonNamespace := strings.TrimSpace(definition.TektonNamespace)
		if tektonNamespace == "" {
			tektonNamespace = kubernetesNamespace
		}
		displayName := strings.TrimSpace(definition.DisplayName)
		if displayName == "" {
			displayName = definition.Name
		}
		out = append(out, map[string]any{
			"name":                definition.Name,
			"displayName":         displayName,
			"enabled":             definition.Enabled,
			"kubernetesNamespace": kubernetesNamespace,
			"tektonNamespace":     tektonNamespace,
		})
	}
	return out
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

func latestValidationEvidence(items []map[string]any) map[string]any {
	for _, item := range items {
		if strings.EqualFold(str(get(item, "evidenceType")), "validation") || strings.EqualFold(str(get(item, "name")), "tekton-validation-evidence") {
			return item
		}
	}
	return nil
}

func validationField(ev map[string]any, key string) any {
	if ev == nil || strings.TrimSpace(key) == "" {
		return ""
	}
	if key == "summary" {
		if summary := str(get(ev, "summary")); summary != "" {
			return summary
		}
	}
	if value := get(ev, key); str(value) != "" {
		return value
	}
	payload, _ := get(ev, "payload").(map[string]any)
	if payload == nil {
		return ""
	}
	if value := payload[key]; str(value) != "" {
		return value
	}
	for _, sectionName := range []string{"tekton", "gitops", "diagnostics"} {
		section, _ := payload[sectionName].(map[string]any)
		if section == nil {
			continue
		}
		if value := section[key]; str(value) != "" {
			return value
		}
	}
	if key == "tektonNamespace" {
		if value := validationField(ev, "namespace"); str(value) != "" {
			return value
		}
	}
	return ""
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
		uiAction("start-build", "Start Build", "Build the immutable application source revision and publish a change-scoped image tag.", false),
		uiAction("check-build", "Check Build", "Read the application build PipelineRun result and record the immutable image digest.", false),
		uiAction("update-gitops", "Update GitOps", "Write the built image digest into the GitOps repository and open a review request.", false),
		uiAction("check-validation", "Check Validation", "Poll latest Tekton PipelineRun result.", false),
		uiAction("check-deployment", "Check Deployment", "Check Argo CD sync and health state.", false),
		uiAction("collect-evidence", "Collect Evidence", "Collect post-deployment Kubernetes/OpenShift evidence.", false),
		uiAction("create-branch", "Create Branch", "Create the Git change branch.", false),
		uiAction("update-files", "Update Files", "Commit generated GitOps files on the change branch.", false),
		uiAction("open-merge-request", "Open Review Request", "Open the Git review request.", false),
		uiAction("merge-request", "Merge Review Request", "Merge the approved Git review request.", false),
	}
}

func withUIActionVisibility(ctx context.Context, change map[string]any) map[string]any {
	if change == nil {
		return change
	}

	copy := map[string]any{}
	for key, value := range change {
		copy[key] = value
	}
	copy["uiCanSeeTechnicalActions"] = userCanExecuteTechnicalActions(ctx)
	return copy
}

func userCanSeeTechnicalActions(change map[string]any) bool {
	allowed, ok := change["uiCanSeeTechnicalActions"].(bool)
	return ok && allowed
}

func userCanExecuteTechnicalActions(ctx context.Context) bool {
	identity, ok := ctx.Value(identityContextKey).(authIdentity)
	if !ok {
		return false
	}
	return identity.Roles["operator"] || identity.Roles["admin"]
}

func environmentAllowsTechnicalActions(change map[string]any) bool {
	catalog := appsvc.DefaultEnvironmentCatalog()
	return catalog.AllowsTechnicalActions(str(get(change, "targetEnvironment")))
}

func environmentActionWarning(change map[string]any) string {
	catalog := appsvc.DefaultEnvironmentCatalog()
	targetEnvironment := strings.TrimSpace(str(get(change, "targetEnvironment")))
	if targetEnvironment == "" {
		return "Target environment is empty. Technical actions are not available until the ChangeRequest has a valid target environment."
	}

	definition, ok := catalog.Resolve(targetEnvironment)
	if !ok {
		return "Target environment " + targetEnvironment + " is not configured. Technical actions are not available for this historical ChangeRequest."
	}
	if !definition.Enabled {
		return "Target environment " + targetEnvironment + " is currently disabled. Technical actions are not available for this historical ChangeRequest."
	}
	if !definition.AllowTechnicalActions {
		return "Target environment " + targetEnvironment + " does not allow technical actions."
	}

	return ""
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
		return []map[string]any{uiAction("open-merge-request", "Open Review Request", "Open the Git review request.", true)}
	case "mergerequestopened":
		return []map[string]any{uiAction("merge-request", "Merge Review Request", "Merge the review request when governance allows it.", true)}
	case "mergerequestmerged":
		return []map[string]any{uiAction("start-build", "Start Build", "Build the merged immutable application source revision.", true)}
	case "buildrunning":
		return []map[string]any{uiAction("check-build", "Check Build", "Read the application build result and record the image digest.", true)}
	case "buildsucceeded":
		return []map[string]any{uiAction("update-gitops", "Update GitOps", "Write the built image digest into the GitOps repository.", true)}
	case "gitopsupdated":
		return []map[string]any{uiAction("check-deployment", "Check Deployment", "Verify Argo CD sync and application health for the new image.", true)}
	case "buildfailed":
		return []map[string]any{uiAction("start-build", "Start Build", "Rebuild the application source after remediation.", true)}
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
/* DevOps Control Plane - Enterprise UI Design System */

/* Color palette and design tokens */
:root {
  --color-primary: #0f172a;
  --color-primary-light: #1e293b;
  --color-primary-dark: #020617;
  --color-accent: #2563eb;
  --color-accent-light: #3b82f6;
  --color-accent-darker: #1d4ed8;
  --color-success: #16a34a;
  --color-success-bg: #dcfce7;
  --color-success-fg: #15803d;
  --color-warning: #d97706;
  --color-warning-bg: #fef3c7;
  --color-warning-fg: #b45309;
  --color-critical: #dc2626;
  --color-critical-bg: #fee2e2;
  --color-critical-fg: #b91c1c;
  --color-info: #1d4ed8;
  --color-info-bg: #dbeafe;
  --color-info-fg: #1d4ed8;

  --sidebar-bg: #071b33;
  --sidebar-bg-accent: #0b2646;
  --sidebar-text: #ffffff;
  --sidebar-text-secondary: #cbd5e1;

  --text-primary: #0f172a;
  --text-secondary: #475569;
  --text-tertiary: #64748b;
  --text-muted: #94a3b8;

  --bg-primary: #ffffff;
  --bg-secondary: #f8fafc;
  --bg-tertiary: #f1f5f9;

  --border-color: #e2e8f0;
  --border-color-light: #cbd5e1;
  --border-color-darker: #94a3b8;

  --shadow-sm: 0 1px 2px rgba(15, 23, 42, 0.05);
  --shadow-md: 0 4px 6px rgba(15, 23, 42, 0.08);
  --shadow-lg: 0 8px 20px rgba(15, 23, 42, 0.1);
  --shadow-xl: 0 12px 32px rgba(15, 23, 42, 0.12);

  --transition-fast: 150ms cubic-bezier(0.4, 0, 0.2, 1);
  --transition-normal: 200ms cubic-bezier(0.4, 0, 0.2, 1);
  --transition-slow: 300ms cubic-bezier(0.4, 0, 0.2, 1);
}

/* Global styles */
* {
  box-sizing: border-box;
}

html {
  scroll-behavior: smooth;
}

body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  font-size: 14px;
  line-height: 1.6;
  background-color: var(--bg-secondary);
  color: var(--text-primary);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}

/* Links */
a {
  color: var(--color-accent);
  text-decoration: none;
  transition: color var(--transition-fast);
}

a:hover {
  color: var(--color-accent-light);
  text-decoration: underline;
}

a:focus {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}

/* Layout container */
.app {
  display: flex;
  min-height: 100vh;
  background: var(--bg-secondary);
}

/* ===== SIDEBAR NAVIGATION ===== */

.sidebar {
  width: 280px;
  background: linear-gradient(180deg, var(--sidebar-bg) 0%, var(--sidebar-bg-accent) 100%);
  color: var(--sidebar-text);
  padding: 28px 20px;
  display: flex;
  flex-direction: column;
  gap: 32px;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 100;
  overflow-y: auto;
  box-shadow: var(--shadow-lg);
}

.sidebar::-webkit-scrollbar {
  width: 6px;
}

.sidebar::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
}

.sidebar::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.sidebar::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* Brand section */
.brand {
  display: flex;
  align-items: center;
  gap: 14px;
  font-weight: 800;
  font-size: 16px;
  letter-spacing: -0.5px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.brand-icon {
  width: 40px;
  height: 40px;
  border: 2px solid var(--color-accent-light);
  border-radius: 10px;
  display: grid;
  place-items: center;
  color: #60a5fa;
  font-size: 20px;
  flex-shrink: 0;
}

/* Navigation groups */
.nav-group {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-group:not(:first-child) {
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding-top: 14px;
}

.nav {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.nav a {
  display: flex;
  align-items: center;
  gap: 12px;
  color: rgba(255, 255, 255, 0.7);
  padding: 12px 14px;
  border-radius: 8px;
  font-weight: 600;
  font-size: 13px;
  transition: all var(--transition-fast);
  cursor: pointer;
  position: relative;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav a:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--sidebar-text);
}

.nav a:focus {
  outline: 2px solid var(--color-accent-light);
  outline-offset: -2px;
}

.nav a.active {
  background: var(--color-accent-darker);
  color: var(--sidebar-text);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.1);
  font-weight: 700;
}

/* System status box */
.sys {
  margin-top: auto;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.03);
  backdrop-filter: blur(10px);
  font-size: 12px;
}

.sys h4 {
  margin: 0 0 14px;
  font-size: 12px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: rgba(255, 255, 255, 0.5);
}

.sys-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.sys-row:last-child {
  border-bottom: none;
}

.sys-row span {
  color: rgba(255, 255, 255, 0.6);
  display: flex;
  align-items: center;
  gap: 8px;
}

.sys-row b {
  color: var(--color-success-bg);
  font-weight: 700;
}

.dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-success);
  box-shadow: 0 0 8px rgba(22, 163, 74, 0.4);
  flex-shrink: 0;
}

/* Version info */
.version {
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  text-align: center;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
  line-height: 1.6;
}

/* ===== MAIN CONTENT ===== */

.main {
  margin-left: 280px;
  width: calc(100% - 280px);
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
}

/* Top navigation bar */
.topbar {
  min-height: 80px;
  height: auto;
  padding: 20px 28px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  box-shadow: var(--shadow-sm);
  position: relative;
  z-index: 50;
}

/* Page title section */
.title {
  flex: 1 1 auto;
  min-width: 220px;
}

.title h1 {
  margin: 0;
  padding: 0;
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.5px;
}

.title p {
  margin: 6px 0 0;
  padding: 0;
  color: var(--text-tertiary);
  font-size: 13px;
  font-weight: 500;
}

/* User/environment area */
.user {
  display: flex;
  align-items: stretch;
  gap: 0;
  flex: 0 1 auto;
  min-width: 0;
}

.environment-summary {
  min-width: 500px;
  max-width: 660px;
  flex-shrink: 0;
  align-self: stretch;
  border: 1px solid var(--border-color);
  border-radius: 8px 0 0 8px;
  background: var(--bg-primary);
  padding: 12px 16px;
  box-shadow: var(--shadow-sm);
  line-height: 1.4;
}

.environment-summary-title {
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--text-secondary);
}

.environment-summary-row {
  display: grid;
  grid-template-columns: 100px 1fr 1fr;
  gap: 8px;
  font-size: 12px;
  color: var(--text-tertiary);
  white-space: nowrap;
  padding: 6px 0;
  align-items: center;
}

.environment-summary-row b {
  color: var(--text-secondary);
  font-weight: 600;
  text-overflow: ellipsis;
  overflow: hidden;
}

.environment-summary-row span {
  color: var(--text-tertiary);
  font-size: 11px;
  text-overflow: ellipsis;
  overflow: hidden;
}

.user-summary {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  align-self: stretch;
  border: 1px solid var(--border-color);
  border-left: none;
  border-radius: 0 8px 8px 0;
  background: var(--bg-primary);
  padding: 12px 16px;
  box-shadow: var(--shadow-sm);
  min-height: 56px;
  min-width: 160px;
  flex-shrink: 0;
}

.user-summary b {
  color: var(--text-secondary);
  font-weight: 600;
  font-size: 13px;
}

.avatar {
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  font-weight: 700;
  color: var(--bg-primary);
  background: var(--color-accent);
  flex-shrink: 0;
  font-size: 14px;
}

.select {
  border: 1px solid var(--border-color-light);
  border-radius: 8px;
  background: var(--bg-primary);
  padding: 10px 14px;
  font-size: 13px;
  cursor: pointer;
  color: var(--text-primary);
}

/* ===== CONTENT AREA ===== */

.content {
  flex: 1;
  padding: 28px;
  background: var(--bg-secondary);
  overflow-y: auto;
}

/* KPI Cards */
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 18px;
  margin-bottom: 24px;
}

.card {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: var(--shadow-md);
  transition: all var(--transition-normal);
}

.card:hover {
  box-shadow: var(--shadow-lg);
  border-color: var(--border-color-light);
}

.card:focus-within {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}

.metric {
  padding: 20px;
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.metric .icon {
  width: 50px;
  height: 50px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  font-weight: 800;
  font-size: 20px;
  flex-shrink: 0;
}

.metric > div {
  flex: 1;
  min-width: 0;
}

.kpi-title-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  flex-wrap: wrap;
  width: 100%;
}

.kpi-title {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 700;
}

.kpi-counter {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 26px;
  padding: 0 8px;
  border: 1px solid var(--text-secondary);
  border-radius: 4px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 700;
  line-height: 1;
  flex-shrink: 0;
}

.metric > div > span {
  display: block;
  color: var(--text-tertiary);
  font-size: 12px;
  line-height: 1.4;
}

/* Grid layout for main sections */
.grid {
  display: grid;
  grid-template-columns: 1fr 1.5fr 1.3fr;
  gap: 20px;
  margin-top: 24px;
}

.grid .full {
  grid-column: 1 / -1;
}

/* Panels and sections */
.panel {
  padding: 0;
  border-radius: 12px;
  overflow: hidden;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-md);
}

.panel h3 {
  font-size: 15px;
  font-weight: 700;
  margin: 0;
  padding: 18px;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.panel h3 a {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-accent);
}

/* Application grouping */
.logical-app {
  border-bottom: 1px solid var(--border-color);
}

.logical-app:last-child {
  border-bottom: none;
}

.logical-app-title,
.standalone-title,
.logical-app-table-title {
  padding: 14px 18px;
  font-weight: 700;
  color: var(--text-secondary);
  background: var(--bg-secondary);
  font-size: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-color);
}

.environment-instance {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-top: 1px solid var(--border-color);
  border-top: 0;
  gap: 16px;
  transition: background var(--transition-fast);
}

.environment-instance:hover {
  background: var(--bg-secondary);
}

.environment-instance > div {
  flex: 1;
  min-width: 0;
}

.environment-instance b {
  display: block;
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: 2px;
}

.environment-instance .small {
  font-size: 12px;
  color: var(--text-tertiary);
}

.standalone-title {
  border-top: 1px solid var(--border-color);
}

/* Empty state */
.empty-state {
  padding: 20px 18px;
  text-align: center;
  color: var(--text-tertiary);
  background: var(--bg-secondary);
}

/* Lists */
.list {
  padding: 0;
  margin: 0;
  list-style: none;
}

.list li {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-color);
  gap: 12px;
  transition: background var(--transition-fast);
}

.list li:hover {
  background: var(--bg-secondary);
}

.list li > div {
  flex: 1;
  min-width: 0;
}

.list li b {
  display: block;
  color: var(--text-primary);
  font-weight: 600;
}

.small {
  font-size: 12px;
  color: var(--text-tertiary);
  line-height: 1.4;
}

/* ===== BADGES ===== */

.badge {
  border-radius: 6px;
  padding: 5px 10px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.3px;
  display: inline-block;
  white-space: nowrap;
  vertical-align: middle;
  border: 1px solid;
}

.badge-ok {
  background: var(--color-success-bg);
  color: var(--color-success-fg);
  border-color: var(--color-success);
}

.badge-bad {
  background: var(--color-critical-bg);
  color: var(--color-critical-fg);
  border-color: var(--color-critical);
}

.badge-warn {
  background: var(--color-warning-bg);
  color: var(--color-warning-fg);
  border-color: var(--color-warning);
}

.badge-info {
  background: var(--color-info-bg);
  color: var(--color-info-fg);
  border-color: var(--color-accent);
}

.badge-muted {
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  border-color: var(--border-color-light);
}

/* ===== DETAIL VIEWS ===== */

.detail {
  padding: 24px;
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  box-shadow: var(--shadow-md);
}

.detail h3 {
  font-size: 15px;
  font-weight: 700;
  margin: 0 0 18px;
  color: var(--text-primary);
}

.detail-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  gap: 16px;
}

.detail-head h3 {
  margin: 0;
  flex: 1;
}

/* Key-value pairs */
.kv {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px 24px;
  margin-bottom: 24px;
}

.kv > div {
  display: flex;
  flex-direction: column;
}

.kv .label {
  color: var(--text-tertiary);
  font-size: 12px;
  margin-bottom: 6px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.3px;
}

.kv b {
  color: var(--text-primary);
  font-weight: 600;
  word-break: break-word;
}

/* Sections */
.section {
  border-top: 1px solid var(--border-color);
  margin-top: 24px;
  padding-top: 24px;
}

.section h3 {
  font-size: 14px;
  font-weight: 700;
  margin: 0 0 12px;
  color: var(--text-primary);
}

.section p {
  margin: 0 0 12px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.section p:last-child {
  margin-bottom: 0;
}

/* ===== ACTIONS ===== */

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin: 12px 0;
}

.action-groups {
  display: grid;
  gap: 18px;
}

.action-card {
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 14px;
  background: var(--bg-secondary);
  max-width: 280px;
  transition: all var(--transition-normal);
}

.action-card:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--color-accent);
  background: var(--bg-primary);
}

.action-card form {
  margin: 0 0 8px;
}

.action-desc {
  line-height: 1.5;
  font-size: 12px;
  color: var(--text-tertiary);
}

/* Buttons */
.btn {
  border: 1px solid var(--color-accent);
  color: var(--color-accent);
  background: white;
  padding: 10px 16px;
  border-radius: 8px;
  font-weight: 700;
  font-size: 13px;
  cursor: pointer;
  transition: all var(--transition-fast);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
  user-select: none;
  text-decoration: none;
}

.btn:hover {
  background: var(--bg-secondary);
  border-color: var(--color-accent-darker);
  color: var(--color-accent-darker);
  text-decoration: none;
}

.btn:focus {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}

.btn:active {
  transform: scale(0.98);
}

.btn.primary {
  background: var(--color-accent);
  color: white;
  border-color: var(--color-accent);
}

.btn.primary:hover {
  background: var(--color-accent-darker);
  border-color: var(--color-accent-darker);
  color: white;
}

/* ===== TIMELINE & STATUS ===== */

.timeline {
  padding: 20px;
}

.step {
  display: flex;
  gap: 16px;
  margin: 0 0 20px;
  align-items: flex-start;
}

.circle {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border-color-light);
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 1px;
  background: white;
  transition: all var(--transition-normal);
}

.circle.done {
  background: var(--color-success);
  border-color: var(--color-success);
  box-shadow: 0 0 0 3px rgba(22, 163, 74, 0.1);
}

.step > div {
  flex: 1;
}

.step b {
  display: block;
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: 4px;
}

.step .small {
  font-size: 12px;
  color: var(--text-tertiary);
}

/* Evidence section */
.evidence {
  padding: 20px;
}

.ev-row {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 0;
  border-bottom: 1px solid var(--border-color);
  align-items: flex-start;
}

.ev-row:last-child {
  border-bottom: none;
}

.ev-row > div:first-child {
  flex: 1;
  min-width: 0;
}

.ev-row b {
  display: block;
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: 2px;
}

.ev-row a {
  white-space: nowrap;
  font-weight: 600;
}

/* ===== TABLES ===== */

.table {
  width: 100%;
  border-collapse: collapse;
  background: var(--bg-primary);
  font-size: 13px;
}

.table thead {
  background: var(--bg-secondary);
  position: sticky;
  top: 0;
  z-index: 10;
}

.table th {
  font-size: 11px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 700;
  padding: 14px 16px;
  text-align: left;
  border-bottom: 2px solid var(--border-color);
}

.table td {
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
}

.table tbody tr:hover {
  background: var(--bg-secondary);
}

.table tbody tr:focus-within {
  outline: 2px inset var(--color-accent);
  outline-offset: -1px;
}

.table tbody tr td:first-child a,
.table tbody tr td b {
  font-weight: 600;
  color: var(--text-primary);
}

.table .badge {
  margin: 0;
}

/* ===== EVIDENCE GRID ===== */

.evidence-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 14px;
  margin-top: 14px;
}

.evidence-card {
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 14px;
  background: var(--bg-secondary);
  transition: all var(--transition-normal);
}

.evidence-card:hover {
  border-color: var(--border-color-light);
  box-shadow: var(--shadow-sm);
}

.evidence-card h4 {
  margin: 0 0 10px;
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
}

.evidence-kv {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding: 6px 0;
  color: var(--text-secondary);
  font-size: 12px;
  align-items: baseline;
}

.evidence-kv span {
  font-weight: 600;
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.evidence-kv b {
  color: var(--text-primary);
  font-weight: 600;
  text-align: right;
  word-break: break-word;
  flex: 1;
  min-width: 0;
}

.pod-list {
  margin: 8px 0 0;
  padding-left: 20px;
  list-style: disc;
}

.pod-list li {
  margin: 4px 0;
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.5;
}

/* Details/Summary */
details {
  margin-top: 16px;
}

details > summary {
  cursor: pointer;
  padding: 10px 0;
  color: var(--color-accent);
  font-weight: 700;
  font-size: 12px;
  transition: color var(--transition-fast);
  user-select: none;
}

details > summary:hover {
  color: var(--color-accent-darker);
}

details > summary:focus {
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}

details[open] {
  margin-top: 16px;
  margin-bottom: 12px;
}

details[open] > summary {
  margin-bottom: 12px;
}

/* ===== ALERTS ===== */

.alert {
  padding: 14px 16px;
  border-radius: 10px;
  margin-bottom: 18px;
  font-weight: 700;
  font-size: 13px;
  border: 1px solid;
  line-height: 1.5;
}

.alert-ok {
  background: var(--color-success-bg);
  color: var(--color-success-fg);
  border-color: var(--color-success);
}

.alert-error {
  background: var(--color-critical-bg);
  color: var(--color-critical-fg);
  border-color: var(--color-critical);
}

.alert-warn {
  background: var(--color-warning-bg);
  color: var(--color-warning-fg);
  border-color: var(--color-warning);
}

/* ===== JSON DISPLAY ===== */

.json {
  white-space: pre-wrap;
  word-wrap: break-word;
  background: #0f172a;
  color: #dbeafe;
  border-radius: 10px;
  padding: 16px;
  max-height: 420px;
  overflow: auto;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'source-code-pro', monospace;
  font-size: 11px;
  line-height: 1.6;
  border: 1px solid #1e293b;
}

.json::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.json::-webkit-scrollbar-track {
  background: #0f172a;
}

.json::-webkit-scrollbar-thumb {
  background: #334155;
  border-radius: 3px;
}

.json::-webkit-scrollbar-thumb:hover {
  background: #475569;
}

/* ===== FULL WIDTH ===== */

.full {
  grid-column: 1 / -1;
}

/* ===== FOOTER ===== */

.footer {
  text-align: center;
  color: var(--text-tertiary);
  font-size: 12px;
  margin-top: 40px;
  padding: 24px 0;
  border-top: 1px solid var(--border-color);
}

/* ===== RESPONSIVE DESIGN ===== */

/* Tablet layout */
@media (max-width: 1400px) {
  .grid {
    grid-template-columns: 1fr 1.2fr;
  }

  .cards {
    grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  }
}

/* Tablet - stacked layout */
@media (max-width: 1100px) {
  .sidebar {
    position: static;
    width: 100%;
    flex-direction: row;
    gap: 20px;
    padding: 16px 20px;
    align-items: center;
    border-bottom: 1px solid var(--border-color);
  }

  .brand {
    flex-shrink: 0;
    border-bottom: none;
    padding-bottom: 0;
    border-right: 1px solid rgba(255, 255, 255, 0.1);
    padding-right: 20px;
    margin-right: 0;
  }

  .nav {
    display: none;
  }

  .sys {
    display: none;
  }

  .version {
    display: none;
  }

  .main {
    margin-left: 0;
    width: 100%;
  }

  .topbar {
    flex-wrap: wrap;
    gap: 12px;
  }

  .user {
    flex-wrap: wrap;
    width: 100%;
  }

  .environment-summary {
    min-width: 100%;
    border-radius: 8px;
  }

  .user-summary {
    border-radius: 8px;
    border-left: 1px solid var(--border-color);
  }

  .grid {
    grid-template-columns: 1fr;
  }

  .evidence-grid {
    grid-template-columns: 1fr;
  }

  .cards {
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  }
}

/* Mobile layout */
@media (max-width: 768px) {
  .sidebar {
    flex-direction: column;
    gap: 12px;
    padding: 12px 16px;
  }

  .brand {
    border-right: none;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    padding-right: 0;
    padding-bottom: 12px;
  }

  .content {
    padding: 16px;
  }

  .topbar {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
    min-height: auto;
    padding: 16px;
  }

  .title h1 {
    font-size: 20px;
  }

  .user {
    flex-direction: column;
    width: 100%;
  }

  .environment-summary {
    min-width: 100%;
    border-radius: 8px;
  }

  .user-summary {
    width: 100%;
    border-radius: 8px;
    border-left: 1px solid var(--border-color);
    justify-content: flex-start;
  }

  .cards {
    grid-template-columns: 1fr;
    gap: 12px;
  }

  .kv {
    grid-template-columns: 1fr;
  }

  .actions {
    gap: 8px;
  }

  .btn {
    width: 100%;
  }

  .table {
    font-size: 12px;
  }

  .table th,
  .table td {
    padding: 10px 8px;
  }
}
</style></head><body><div class="app"><aside class="sidebar"><div class="brand"><div class="brand-icon">☁</div><div>DevOps Control Plane</div></div><nav class="nav"><a class="{{if eq .ActiveNav "dashboard"}}active{{end}}" href="/">▣ Dashboard</a><a class="{{if eq .ActiveNav "applications"}}active{{end}}" href="/ui/applications">▧ Applications</a><div class="nav-group"><a class="{{if eq .ActiveNav "change-requests"}}active{{end}}" href="/ui/changes">◌ Change Requests</a><a class="{{if eq .ActiveNav "all-changes"}}active{{end}}" href="/ui/changes">All changes</a><a class="{{if eq .ActiveNav "changes-api"}}active{{end}}" href="/ui/changes-api">Changes API</a></div><div class="nav-group"><a class="{{if eq .ActiveNav "evidence"}}active{{end}}" href="/ui/changes/CHG-2026-0005/evidence">▤ Evidence</a><a class="{{if eq .ActiveNav "audit-log"}}active{{end}}" href="/ui/changes/CHG-2026-0005/events">☷ Audit log</a></div><div class="nav-group"><a class="{{if eq .ActiveNav "settings"}}active{{end}}" href="/ui/settings">⚙ Settings</a></div></nav><div class="sys"><h4>System status</h4><div class="sys-row"><span><i class="dot"></i>API</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>Database</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>Git Providers</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>Tekton</span><b>OK</b></div><div class="sys-row"><span><i class="dot"></i>Argo CD</span><b>OK</b></div></div><div class="version">DevOps Control Plane<br>v0.1.0</div></aside><main class="main"><header class="topbar"><div class="title"><h1>{{.Title}}</h1><p>{{.Subtitle}}</p></div><div class="user"><div class="environment-summary"><div class="environment-summary-title">Environments / Namespaces</div>{{range environmentSummaries}}<div class="environment-summary-row"><b>{{get . "name"}}</b><span>k8s: {{get . "kubernetesNamespace"}}</span><span>tekton: {{get . "tektonNamespace"}}</span></div>{{end}}</div><div class="user-summary"><div class="avatar">A</div><b>admin</b></div></div></header><section class="content">{{if .Flash}}<div class="alert alert-ok">{{.Flash}}</div>{{end}}{{if .ActionError}}<div class="alert alert-error">{{.ActionError}}</div>{{end}}{{if .Error}}<div class="card detail"><b>Error:</b> {{.Error}}</div>{{else}}{{if eq .Mode "changeEvents"}}{{template "changeEventsPage" .}}{{else if eq .Mode "changeEvidence"}}{{template "changeEvidencePage" .}}{{else if eq .Active "settings"}}{{template "settingsPage" .}}{{else if eq .Mode "changesAPI"}}{{template "changesAPIPage" .}}{{else if eq .Active "changes"}}{{if .SelectedChange}}{{template "changeDetail" .}}{{else}}{{template "changesList" .}}{{end}}{{else if eq .Active "applications"}}{{if .SelectedApplication}}{{template "applicationDetail" .}}{{else}}{{template "applicationsList" .}}{{end}}{{else}}{{template "dashboard" .}}{{end}}{{end}}<div class="footer">© 2026 DevOps Control Plane <span style="float:right">v0.1.0</span></div></section></main></div></body></html>
{{define "dashboard"}}<div class="cards"><div class="card metric"><div class="icon" style="background:#dbeafe;color:#2563eb">□</div><div><div class="kpi-title-line"><span class="kpi-title">Logical Applications</span><b class="kpi-counter">{{get .Stats "logicalApplications"}}</b></div><span>{{get .Stats "environmentInstances"}} environment instances</span></div></div><div class="card metric"><div class="icon" style="background:#dcfce7;color:#16a34a">✓</div><div><div class="kpi-title-line"><span class="kpi-title">Completed changes</span><b class="kpi-counter">{{get .Stats "completed"}}</b></div><span>Last 30 days</span></div></div><div class="card metric"><div class="icon" style="background:#fef3c7;color:#d97706">◷</div><div><div class="kpi-title-line"><span class="kpi-title">Running changes</span><b class="kpi-counter">{{get .Stats "running"}}</b></div><span>Currently running</span></div></div><div class="card metric"><div class="icon" style="background:#fee2e2;color:#dc2626">!</div><div><div class="kpi-title-line"><span class="kpi-title">Failed changes</span><b class="kpi-counter">{{get .Stats "failed"}}</b></div><span>Last 30 days</span></div></div><div class="card metric"><div class="icon" style="background:#ede9fe;color:#7c3aed">▤</div><div><div class="kpi-title-line"><span class="kpi-title">Collected evidence</span><b class="kpi-counter">{{get .Stats "evidence"}}</b></div><span>For selected change</span></div></div></div><div class="grid"><div><div class="card panel"><h3>Logical Applications <a style="float:right;font-size:13px" href="/ui/applications">View all</a></h3>{{range .LogicalApplications}}<div class="logical-app"><div class="logical-app-title"><b>{{get . "name"}}</b></div>{{range get . "environments"}}<div class="environment-instance"><div><a href="/ui/applications/{{get . "argocdApplicationName"}}"><b>{{get . "environment"}}</b></a><div class="small">{{get . "argocdApplicationName"}} · {{get . "kubernetesNamespace"}}</div></div><span class="badge {{badgeClass (get . "healthStatus")}}">{{get . "healthStatus"}}</span></div>{{end}}</div>{{else}}<div class="small empty-state">No logical application bindings configured.</div>{{end}}{{if .StandaloneApplications}}<div class="standalone-title">Standalone Argo CD Applications</div>{{range .StandaloneApplications}}<div class="environment-instance"><div><a href="/ui/applications/{{get . "name"}}"><b>{{get . "name"}}</b></a><div class="small">{{get . "targetNamespace"}}</div></div><span class="badge {{badgeClass (get . "healthStatus")}}">{{get . "healthStatus"}}</span></div>{{end}}{{end}}</div><div class="card panel" style="margin-top:18px"><h3>Recent changes <a style="float:right;font-size:13px" href="/ui/changes">View all</a></h3><ul class="list">{{range recentChanges .Changes}}<li><div><a href="/ui/changes/{{changeNumberOrID .}}"><b>{{changeNumberOrID .}}</b></a><div class="small">{{get . "applicationName"}} · Environment: {{get . "targetEnvironment"}} · Requested by: {{get . "requestedBy"}}</div></div><span class="badge {{badgeClass (get . "runtimeStatus")}}">{{get . "runtimeStatus"}}</span></li>{{end}}</ul></div></div><div>{{template "changeCard" .}}</div><div><div class="card timeline"><h3>Workflow Change</h3>{{range .Events}}<div class="step"><span class="circle done"></span><div><b>{{eventStep .}}</b><div class="small">{{get . "createdAt"}}</div></div></div>{{else}}<div class="small">No events available</div>{{end}}</div><div class="card evidence" style="margin-top:18px"><h3>Available evidence</h3>{{range .Evidence}}<div class="ev-row"><div><b>{{get . "name"}}</b><div class="small">{{get . "summary"}}</div></div><a href="/ui/changes/{{get . "changeNumber"}}/evidence">View</a></div>{{else}}<div class="small">No evidence available</div>{{end}}</div></div></div>{{end}}
{{define "changeCard"}}<div class="card detail">{{if .SelectedChange}}<div class="detail-head"><h3>Change Request: {{changeNumberOrID .SelectedChange}}</h3><span class="badge {{badgeClass (get .SelectedChange "runtimeStatus")}}">{{get .SelectedChange "runtimeStatus"}}</span></div><div class="kv"><div><div class="label">Application</div><b>{{get .SelectedChange "applicationName"}}</b></div><div><div class="label">Requester</div><b>{{get .SelectedChange "requestedBy"}}</b></div><div><div class="label">Environment</div><b>{{get .SelectedChange "targetEnvironment"}}</b></div><div><div class="label">Process lifecycle status</div><span class="badge {{badgeClass (get .SelectedChange "status")}}">{{get .SelectedChange "status"}}</span></div><div><div class="label">Technical runtime status</div><span class="badge {{badgeClass (get .SelectedChange "runtimeStatus")}}">{{get .SelectedChange "runtimeStatus"}}</span></div><div><div class="label">Created at</div><b>{{get .SelectedChange "createdAt"}}</b></div></div><div class="section"><h3>Status meaning</h3><p class="small">Process lifecycle status tracks the governance and approval state of the ChangeRequest. Technical runtime status tracks the latest automation, validation, deployment or evidence observation.</p></div><div class="section"><h3>Description</h3><p>{{get .SelectedChange "description"}}</p><div class="actions"><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}/evidence">View evidence</a><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}/events">View audit events</a></div></div><div class="section"><h3>Technical actions</h3>{{with environmentActionWarning .SelectedChange}}<div class="alert alert-error">{{.}}</div>{{end}}{{if and (environmentAllowsTechnicalActions .SelectedChange) (userCanSeeTechnicalActions .SelectedChange)}}<div class="action-groups"><div><div class="small" style="margin-bottom:8px">Recommended next actions</div><div class="actions">{{range recommendedActions .SelectedChange}}<div class="action-card"><form method="post" action="/ui/changes/{{changeNumberOrID $.SelectedChange}}/actions/{{get . "name"}}"><button class="btn {{if get . "primary"}}primary{{end}}">{{get . "label"}}</button></form><div class="small action-desc">{{get . "description"}}</div></div>{{else}}<div class="small">No recommended technical action for the current state.</div>{{end}}</div></div><details><summary>Advanced/manual actions</summary><div class="actions" style="margin-top:10px">{{range advancedActions .SelectedChange}}<div class="action-card"><form method="post" action="/ui/changes/{{changeNumberOrID $.SelectedChange}}/actions/{{get . "name"}}"><button class="btn">{{get . "label"}}</button></form><div class="small action-desc">{{get . "description"}}</div></div>{{end}}</div></details></div>{{else}}<div class="small">Technical actions are not available because the target environment is not currently enabled for automation.</div>{{end}}</div><div class="section"><h3>Technical runtime state</h3><div class="evidence-grid"><div class="evidence-card"><h4>Source repository</h4>{{with get .ChangeRuntimeState "source"}}<div class="evidence-kv"><span>Provider</span><b>{{get . "provider"}}</b></div><div class="evidence-kv"><span>Repository</span><b>{{get . "repositoryURL"}}</b></div><div class="evidence-kv"><span>Branch</span><b>{{get . "branch"}}</b></div><div class="evidence-kv"><span>Commit</span><b>{{short (get . "commitSHA")}}</b></div>{{else}}<div class="small">No runtime state recorded</div>{{end}}</div><div class="evidence-card"><h4>Build artifact state</h4>{{with get .ChangeRuntimeState "artifact"}}<div class="evidence-kv"><span>Provider</span><b>{{get . "provider"}}</b></div><div class="evidence-kv"><span>PipelineRun</span><b>{{get . "pipelineRunName"}}</b></div><div class="evidence-kv"><span>Source commit</span><b>{{short (get . "sourceCommitSHA")}}</b></div><div class="evidence-kv"><span>Image</span><b>{{get . "imageRepository"}}:{{get . "imageTag"}}</b></div><div class="evidence-kv"><span>Digest</span><b>{{short (get . "imageDigest")}}</b></div><div class="evidence-kv"><span>Status</span><b><span class="badge {{badgeClass (get . "status")}}">{{get . "status"}}</span></b></div><div class="evidence-kv"><span>Reason</span><b>{{get . "reason"}}</b></div>{{else}}<div class="small">No runtime state recorded</div>{{end}}</div><div class="evidence-card"><h4>GitOps repository</h4>{{with get .ChangeRuntimeState "gitops"}}<div class="evidence-kv"><span>Provider</span><b>{{get . "provider"}}</b></div><div class="evidence-kv"><span>Repository</span><b>{{get . "repositoryURL"}}</b></div><div class="evidence-kv"><span>Revision</span><b>{{get . "revision"}}</b></div><div class="evidence-kv"><span>Commit</span><b>{{short (get . "commitSHA")}}</b></div>{{else}}<div class="small">No runtime state recorded</div>{{end}}</div><div class="evidence-card"><h4>Tekton validation state</h4>{{with get .ChangeRuntimeState "tekton"}}<div class="evidence-kv"><span>PipelineRun</span><b>{{get . "pipelineRunName"}}</b></div><div class="evidence-kv"><span>Namespace</span><b>{{get . "namespace"}}</b></div><div class="evidence-kv"><span>Status</span><b><span class="badge {{badgeClass (get . "status")}}">{{get . "status"}}</span></b></div><div class="evidence-kv"><span>Reason</span><b>{{get . "reason"}}</b></div>{{else}}<div class="small">No runtime state recorded</div>{{end}}</div><div class="evidence-card"><h4>Argo CD deployment state</h4>{{with get .ChangeRuntimeState "argocd"}}<div class="evidence-kv"><span>Application</span><b>{{get . "applicationName"}}</b></div><div class="evidence-kv"><span>Sync</span><b>{{get . "syncStatus"}}</b></div><div class="evidence-kv"><span>Health</span><b>{{get . "healthStatus"}}</b></div><div class="evidence-kv"><span>Correlation</span><b>{{get . "correlationStatus"}}</b></div>{{else}}<div class="small">No runtime state recorded</div>{{end}}</div><div class="evidence-card"><h4>Kubernetes runtime state</h4>{{with get .ChangeRuntimeState "runtime"}}<div class="evidence-kv"><span>Cluster</span><b>{{get . "clusterName"}}</b></div><div class="evidence-kv"><span>Namespace</span><b>{{get . "namespace"}}</b></div><div class="evidence-kv"><span>Resource</span><b>{{get . "resourceKind"}} / {{get . "resourceName"}}</b></div><div class="evidence-kv"><span>Status</span><b><span class="badge {{badgeClass (get . "status")}}">{{get . "status"}}</span></b></div><div class="evidence-kv"><span>Reason</span><b>{{get . "reason"}}</b></div>{{else}}<div class="small">No runtime state recorded</div>{{end}}</div></div></div>{{with latestValidationEvidence .Evidence}}<div class="section"><h3>Tekton validation</h3><div class="evidence-card" style="margin-top:12px"><h4>Latest validation evidence</h4><div class="evidence-kv"><span>PipelineRun</span><b>{{validationField . "pipelineRunName"}}</b></div><div class="evidence-kv"><span>Tekton namespace</span><b>{{validationField . "tektonNamespace"}}</b></div><div class="evidence-kv"><span>Pipeline</span><b>{{validationField . "pipelineName"}}</b></div><div class="evidence-kv"><span>Git revision</span><b>{{validationField . "revision"}}</b></div><div class="evidence-kv"><span>Validation path</span><b>{{validationField . "validationPath"}}</b></div><div class="evidence-kv"><span>Status</span><b><span class="badge {{badgeClass (validationField . "status")}}">{{validationField . "status"}}</span></b></div><div class="evidence-kv"><span>Reason</span><b>{{validationField . "reason"}}</b></div><div class="evidence-kv"><span>Failed tasks</span><b>{{validationField . "failedTaskCount"}}</b></div><div class="evidence-kv"><span>Summary</span><b>{{validationField . "summary"}}</b></div><div class="evidence-kv"><span>Evidence sanitized</span><b><span class="badge {{badgeClass (get . "sanitized")}}">{{get . "sanitized"}}</span></b></div></div><details><summary>View raw validation evidence</summary><pre class="json">{{jsonPretty .}}</pre></details></div>{{end}}{{with latestEvidence .Evidence}}<div class="section"><h3>Latest runtime evidence</h3><div class="small">{{get . "summary"}}</div>{{with diagnosticsSummary .}}<div class="evidence-card" style="margin-top:12px"><h4>Deployment diagnostics</h4><div class="evidence-kv"><span>Summary</span><b>{{get . "summary"}}</b></div><div class="evidence-kv"><span>Argo CD synced</span><b>{{get . "argocdSynced"}}</b></div><div class="evidence-kv"><span>Argo CD healthy</span><b>{{get . "argocdHealthy"}}</b></div><div class="evidence-kv"><span>Deployment ready</span><b>{{get . "deploymentReady"}}</b></div><div class="evidence-kv"><span>Replicas</span><b>{{get . "readyReplicas"}}</b></div><div class="evidence-kv"><span>Pods</span><b>{{get . "podsReady"}}</b></div><div class="evidence-kv"><span>Restarts</span><b>{{get . "totalRestarts"}}</b></div><div class="evidence-kv"><span>Service available</span><b>{{get . "serviceAvailable"}}</b></div><div class="evidence-kv"><span>Route available</span><b>{{get . "routeAvailable"}}</b></div>{{with get . "warnings"}}<div class="small" style="margin-top:8px"><b>Warnings</b><ul class="pod-list">{{range .}}<li>{{.}}</li>{{end}}</ul></div>{{end}}</div>{{end}}<div class="evidence-grid">{{with get (kubeSummary .) "deployment"}}<div class="evidence-card"><h4>Deployment</h4><div class="evidence-kv"><span>Name</span><b>{{get . "name"}}</b></div><div class="evidence-kv"><span>Namespace</span><b>{{get . "namespace"}}</b></div><div class="evidence-kv"><span>Ready</span><b>{{get . "readyReplicas"}}/{{get . "desiredReplicas"}}</b></div><div class="evidence-kv"><span>Available</span><b>{{get . "availableReplicas"}}</b></div><div class="evidence-kv"><span>Updated</span><b>{{get . "updatedReplicas"}}</b></div></div>{{end}}{{with get (kubeSummary .) "service"}}<div class="evidence-card"><h4>Service</h4><div class="evidence-kv"><span>Name</span><b>{{get . "name"}}</b></div><div class="evidence-kv"><span>Type</span><b>{{get . "type"}}</b></div><div class="evidence-kv"><span>Cluster IP</span><b>{{get . "clusterIP"}}</b></div></div>{{end}}{{with get (kubeSummary .) "route"}}<div class="evidence-card"><h4>Route</h4><div class="evidence-kv"><span>Host</span><b>{{get . "host"}}</b></div><div class="evidence-kv"><span>TLS</span><b>{{get . "tlsTermination"}}</b></div><div class="evidence-kv"><span>To</span><b>{{get . "to"}}</b></div></div>{{end}}{{with get (kubeSummary .) "pods"}}<div class="evidence-card"><h4>Pods</h4><ul class="pod-list">{{range .}}<li><b>{{get . "name"}}</b> - {{get . "phase"}}, ready={{get . "ready"}}, restarts={{get . "restartCount"}}, node={{get . "nodeName"}}</li>{{end}}</ul></div>{{end}}</div><details><summary>View raw deployment evidence</summary><pre class="json">{{jsonPretty .}}</pre></details></div>{{end}}{{else}}<p>No ChangeRequest available.</p>{{end}}</div>{{end}}
{{define "changesList"}}<div class="card panel full"><h3>Change Requests</h3><table class="table"><thead><tr><th>Change</th><th>Application</th><th>Requested by</th><th>Environment</th><th>Process lifecycle</th><th>Technical runtime</th><th>Action</th></tr></thead><tbody>{{range .Changes}}<tr><td><b>{{changeNumberOrID .}}</b></td><td>{{get . "applicationName"}}</td><td>{{get . "requestedBy"}}</td><td>{{get . "targetEnvironment"}}</td><td><span class="badge {{badgeClass (get . "status")}}">{{get . "status"}}</span></td><td><span class="badge {{badgeClass (get . "runtimeStatus")}}">{{get . "runtimeStatus"}}</span></td><td><a href="/ui/changes/{{changeNumberOrID .}}">Open</a></td></tr>{{end}}</tbody></table></div>{{end}}

{{define "settingsPage"}}<div class="grid"><div class="full"><div class="card detail"><div class="detail-head"><h3>Settings</h3><span class="badge badge-info">MVP</span></div><div class="kv"><div><div class="label">Readiness</div><b><a href="/readyz">/readyz</a></b><div class="small">Technical readiness endpoint for API and database checks.</div></div><div><div class="label">Environment</div><b>dev</b><div class="small">Static display only. Multi-environment selection is not implemented yet.</div></div><div><div class="label">User</div><b>admin</b><div class="small">Static placeholder only. Authentication and authorization are planned for Phase 9.3.</div></div><div><div class="label">Version</div><b>v0.1.0</b><div class="small">UI MVP for lab validation.</div></div></div><div class="section"><h3>Production readiness notes</h3><ul class="pod-list"><li>AuthN/AuthZ is intentionally deferred to Phase 9.3.</li><li>Environment selector is intentionally static until multi-environment support is implemented.</li><li>Readiness remains available as JSON at <a href="/readyz">/readyz</a>.</li><li>OpenShift deployment, TLS, secrets and RBAC will be handled in Phase 8 and Phase 9.</li></ul></div></div></div></div>{{end}}
{{define "applicationsList"}}<div class="card panel full"><h3>Logical Applications by Environment</h3>{{range .LogicalApplications}}<div class="logical-app-table-title">{{get . "name"}}</div><table class="table"><thead><tr><th>Environment</th><th>Argo CD Application</th><th>Namespace</th><th>Sync</th><th>Health</th><th>Action</th></tr></thead><tbody>{{range get . "environments"}}<tr><td><b>{{get . "environment"}}</b></td><td>{{get . "argocdApplicationName"}}</td><td>{{get . "kubernetesNamespace"}}</td><td><span class="badge {{badgeClass (get . "syncStatus")}}">{{get . "syncStatus"}}</span></td><td><span class="badge {{badgeClass (get . "healthStatus")}}">{{get . "healthStatus"}}</span></td><td><a href="/ui/applications/{{get . "argocdApplicationName"}}">Open</a></td></tr>{{end}}</tbody></table>{{else}}<div class="small empty-state">No logical application bindings configured.</div>{{end}}</div>{{if .StandaloneApplications}}<div class="card panel full" style="margin-top:18px"><h3>Standalone Argo CD Applications</h3><table class="table"><thead><tr><th>Name</th><th>Namespace</th><th>Sync</th><th>Health</th><th>Action</th></tr></thead><tbody>{{range .StandaloneApplications}}<tr><td><b>{{get . "name"}}</b></td><td>{{get . "targetNamespace"}}</td><td><span class="badge {{badgeClass (get . "syncStatus")}}">{{get . "syncStatus"}}</span></td><td><span class="badge {{badgeClass (get . "healthStatus")}}">{{get . "healthStatus"}}</span></td><td><a href="/ui/applications/{{get . "name"}}">Open</a></td></tr>{{end}}</tbody></table></div>{{end}}{{end}}
{{define "applicationDetail"}}<div class="grid"><div class="full"><div class="card detail"><div class="detail-head"><h3>Application: {{get .SelectedApplication "name"}}</h3><span class="badge {{badgeClass (get .SelectedApplication "healthStatus")}}">{{get .SelectedApplication "healthStatus"}}</span></div><div class="kv"><div><div class="label">Namespace</div><b>{{get .SelectedApplication "targetNamespace"}}</b></div><div><div class="label">Sync status</div><span class="badge {{badgeClass (get .SelectedApplication "syncStatus")}}">{{get .SelectedApplication "syncStatus"}}</span></div><div><div class="label">Health status</div><span class="badge {{badgeClass (get .SelectedApplication "healthStatus")}}">{{get .SelectedApplication "healthStatus"}}</span></div><div><div class="label">Revision</div><b>{{short (get .SelectedApplication "revision")}}</b></div><div><div class="label">Repository</div><b>{{get .SelectedApplication "repoURL"}}</b></div><div><div class="label">Path</div><b>{{get .SelectedApplication "path"}}</b></div></div><div class="section"><h3>Runtime summary</h3><pre class="json">{{jsonPretty .Runtime}}</pre></div></div></div><div class="card panel full"><h3>Resources</h3><table class="table"><thead><tr><th>Kind</th><th>Name</th><th>Namespace</th><th>Status</th></tr></thead><tbody>{{range .Resources}}<tr><td>{{get . "kind"}}</td><td>{{get . "name"}}</td><td>{{get . "namespace"}}</td><td><span class="badge {{badgeClass (get . "status")}}">{{get . "status"}}</span></td></tr>{{else}}<tr><td colspan="4" class="small">No resource details available.</td></tr>{{end}}</tbody></table></div><div class="card panel full"><h3>Deployment history</h3><table class="table"><thead><tr><th>Revision</th><th>Status</th><th>Deployed at</th></tr></thead><tbody>{{range .History}}<tr><td>{{short (get . "revision")}}</td><td><span class="badge {{badgeClass (get . "status")}}">{{get . "status"}}</span></td><td>{{get . "deployedAt"}}</td></tr>{{else}}<tr><td colspan="3" class="small">No deployment history available.</td></tr>{{end}}</tbody></table></div></div>{{end}}

{{define "changeEvidencePage"}}<div class="grid"><div class="full"><div class="card detail"><div class="detail-head"><h3>Evidence for {{changeNumberOrID .SelectedChange}}</h3><span class="badge {{badgeClass (get .SelectedChange "runtimeStatus")}}">{{get .SelectedChange "runtimeStatus"}}</span></div><div class="actions"><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}">Back to change</a><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}/events">View audit events</a></div></div></div><div class="card panel full"><h3>Collected evidence</h3><table class="table"><thead><tr><th>Name</th><th>Type</th><th>Summary</th><th>Sanitized</th><th>Created</th></tr></thead><tbody>{{range .Evidence}}<tr><td><b>{{get . "name"}}</b></td><td>{{get . "evidenceType"}}</td><td>{{get . "summary"}}</td><td><span class="badge {{badgeClass (get . "sanitized")}}">{{get . "sanitized"}}</span></td><td>{{get . "createdAt"}}</td></tr>{{else}}<tr><td colspan="5" class="small">No evidence available.</td></tr>{{end}}</tbody></table></div>{{range .Evidence}}<div class="card panel full"><h3>{{get . "name"}}</h3><pre class="json">{{jsonPretty .}}</pre></div>{{end}}</div>{{end}}

{{define "changesAPIPage"}}<div class="grid"><div class="full"><div class="card detail"><div class="detail-head"><h3>Changes API</h3><span class="badge badge-info">UI wrapper</span></div><p>This page shows the Change Requests API data inside the web UI, so users do not land directly on a raw JSON-only browser page.</p><div class="actions"><a class="btn primary" href="/">Back to dashboard</a><a class="btn" href="/ui/changes">Back to changes</a><a class="btn" href="/api/v1/changes">Open raw JSON API</a></div></div></div><div class="card panel full"><h3>Change Requests API preview</h3><pre class="json">{{jsonPretty .Changes}}</pre></div></div>{{end}}
{{define "changeEventsPage"}}<div class="grid"><div class="full"><div class="card detail"><div class="detail-head"><h3>Audit events for {{changeNumberOrID .SelectedChange}}</h3><span class="badge {{badgeClass (get .SelectedChange "status")}}">{{get .SelectedChange "status"}}</span></div><div class="actions"><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}">Back to change</a><a class="btn" href="/ui/changes/{{changeNumberOrID .SelectedChange}}/evidence">View evidence</a></div></div></div><div class="card panel full"><h3>Audit trail</h3><table class="table"><thead><tr><th>Event</th><th>Step</th><th>Previous</th><th>New</th><th>Source</th><th>Created</th></tr></thead><tbody>{{range .Events}}<tr><td>{{get . "eventType"}}</td><td>{{eventStep .}}</td><td>{{get . "previousStatus"}}</td><td>{{get . "newStatus"}}</td><td>{{get . "source"}}</td><td>{{get . "createdAt"}}</td></tr>{{else}}<tr><td colspan="6" class="small">No audit events available.</td></tr>{{end}}</tbody></table></div>{{range .Events}}<div class="card panel full"><h3>{{get . "eventType"}} - {{eventStep .}}</h3><pre class="json">{{jsonPretty .}}</pre></div>{{end}}</div>{{end}}
{{define "changeDetail"}}<div class="grid"><div class="full">{{template "changeCard" .}}</div><div class="card panel full"><h3>Audit events</h3><table class="table"><thead><tr><th>Event</th><th>Step</th><th>Created</th></tr></thead><tbody>{{range .Events}}<tr><td>{{get . "eventType"}}</td><td>{{eventStep .}}</td><td>{{get . "createdAt"}}</td></tr>{{end}}</tbody></table></div></div>{{end}}`
