package wecom_test

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"sort"
	"strings"
	"testing"
)

// explainerTitles lists official pages that describe behaviour instead of
// exposing an API; they are listed without a checkbox.
var explainerTitles = map[string]bool{
	"概述": true, "开发前必读": true, "基本概念介绍": true, "简易教程": true,
	"常见问题": true, "常见问题解答": true, "开发案例演示": true, "开始开发": true,
	"消息格式": true, "事件格式": true, "被动回复消息格式": true, "回调配置": true,
	"签名算法": true, "标准年级对照表": true, "通讯录同步接口调整": true,
	"加解密方案说明": true, "API模式机器人文档使用说明": true, "模板卡片类型": true,
	"回调和回复的加解密方案": true, "接入指引": true, "专区程序开发指引": true,
	"专区程序示例": true, "专区程序SDK和示例下载": true, "镜像文件配置指引": true,
	"会话展示组件": true, "Web登录组件": true, "文件内容存档": true,
	"调试说明": true, "使用前帮助": true, "使用说明": true,
}

// fetchOfficialTree downloads the documentation page and parses its sidebar.
func fetchOfficialTree(t *testing.T) []*navNode {
	t.Helper()
	client := &http.Client{Timeout: fetchTimeout}
	request, err := http.NewRequest(http.MethodGet, officialNavURL, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; go-wecom coverage generator)")
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("fetch %s: %v", officialNavURL, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("fetch %s: unexpected status %s", officialNavURL, response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	tree, err := parseSidebar(string(body))
	if err != nil {
		t.Fatalf("parse sidebar: %v", err)
	}
	return tree
}

// frame tracks one open element while parsing.
type frame struct {
	tag   string
	wrap  *navNode
	owner *navNode
}

// parseSidebar locates every ep-doc-select block and keeps the largest tree:
// the page also renders small mobile drawers that only hold a couple of links.
func parseSidebar(page string) ([]*navNode, error) {
	page = commentPattern.ReplaceAllString(page, "")

	var best []*navNode
	for offset := 0; offset < len(page); {
		at := strings.Index(page[offset:], `class="ep-doc-select"`)
		if at < 0 {
			break
		}
		at += offset
		start := strings.LastIndex(page[:at], "<div")
		if start < 0 {
			break
		}
		tree, err := parseSelect(page[start:])
		if err == nil && countLeaves(tree) > countLeaves(best) {
			best = tree
		}
		offset = at + len(`class="ep-doc-select"`)
	}
	if best == nil {
		return nil, fmt.Errorf(`no usable class="ep-doc-select" block found`)
	}
	return best, nil
}

// parseSelect parses one sidebar block, which ends when its outer div closes.
func parseSelect(fragment string) ([]*navNode, error) {
	var (
		roots    []*navNode
		levels   []*navNode
		stack    []*frame
		divDepth int
	)
	cursor := 0
	for _, m := range tagPattern.FindAllStringSubmatchIndex(fragment, -1) {
		if text := fragment[cursor:m[0]]; text != "" {
			for i := len(stack) - 1; i >= 0; i-- {
				if stack[i].owner != nil {
					stack[i].owner.title += text
					break
				}
			}
		}
		cursor = m[1]

		closing := m[3] > m[2]
		tag := fragment[m[4]:m[5]]
		attrs := ""
		if m[6] >= 0 {
			attrs = fragment[m[6]:m[7]]
		}

		if tag == "div" {
			if closing {
				divDepth--
			} else {
				divDepth++
			}
			if divDepth == 0 {
				break
			}
		}

		if closing {
			for len(stack) > 0 {
				last := len(stack) - 1
				popped := stack[last]
				stack = stack[:last]
				if popped.tag == tag {
					break
				}
			}
			continue
		}

		class := ""
		if cm := classPattern.FindStringSubmatch(attrs); cm != nil {
			class = cm[1]
		}

		switch {
		case tag == "div" && strings.Contains(class, "ep-doc-wrap"):
			level := 0
			if lm := levelPattern.FindStringSubmatch(attrs); lm != nil {
				fmt.Sscanf(lm[1], "%d", &level)
			}
			node := &navNode{}
			if level == 0 {
				roots = append(roots, node)
			} else {
				if level-1 >= len(levels) || levels[level-1] == nil {
					return nil, fmt.Errorf("sidebar level %d has no parent", level)
				}
				parent := levels[level-1]
				parent.kids = append(parent.kids, node)
			}
			if level < len(levels) {
				levels = levels[:level]
			}
			levels = append(levels, node)
			stack = append(stack, &frame{tag: "div", wrap: node, owner: node})
		case tag == "div" && strings.Contains(class, "ep-doc-special"):
			owner := nearestWrap(stack)
			if owner != nil {
				owner.group = true
			}
			stack = append(stack, &frame{tag: "div", owner: owner})
		case tag == "div" && strings.Contains(class, "ep-doc-item"):
			stack = append(stack, &frame{tag: "div", owner: nearestWrap(stack)})
		case tag == "a" && strings.Contains(class, "ep-doc-item"):
			owner := nearestWrap(stack)
			if owner != nil {
				if hm := hrefPattern.FindStringSubmatch(attrs); hm != nil {
					if dm := pathIDPattern.FindStringSubmatch(hm[1]); dm != nil {
						owner.id = dm[1]
					}
				}
			}
			stack = append(stack, &frame{tag: "a", owner: owner})
		case tag == "span":
			stack = append(stack, &frame{tag: "span"})
		case tag == "img":
			// void element, nothing to track
		default:
			stack = append(stack, &frame{tag: tag})
		}
	}

	if len(roots) == 0 {
		return nil, fmt.Errorf("empty sidebar block")
	}
	for _, node := range roots {
		normalize(node)
	}
	roots, err := dropDuplicateCopy(roots)
	if err != nil {
		return nil, err
	}
	return groupSections(roots), nil
}

// normalize trims titles collected from the markup.
func normalize(node *navNode) {
	node.title = html.UnescapeString(strings.Join(strings.Fields(node.title), " "))
	for _, kid := range node.kids {
		normalize(kid)
	}
}

// dropDuplicateCopy cuts a repeated copy of the navigation, including the
// trailing 开发资源 block that follows it. Some layouts split the copies into
// separate ep-doc-select blocks, in which case there is nothing to cut.
func dropDuplicateCopy(roots []*navNode) ([]*navNode, error) {
	if len(roots) == 0 || roots[0].title == "" {
		return nil, fmt.Errorf("sidebar parse produced %d top-level entries", len(roots))
	}
	first := roots[0].title
	for i := 10; i < len(roots); i++ {
		if roots[i].title == first {
			return roots[:i], nil
		}
	}
	return roots, nil
}

// groupSections folds the flat top-level sequence into its groups: the official
// sidebar uses divider headings, and the categories that follow belong to them.
func groupSections(roots []*navNode) []*navNode {
	var out []*navNode
	for _, node := range roots {
		if node.group {
			out = append(out, node)
			continue
		}
		if n := len(out); n > 0 && out[n-1].group {
			last := out[n-1]
			last.kids = append(last.kids, node)
			continue
		}
		out = append(out, node)
	}
	return out
}

func nearestWrap(stack []*frame) *navNode {
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i].wrap != nil {
			return stack[i].wrap
		}
	}
	return nil
}

