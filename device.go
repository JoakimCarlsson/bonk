package bonk

import "github.com/joakimcarlsson/bonk/proto"

// Device describes a device profile for emulation.
type Device struct {
	Name              string
	UserAgent         string
	ViewportWidth     int
	ViewportHeight    int
	DeviceScaleFactor float64
	IsMobile          bool
	HasTouch          bool
}

// IPhone15, IPhone15ProMax, Pixel7, Pixel8, IPadPro11, IPadAir, and GalaxyS23 are preset device profiles for Emulate.
var (
	IPhone15 = Device{
		Name:              "iPhone 15",
		UserAgent:         "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		ViewportWidth:     393,
		ViewportHeight:    852,
		DeviceScaleFactor: 3,
		IsMobile:          true,
		HasTouch:          true,
	}

	IPhone15ProMax = Device{
		Name:              "iPhone 15 Pro Max",
		UserAgent:         "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		ViewportWidth:     430,
		ViewportHeight:    932,
		DeviceScaleFactor: 3,
		IsMobile:          true,
		HasTouch:          true,
	}

	Pixel7 = Device{
		Name:              "Pixel 7",
		UserAgent:         "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Mobile Safari/537.36",
		ViewportWidth:     412,
		ViewportHeight:    915,
		DeviceScaleFactor: 2.625,
		IsMobile:          true,
		HasTouch:          true,
	}

	Pixel8 = Device{
		Name:              "Pixel 8",
		UserAgent:         "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36",
		ViewportWidth:     412,
		ViewportHeight:    915,
		DeviceScaleFactor: 2.625,
		IsMobile:          true,
		HasTouch:          true,
	}

	IPadPro11 = Device{
		Name:              "iPad Pro 11",
		UserAgent:         "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		ViewportWidth:     834,
		ViewportHeight:    1194,
		DeviceScaleFactor: 2,
		IsMobile:          true,
		HasTouch:          true,
	}

	IPadAir = Device{
		Name:              "iPad Air",
		UserAgent:         "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		ViewportWidth:     820,
		ViewportHeight:    1180,
		DeviceScaleFactor: 2,
		IsMobile:          true,
		HasTouch:          true,
	}

	GalaxyS23 = Device{
		Name:              "Galaxy S23",
		UserAgent:         "Mozilla/5.0 (Linux; Android 13; SM-S911B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Mobile Safari/537.36",
		ViewportWidth:     360,
		ViewportHeight:    780,
		DeviceScaleFactor: 3,
		IsMobile:          true,
		HasTouch:          true,
	}
)

// Emulate applies a device profile to the page, setting viewport,
// user agent, device scale factor, and touch emulation.
func (p *Page) Emulate(d Device) error {
	if err := proto.EmulationSetDeviceMetricsOverride(
		int64(d.ViewportWidth),
		int64(d.ViewportHeight),
		d.DeviceScaleFactor,
		d.IsMobile,
	).Do(p.execCtx); err != nil {
		return err
	}

	if err := proto.EmulationSetTouchEmulationEnabled(d.HasTouch).
		Do(p.execCtx); err != nil {
		return err
	}

	return proto.EmulationSetUserAgentOverride(d.UserAgent).Do(p.execCtx)
}
