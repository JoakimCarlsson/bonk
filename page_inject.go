package bonk

import (
	"encoding/json"
	"fmt"
)

// ScriptTagOption configures AddScriptTag behavior.
type ScriptTagOption func(*scriptTagConfig)

type scriptTagConfig struct {
	url     string
	content string
	typ     string
}

// ScriptTagURL sets the script src URL. The tag is appended to
// <head> and the call blocks until the script has loaded.
func ScriptTagURL(url string) ScriptTagOption {
	return func(c *scriptTagConfig) { c.url = url }
}

// ScriptTagContent sets inline script content.
func ScriptTagContent(content string) ScriptTagOption {
	return func(c *scriptTagConfig) { c.content = content }
}

// ScriptTagType sets the script type attribute (e.g. "module").
func ScriptTagType(t string) ScriptTagOption {
	return func(c *scriptTagConfig) { c.typ = t }
}

// AddScriptTag injects a <script> tag into the page's <head>.
// Provide either ScriptTagURL or ScriptTagContent; if a URL is
// given the call blocks until the script has loaded.
func (p *Page) AddScriptTag(
	opts ...ScriptTagOption,
) error {
	var cfg scriptTagConfig
	for _, o := range opts {
		o(&cfg)
	}

	if cfg.url == "" && cfg.content == "" {
		return fmt.Errorf(
			"bonk: AddScriptTag requires a URL or content",
		)
	}

	urlJSON, _ := json.Marshal(cfg.url)
	contentJSON, _ := json.Marshal(cfg.content)
	typeJSON, _ := json.Marshal(cfg.typ)

	js := fmt.Sprintf(`(function(){
var s=document.createElement('script');
var url=%s,content=%s,typ=%s;
if(typ) s.type=typ;
if(url){
  s.src=url;
  return new Promise(function(ok,fail){
    s.onload=ok;s.onerror=fail;
    document.head.appendChild(s);
  });
}
s.textContent=content;
document.head.appendChild(s);
})()`, urlJSON, contentJSON, typeJSON)

	_, err := p.Evaluate(js)
	return err
}

// StyleTagOption configures AddStyleTag behavior.
type StyleTagOption func(*styleTagConfig)

type styleTagConfig struct {
	url     string
	content string
}

// StyleTagURL sets the stylesheet href URL. The tag is appended to
// <head> and the call blocks until the stylesheet has loaded.
func StyleTagURL(url string) StyleTagOption {
	return func(c *styleTagConfig) { c.url = url }
}

// StyleTagContent sets inline style content.
func StyleTagContent(content string) StyleTagOption {
	return func(c *styleTagConfig) { c.content = content }
}

// AddStyleTag injects a <style> or <link rel="stylesheet"> tag into
// the page's <head>. Provide either StyleTagURL or StyleTagContent;
// if a URL is given the call blocks until the stylesheet has loaded.
func (p *Page) AddStyleTag(
	opts ...StyleTagOption,
) error {
	var cfg styleTagConfig
	for _, o := range opts {
		o(&cfg)
	}

	if cfg.url == "" && cfg.content == "" {
		return fmt.Errorf(
			"bonk: AddStyleTag requires a URL or content",
		)
	}

	urlJSON, _ := json.Marshal(cfg.url)
	contentJSON, _ := json.Marshal(cfg.content)

	js := fmt.Sprintf(`(function(){
var url=%s,content=%s;
if(url){
  var l=document.createElement('link');
  l.rel='stylesheet';l.href=url;
  return new Promise(function(ok,fail){
    l.onload=ok;l.onerror=fail;
    document.head.appendChild(l);
  });
}
var s=document.createElement('style');
s.textContent=content;
document.head.appendChild(s);
})()`, urlJSON, contentJSON)

	_, err := p.Evaluate(js)
	return err
}
