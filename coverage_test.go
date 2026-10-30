package wecom_test

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

// docs/coverage.md is generated: the tree comes from the official documentation
// sidebar, the check marks come from the `// 文档:` comments in this module.
// Run `go test -run TestCoverageDocument -update-coverage .` to fetch the live
// tree and rewrite the file.
var updateCoverage = flag.Bool("update-coverage", false,
	"fetch the official documentation tree and rewrite docs/coverage.md")

var updateCoverageSources = flag.Bool("update-coverage-sources", false,
	"refresh source links in docs/coverage.md without fetching the official tree")

const (
	coverageDocPath = "docs/coverage.md"
	officialNavURL  = "https://developer.work.weixin.qq.com/document/path/90664"
	docBaseURL      = "https://developer.work.weixin.qq.com/document/path/"
	fetchTimeout    = 30 * time.Second
)

var (
	// linkPattern matches one generated bullet: [title](url).
	linkPattern = regexp.MustCompile(`^\[([^\]]+)\]\((https?://[^)]+)\)`)
	// docLinkPattern matches the canonical documentation comment carried by
	// every API method.
	docLinkPattern = regexp.MustCompile(`https://developer\.work\.weixin\.qq\.com/document/path/(\d+)`)
	// pathIDPattern matches the sidebar's relative hrefs.
	pathIDPattern = regexp.MustCompile(`/document/path/(\d+)`)
	// funcPattern identifies a declaration so a documentation comment can be
	// attributed to the method that follows it.
	funcPattern = regexp.MustCompile(`^func (?:\(.*?\*(\w+)\) )?(\w+)\(`)
	// tagPattern tokenises the sidebar markup; its attributes never contain an
	// unescaped '>'.
	tagPattern   = regexp.MustCompile(`<(/?)(div|a|span|img)\b([^>]*)>`)
	hrefPattern  = regexp.MustCompile(`href="([^"]*)"`)
	classPattern = regexp.MustCompile(`class="([^"]*)"`)
	levelPattern = regexp.MustCompile(`level="(\d+)"`)
	// commentPattern strips the Vue placeholders ("<!---->") that sit next to
	// every sidebar entry.
	commentPattern = regexp.MustCompile(`(?s)<!--.*?-->`)
)

// navNode is one entry of the official documentation tree.
type navNode struct {
	title string
	id    string // documentation path id; empty for a category heading
	group bool   // top-level divider such as 基础 / 连接微信 / 办公
	kids  []*navNode
}

