package bonk

import (
	"time"

	"github.com/joakimcarlsson/bonk/proto"
)

// Mouse provides low-level mouse control on a page.
type Mouse struct {
	page *Page
	x, y float64
}

// Mouse returns the page's Mouse controller.
func (p *Page) Mouse() *Mouse {
	return &Mouse{page: p}
}

// Move moves the mouse to the given coordinates.
func (m *Mouse) Move(x, y float64) error {
	m.x = x
	m.y = y
	return proto.InputDispatchMouseEvent(
		proto.InputDispatchMouseEventTypeMouseMoved, x, y,
	).Do(m.page.execCtx)
}

// Click moves to the coordinates and performs a click.
func (m *Mouse) Click(x, y float64) error {
	if err := m.Move(x, y); err != nil {
		return err
	}
	if err := m.Down(); err != nil {
		return err
	}
	return m.Up()
}

// Down presses the left mouse button at the current position.
func (m *Mouse) Down() error {
	return proto.InputDispatchMouseEvent(
		proto.InputDispatchMouseEventTypeMousePressed, m.x, m.y,
	).WithButton(proto.InputMouseButtonLeft).
		WithClickCount(1).
		Do(m.page.execCtx)
}

// Up releases the left mouse button at the current position.
func (m *Mouse) Up() error {
	return proto.InputDispatchMouseEvent(
		proto.InputDispatchMouseEventTypeMouseReleased, m.x, m.y,
	).WithButton(proto.InputMouseButtonLeft).
		WithClickCount(1).
		Do(m.page.execCtx)
}

// DragTo performs a drag from one point to another.
func (m *Mouse) DragTo(fromX, fromY, toX, toY float64) error {
	if err := m.Move(fromX, fromY); err != nil {
		return err
	}
	if err := m.Down(); err != nil {
		return err
	}

	steps := 10
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := fromX + (toX-fromX)*t
		y := fromY + (toY-fromY)*t
		if err := m.Move(x, y); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
	}

	return m.Up()
}
