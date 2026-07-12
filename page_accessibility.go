package bonk

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/joakimcarlsson/bonk/proto"
)

// AXNode represents a node in the accessibility tree.
type AXNode struct {
	Role          string
	Name          string
	Value         string
	Disabled      bool
	Focused       bool
	Checked       string
	Selected      bool
	Expanded      string
	Level         int
	BackendNodeID proto.BackendNodeID
	Children      []*AXNode
}

var interactiveRoles = map[string]bool{
	"button":           true,
	"link":             true,
	"textbox":          true,
	"searchbox":        true,
	"combobox":         true,
	"listbox":          true,
	"option":           true,
	"checkbox":         true,
	"radio":            true,
	"switch":           true,
	"slider":           true,
	"spinbutton":       true,
	"tab":              true,
	"menuitem":         true,
	"menuitemradio":    true,
	"menuitemcheckbox": true,
	"treeitem":         true,
}

var leafTextRoles = map[string]bool{
	"StaticText":    true,
	"InlineTextBox": true,
}

// AccessibilityTree returns the page's accessibility tree.
func (p *Page) AccessibilityTree() ([]*AXNode, error) {
	proto.AccessibilityEnable().Do(p.execCtx)
	defer proto.AccessibilityDisable().Do(p.execCtx)

	res, err := proto.AccessibilityGetFullAXTree().
		Do(p.execCtx)
	if err != nil {
		return nil, fmt.Errorf(
			"get accessibility tree: %w", err,
		)
	}

	lookup := make(
		map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode,
		len(res.Nodes),
	)
	for i := range res.Nodes {
		lookup[res.Nodes[i].NodeID] = &res.Nodes[i]
	}

	var roots []proto.AccessibilityAXNodeID
	for i := range res.Nodes {
		if res.Nodes[i].ParentID == "" {
			roots = append(roots, res.Nodes[i].NodeID)
		}
	}

	var result []*AXNode
	for _, id := range roots {
		if node := buildAXNode(lookup, id); node != nil {
			result = append(result, node)
		}
	}
	return result, nil
}

// ElementForNode resolves an accessibility node to a live Element via its
// backend DOM node id, so a node read from AccessibilityTree can be acted on
// directly without re-locating it by accessible name.
func (p *Page) ElementForNode(n *AXNode) (*Element, error) {
	if n == nil || n.BackendNodeID == 0 {
		return nil, fmt.Errorf("bonk: accessibility node has no backend node id")
	}
	res, err := proto.DOMResolveNode().
		WithBackendNodeID(n.BackendNodeID).
		Do(p.execCtx)
	if err != nil {
		return nil, fmt.Errorf("bonk: resolve accessibility node: %w", err)
	}
	if res.Object.ObjectID == "" {
		return nil, fmt.Errorf("bonk: accessibility node did not resolve to an element")
	}
	return &Element{page: p, objectID: res.Object.ObjectID}, nil
}

// ClickNode resolves an accessibility node and clicks it by dispatching a native
// DOM click on the element. Unlike a coordinate-based click, this reliably
// triggers links and buttons even when an overlay covers the element's centre or
// hit-testing would otherwise send the click elsewhere.
func (p *Page) ClickNode(n *AXNode) error {
	el, err := p.ElementForNode(n)
	if err != nil {
		return err
	}
	_, err = p.EvaluateOn(el, "function(){ this.click() }")
	return err
}

