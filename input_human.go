package bonk

import (
	"math"
	"math/rand/v2"
	"sync"
	"time"
	"unicode"

	"github.com/joakimcarlsson/bonk/proto"
)

// cursorState is where the page's mouse pointer last was and which buttons
// it holds, so consecutive gestures continue from one place the way a real
// pointer does.
type cursorState struct {
	mu      sync.Mutex
	x, y    float64
	buttons int64
	placed  bool
}

// position returns the pointer's last coordinates.
func (c *cursorState) position() (x, y float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.x, c.y
}

// origin returns where a movement towards x, y starts. Before the pointer
// has been anywhere it is put at a random spot a few hundred pixels away,
// since a real pointer never starts at the viewport's corner.
func (c *cursorState) origin(x, y float64) (sx, sy float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.placed {
		angle := rand.Float64() * 2 * math.Pi
		r := 150 + rand.Float64()*350
		c.x = math.Max(0, x+math.Cos(angle)*r)
		c.y = math.Max(0, y+math.Sin(angle)*r)
		c.placed = true
	}
	return c.x, c.y
}

// pause sleeps for a random duration between lo and hi milliseconds, the
// gap a human leaves between two input events.
func pause(lo, hi int) {
	time.Sleep(time.Duration(lo+rand.IntN(hi-lo+1)) * time.Millisecond)
}

// dispatchMouse sends one mouse event at x, y with the button state the
// cursor holds, recording the new position. button is the button the
// event concerns and clickCount the click number of a press or release.
// It returns the error of the CDP command.
func (p *Page) dispatchMouse(
	typ proto.InputDispatchMouseEventType,
	x, y float64,
	button proto.InputMouseButton,
	clickCount int64,
) error {
	p.cursor.mu.Lock()
	switch typ {
	case proto.InputDispatchMouseEventTypeMousePressed:
		p.cursor.buttons |= 1
	case proto.InputDispatchMouseEventTypeMouseReleased:
		p.cursor.buttons &^= 1
	}
	p.cursor.x, p.cursor.y, p.cursor.placed = x, y, true
	buttons := p.cursor.buttons
	p.cursor.mu.Unlock()

	params := proto.InputDispatchMouseEvent(typ, x, y).
		WithButtons(buttons).
		WithPointerType(proto.InputDispatchMouseEventPointerTypeMouse)
	if typ == proto.InputDispatchMouseEventTypeMouseMoved {
		if buttons&1 != 0 {
			params = params.WithButton(proto.InputMouseButtonLeft)
		}
	} else {
		params = params.WithButton(button).WithClickCount(clickCount)
		if typ == proto.InputDispatchMouseEventTypeMousePressed {
			params = params.WithForce(0.5)
		}
	}
	return params.Do(p.execCtx)
}

// moveCursor moves the pointer to x, y. A stealth page travels there along
// a curved, eased path of many small steps with human timing; any other
// page jumps there in one event. It returns the first CDP error.
func (p *Page) moveCursor(x, y float64) error {
	if !p.stealth {
		return p.dispatchMouse(
			proto.InputDispatchMouseEventTypeMouseMoved, x, y,
			proto.InputMouseButtonNone, 0,
		)
	}
	sx, sy := p.cursor.origin(x, y)
	for _, pt := range humanPath(sx, sy, x, y) {
		if err := p.dispatchMouse(
			proto.InputDispatchMouseEventTypeMouseMoved, pt[0], pt[1],
			proto.InputMouseButtonNone, 0,
		); err != nil {
			return err
		}
		pause(6, 16)
	}
	return nil
}

// humanPath returns the points of a pointer movement from sx, sy to x, y:
// a cubic Bézier curve with randomly bent control points, eased so it
// accelerates and settles, with a little jitter on the way and the exact
// target as the last point. Short moves get few points, long ones more.
func humanPath(sx, sy, x, y float64) [][2]float64 {
	dx, dy := x-sx, y-sy
	dist := math.Hypot(dx, dy)
	if dist < 2 {
		return [][2]float64{{x, y}}
	}
	steps := int(math.Min(45, math.Max(8, dist/18)))
	bend := dist * (0.08 + rand.Float64()*0.2)
	nx, ny := -dy/dist, dx/dist
	side := 1.0
	if rand.IntN(2) == 0 {
		side = -1
	}
	c1x := sx + dx*0.3 + nx*bend*side*rand.Float64()
	c1y := sy + dy*0.3 + ny*bend*side*rand.Float64()
	c2x := sx + dx*0.7 + nx*bend*side*rand.Float64()
	c2y := sy + dy*0.7 + ny*bend*side*rand.Float64()

	pts := make([][2]float64, 0, steps)
	for i := 1; i <= steps; i++ {
		t := float64(i) / float64(steps)
		e := t * t * (3 - 2*t)
		u := 1 - e
		px := u*u*u*sx + 3*u*u*e*c1x + 3*u*e*e*c2x + e*e*e*x
		py := u*u*u*sy + 3*u*u*e*c1y + 3*u*e*e*c2y + e*e*e*y
		if i < steps {
			px += rand.NormFloat64() * 0.6
			py += rand.NormFloat64() * 0.6
		}
		pts = append(pts, [2]float64{px, py})
	}
	return pts
}

