package api

import (
	"strings"
	"testing"
)

func TestUIDataActiveNavField(t *testing.T) {
	tests := []struct {
		name        string
		data        uiData
		expectedNav string
	}{
		{
			name:        "dashboard sets ActiveNav to dashboard",
			data:        uiData{ActiveNav: "dashboard"},
			expectedNav: "dashboard",
		},
		{
			name:        "applications sets ActiveNav to applications",
			data:        uiData{ActiveNav: "applications"},
			expectedNav: "applications",
		},
		{
			name:        "all changes sets ActiveNav to all-changes",
			data:        uiData{ActiveNav: "all-changes"},
			expectedNav: "all-changes",
		},
		{
			name:        "change requests sets ActiveNav to change-requests",
			data:        uiData{ActiveNav: "change-requests"},
			expectedNav: "change-requests",
		},
		{
			name:        "changes API sets ActiveNav to changes-api",
			data:        uiData{ActiveNav: "changes-api"},
			expectedNav: "changes-api",
		},
		{
			name:        "evidence sets ActiveNav to evidence",
			data:        uiData{ActiveNav: "evidence"},
			expectedNav: "evidence",
		},
		{
			name:        "audit log sets ActiveNav to audit-log",
			data:        uiData{ActiveNav: "audit-log"},
			expectedNav: "audit-log",
		},
		{
			name:        "settings sets ActiveNav to settings",
			data:        uiData{ActiveNav: "settings"},
			expectedNav: "settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data.ActiveNav != tt.expectedNav {
				t.Errorf("ActiveNav = %s, want %s", tt.data.ActiveNav, tt.expectedNav)
			}
		})
	}
}

func TestTemplateUsesActiveNav(t *testing.T) {
	testCases := []struct {
		name     string
		navItem  string
		expected string
	}{
		{
			name:     "Dashboard link uses ActiveNav",
			navItem:  "Dashboard",
			expected: `class="{{if eq .ActiveNav "dashboard"}}active{{end}}"`,
		},
		{
			name:     "Applications link uses ActiveNav",
			navItem:  "Applications",
			expected: `class="{{if eq .ActiveNav "applications"}}active{{end}}"`,
		},
		{
			name:     "Change Requests parent link uses ActiveNav",
			navItem:  "Change Requests",
			expected: `class="{{if eq .ActiveNav "change-requests"}}active{{end}}"`,
		},
		{
			name:     "All changes link uses ActiveNav",
			navItem:  "All changes",
			expected: `class="{{if eq .ActiveNav "all-changes"}}active{{end}}"`,
		},
		{
			name:     "Changes API link uses ActiveNav",
			navItem:  "Changes API",
			expected: `class="{{if eq .ActiveNav "changes-api"}}active{{end}}"`,
		},
		{
			name:     "Evidence link uses ActiveNav",
			navItem:  "Evidence",
			expected: `class="{{if eq .ActiveNav "evidence"}}active{{end}}"`,
		},
		{
			name:     "Audit log link uses ActiveNav",
			navItem:  "Audit log",
			expected: `class="{{if eq .ActiveNav "audit-log"}}active{{end}}"`,
		},
		{
			name:     "Settings link uses ActiveNav",
			navItem:  "Settings",
			expected: `class="{{if eq .ActiveNav "settings"}}active{{end}}"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(uiTemplate, tc.expected) {
				t.Errorf("Template does not contain expected pattern for %s: %s", tc.navItem, tc.expected)
			}
		})
	}
}

func TestTemplateNoActiveFieldInNavigation(t *testing.T) {
	navigationSection := extractTemplateSection(uiTemplate, `<nav class="nav">`, `</nav>`)

	forbiddenPatterns := []string{
		`class="{{if eq .Active "changes"}}active{{end}}"`,
		`class="{{if eq .Active "dashboard"}}active{{end}}"`,
		`class="{{if eq .Active "applications"}}active{{end}}"`,
		`class="{{if eq .Active "settings"}}active{{end}}"`,
	}

	for _, pattern := range forbiddenPatterns {
		if strings.Contains(navigationSection, pattern) {
			t.Errorf("Navigation section should not use .Active field, found: %s", pattern)
		}
	}
}

func TestExactlyOneActiveNavLinkPerState(t *testing.T) {
	navigationSection := extractTemplateSection(uiTemplate, `<nav class="nav">`, `</nav>`)

	navStates := []string{
		"dashboard",
		"applications",
		"change-requests",
		"all-changes",
		"changes-api",
		"evidence",
		"audit-log",
		"settings",
	}

	for _, state := range navStates {
		pattern := `class="{{if eq .ActiveNav "` + state + `"}}active{{end}}"`
		count := strings.Count(navigationSection, pattern)
		if count != 1 {
			t.Errorf("Expected exactly 1 link for ActiveNav state %q, found %d instances of pattern: %s", state, count, pattern)
		}
	}
}

func extractTemplateSection(template, start, end string) string {
	startIdx := strings.Index(template, start)
	if startIdx == -1 {
		return ""
	}
	endIdx := strings.Index(template[startIdx:], end)
	if endIdx == -1 {
		return ""
	}
	return template[startIdx : startIdx+endIdx+len(end)]
}