func countLeaves(tree []*navNode) int {
	total := 0
	var walk func([]*navNode)
	walk = func(nodes []*navNode) {
		for _, node := range nodes {
			if len(node.kids) == 0 {
				total++
				continue
			}
			walk(node.kids)
		}
	}
	walk(tree)
	return total
}

// coverageInput is what rendering needs: the ledger decides which pages are
// supported, the scan supplies the jump target for each one.
type coverageInput struct {
	caps    map[string]capability
	methods map[string][]sourceRef
}

// renderCoverage renders the complete markdown document.
func renderCoverage(tree []*navNode, in coverageInput) string {
	r := &coverageRenderer{caps: in.caps, methods: in.methods}
	var b strings.Builder

	b.WriteString("# 功能覆盖\n\n")
	b.WriteString("本文件按[企业微信官方文档目录](https://developer.work.weixin.qq.com/document)组织，" +
		"分类、子分类与接口名称均取自官方侧边栏，逐项标注 SDK 的实现情况。\n\n")
	b.WriteString("每个已实现的方法注释都附有官方文档地址，可与本文件交叉核对；" +
		"已实现的条目还标注了实现该接口的源码位置，可直接跳转。\n\n")
	b.WriteString("本文件由 `coverage_test.go` 抓取官网目录后生成，请勿手工编辑；" +
		"联网刷新目录：`go test -run TestCoverageDocument -update-coverage .`；" +
		"离线刷新源码链接：`go test -run TestCoverageDocument -update-coverage-sources .`。\n\n")
	b.WriteString("支持状态来自 `docs/capabilities.json` 能力台账，" +
		"源码位置由扫描方法注释得出，两者与本文不一致时测试失败。\n\n")
	b.WriteString("## 标记说明\n\n")
	b.WriteString("| 标记 | 含义 |\n| --- | --- |\n")
	b.WriteString("| `[x]` | SDK 已提供对应能力；具体支持范围以源码中的方法和类型为准 |\n")
	b.WriteString("| `[ ]` | 尚未标记为已支持 |\n")
	b.WriteString("| 无勾选 | 说明性文档 |\n")
	b.WriteString("| **部分支持** | 子项只实现了一部分 |\n")

	b.WriteString("\n## 快速导航\n\n")
	for i, section := range tree {
		if i > 0 {
			b.WriteString(" · ")
		}
		fmt.Fprintf(&b, "[%s](#%s)", section.title, section.title)
	}
	b.WriteString("\n\n常用能力：")
	first := true
	for _, section := range tree {
		for _, child := range section.kids {
			switch child.title {
			case "消息接收与发送", "微信客服", "文档", "邮件", "素材管理":
				if !first {
					b.WriteString(" · ")
				}
				first = false
				fmt.Fprintf(&b, "[%s](#%s)", child.title, child.title)
			}
		}
	}
	b.WriteString("\n")

	for _, section := range tree {
		if section.group {
			fmt.Fprintf(&b, "\n## %s\n", section.title)
			for _, child := range section.kids {
				fmt.Fprintf(&b, "\n### %s\n\n", child.title)
				r.writeNodes(&b, child.kids, "")
			}
			continue
		}
		fmt.Fprintf(&b, "\n## %s\n\n", section.title)
		r.writeNodes(&b, section.kids, "")
	}

	return b.String()
}

