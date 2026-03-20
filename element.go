package bonk

import (
	"github.com/joakimcarlsson/bonk/proto"
)

// Element represents a DOM element on a page.
type Element struct {
	page     *Page
	objectID proto.RemoteObjectID
	selector string
}
