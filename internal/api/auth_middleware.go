package api

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type authContextKey string

const identityContextKey authContextKey = "authIdentity"

type authIdentity struct {
	Username string
	Groups   []string
	Roles    map[string]bool
	Source   string
}

func authenticatedUsernameFromContext(ctx context.Context) (string, bool) {
	identity, ok := ctx.Value(identityContextKey).(authIdentity)
	if !ok {
		return "", false
	}

	username := strings.TrimSpace(identity.Username)
	if username == "" {
		return "", false
	}

	return username, true
}

func withAuthMiddleware(next http.Handler) http.Handler {
	if !getBoolEnv("AUTH_ENABLED", false) {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicEndpoint(r) {
			next.ServeHTTP(w, r)
			return
		}

		requiredRoles, classified := requiredRolesForRequest(r)
		if !classified {
			http.Error(w, "forbidden: endpoint is not classified", http.StatusForbidden)
			return
		}

		identity, ok := identityFromHeaders(r)
		if !ok {
			http.Error(w, "unauthorized: missing trusted identity headers", http.StatusUnauthorized)
			return
		}

		if !hasAnyRole(identity, requiredRoles) {
			http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), identityContextKey, identity)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isPublicEndpoint(r *http.Request) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}
	return r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/livez"
}

func identityFromHeaders(r *http.Request) (authIdentity, bool) {
	userHeader := getEnv("AUTH_HEADER_USER", "X-Forwarded-User")
	groupsHeader := getEnv("AUTH_HEADER_GROUPS", "X-Forwarded-Groups")
	altUserHeader := getEnv("AUTH_HEADER_ALT_USER", "X-Auth-Request-User")

	username := strings.TrimSpace(r.Header.Get(userHeader))
	if username == "" && altUserHeader != "" {
		username = strings.TrimSpace(r.Header.Get(altUserHeader))
	}
	if username == "" {
		return authIdentity{}, false
	}

	groups := splitCSV(r.Header.Get(groupsHeader))
	source := "header"

	if len(groups) == 0 && getBoolEnv("AUTH_OPENSHIFT_GROUP_LOOKUP_ENABLED", false) {
		resolvedGroups, err := openShiftGroupsForUser(r.Context(), username)
		if err == nil && len(resolvedGroups) > 0 {
			groups = resolvedGroups
			source = "openshift-groups"
		}
	}

	roles := rolesFromGroups(groups)

	return authIdentity{Username: username, Groups: groups, Roles: roles, Source: source}, true
}

type openShiftGroupList struct {
	Items []struct {
		Metadata struct {
			Name string `json:"name"`
		} `json:"metadata"`
		Users []string `json:"users"`
	} `json:"items"`
}

func openShiftGroupsForUser(ctx context.Context, username string) ([]string, error) {
	apiURL := strings.TrimRight(getEnv("AUTH_OPENSHIFT_API_URL", ""), "/")
	if apiURL == "" {
		host := strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_HOST"))
		port := strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_PORT"))
		if host == "" {
			return nil, fmt.Errorf("openshift api host is not configured")
		}
		if port == "" {
			port = "443"
		}
		apiURL = "https://" + host + ":" + port
	}

	tokenPath := getEnv("AUTH_OPENSHIFT_TOKEN_FILE", "/var/run/secrets/kubernetes.io/serviceaccount/token")
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("read service account token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return nil, fmt.Errorf("service account token is empty")
	}

	httpClient, err := openShiftHTTPClient()
	if err != nil {
		return nil, err
	}

	requestURL := apiURL + "/apis/user.openshift.io/v1/groups"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("openshift groups lookup failed: status=%d body=%q", resp.StatusCode, string(body))
	}

	var groupList openShiftGroupList
	if err := json.NewDecoder(resp.Body).Decode(&groupList); err != nil {
		return nil, err
	}

	var groups []string
	for _, item := range groupList.Items {
		for _, user := range item.Users {
			if user == username {
				groups = append(groups, item.Metadata.Name)
				break
			}
		}
	}

	return groups, nil
}