// clickAt moves the pointer to x, y and clicks count times, with a human
// pause before pressing and a human hold of the button on a stealth page.
// It returns the first CDP error.
func (p *Page) clickAt(x, y float64, count int64) error {
	if err := p.moveCursor(x, y); err != nil {
		return err
	}
	for i := int64(1); i <= count; i++ {
		if p.stealth {
			pause(40, 120)
		}
		if err := p.dispatchMouse(
			proto.InputDispatchMouseEventTypeMousePressed, x, y,
			proto.InputMouseButtonLeft, i,
		); err != nil {
			return err
		}
		if p.stealth {
			pause(45, 110)
		}
		if err := p.dispatchMouse(
			proto.InputDispatchMouseEventTypeMouseReleased, x, y,
			proto.InputMouseButtonLeft, i,
		); err != nil {
			return err
		}
	}
	return nil
}

// targetPoint picks where in box to put the pointer: its exact centre on a
// plain page, a random spot in the middle part of it on a stealth page, as
// people rarely hit dead centre.
func (p *Page) targetPoint(box *Box) (x, y float64) {
	x, y = box.X+box.Width/2, box.Y+box.Height/2
	if !p.stealth {
		return x, y
	}
	x += (rand.Float64() - 0.5) * box.Width * 0.5
	y += (rand.Float64() - 0.5) * box.Height * 0.5
	return x, y
}

// keyDef describes one physical key as a keyboard event reports it.
type keyDef struct {
	key   string
	code  string
	vk    int64
	text  string
	shift bool
}

// namedKeys are the non-printing keys by their KeyboardEvent.key name.
var namedKeys = map[string]keyDef{
	"Enter":      {key: "Enter", code: "Enter", vk: 13, text: "\r"},
	"Tab":        {key: "Tab", code: "Tab", vk: 9},
	"Backspace":  {key: "Backspace", code: "Backspace", vk: 8},
	"Delete":     {key: "Delete", code: "Delete", vk: 46},
	"Escape":     {key: "Escape", code: "Escape", vk: 27},
	"ArrowLeft":  {key: "ArrowLeft", code: "ArrowLeft", vk: 37},
	"ArrowUp":    {key: "ArrowUp", code: "ArrowUp", vk: 38},
	"ArrowRight": {key: "ArrowRight", code: "ArrowRight", vk: 39},
	"ArrowDown":  {key: "ArrowDown", code: "ArrowDown", vk: 40},
	"Home":       {key: "Home", code: "Home", vk: 36},
	"End":        {key: "End", code: "End", vk: 35},
	"PageUp":     {key: "PageUp", code: "PageUp", vk: 33},
	"PageDown":   {key: "PageDown", code: "PageDown", vk: 34},
	"Shift":      {key: "Shift", code: "ShiftLeft", vk: 16},
	"Control":    {key: "Control", code: "ControlLeft", vk: 17},
	"Alt":        {key: "Alt", code: "AltLeft", vk: 18},
	"Meta":       {key: "Meta", code: "MetaLeft", vk: 91},
	" ":          {key: " ", code: "Space", vk: 32, text: " "},
}