// coverageRenderer maps tree entries to the methods that implement them.
type coverageRenderer struct {
	caps    map[string]capability
	methods map[string][]sourceRef
}

// sourceRef points at the line where one documented method is declared.
type sourceRef struct {
	file        string
	line        int
	constructor bool
}

// source renders links to the files implementing one documentation page.
// Constructors are ignored while a real call exists; several methods in the same
// file collapse into one link.
func (r *coverageRenderer) source(id string) string {
	refs := r.methods[id]
	kept := make([]sourceRef, 0, len(refs))
	for _, ref := range refs {
		if !ref.constructor {
			kept = append(kept, ref)
		}
	}
	if len(kept) == 0 {
		kept = refs
	}

	byFile := map[string]sourceRef{}
	for _, ref := range kept {
		if current, ok := byFile[ref.file]; !ok || ref.line < current.line {
			byFile[ref.file] = ref
		}
	}
	files := make([]string, 0, len(byFile))
	for file := range byFile {
		files = append(files, file)
	}
	sort.Strings(files)

	links := make([]string, 0, len(files))
	for _, file := range files {
		ref := byFile[file]
		links = append(links, fmt.Sprintf("[%s:%d](../%s#L%d)", file, ref.line, file, ref.line))
	}
	if len(links) > 3 {
		return strings.Join(links[:3], "、") + fmt.Sprintf(" 等 %d 个文件", len(links))
	}
	return strings.Join(links, "、")
}

