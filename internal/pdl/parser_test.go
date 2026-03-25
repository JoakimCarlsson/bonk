package pdl

import (
	"testing"
)

const testPDL = `
version
  major 1
  minor 3

# Inspector domain.
experimental domain Inspector
  # Enables inspector domain notifications.
  command enable

  # Disables inspector domain notifications.
  command disable

  # Fired when remote debugging connection is about to be terminated.
  event detached
    parameters
      # The reason why connection has been terminated.
      string reason

  event targetCrashed

  event targetReloadedAfterCrash

domain Network
  depends on Runtime
  depends on Security

  # Resource type as it was perceived by the rendering engine.
  type ResourceType extends string
    enum
      Document
      Stylesheet
      Image
      Script

  # Unique request identifier.
  type RequestId extends string

  # HTTP request data.
  type Request extends object
    properties
      string url
      optional string urlFragment
      string method
      # HTTP request headers.
      object headers
      optional string postData
      optional boolean hasPostData
      experimental optional array of PostDataEntry postDataEntries
      enum referrerPolicy
        unsafe-url
        no-referrer
        origin

  # An HTTP response.
  type Response extends object
    properties
      string url
      integer status
      string statusText
      optional Network.RequestId requestId

  # Sends a request.
  command sendRequest
    parameters
      string url
      optional string method
      optional object headers
    returns
      RequestId requestId
      optional string errorText

  experimental command getSecurityDetails
    parameters
      RequestId requestId

  deprecated event requestWillBeSent
    parameters
      RequestId requestId
      string url
      Request request

  # Fired when page is about to send HTTP request.
  event responseReceived
    parameters
      RequestId requestId
      Response response

  command navigate
    parameters
      string url
      optional string referrer
    returns
      string frameId
      optional string errorText
    redirect Page

domain Page
  depends on Network

  # Unique frame identifier.
  type FrameId extends string

  experimental type AdFrameType extends string
    enum
      none
      child
      root

  command enable

  command navigate
    parameters
      string url
      optional string referrer
      optional Network.RequestId requestId
    returns
      FrameId frameId
`