// TestCoverageDocument regenerates docs/coverage.md from the live official tree
// when -update-coverage is set, and otherwise verifies that the checked-in
// document still agrees with the code.
func TestCoverageDocument(t *testing.T) {
	if *updateCoverage && *updateCoverageSources {
		t.Fatal("choose only one update mode")
	}
	if *updateCoverageSources {
		raw, err := os.ReadFile(coverageDocPath)
		if err != nil {
			t.Fatal(err)
		}
		rendered, err := refreshCoverageSources(string(raw), coverageInput{
			caps: loadCapabilities(t), methods: scanDocumentedMethods(t),
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(coverageDocPath, []byte(rendered), 0o644); err != nil {
			t.Fatal(err)
		}
		verifyCoverageDocument(t)
		return
	}
	if !*updateCoverage {
		verifyCoverageDocument(t)
		return
	}

	tree := fetchOfficialTree(t)
	in := coverageInput{caps: loadCapabilities(t), methods: scanDocumentedMethods(t)}
	if err := validateOfficialTree(tree, in); err != nil {
		t.Fatal(err)
	}
	rendered := renderCoverage(tree, in)
	if err := os.WriteFile(coverageDocPath, []byte(rendered), 0o644); err != nil {
		t.Fatalf("write %s: %v", coverageDocPath, err)
	}
	t.Logf("wrote %s: %d pages, %d lines", coverageDocPath, countLeaves(tree), strings.Count(rendered, "\n"))
}

// verifyCoverageDocument checks the checked-in document against the code
// without touching the network, so `go test ./...` works offline.
func verifyCoverageDocument(t *testing.T) {
	t.Helper()
	covered, pending, explainer, total := readCoverageDocument(t, coverageDocPath)
	in := coverageInput{caps: loadCapabilities(t), methods: scanDocumentedMethods(t)}
	inCode := make(map[string]bool, len(in.caps))
	for id, entry := range in.caps {
		if capabilityDone(entry) {
			inCode[id] = true
		}
	}

	if total < 700 {
		t.Errorf("%s looks truncated: only %d documented pages", coverageDocPath, total)
	}
	for _, id := range difference(inCode, covered) {
		t.Errorf("implemented but not checked off: path/%s", id)
	}
	for _, id := range difference(covered, inCode) {
		t.Errorf("checked off but no method documents it: path/%s", id)
	}
	for _, id := range intersection(inCode, pending) {
		t.Errorf("implemented but left unchecked: path/%s", id)
	}
	for _, id := range intersection(inCode, explainer) {
		t.Errorf("implemented but listed as an explainer page: path/%s", id)
	}
	raw, err := os.ReadFile(coverageDocPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateCoverageSources(string(raw), in); err != nil {
		t.Errorf("%v (refresh: go test -run TestCoverageDocument -update-coverage-sources .)", err)
	}
}

// readCoverageDocument parses the generated document and buckets documentation
// ids by their checkbox state.
func readCoverageDocument(t *testing.T, path string) (covered, pending, explainer map[string]bool, total int) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s (run go test -run TestCoverageDocument -update-coverage .): %v", path, err)
	}
	covered, pending, explainer = map[string]bool{}, map[string]bool{}, map[string]bool{}
	seen := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		matches := docLinkPattern.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}
		if len(matches) > 1 {
			t.Fatalf("one item must reference a single page: %s", line)
		}
		id := matches[0][1]
		if previous, ok := seen[id]; ok {
			t.Fatalf("%s lists path/%s twice:\n%s\n%s", path, id, previous, line)
		}
		seen[id] = line
		total++
		switch {
		case strings.Contains(line, "[x]"):
			covered[id] = true
		case strings.Contains(line, "[ ]"):
			pending[id] = true
		default:
			explainer[id] = true
		}
	}
	return covered, pending, explainer, total
}

// scanDocumentedMethods records, for every documentation path, the locations of
// the methods that document it.
func scanDocumentedMethods(t *testing.T) map[string][]sourceRef {
	t.Helper()
	found := map[string][]sourceRef{}
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			name := entry.Name()
			if path != "." && (strings.HasPrefix(name, ".") || name == "vendor" || name == "docs") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if !strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			match := docLinkPattern.FindStringSubmatch(line)
			if match == nil {
				continue
			}
			if decl, line := declarationAfter(lines, i); decl != "" {
				found[match[1]] = append(found[match[1]], sourceRef{
					file:        filepath.ToSlash(path),
					line:        line,
					constructor: isConstructor(decl),
				})
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan source: %v", err)
	}
	if len(found) == 0 {
		t.Fatal("no documented methods found")
	}
	return found
}

// declarationAfter finds the function declared after a documentation comment and
// returns its name with its 1-based line number.
func declarationAfter(lines []string, comment int) (name string, line int) {
	for i := comment + 1; i < len(lines) && i < comment+30; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		if match := funcPattern.FindStringSubmatch(lines[i]); match != nil {
			return match[2], i + 1
		}
		return "", 0
	}
	return "", 0
}

// isConstructor reports whether a name only builds a client instead of calling
// the documented interface.
func isConstructor(name string) bool {
	return name == "New" || strings.HasPrefix(name, "New")
}

