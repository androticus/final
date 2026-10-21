/*
Final Logging words
*/

/*
Will a log get sent at Fatal level?
( -- b )
*/
: logFatal?
  LogLevel.Fatal logger.willLog^
;

/*
Will a log get sent at Error level?
( -- b )
*/
: logError?
  LogLevel.Error logger.willLog^
;

/*
Will a log get sent at Warn level?
( -- b )
*/
: logWarn?
  LogLevel.Warn logger.willLog^
;

/*
Will a log get sent at Info level?
( -- b )
*/
: logInfo?
  LogLevel.Info logger.willLog^
;

/*
Will a log get sent at Debug level?
( -- b )
*/
: logDebug?
  LogLevel.Debug logger.willLog^
;

/*
Will a log get sent at Trace level?
( -- b )
*/
: logTrace?
  LogLevel.Trace logger.willLog^
;

/*
Send a log at Fatal level, if enabled.
( msg:s data:a -- )
*/
: logFatal
  LogLevel.Fatal logAt
;

/*
Send a log at Error level, if enabled.
( msg:s data:a -- )
*/
: logError
  LogLevel.Error logAt
;

/*
Send a log at Warn level, if enabled.
( msg:s data:a -- )
*/
: logWarn
  LogLevel.Warn logAt
;

/*
Send a log at Info level, if enabled.
( msg:s data:a -- )
*/
: logInfo
  LogLevel.Info logAt
;

/*
Send a log at Debug level, if enabled.
( msg:s data:a -- )
*/
: logDebug
  LogLevel.Debug logAt
;

/*
Send a log at Trace level, if enabled.
( msg:s data:a -- )
*/
: logTrace
  LogLevel.Trace logAt
;
