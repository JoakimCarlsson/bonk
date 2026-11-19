package bonk

import (
	"github.com/joakimcarlsson/bonk/proto"
)

// Mouse provides low-level mouse control on a page. Its position is the
// page's, so it carries over between Mouse values and element actions.
type Mouse struct {
	page *Page
}

// Mouse returns the page's Mouse controller.
func (p *Page) Mouse() *Mouse {
	return &Mouse{page: p}
}

// Move moves the mouse to the given coordinates, along a human path on a
// stealth page. It returns the first CDP error.
func (m *Mouse) Move(x, y float64) error {
	return m.page.moveCursor(x, y)
}

// Click moves to the coordinates and performs a click. It returns the first
// CDP error.
func (m *Mouse) Click(x, y float64) error {
	return m.page.clickAt(x, y, 1)
}

// Down presses the left mouse button at the current position. It returns
// the CDP error.
func (m *Mouse) Down() error {
	x, y := m.page.cursor.position()
	return m.page.dispatchMouse(
		proto.InputDispatchMouseEventTypeMousePressed, x, y,
		proto.InputMouseButtonLeft, 1,
	)
}

// Up releases the left mouse button at the current position. It returns
// the CDP error.
func (m *Mouse) Up() error {
	x, y := m.page.cursor.position()
	return m.page.dispatchMouse(
		proto.InputDispatchMouseEventTypeMouseReleased, x, y,
		proto.InputMouseButtonLeft, 1,
	)
}

// DragTo drags from one point to another with the left button held. It
// returns the first CDP error.
func (m *Mouse) DragTo(fromX, fromY, toX, toY float64) error {
	if err := m.Move(fromX, fromY); err != nil {
		return err
	}
	if err := m.Down(); err != nil {
		return err
	}
	if !m.page.stealth {
		const steps = 10
		for i := 1; i <= steps; i++ {
			t := float64(i) / steps
			if err := m.page.dispatchMouse(
				proto.InputDispatchMouseEventTypeMouseMoved,
				fromX+(toX-fromX)*t, fromY+(toY-fromY)*t,
				proto.InputMouseButtonNone, 0,
			); err != nil {
				return err
			}
		}
	} else if err := m.Move(toX, toY); err != nil {
		return err
	}
	return m.Up()
}
