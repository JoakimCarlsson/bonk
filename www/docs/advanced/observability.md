# Observability

When automation hangs or a tab dies, bonk's default error (`bonk: timeout`)
doesn't say *why*. These options surface the signals Chrome actually emits —
stderr, renderer crashes, and Crashpad minidumps — so you can debug failures
that happen below the CDP layer.

## Chrome stderr

By default, bonk reads Chrome's stderr only until the `DevTools listening on …`
line, then drops the rest. To keep it, pass a writer:

```go
f, _ := os.Create("chrome.stderr.log")
defer f.Close()

b, err := bonk.Launch(
    bonk.WithStderrSink(f),
)
```

This catches lines like `[ERROR:google_apis\gcm\engine\registration_request.cc:291] ...`,
GPU process failures, and the `[FATAL]` lines that precede a forced exit. Combine
with `--enable-logging=stderr --v=1` via `bonk.Args` for verbose Chromium logs.

## Renderer crashes

Subscribe to `Inspector.targetCrashed` per page:

```go
page.OnCrash(func(reason string) {
    log.Printf("renderer crashed: %s", reason)
})
```

The CDP event carries no payload, so `reason` is currently the event name
(`"Inspector.targetCrashed"`); use it as a signal to fetch the crash dump
(below) rather than expecting a cause string.

!!! warning "Stealth"
    `OnCrash` enables the `Inspector` CDP domain on the page, which is
    detectable by anti-automation scripts. Skip it in stealth contexts unless
    you specifically need crash signals.

## Crashpad minidumps

Chrome writes minidumps to `<user-data-dir>/Crashpad/reports/` when a process
dies. Bonk snapshots that directory at launch, and on `Browser.Close` reports
any *new* dumps to a handler you provide:

```go
b, err := bonk.Launch(
    bonk.WithCrashpadHandler(func(reports []bonk.CrashReport) {
        for _, r := range reports {
            log.Printf("crash dump: %s (%d bytes) — %s",
                r.Path, r.Size, r.Summary)
        }
    }),
)
```

`CrashReport.Summary` is a best-effort one-line cause pulled from the dump's
embedded annotation strings — e.g. `"v8-oom in Zone (stringify)"`. It is not
a full minidump parse; expect empty strings for crashes that don't include
recognised annotations.

!!! important "Read or copy synchronously"
    The handler runs inside `Close()` *before* the temp user-data-dir is
    deleted. If you launched without `UserDataDir(...)`, the `.dmp` files
    vanish as soon as the handler returns. Copy or read them in the handler
    if you need them afterwards.

## Putting it together

A typical "tell me what happened" launch:

```go
stderr, _ := os.Create("chrome.stderr.log")
defer stderr.Close()

b, err := bonk.Launch(
    bonk.Headless(false),
    bonk.WithStderrSink(stderr),
    bonk.WithCrashpadHandler(func(reports []bonk.CrashReport) {
        for _, r := range reports {
            log.Printf("crash: %s — %s", r.Path, r.Summary)
        }
    }),
)
if err != nil {
    log.Fatal(err)
}
defer b.Close()

page, _ := ctx.NewPage()
page.OnCrash(func(reason string) {
    log.Printf("renderer crashed: %s", reason)
})
```

If the tab dies, you get the `OnCrash` callback immediately, the Chrome stderr
in your file, and a dump path + summary on `Close()`.
