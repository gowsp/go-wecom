package wecom_test

import (
	"strings"
	"testing"
)

func TestCoverageSourceLinks(t *testing.T) {
	in := coverageInput{
		caps: map[string]capability{"99896": {ID: "99896", Status: statusImplemented}},
		methods: map[string][]sourceRef{"99896": {
			{file: "wedoc/client.go", line: 5, constructor: true},
			{file: "wedoc/sheet.go", line: 37},
			{file: "wedoc/sheet.go", line: 50},
		}},
	}
	base := "- [x] [添加子表](" + docBaseURL + "99896)"
	expected := base + " — [wedoc/sheet.go:37](../wedoc/sheet.go#L37)\n"
	if err := validateCoverageSources(expected, in); err != nil {
		t.Fatal(err)
	}

	// Offline refresh normalises anything it can derive from the ledger and the
	// source scan, so a stale document is repaired instead of rejected.
	cases := map[string]string{
		"missing links":  base + "\n",
		"stale line":     strings.ReplaceAll(expected, "37", "36"),
		"missing file":   strings.ReplaceAll(expected, "sheet.go", "missing.go"),
		"duplicate link": strings.TrimSuffix(expected, "\n") + "、[wedoc/sheet.go:37](../wedoc/sheet.go#L37)\n",
		"state changed":  strings.Replace(expected, "[x]", "[ ]", 1),
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if err := validateCoverageSources(input, in); err == nil {
				t.Fatal("stale document accepted as is")
			}
			got, err := refreshCoverageSources(input, in)
			if err != nil {
				t.Fatal(err)
			}
			if got != expected {
				t.Fatalf("refresh = %q, want %q", got, expected)
			}
		})
	}

	// Pages outside the ledger are left untouched, including explainer links.
	explainer := "- [概述](" + docBaseURL + "90000)\n"
	if got, err := refreshCoverageSources(explainer+expected, in); err != nil || got != explainer+expected {
		t.Fatalf("explainer rewritten: %q %v", got, err)
	}

	// A supported page that vanished from the document needs a network refresh.
	for name, input := range map[string]string{
		"missing page":   "# empty\n",
		"duplicate page": expected + expected,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := refreshCoverageSources(input, in); err == nil {
				t.Fatal("invalid coverage accepted")
			}
		})
	}
}

func TestCoverageRendering(t *testing.T) {
	tree := []*navNode{{title: "基础", group: true, kids: []*navNode{{title: "素材管理", kids: []*navNode{
		{title: "上传素材", id: "1"}, {title: "下载素材", id: "2"}, {title: "概述", id: "3"},
	}}}}}
	got := renderCoverage(tree, coverageInput{
		caps:    map[string]capability{"1": {ID: "1", Status: statusImplemented}},
		methods: map[string][]sourceRef{"1": {{file: "media/upload.go", line: 12}}},
	})
	for _, want := range []string{
		"[基础](#基础)", "[素材管理](#素材管理)",
		"- [x] [上传素材](" + docBaseURL + "1) — [media/upload.go:12](../media/upload.go#L12)",
		"- [ ] [下载素材](" + docBaseURL + "2)",
		"- [概述](" + docBaseURL + "3)",
		"具体支持范围以源码中的方法和类型为准",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, unwanted := range []string{"（说明页）", "并有单元测试", "补充未覆盖能力"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("unexpected %q", unwanted)
		}
	}
	if err := validateOfficialTree(tree, coverageInput{}); err == nil {
		t.Fatal("accepted truncated tree")
	}
}

func TestSidebarBoundaries(t *testing.T) {
	page := `<div class="ep-doc-select"><div level="0" class="ep-doc-wrap"><a class="ep-doc-item" href="/document/path/1">概述</a></div></div>` +
		`<div class="ep-doc-select"><div level="0" class="ep-doc-wrap"><div class="ep-doc-item">开发指南</div></div><div level="1" class="ep-doc-wrap"><a class="ep-doc-item" href="/document/path/2">获取&lt;token&gt;<!----></a></div><div level="1" class="ep-doc-wrap"><a class="ep-doc-item" href="/document/path/3">基本概念</a></div></div>` +
		`<div class="ep-doc-wrap"><a class="ep-doc-item" href="/document/path/4">不应包含</a></div>`
	tree, err := parseSidebar(page)
	if err != nil {
		t.Fatal(err)
	}
	if countLeaves(tree) != 2 || tree[0].kids[0].title != "获取<token>" {
		t.Fatalf("wrong sidebar: %+v", tree)
	}
}