func openShiftHTTPClient() (*http.Client, error) {
	caFile := getEnv("AUTH_OPENSHIFT_CA_FILE", "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	caBytes, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read openshift ca file: %w", err)
	}

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(caBytes) {
		return nil, fmt.Errorf("openshift ca file does not contain valid PEM certificates")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    roots,
			MinVersion: tls.VersionTLS12,
		},
	}

	return &http.Client{
		Timeout:   5 * time.Second,
		Transport: transport,
	}, nil
}

func rolesFromGroups(groups []string) map[string]bool {
	roles := map[string]bool{}
	groupSet := map[string]bool{}
	for _, group := range groups {
		groupSet[group] = true
	}

	addRoleIfGroupMatches(roles, groupSet, "viewer", getEnv("AUTH_GROUP_VIEWER", ""))
	addRoleIfGroupMatches(roles, groupSet, "operator", getEnv("AUTH_GROUP_OPERATOR", ""))
	addRoleIfGroupMatches(roles, groupSet, "approver", getEnv("AUTH_GROUP_APPROVER", ""))
	addRoleIfGroupMatches(roles, groupSet, "admin", getEnv("AUTH_GROUP_ADMIN", ""))

	if roles["operator"] || roles["approver"] || roles["admin"] {
		roles["viewer"] = true
	}
	if roles["admin"] {
		roles["operator"] = true
		roles["approver"] = true
	}

	return roles
}

func addRoleIfGroupMatches(roles map[string]bool, groupSet map[string]bool, role string, configuredGroups string) {
	for _, group := range splitCSV(configuredGroups) {
		if groupSet[group] {
			roles[role] = true
			return
		}
	}
}

func hasAnyRole(identity authIdentity, requiredRoles []string) bool {
	for _, role := range requiredRoles {
		if identity.Roles[role] {
			return true
		}
	}
	return false
}

func requiredRolesForRequest(r *http.Request) ([]string, bool) {
	path := r.URL.Path
	method := r.Method

	if method == http.MethodGet {
		switch {
		case path == "/" || path == "/ui" || path == "/ui/dashboard":
			return []string{"viewer", "operator", "approver", "admin"}, true
		case path == "/ui/settings":
			return []string{"admin"}, true
		case path == "/ui/changes-api":
			return []string{"viewer", "operator", "approver", "admin"}, true
		case path == "/ui/changes" || strings.HasPrefix(path, "/ui/changes/"):
			return []string{"viewer", "operator", "approver", "admin"}, true
		case path == "/ui/applications" || strings.HasPrefix(path, "/ui/applications/"):
			return []string{"viewer", "operator", "approver", "admin"}, true
		case path == "/api/v1/applications" || strings.HasPrefix(path, "/api/v1/applications/"):
			return []string{"viewer", "operator", "approver", "admin"}, true
		case path == "/api/v1/changes" || strings.HasPrefix(path, "/api/v1/changes/"):
			return []string{"viewer", "operator", "approver", "admin"}, true
		}
	}

	if method == http.MethodPost {
		if path == "/api/v1/changes" {
			return []string{"operator", "admin"}, true
		}
		if strings.HasPrefix(path, "/ui/changes/") && strings.Contains(path, "/actions/") {
			action := path[strings.LastIndex(path, "/")+1:]
			return requiredRolesForAction(action)
		}
		if strings.HasPrefix(path, "/api/v1/changes/") {
			action := path[strings.LastIndex(path, "/")+1:]
			return requiredRolesForAction(action)
		}
	}

	return nil, false
}

func requiredRolesForAction(action string) ([]string, bool) {
	switch action {
	case "validate", "check-validation", "check-deployment", "collect-evidence", "create-branch", "update-files", "open-merge-request", "sync":
		return []string{"operator", "admin"}, true
	case "submit", "approve", "reject", "start-execution", "complete-execution", "fail-execution", "close", "cancel", "merge-request":
		return []string{"approver", "admin"}, true
	default:
		return nil, false
	}
}

func splitCSV(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getBoolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