// punctKeys maps US layout punctuation to its code, virtual key code and
// whether Shift produces it.
var punctKeys = map[rune]keyDef{
	'-': {code: "Minus", vk: 189}, '_': {code: "Minus", vk: 189, shift: true},
	'=': {code: "Equal", vk: 187}, '+': {code: "Equal", vk: 187, shift: true},
	'[': {code: "BracketLeft", vk: 219}, '{': {code: "BracketLeft", vk: 219, shift: true},
	']': {code: "BracketRight", vk: 221}, '}': {code: "BracketRight", vk: 221, shift: true},
	'\\': {code: "Backslash", vk: 220}, '|': {code: "Backslash", vk: 220, shift: true},
	';': {code: "Semicolon", vk: 186}, ':': {code: "Semicolon", vk: 186, shift: true},
	'\'': {code: "Quote", vk: 222}, '"': {code: "Quote", vk: 222, shift: true},
	',': {code: "Comma", vk: 188}, '<': {code: "Comma", vk: 188, shift: true},
	'.': {code: "Period", vk: 190}, '>': {code: "Period", vk: 190, shift: true},
	'/': {code: "Slash", vk: 191}, '?': {code: "Slash", vk: 191, shift: true},
	'`': {code: "Backquote", vk: 192}, '~': {code: "Backquote", vk: 192, shift: true},
	'!': {code: "Digit1", vk: 49, shift: true}, '@': {code: "Digit2", vk: 50, shift: true},
	'#': {code: "Digit3", vk: 51, shift: true}, '$': {code: "Digit4", vk: 52, shift: true},
	'%': {code: "Digit5", vk: 53, shift: true}, '^': {code: "Digit6", vk: 54, shift: true},
	'&': {code: "Digit7", vk: 55, shift: true}, '*': {code: "Digit8", vk: 56, shift: true},
	'(': {code: "Digit9", vk: 57, shift: true}, ')': {code: "Digit0", vk: 48, shift: true},
}

// lookupKey returns the key definition for a key name or a single
// character. Characters outside the US layout get only key and text.
func lookupKey(key string) keyDef {
	if d, ok := namedKeys[key]; ok {
		return d
	}
	r := []rune(key)
	if len(r) != 1 {
		return keyDef{key: key}
	}
	ch := r[0]
	d := keyDef{key: key, text: key}
	switch {
	case ch >= 'a' && ch <= 'z':
		d.code, d.vk = "Key"+string(unicode.ToUpper(ch)), int64(unicode.ToUpper(ch))
	case ch >= 'A' && ch <= 'Z':
		d.code, d.vk, d.shift = "Key"+string(ch), int64(ch), true
	case ch >= '0' && ch <= '9':
		d.code, d.vk = "Digit"+string(ch), int64(ch)
	default:
		if pd, ok := punctKeys[ch]; ok {
			d.code, d.vk, d.shift = pd.code, pd.vk, pd.shift
		}
	}
	return d
}

// modifierShift is the CDP modifier bit for Shift.
const modifierShift = 8

// dispatchKey sends a key down or key up event for d, with its code,
// virtual key code and Shift modifier, and its text on key down. It
// returns the error of the CDP command.
func (p *Page) dispatchKey(down bool, d keyDef) error {
	typ := proto.InputDispatchKeyEventTypeKeyUp
	if down {
		typ = proto.InputDispatchKeyEventTypeRawKeyDown
		if d.text != "" {
			typ = proto.InputDispatchKeyEventTypeKeyDown
		}
	}
	params := proto.InputDispatchKeyEvent(typ).WithKey(d.key)
	if d.code != "" {
		params = params.WithCode(d.code)
	}
	if d.vk != 0 {
		params = params.WithWindowsVirtualKeyCode(d.vk).
			WithNativeVirtualKeyCode(d.vk)
	}
	if d.shift {
		params = params.WithModifiers(modifierShift)
	}
	if down && d.text != "" {
		params = params.WithText(d.text).WithUnmodifiedText(d.text)
	}
	return params.Do(p.execCtx)
}

// pressKey presses and releases one key, holding Shift around it when the
// character needs it, with human hold and gap times on a stealth page. It
// returns the first CDP error.
func (p *Page) pressKey(key string) error {
	d := lookupKey(key)
	shift := namedKeys["Shift"]
	if d.shift {
		if err := p.dispatchKey(true, shift); err != nil {
			return err
		}
	}
	if err := p.dispatchKey(true, d); err != nil {
		return err
	}
	if p.stealth {
		pause(25, 85)
	}
	if err := p.dispatchKey(false, d); err != nil {
		return err
	}
	if d.shift {
		return p.dispatchKey(false, shift)
	}
	return nil
}

// typeText types text one key at a time. delay is the gap between keys;
// zero on a stealth page means a random human gap. It returns the first
// CDP error.
func (p *Page) typeText(text string, delay time.Duration) error {
	for _, ch := range text {
		if err := p.pressKey(string(ch)); err != nil {
			return err
		}
		switch {
		case delay > 0:
			time.Sleep(delay)
		case p.stealth:
			pause(30, 140)
		}
	}
	return nil
}
