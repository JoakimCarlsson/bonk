package bonk

import (
	"fmt"
	"strings"
)

// LocatorFilter specifies criteria for narrowing locator matches.
type LocatorFilter struct {
	HasText    string
	HasNotText string
	Has        *Locator
	HasNot     *Locator
}

// Filter returns a new Locator that narrows matches by the given criteria.
// All specified criteria must be satisfied (AND semantics).
func (l *Locator) Filter(f LocatorFilter) *Locator {
	baseAll := l.allExpr()

	var preamble []string
	var conditions []string
	var descParts []string
	varIdx := 0

	if f.HasText != "" {
		conditions = append(conditions,
			fmt.Sprintf("el.textContent.includes(%s)", jsString(f.HasText)))
		descParts = append(descParts,
			fmt.Sprintf("hasText=%s", jsString(f.HasText)))
	}
	if f.HasNotText != "" {
		conditions = append(conditions,
			fmt.Sprintf("!el.textContent.includes(%s)", jsString(f.HasNotText)))
		descParts = append(descParts,
			fmt.Sprintf("hasNotText=%s", jsString(f.HasNotText)))
	}
	if f.Has != nil {
		cond, pre := subLocatorCondition(f.Has, true, &varIdx)
		preamble = append(preamble, pre...)
		conditions = append(conditions, cond)
		descParts = append(descParts,
			fmt.Sprintf("has=%s", f.Has.description()))
	}
	if f.HasNot != nil {
		cond, pre := subLocatorCondition(f.HasNot, false, &varIdx)
		preamble = append(preamble, pre...)
		conditions = append(conditions, cond)
		descParts = append(descParts,
			fmt.Sprintf("hasNot=%s", f.HasNot.description()))
	}

	filterExpr := "true"
	if len(conditions) > 0 {
		filterExpr = strings.Join(conditions, "&&")
	}

	preStr := ""
	if len(preamble) > 0 {
		preStr = strings.Join(preamble, ";") + ";"
	}

	jsAllExpr := fmt.Sprintf(
		"(()=>{%sconst all=%s;return all.filter(el=>%s)})()",
		preStr, baseAll, filterExpr)
	jsExpr := fmt.Sprintf(
		"(()=>{%sconst all=%s;return all.filter(el=>%s)[0]||null})()",
		preStr, baseAll, filterExpr)

	desc := fmt.Sprintf("%s.filter(%s)",
		l.description(), strings.Join(descParts, ", "))

	return &Locator{
		page:      l.page,
		frame:     l.frame,
		jsExpr:    jsExpr,
		jsAllExpr: jsAllExpr,
		desc:      desc,
		nth:       -1,
	}
}

// And returns a new Locator matching elements that satisfy both this
// locator and the other locator.
func (l *Locator) And(other *Locator) *Locator {
	leftAll := l.allExpr()
	rightAll := other.allExpr()

	jsAllExpr := fmt.Sprintf(
		"(()=>{const a=%s;const b=new Set(%s);return a.filter(el=>b.has(el))})()",
		leftAll,
		rightAll,
	)
	jsExpr := fmt.Sprintf(
		"(()=>{const a=%s;const b=new Set(%s);return a.filter(el=>b.has(el))[0]||null})()",
		leftAll,
		rightAll,
	)

	desc := fmt.Sprintf("%s.and(%s)", l.description(), other.description())

	return &Locator{
		page:      l.page,
		frame:     l.frame,
		jsExpr:    jsExpr,
		jsAllExpr: jsAllExpr,
		desc:      desc,
		nth:       -1,
	}
}

// Or returns a new Locator matching elements that satisfy either this
// locator or the other locator. Results are in document order.
func (l *Locator) Or(other *Locator) *Locator {
	leftAll := l.allExpr()
	rightAll := other.allExpr()

	body := fmt.Sprintf(
		`const a=%s;const b=%s;`+
			`const s=new Set(a);b.forEach(el=>{if(!s.has(el))a.push(el)});`+
			`a.sort((x,y)=>x.compareDocumentPosition(y)&2?1:-1)`,
		leftAll, rightAll)

	jsAllExpr := fmt.Sprintf("(()=>{%s;return a})()", body)
	jsExpr := fmt.Sprintf("(()=>{%s;return a[0]||null})()", body)

	desc := fmt.Sprintf("%s.or(%s)", l.description(), other.description())

	return &Locator{
		page:      l.page,
		frame:     l.frame,
		jsExpr:    jsExpr,
		jsAllExpr: jsAllExpr,
		desc:      desc,
		nth:       -1,
	}
}

func (l *Locator) allExpr() string {
	if l.jsAllExpr != "" {
		return l.jsAllExpr
	}
	return fmt.Sprintf(
		"[...document.querySelectorAll(%s)]",
		jsString(l.selector),
	)
}

func subLocatorCondition(sub *Locator, want bool, idx *int) (string, []string) {
	if sub.selector != "" && sub.jsAllExpr == "" {
		if want {
			return fmt.Sprintf(
				"el.querySelector(%s)!==null",
				jsString(sub.selector),
			), nil
		}
		return fmt.Sprintf(
			"el.querySelector(%s)===null",
			jsString(sub.selector),
		), nil
	}

	varName := fmt.Sprintf("_sub%d", *idx)
	*idx++
	preamble := []string{
		fmt.Sprintf("const %s=%s", varName, sub.allExpr()),
	}
	if want {
		return fmt.Sprintf("%s.some(s=>el.contains(s))", varName), preamble
	}
	return fmt.Sprintf("!%s.some(s=>el.contains(s))", varName), preamble
}
