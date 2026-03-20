# Device Emulation

bonk can emulate mobile and tablet devices by setting viewport dimensions, device scale factor, touch support, and user agent.

## Emulate a Device

```go
err = page.Emulate(bonk.IPhone15)
```

This sets:

- Viewport dimensions
- Device scale factor
- Mobile flag
- Touch emulation
- User agent string

## Built-in Presets

| Device | Viewport | Scale | Mobile | Touch |
|--------|----------|-------|--------|-------|
| `IPhone15` | 393x852 | 3x | Yes | Yes |
| `IPhone15ProMax` | 430x932 | 3x | Yes | Yes |
| `Pixel7` | 412x915 | 2.625x | Yes | Yes |
| `Pixel8` | 412x915 | 2.625x | Yes | Yes |
| `IPadPro11` | 834x1194 | 2x | Yes | Yes |
| `IPadAir` | 820x1180 | 2x | Yes | Yes |
| `GalaxyS23` | 360x780 | 3x | Yes | Yes |

## Custom Devices

Create your own device profile:

```go
myDevice := bonk.Device{
    Name:              "Custom Tablet",
    UserAgent:         "Mozilla/5.0 (Linux; Android 13; ...) ...",
    ViewportWidth:     800,
    ViewportHeight:    1280,
    DeviceScaleFactor: 2,
    IsMobile:          true,
    HasTouch:          true,
}

page.Emulate(myDevice)
```

## Manual Viewport

Set viewport without emulating a full device:

```go
page.SetViewport(1440, 900)
```
