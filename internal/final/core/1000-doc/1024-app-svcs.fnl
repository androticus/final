// Application Services

// Configuration

// Logging

// Session

/*
The Session object.
( -- sess:Session )
*/
:doc sess

/*
Session provides access to Session-level members.
These values apply to all Pips in the Session
*/
:doc Session

/*
The JSON string that will create .config
if initConfig is called.
*/
:doc Session#configJson

/*
The Session config, if it has been set via
configJson and initConfig, otherwise empty.
*/
:doc Session#config

/*
Get the application Logger intance.
NOTE: also available via Final word `logger`.
( -- logger:Logger )
*/
:doc Session#logger


/*
Initialize config by parsing configJson.
Can be called in Final startup.
Called in an embedding call to Session.InitConfig
in a Send.
( -- )
*/
:doc Session#initConfig


// Logger

/*
The Session-wide Logger object.
( -- logger:Logger )
*/
:doc logger

/*
Logger is the Final-level logging object.
It contains the properties shared by all logs.
*/
:doc Logger

/*
An optional Instance ID, such as a serial number, UUID, etc.
LogProviders must accept an empty value.
*/
:doc Logger#instanceId

/*
The Service name, eg "final" "MyApp" "MyApp.MyModule".
Non-empty value may be required by LogProviders
*/
:doc Logger#service

/*
The current logging level.
Only logs at this level or lower get logged.
0 = Off.
*/
:doc Logger#level

/*
Write a a log, if logging is enabled, and level is at or
below current Logger.level.
Note: normally use Final log* words to log.d
This can be used to write logs into another Session.
( msg:s data:a level:i -- )
*/
:doc Logger#log

/*
Initialize the Logger, and set a LogProvider.
( service:s instanceId:s level:i lp:LogProvider -- )
*/
:doc Logger#init

/*
Change the logging level (0 = Off).
( level:i -- )
*/
:doc Logger#setLevel

/*
Set only the LogProvider; `null` disables logging.
Normally use .init to set all values.
( lp:LogProvider -- )
*/
:doc Logger#setLogProvider

/*
Will the logger log at this level?
( level:i -- b )
*/
:doc Logger#willLog


// LogLevel

/*
LogLevel declares the LogLevel constants.
*/
:doc LogLevel

/*
Logging is Off.
*/
:doc LogLevel#Off

/*
Logging at Fatal Level.
*/
:doc LogLevel#Fatal

/*
Logging at Error Level.
*/
:doc LogLevel#Error

/*
Logging at Warn Level.
*/
:doc LogLevel#Warn

/*
Logging at Info Level (typical value).
*/
:doc LogLevel#Info

/*
Logging at Debug Level.
*/
:doc LogLevel#Debug

/*
Logging at Trace Level.
*/
:doc LogLevel#Trace

// LogProvider

/*
Factory object to get LogProvider instances.
*/
:doc LogProvider

/*
LogProvider that writes logs to stdout.
( -- lp:LogProvider )
*/
:doc LogProvider#newStdOutLogProvider

/*
LogProvider that writes logs to stderr.
( -- lp:LogProvider )
*/
:doc LogProvider#newStdErrLogProvider

/*
LogProvider that writes logs to a file.
File is opened for Append or Create.
( path:s -- lp:LogProvider )
*/
:doc LogProvider#newFileLogProvider


// Final Logging axioms

/*
Log the msg and data at the indicated level, if enabled.
( data:a msg:s level:i )
*/
:doc logAt