func difference(a, b map[string]bool) []string {
	var out []string
	for key := range a {
		if !b[key] {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

func intersection(a, b map[string]bool) []string {
	var out []string
	for key := range a {
		if b[key] {
			out = append(out, key)
		}
	}
	sort.Strings(out)
	return out
}

// docNode is one bullet of the generated document, kept so an offline refresh
// can recompute both leaf rows and category rows from the ledger.
type docNode struct {
	line   int
	indent int
	marker string // "x", " " or "" for explainer pages
	body   string // title and official link, without checkbox and links
	title  string
	id     string
	kids   []*docNode
}

// refreshCoverageSources re-renders every bullet already present in the tree. The
// ledger owns the checkbox, so pages and categories can be updated offline;
// adding or removing official pages still needs a full network refresh.
func refreshCoverageSources(document string, in coverageInput) (string, error) {
	r := &coverageRenderer{caps: in.caps, methods: in.methods}
	lines := strings.Split(document, "\n")
	roots, err := parseDocNodes(lines)
	if err != nil {
		return "", err
	}

	seen := map[string]bool{}
	var walk func(nodes []*docNode) (done, missing int)
	walk = func(nodes []*docNode) (done, missing int) {
		for _, node := range nodes {
			if len(node.kids) > 0 {
				childDone, childMissing := walk(node.kids)
				done += childDone
				missing += childMissing
				if childDone == 0 && childMissing == 0 {
					// Nothing applicable below this heading: list it plainly.
					lines[node.line] = strings.Repeat(" ", node.indent) + "- " +
						strings.TrimSuffix(node.body, "（**部分支持**）")
					continue
				}
				state, partial := " ", false
				switch {
				case childDone == 0:
				case childMissing == 0:
					state = "x"
				default:
					state, partial = "x", true
				}
				suffix := ""
				if partial {
					suffix = "（**部分支持**）"
				}
				lines[node.line] = strings.Repeat(" ", node.indent) + "- [" + state + "] " +
					strings.TrimSuffix(node.body, "（**部分支持**）") + suffix
				continue
			}
			if node.id == "" {
				continue
			}
			if seen[node.id] {
				err = fmt.Errorf("duplicate path/%s", node.id)
				return
			}
			seen[node.id] = true
			entry, ok := in.caps[node.id]
			if !ok {
				// Explainer pages and pages not yet triaged keep their line as is.
				continue
			}
			switch {
			case capabilityDone(entry):
				done++
				line := strings.Repeat(" ", node.indent) + "- [x] " + node.body
				if source := r.source(node.id); source != "" {
					line += " — " + source
				}
				lines[node.line] = line
			case entry.Status == statusNotApplicable:
				lines[node.line] = strings.Repeat(" ", node.indent) + "- " + node.body
			default:
				missing++
				lines[node.line] = strings.Repeat(" ", node.indent) + "- [ ] " + node.body
			}
		}
		return
	}
	walk(roots)
	if err != nil {
		return "", err
	}
	for id, entry := range in.caps {
		if capabilityDone(entry) && !seen[id] {
			return "", fmt.Errorf("supported path/%s missing from the document; run -update-coverage", id)
		}
	}
	return strings.Join(lines, "\n"), nil
}

// parseDocNodes groups the bullet lines by indentation into a forest.
func parseDocNodes(lines []string) ([]*docNode, error) {
	var roots []*docNode
	var stack []*docNode
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		node := &docNode{line: i, indent: len(line) - len(trimmed)}
		rest := trimmed[2:]
		switch {
		case strings.HasPrefix(rest, "[x] "):
			node.marker, rest = "x", rest[4:]
		case strings.HasPrefix(rest, "[ ] "):
			node.marker, rest = " ", rest[4:]
		}
		if base, _, ok := strings.Cut(rest, " — "); ok {
			rest = base
		}
		node.body = rest
		if match := linkPattern.FindStringSubmatch(rest); match != nil {
			node.title = match[1]
			if id := pathIDPattern.FindStringSubmatch(match[2]); id != nil {
				node.id = id[1]
			}
		} else {
			node.title = rest
		}

		for len(stack) > 0 && stack[len(stack)-1].indent >= node.indent {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			roots = append(roots, node)
		} else {
			parent := stack[len(stack)-1]
			parent.kids = append(parent.kids, node)
		}
		stack = append(stack, node)
	}
	return roots, nil
}

// Exact comparison catches missing links, stale paths/lines, duplicates and links
// on unimplemented entries, using the current source scan as the authority.
func validateCoverageSources(document string, in coverageInput) error {
	refreshed, err := refreshCoverageSources(document, in)
	if err != nil {
		return err
	}
	if refreshed != document {
		return fmt.Errorf("coverage source links are missing, duplicated or stale")
	}
	return nil
}