func (r *coverageRenderer) writeNodes(b *strings.Builder, nodes []*navNode, indent string) {
	for _, node := range nodes {
		r.writeNode(b, node, indent)
	}
}

func (r *coverageRenderer) writeNode(b *strings.Builder, node *navNode, indent string) {
	if len(node.kids) == 0 {
		switch {
		case r.implemented(node.id):
			fmt.Fprintf(b, "%s- [x] [%s](%s%s)", indent, node.title, docBaseURL, node.id)
			if source := r.source(node.id); source != "" {
				fmt.Fprintf(b, " — %s", source)
			}
			b.WriteString("\n")
		case r.notApplicable(node.id) || explainerTitles[node.title]:
			fmt.Fprintf(b, "%s- [%s](%s%s)\n", indent, node.title, docBaseURL, node.id)
		case node.id == "":
			fmt.Fprintf(b, "%s- [ ] %s\n", indent, node.title)
		default:
			fmt.Fprintf(b, "%s- [ ] [%s](%s%s)\n", indent, node.title, docBaseURL, node.id)
		}
		return
	}

	state, partial, applicable := r.branchState(node)
	if !applicable {
		fmt.Fprintf(b, "%s- %s\n", indent, node.title)
		r.writeNodes(b, node.kids, indent+"  ")
		return
	}
	suffix := ""
	if partial {
		suffix = "（**部分支持**）"
	}
	fmt.Fprintf(b, "%s- [%c] %s%s\n", indent, state, node.title, suffix)
	r.writeNodes(b, node.kids, indent+"  ")
}

func (r *coverageRenderer) implemented(id string) bool {
	entry, ok := r.caps[id]
	return ok && capabilityDone(entry)
}

// notApplicable reports a page the official API does not expose, so it is listed
// without a checkbox instead of staying pending forever.
func (r *coverageRenderer) notApplicable(id string) bool {
	entry, ok := r.caps[id]
	return ok && entry.Status == statusNotApplicable
}

// branchState summarises a category: [x] when every entry is implemented, [x]
// with a partial marker when only some are, [ ] when none are.
func (r *coverageRenderer) branchState(node *navNode) (state byte, partial, applicable bool) {
	var done, missing int
	var walk func(*navNode)
	walk = func(n *navNode) {
		for _, kid := range n.kids {
			if len(kid.kids) > 0 {
				walk(kid)
				continue
			}
			switch {
			case r.implemented(kid.id):
				done++
			case r.notApplicable(kid.id), explainerTitles[kid.title]:
			default:
				missing++
			}
		}
	}
	walk(node)

	switch {
	case done == 0 && missing == 0:
		return ' ', false, false
	case done == 0:
		return ' ', false, true
	case missing == 0:
		return 'x', false, true
	default:
		return 'x', true, true
	}
}

// Refuse to replace the document with a partial or unrelated sidebar after a
// site layout change. Validate before writing, including every supported page.
func validateOfficialTree(tree []*navNode, in coverageInput) error {
	seen := map[string]bool{}
	var walk func([]*navNode) error
	walk = func(nodes []*navNode) error {
		for _, node := range nodes {
			if node.title == "" {
				return fmt.Errorf("empty official title")
			}
			if len(node.kids) > 0 {
				if err := walk(node.kids); err != nil {
					return err
				}
				continue
			}
			if node.id == "" || seen[node.id] {
				return fmt.Errorf("missing or duplicate official id %q", node.id)
			}
			seen[node.id] = true
		}
		return nil
	}
	if err := walk(tree); err != nil {
		return err
	}
	if len(seen) < 700 {
		return fmt.Errorf("official tree looks truncated: %d pages", len(seen))
	}
	for id, entry := range in.caps {
		if capabilityDone(entry) && !seen[id] {
			return fmt.Errorf("official tree missing supported path/%s", id)
		}
	}
	return nil
}
