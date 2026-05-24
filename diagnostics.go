package bonk

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// CrashReport describes a Chromium Crashpad dump discovered in the user data
// directory after a session ended.
type CrashReport struct {
	// Path is the absolute path to the .dmp file.
	Path string
	// Size is the dump size in bytes.
	Size int64
	// Summary is a best-effort one-line cause extracted from the dump's
	// embedded annotation strings. Empty if nothing recognisable was found.
	// Example: "v8-oom in Zone (stringify)".
	Summary string
}

func collectCrashReports(dataDir string, seen map[string]struct{}) []CrashReport {
	if dataDir == "" {
		return nil
	}
	reportsDir := filepath.Join(dataDir, "Crashpad", "reports")
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil
	}
	var out []CrashReport
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".dmp") {
			continue
		}
		path := filepath.Join(reportsDir, e.Name())
		if seen != nil {
			if _, ok := seen[path]; ok {
				continue
			}
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, CrashReport{
			Path:    path,
			Size:    info.Size(),
			Summary: summarizeCrashDump(path),
		})
	}
	return out
}

func snapshotCrashReports(dataDir string) map[string]struct{} {
	out := map[string]struct{}{}
	reportsDir := filepath.Join(dataDir, "Crashpad", "reports")
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".dmp") {
			continue
		}
		out[filepath.Join(reportsDir, e.Name())] = struct{}{}
	}
	return out
}

var crashAnnotationKeys = []string{
	"v8-oom-location",
	"v8-oom-stack",
	"crash-keys",
	"chromium-version",
}

// summarizeCrashDump pulls out a handful of well-known annotation values
// from a Crashpad minidump by streaming the file and looking for known keys
// followed by their value. It is a best-effort heuristic; minidumps are a
// binary format and bonk does not parse them properly.
func summarizeCrashDump(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	const maxBytes = 4 << 20
	buf := make([]byte, 0, 64*1024)
	read := int64(0)
	r := bufio.NewReader(f)
	tmp := make([]byte, 8192)
	for read < maxBytes {
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			read += int64(n)
		}
		if err != nil {
			break
		}
	}

	found := map[string]string{}
	for _, key := range crashAnnotationKeys {
		idx := bytes.Index(buf, []byte(key))
		if idx < 0 {
			continue
		}
		start := idx + len(key)
		end := start
		for end < len(buf) && end-start < 256 {
			c := buf[end]
			if c == 0 || c == '\n' || c == '\r' {
				break
			}
			end++
		}
		val := strings.TrimFunc(string(buf[start:end]), func(r rune) bool {
			return r < 0x20 || r == ' '
		})
		val = sanitizeAnnotation(val)
		if val != "" {
			found[key] = val
		}
	}

	if loc, ok := found["v8-oom-location"]; ok {
		stack := found["v8-oom-stack"]
		if stack != "" {
			return "v8-oom in " + loc + " (" + stack + ")"
		}
		return "v8-oom in " + loc
	}
	if ver := found["chromium-version"]; ver != "" {
		return "crash in chromium " + ver
	}
	return ""
}

var nonPrintable = regexp.MustCompile(`[^\x20-\x7e]+`)

func sanitizeAnnotation(s string) string {
	s = nonPrintable.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}
