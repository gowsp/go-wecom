package wecom_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The capability ledger is the authority for what the SDK supports. The coverage
// document only renders it: the ledger decides whether a page is implemented,
// the source scan provides the jump target, and docs/coverage.md must agree with
// both.
const capabilityLedgerPath = "docs/capabilities.json"

// capability is one official documentation page as the SDK sees it.
type capability struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Group       string   `json:"group"`
	Category    string   `json:"category"`
	Package     string   `json:"package"`
	Kind        string   `json:"kind"`
	Auth        string   `json:"auth"`
	Status      string   `json:"status"`
	Integration string   `json:"integration"`
	Methods     []string `json:"methods"`
	Notes       string   `json:"notes"`
}

// Capability states. "partial" means the page exposes several interfaces and only
// some of them are implemented; the gap must be described in Notes.
const (
	statusImplemented   = "implemented"
	statusPartial       = "partial"
	statusPending       = "pending"
	statusNotApplicable = "not_applicable"
)

// Integration states. Production credentials are never part of the repository,
// so "untested" is the honest default for everything landed without a live run.
const (
	integrationUntested = "untested"
	integrationPassed   = "passed"
	integrationBlocked  = "blocked"
)

var (
	capabilityKinds = map[string]bool{
		"api": true, "callback": true, "event": true, "upload": true,
		"download": true, "async": true, "special": true,
	}
	capabilityAuths = map[string]bool{
		"corp": true, "callback": true, "webhook": true, "bot": true,
		"payment": true, "archive": true, "external": true,
	}
	capabilityStatuses = map[string]bool{
		statusImplemented: true, statusPartial: true,
		statusPending: true, statusNotApplicable: true,
	}
	capabilityIntegrations = map[string]bool{
		integrationUntested: true, integrationPassed: true, integrationBlocked: true,
	}
)

// capabilityDone reports whether the ledger claims the SDK supports a page.
func capabilityDone(c capability) bool {
	return c.Status == statusImplemented || c.Status == statusPartial
}

// loadCapabilities reads the ledger and returns it keyed by documentation id.
func loadCapabilities(t *testing.T) map[string]capability {
	t.Helper()
	raw, err := os.ReadFile(capabilityLedgerPath)
	if err != nil {
		t.Fatalf("read %s: %v", capabilityLedgerPath, err)
	}
	var ledger struct {
		Version int          `json:"version"`
		Entries []capability `json:"entries"`
	}
	if err := json.Unmarshal(raw, &ledger); err != nil {
		t.Fatalf("parse %s: %v", capabilityLedgerPath, err)
	}
	if ledger.Version != 1 {
		t.Fatalf("%s: unsupported version %d", capabilityLedgerPath, ledger.Version)
	}
	out := make(map[string]capability, len(ledger.Entries))
	for _, entry := range ledger.Entries {
		if _, dup := out[entry.ID]; dup {
			t.Fatalf("%s: duplicate id %s", capabilityLedgerPath, entry.ID)
		}
		out[entry.ID] = entry
	}
	if len(out) == 0 {
		t.Fatalf("%s: no entries", capabilityLedgerPath)
	}
	return out
}

// TestCapabilityLedger keeps the ledger internally consistent and checks that
// every documented method is accounted for. It runs offline.
func TestCapabilityLedger(t *testing.T) {
	caps := loadCapabilities(t)
	methods := scanDocumentedMethods(t)

	for id, entry := range caps {
		if entry.Title == "" || entry.Category == "" {
			t.Errorf("path/%s: title and category are required", id)
		}
		if entry.Status == statusNotApplicable {
			if entry.Notes == "" {
				t.Errorf("path/%s: not_applicable entries must explain why in notes", id)
			}
		} else if entry.Package == "" {
			t.Errorf("path/%s: package is required", id)
		}
		if !capabilityKinds[entry.Kind] {
			t.Errorf("path/%s: unknown kind %q", id, entry.Kind)
		}
		if !capabilityAuths[entry.Auth] {
			t.Errorf("path/%s: unknown auth %q", id, entry.Auth)
		}
		if !capabilityStatuses[entry.Status] {
			t.Errorf("path/%s: unknown status %q", id, entry.Status)
		}
		if !capabilityIntegrations[entry.Integration] {
			t.Errorf("path/%s: unknown integration state %q", id, entry.Integration)
		}
		if capabilityDone(entry) {
			if len(entry.Methods) == 0 {
				t.Errorf("path/%s: supported entry lists no methods", id)
			}
			if len(methods[id]) == 0 {
				t.Errorf("path/%s: supported entry has no documented method in the source", id)
			}
		}
		if entry.Status == statusPartial && entry.Notes == "" {
			t.Errorf("path/%s: partial support must describe what is missing in notes", id)
		}
		if entry.Status == statusPending && len(methods[id]) > 0 {
			t.Errorf("path/%s: source documents methods but the ledger says pending", id)
		}
	}

	for id := range methods {
		entry, ok := caps[id]
		if !ok {
			t.Errorf("path/%s: documented in the source but missing from %s", id, capabilityLedgerPath)
			continue
		}
		if !capabilityDone(entry) {
			t.Errorf("path/%s: source implements it but the ledger says %q", id, entry.Status)
		}
	}

	// Pending entries exist only once a phase starts working on them; keep the
	// list sorted so the file stays reviewable.
	var pending []string
	for id, entry := range caps {
		if entry.Status == statusPending {
			pending = append(pending, id)
		}
	}
	sort.Strings(pending)
	var skipped int
	for _, entry := range caps {
		if entry.Status == statusNotApplicable {
			skipped++
		}
	}
	t.Logf("ledger: %d entries, %d supported, %d pending, %d not applicable",
		len(caps), len(caps)-len(pending)-skipped, len(pending), skipped)
}

// declaredFunctions returns the function and method names declared by the
// non-test Go files of one package directory.
func declaredFunctions(t *testing.T, dir string) map[string]bool {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read package %s: %v", dir, err)
	}
	names := map[string]bool{}
	decl := regexp.MustCompile(`(?m)^func (?:\([^)]*\) )?([A-Za-z_]\w*)\(`)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") ||
			strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		for _, match := range decl.FindAllStringSubmatch(string(content), -1) {
			names[match[1]] = true
		}
	}
	return names
}

// TestCapabilityMethodsExist keeps the ledger honest about the Go surface: every
// method it lists must be declared by the package it names. It runs offline.
func TestCapabilityMethodsExist(t *testing.T) {
	caps := loadCapabilities(t)
	packages := map[string]map[string]bool{}
	for id, entry := range caps {
		if !capabilityDone(entry) {
			continue
		}
		dir := entry.Package
		if dir == "" || dir == "wecom" {
			dir = "."
		}
		names, ok := packages[dir]
		if !ok {
			names = declaredFunctions(t, dir)
			packages[dir] = names
		}
		if len(names) == 0 {
			t.Errorf("path/%s: package %q declares no functions", id, entry.Package)
			continue
		}
		for _, method := range entry.Methods {
			if !names[method] {
				t.Errorf("path/%s: package %q does not declare %s", id, entry.Package, method)
			}
		}
	}
	t.Logf("verified %d packages", len(packages))
}
