package pdl

import (
	"strings"
	"unicode"
)

var initialisms = map[string]string{
	"id":    "ID",
	"Id":    "ID",
	"url":   "URL",
	"Url":   "URL",
	"uri":   "URI",
	"Uri":   "URI",
	"html":  "HTML",
	"Html":  "HTML",
	"css":   "CSS",
	"Css":   "CSS",
	"dom":   "DOM",
	"Dom":   "DOM",
	"http":  "HTTP",
	"Http":  "HTTP",
	"https": "HTTPS",
	"Https": "HTTPS",
	"ssl":   "SSL",
	"Ssl":   "SSL",
	"tls":   "TLS",
	"Tls":   "TLS",
	"xhr":   "XHR",
	"Xhr":   "XHR",
	"xml":   "XML",
	"Xml":   "XML",
	"json":  "JSON",
	"Json":  "JSON",
	"api":   "API",
	"Api":   "API",
	"cpu":   "CPU",
	"Cpu":   "CPU",
	"gpu":   "GPU",
	"Gpu":   "GPU",
	"uuid":  "UUID",
	"Uuid":  "UUID",
	"ip":    "IP",
	"Ip":    "IP",
	"tcp":   "TCP",
	"Tcp":   "TCP",
	"udp":   "UDP",
	"Udp":   "UDP",
	"dns":   "DNS",
	"Dns":   "DNS",
	"rtc":   "RTC",
	"Rtc":   "RTC",
	"js":    "JS",
	"Js":    "JS",
	"pdf":   "PDF",
	"Pdf":   "PDF",
	"svg":   "SVG",
	"Svg":   "SVG",
	"wasm":  "WASM",
	"Wasm":  "WASM",
	"eof":   "EOF",
	"Eof":   "EOF",
	"rpc":   "RPC",
	"Rpc":   "RPC",
	"pwa":   "PWA",
	"Pwa":   "PWA",
	"sw":    "SW",
	"Sw":    "SW",
}

// DomainToPackage converts a CDP domain name to a Go package name.
func DomainToPackage(domain string) string {
	return strings.ToLower(domain)
}

// ExportedName converts a CDP identifier to an exported Go identifier.
func ExportedName(name string) string {
	words := splitWords(name)
	var b strings.Builder
	for _, w := range words {
		if upper, ok := initialisms[w]; ok {
			b.WriteString(upper)
			continue
		}
		if len(w) > 0 {
			b.WriteRune(unicode.ToUpper(rune(w[0])))
			b.WriteString(w[1:])
		}
	}
	return b.String()
}

// UnexportedName converts a CDP identifier to an unexported Go identifier.
func UnexportedName(name string) string {
	exported := ExportedName(name)
	if exported == "" {
		return ""
	}

	for abbr := range initialisms {
		upper := initialisms[abbr]
		if strings.HasPrefix(exported, upper) {
			return strings.ToLower(upper) + exported[len(upper):]
		}
	}

	return string(unicode.ToLower(rune(exported[0]))) + exported[1:]
}

// EnumValueToIdent converts a CDP enum value to a Go identifier.
func EnumValueToIdent(value string) string {
	value = strings.ReplaceAll(value, "-", " ")
	value = strings.ReplaceAll(value, "_", " ")
	value = strings.ReplaceAll(value, ".", " ")

	words := strings.Fields(value)
	var b strings.Builder
	for _, w := range words {
		if upper, ok := initialisms[w]; ok {
			b.WriteString(upper)
			continue
		}
		if len(w) > 0 {
			b.WriteRune(unicode.ToUpper(rune(w[0])))
			b.WriteString(w[1:])
		}
	}
	return b.String()
}

func splitWords(s string) []string {
	var words []string
	var current strings.Builder

	flush := func() {
		if current.Len() > 0 {
			words = append(words, current.String())
			current.Reset()
		}
	}

	runes := []rune(s)
	for i := range runes {
		r := runes[i]

		if r == '_' || r == '-' || r == '.' {
			flush()
			continue
		}

		if unicode.IsUpper(r) {
			if current.Len() > 0 {
				if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					flush()
				} else if i > 0 && unicode.IsLower(runes[i-1]) {
					flush()
				}
			}
			current.WriteRune(r)
		} else {
			current.WriteRune(r)
		}
	}
	flush()

	return words
}