func TestParsePDL(t *testing.T) {
	proto, err := ParseString(testPDL)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if proto.Version.Major != "1" || proto.Version.Minor != "3" {
		t.Errorf(
			"version = %s.%s, want 1.3",
			proto.Version.Major,
			proto.Version.Minor,
		)
	}

	if len(proto.Domains) != 3 {
		t.Fatalf("domains = %d, want 3", len(proto.Domains))
	}

	inspector := proto.Domains[0]
	if inspector.Name != "Inspector" {
		t.Errorf("domain[0].Name = %q, want Inspector", inspector.Name)
	}
	if !inspector.Experimental {
		t.Error("Inspector should be experimental")
	}
	if inspector.Description != "Inspector domain." {
		t.Errorf("Inspector.Description = %q", inspector.Description)
	}
	if len(inspector.Commands) != 2 {
		t.Errorf("Inspector commands = %d, want 2", len(inspector.Commands))
	}
	if len(inspector.Events) != 3 {
		t.Errorf("Inspector events = %d, want 3", len(inspector.Events))
	}

	detached := inspector.Events[0]
	if detached.Name != "detached" {
		t.Errorf("event[0].Name = %q, want detached", detached.Name)
	}
	if len(detached.Parameters) != 1 {
		t.Fatalf("detached params = %d, want 1", len(detached.Parameters))
	}
	if detached.Parameters[0].Name != "reason" {
		t.Errorf(
			"detached param name = %q, want reason",
			detached.Parameters[0].Name,
		)
	}

	network := proto.Domains[1]
	if network.Name != "Network" {
		t.Errorf("domain[1].Name = %q, want Network", network.Name)
	}
	if len(network.Dependencies) != 2 {
		t.Errorf("Network deps = %d, want 2", len(network.Dependencies))
	}

	if len(network.Types) != 4 {
		t.Fatalf("Network types = %d, want 4", len(network.Types))
	}

	resourceType := network.Types[0]
	if resourceType.ID != "ResourceType" {
		t.Errorf("type[0].ID = %q, want ResourceType", resourceType.ID)
	}
	if resourceType.BaseType != "string" {
		t.Errorf(
			"ResourceType.BaseType = %q, want string",
			resourceType.BaseType,
		)
	}
	if len(resourceType.Enum) != 4 {
		t.Errorf("ResourceType enum = %d, want 4", len(resourceType.Enum))
	}

	request := network.Types[2]
	if request.ID != "Request" {
		t.Errorf("type[2].ID = %q, want Request", request.ID)
	}
	if len(request.Properties) != 8 {
		t.Fatalf("Request properties = %d, want 8", len(request.Properties))
	}

	postDataEntries := request.Properties[6]
	if postDataEntries.Name != "postDataEntries" {
		t.Errorf(
			"prop[6].Name = %q, want postDataEntries",
			postDataEntries.Name,
		)
	}
	if !postDataEntries.Optional {
		t.Error("postDataEntries should be optional")
	}
	if !postDataEntries.Experimental {
		t.Error("postDataEntries should be experimental")
	}
	if postDataEntries.Ref == nil || postDataEntries.Ref.Items == nil {
		t.Fatal("postDataEntries should be an array")
	}
	if postDataEntries.Ref.Items.Name != "PostDataEntry" {
		t.Errorf(
			"postDataEntries item type = %q, want PostDataEntry",
			postDataEntries.Ref.Items.Name,
		)
	}

	referrerPolicy := request.Properties[7]
	if referrerPolicy.Name != "referrerPolicy" {
		t.Errorf("prop[6].Name = %q, want referrerPolicy", referrerPolicy.Name)
	}
	if len(referrerPolicy.Enum) != 3 {
		t.Errorf("referrerPolicy enum = %d, want 3", len(referrerPolicy.Enum))
	}

	response := network.Types[3]
	if response.ID != "Response" {
		t.Errorf("type[3].ID = %q, want Response", response.ID)
	}
	requestIDProp := response.Properties[3]
	if requestIDProp.Ref == nil || requestIDProp.Ref.Domain != "Network" ||
		requestIDProp.Ref.Name != "RequestId" {
		t.Errorf(
			"Response.requestId ref = %+v, want Network.RequestId",
			requestIDProp.Ref,
		)
	}

	if len(network.Commands) != 3 {
		t.Fatalf("Network commands = %d, want 3", len(network.Commands))
	}

	sendReq := network.Commands[0]
	if sendReq.Name != "sendRequest" {
		t.Errorf("cmd[0].Name = %q, want sendRequest", sendReq.Name)
	}
	if len(sendReq.Parameters) != 3 {
		t.Errorf("sendRequest params = %d, want 3", len(sendReq.Parameters))
	}
	if len(sendReq.Returns) != 2 {
		t.Errorf("sendRequest returns = %d, want 2", len(sendReq.Returns))
	}

	getSec := network.Commands[1]
	if !getSec.Experimental {
		t.Error("getSecurityDetails should be experimental")
	}

	if len(network.Events) != 2 {
		t.Fatalf("Network events = %d, want 2", len(network.Events))
	}
	if !network.Events[0].Deprecated {
		t.Error("requestWillBeSent should be deprecated")
	}

	navigate := network.Commands[2]
	if navigate.Redirect != "Page" {
		t.Errorf("navigate.Redirect = %q, want Page", navigate.Redirect)
	}

	page := proto.Domains[2]
	if page.Name != "Page" {
		t.Errorf("domain[2].Name = %q, want Page", page.Name)
	}

	adFrame := page.Types[1]
	if !adFrame.Experimental {
		t.Error("AdFrameType should be experimental")
	}
	if len(adFrame.Enum) != 3 {
		t.Errorf("AdFrameType enum = %d, want 3", len(adFrame.Enum))
	}

	pageNav := page.Commands[1]
	if len(pageNav.Parameters) != 3 {
		t.Errorf("Page.navigate params = %d, want 3", len(pageNav.Parameters))
	}
	crossRef := pageNav.Parameters[2]
	if crossRef.Ref == nil || crossRef.Ref.Domain != "Network" ||
		crossRef.Ref.Name != "RequestId" {
		t.Errorf(
			"Page.navigate param[2] ref = %+v, want Network.RequestId",
			crossRef.Ref,
		)
	}
}

func TestParseIncludes(t *testing.T) {
	proto, err := ParseString(`
include domains/Network.pdl
include domains/Page.pdl
`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(proto.Includes) != 2 {
		t.Fatalf("includes = %d, want 2", len(proto.Includes))
	}
	if proto.Includes[0] != "domains/Network.pdl" {
		t.Errorf("include[0] = %q, want domains/Network.pdl", proto.Includes[0])
	}
}