func buildAXNode(
	lookup map[proto.AccessibilityAXNodeID]*proto.AccessibilityAXNode,
	id proto.AccessibilityAXNodeID,
) *AXNode {
	raw, ok := lookup[id]
	if !ok {
		return nil
	}

	if raw.Ignored {
		var children []*AXNode
		for _, childID := range raw.ChildIds {
			if c := buildAXNode(lookup, childID); c != nil {
				children = append(children, c)
			}
		}
		if len(children) == 1 {
			return children[0]
		}
		if len(children) > 1 {
			return &AXNode{Children: children}
		}
		return nil
	}

	role := axValueString(raw.Role)
	if leafTextRoles[role] {
		return nil
	}
	if role == "none" || role == "generic" {
		var children []*AXNode
		for _, childID := range raw.ChildIds {
			if c := buildAXNode(lookup, childID); c != nil {
				children = append(children, c)
			}
		}
		if len(children) == 1 {
			return children[0]
		}
		if len(children) > 1 {
			return &AXNode{Children: children}
		}
		return nil
	}

	node := &AXNode{
		Role:          role,
		Name:          axValueString(raw.Name),
		Value:         axValueString(raw.Value),
		BackendNodeID: raw.BackendDOMNodeID,
	}

	for _, prop := range raw.Properties {
		switch prop.Name {
		case proto.AccessibilityAXPropertyNameDisabled:
			node.Disabled = axValueBool(prop.Value)
		case proto.AccessibilityAXPropertyNameFocused:
			node.Focused = axValueBool(prop.Value)
		case proto.AccessibilityAXPropertyNameChecked:
			node.Checked = axValueTristate(prop.Value)
		case proto.AccessibilityAXPropertyNameSelected:
			node.Selected = axValueBool(prop.Value)
		case proto.AccessibilityAXPropertyNameExpanded:
			node.Expanded = axValueTristate(prop.Value)
		case proto.AccessibilityAXPropertyNameLevel:
			node.Level = axValueInt(prop.Value)
		}
	}

	for _, childID := range raw.ChildIds {
		if c := buildAXNode(lookup, childID); c != nil {
			node.Children = append(node.Children, c)
		}
	}

	return node
}

// FormatAccessibilityTree formats the tree as indexed text
// for LLM consumption. Interactive elements get numbered
// indices; non-interactive elements are shown without indices.
func FormatAccessibilityTree(nodes []*AXNode) string {
	text, _ := FormatAccessibilityTreeIndexed(nodes)
	return text
}

// FormatAccessibilityTreeIndexed formats the tree like FormatAccessibilityTree
// and also returns the interactive nodes in index order: the element shown as
// [i] in the text is refs[i-1]. Pass a ref to Page.ClickNode or
// Page.ElementForNode to act on exactly the element that was shown, instead of
// re-locating it by accessible name.
func FormatAccessibilityTreeIndexed(nodes []*AXNode) (string, []*AXNode) {
	var b strings.Builder
	var refs []*AXNode
	idx := 1
	for _, n := range nodes {
		formatNode(&b, n, 0, &idx, &refs)
	}
	return b.String(), refs
}

func formatNode(
	b *strings.Builder,
	n *AXNode,
	depth int,
	idx *int,
	refs *[]*AXNode,
) {
	if n.Role == "" && len(n.Children) > 0 {
		for _, c := range n.Children {
			formatNode(b, c, depth, idx, refs)
		}
		return
	}

	if n.Role == "" {
		return
	}

	indent := strings.Repeat("  ", depth)

	if interactiveRoles[n.Role] {
		fmt.Fprintf(b, "%s[%d] %s", indent, *idx, n.Role)
		*idx++
		*refs = append(*refs, n)
	} else {
		fmt.Fprintf(b, "%s%s", indent, n.Role)
	}

	if n.Name != "" {
		fmt.Fprintf(b, " %q", n.Name)
	}
	if n.Value != "" {
		fmt.Fprintf(b, " value=%q", n.Value)
	}

	var flags []string
	if n.Disabled {
		flags = append(flags, "disabled")
	}
	if n.Focused {
		flags = append(flags, "focused")
	}
	if n.Checked != "" {
		flags = append(flags, "checked="+n.Checked)
	}
	if n.Selected {
		flags = append(flags, "selected")
	}
	if n.Expanded != "" {
		flags = append(flags, "expanded="+n.Expanded)
	}
	if n.Level > 0 {
		flags = append(flags, fmt.Sprintf("level=%d", n.Level))
	}
	if len(flags) > 0 {
		fmt.Fprintf(b, " (%s)", strings.Join(flags, ", "))
	}

	b.WriteByte('\n')

	for _, c := range n.Children {
		formatNode(b, c, depth+1, idx, refs)
	}
}

func axValueString(v proto.AccessibilityAXValue) string {
	if len(v.Value) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(v.Value, &s); err != nil {
		return strings.Trim(string(v.Value), `"`)
	}
	return s
}

func axValueBool(v proto.AccessibilityAXValue) bool {
	if len(v.Value) == 0 {
		return false
	}
	var b bool
	json.Unmarshal(v.Value, &b)
	return b
}

func axValueTristate(v proto.AccessibilityAXValue) string {
	if len(v.Value) == 0 {
		return ""
	}
	return strings.Trim(string(v.Value), `"`)
}

func axValueInt(v proto.AccessibilityAXValue) int {
	if len(v.Value) == 0 {
		return 0
	}
	var n int
	json.Unmarshal(v.Value, &n)
	return n
}
