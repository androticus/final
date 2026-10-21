forget: reset-set-log-provider
`
This example shows how to set one of Final's simple built-in LogProviders
to the Session logger.
Try:

set-stderr-provider
write-log

set-stdout-provider
write-log

set-file-provider // goes to tmp/log-example.log
write-log

This example works best in headless mode: final-tui --headless
copy this into the terminal to compile:
"examples/logging/set-log-provider.fnl" compileFromFile
`
const: set-log-provider-blurb

/*
Set the provider into the Logger.
Pass null to clear the provider, and stop logging
( lp:LogProvider -- )
*/
: set-provider
  >ts "example" "" LogLevel.Info ts>
  logger.init^
;

: set-stderr-provider
  LogProvider.newStdErrLogProvider^
  set-provider
;

: set-stdout-provider
  LogProvider.newStdOutLogProvider^
  set-provider
;

: set-file-provider
  "tmp/log-example.log"
  LogProvider.newFileLogProvider^
  set-provider
;

// write a log with some json data at Info Level
: write-log
  "A log with data"  { 1 "param1" 2 "param2" } >json  logInfo
;

set-log-provider-blurb $
