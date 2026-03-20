package pdl

import "testing"

func TestDomainToPackage(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"DOM", "dom"},
		{"Page", "page"},
		{"Network", "network"},
		{"HeapProfiler", "heapprofiler"},
		{"CSS", "css"},
		{"DOMDebugger", "domdebugger"},
		{"IO", "io"},
	}
	for _, tt := range tests {
		if got := DomainToPackage(tt.in); got != tt.want {
			t.Errorf("DomainToPackage(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestExportedName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"nodeId", "NodeID"},
		{"backendNodeId", "BackendNodeID"},
		{"url", "URL"},
		{"frameId", "FrameID"},
		{"requestId", "RequestID"},
		{"loaderId", "LoaderID"},
		{"cssText", "CSSText"},
		{"domContentLoaded", "DOMContentLoaded"},
		{"htmlContent", "HTMLContent"},
		{"xmlHttpRequest", "XMLHTTPRequest"},
		{"sslCertificate", "SSLCertificate"},
		{"getDocument", "GetDocument"},
		{"querySelector", "QuerySelector"},
		{"enable", "Enable"},
		{"disable", "Disable"},
		{"navigateToHistoryEntry", "NavigateToHistoryEntry"},
		{"setAttributeValue", "SetAttributeValue"},
		{"mixedContentType", "MixedContentType"},
		{"postData", "PostData"},
		{"statusText", "StatusText"},
		{"remoteObjectId", "RemoteObjectID"},
		{"browserContextId", "BrowserContextID"},
		{"timestamp", "Timestamp"},
	}
	for _, tt := range tests {
		if got := ExportedName(tt.in); got != tt.want {
			t.Errorf("ExportedName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestEnumValueToIdent(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"first-line", "FirstLine"},
		{"first-letter", "FirstLetter"},
		{"Document", "Document"},
		{"Stylesheet", "Stylesheet"},
		{"unsafe-url", "UnsafeURL"},
		{"no-referrer", "NoReferrer"},
		{"no-referrer-when-downgrade", "NoReferrerWhenDowngrade"},
		{"origin-when-cross-origin", "OriginWhenCrossOrigin"},
		{"same-origin", "SameOrigin"},
		{"strict-origin", "StrictOrigin"},
		{"Low", "Low"},
		{"Medium", "Medium"},
		{"none", "None"},
		{"child", "Child"},
		{"root", "Root"},
		{"CSPViolation", "CSPViolation"},
	}
	for _, tt := range tests {
		if got := EnumValueToIdent(tt.in); got != tt.want {
			t.Errorf("EnumValueToIdent(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestUnexportedName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"nodeId", "nodeID"},
		{"url", "url"},
		{"frameId", "frameID"},
		{"enable", "enable"},
		{"getDocument", "getDocument"},
		{"postData", "postData"},
	}
	for _, tt := range tests {
		if got := UnexportedName(tt.in); got != tt.want {
			t.Errorf("UnexportedName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
