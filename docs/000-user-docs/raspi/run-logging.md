# Logging

Prefer Final’s logger over ad-hoc `$` prints once a controller leaves the bench. Full API: [Logging](../intro/app-svcs/logging.md). Example: `examples/app-svcs/logging.fnl`.

Quick use:

```final
  "btn enable" "" logInfo
  "duty changed" { o:.duty@ "duty" } >json logDebug
```

Levels: Off, Fatal, Error, Warn, Info, Debug, Trace. Set level and provider once at startup (`logger.init`). For costly message builds, gate with `logDebug? if ... then`.
