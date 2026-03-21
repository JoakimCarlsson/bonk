package bonk

import "fmt"

func getByTextJSExpr(text string, exact bool) (string, string) {
	t := jsString(text)
	matchFn := fmt.Sprintf("s.includes(%s)", t)
	if exact {
		matchFn = fmt.Sprintf("s.trim()===%s", t)
	}

	body := fmt.Sprintf(`(()=>{`+
		`const m=s=>%s;`+
		`const all=[];`+
		`const walk=n=>{`+
		`for(const c of n.children)walk(c);`+
		`const hasText=[...n.childNodes].some(c=>c.nodeType===3&&m(c.textContent));`+
		`if(hasText||(n.children.length===0&&n.textContent&&m(n.textContent)))all.push(n)};`+
		`walk(document.body);`, matchFn)

	jsExpr := body + `return all[0]||null})()`
	jsAllExpr := body + `return all})()`
	return jsExpr, jsAllExpr
}

func getByLabelJSExpr(text string, exact bool) (string, string) {
	t := jsString(text)
	matchFn := fmt.Sprintf("s.includes(%s)", t)
	if exact {
		matchFn = fmt.Sprintf("s.trim()===%s", t)
	}

	body := fmt.Sprintf(`(()=>{`+
		`const m=s=>%s;`+
		`const all=[];const seen=new Set();`+
		`document.querySelectorAll('label').forEach(l=>{`+
		`if(m(l.textContent)&&l.control&&!seen.has(l.control)){seen.add(l.control);all.push(l.control)}});`+
		`document.querySelectorAll('[aria-label]').forEach(e=>{`+
		`if(!seen.has(e)&&m(e.getAttribute('aria-label'))){seen.add(e);all.push(e)}});`+
		`document.querySelectorAll('[aria-labelledby]').forEach(e=>{`+
		`if(seen.has(e))return;`+
		`const ids=e.getAttribute('aria-labelledby').split(/\s+/);`+
		`const t=ids.map(id=>{const r=document.getElementById(id);return r?r.textContent:''}).join(' ');`+
		`if(m(t)){seen.add(e);all.push(e)}});`, matchFn)

	jsExpr := body + `return all[0]||null})()`
	jsAllExpr := body + `return all})()`
	return jsExpr, jsAllExpr
}

var implicitRoleSelectors = map[string]string{
	"button":        "button,input[type=button],input[type=submit],input[type=reset]",
	"link":          "a[href]",
	"heading":       "h1,h2,h3,h4,h5,h6",
	"textbox":       "input:not([type]),input[type=text],input[type=email],input[type=password],input[type=search],input[type=tel],input[type=url],textarea",
	"checkbox":      "input[type=checkbox]",
	"radio":         "input[type=radio]",
	"img":           "img[alt]",
	"list":          "ul,ol",
	"listitem":      "li",
	"navigation":    "nav",
	"main":          "main",
	"complementary": "aside",
	"table":         "table",
	"row":           "tr",
	"cell":          "td",
	"columnheader":  "th",
}

func getByRoleJSExpr(
	role string,
	name string,
	hasName bool,
	exact bool,
) (string, string) {
	r := jsString(role)
	implicit, ok := implicitRoleSelectors[role]

	sel := fmt.Sprintf("[role=%s]", r)
	if ok {
		sel += "," + implicit
	}

	if !hasName {
		body := fmt.Sprintf(
			`(()=>{const all=[...document.querySelectorAll(%s)];`,
			jsString(sel),
		)
		jsExpr := body + `return all[0]||null})()`
		jsAllExpr := body + `return all})()`
		return jsExpr, jsAllExpr
	}

	n := jsString(name)
	matchFn := fmt.Sprintf("s.includes(%s)", n)
	if exact {
		matchFn = fmt.Sprintf("s.trim()===%s", n)
	}

	body := fmt.Sprintf(`(()=>{`+
		`const m=s=>s!=null&&(%s);`+
		`const an=e=>{`+
		`const l=e.getAttribute('aria-label');if(l)return l;`+
		`const lby=e.getAttribute('aria-labelledby');`+
		`if(lby){const t=lby.split(/\s+/).map(id=>{const r=document.getElementById(id);return r?r.textContent:''}).join(' ');if(t)return t}`+
		`return e.textContent};`+
		`const all=[...document.querySelectorAll(%s)].filter(e=>m(an(e)));`,
		matchFn, jsString(sel))

	jsExpr := body + `return all[0]||null})()`
	jsAllExpr := body + `return all})()`
	return jsExpr, jsAllExpr
}

func getByAttrJSExpr(attr, text string, exact bool) (string, string) {
	t := jsString(text)
	a := jsString(attr)

	if exact {
		sel := fmt.Sprintf("[%s=%s]", attr, t)
		body := fmt.Sprintf(
			`(()=>{const all=[...document.querySelectorAll(%s)];`,
			jsString(sel),
		)
		jsExpr := body + `return all[0]||null})()`
		jsAllExpr := body + `return all})()`
		return jsExpr, jsAllExpr
	}

	body := fmt.Sprintf(`(()=>{`+
		`const all=[...document.querySelectorAll('['+%s+']')].filter(e=>e.getAttribute(%s).includes(%s));`,
		a, a, t)
	jsExpr := body + `return all[0]||null})()`
	jsAllExpr := body + `return all})()`
	return jsExpr, jsAllExpr
}
