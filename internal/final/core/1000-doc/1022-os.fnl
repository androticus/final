// Moniker

/*
A type for encapsulating a secret value that shouldn't leak into the runtime.
This type can only have its key values filled by a provider, eg. see Env. subMoniker
*/
:doc Moniker

/*
Create a moniker, with all keynames as $keyname, to be filled by a provider, eg. Env
( pattern:s -- mon:Moniker )
*/
:doc Moniker#new

/*
Env is a class-only type that provides environment variable loading.
*/
:doc Env 

/*
Load an environment variable from an .env file or the environment.
.env file if used must exist in cwd when the command runs.
Returns a value and true if resolved, false otherwise
( key:String -- val:String true | false )
*/
:doc Env#get

/*
Substitutes all the $keyname params into the Moniker.
Returns true if key found (all $keyname occurrences replaced); false if missing (mon unchanged).
( mon:Moniker keyname:String -- b:Bool )
*/
:doc Env#subMoniker

// Sys

/*
A global object used to read system-wide attributes,
and call system-level methods.
Access it via the axiom `sys`.
*/
:doc Sys

/*
The arguments passed to the Final executable.
argv[0] is the path of the Final executable itself.
*/
:doc Sys#argv

/*
Logical CPU count available to this process.
Accurate on every supported OS (Windows, macOS, Linux, Pi),
including CPU affinity. Does not reflect container CPU quotas.
*/
:doc Sys#coreCount

/*
Is the program running an interactive terminal session.
True for normal tui or repl execution, or when executing
a command file.
False for piped input.
*/
:doc Sys#isTty

/*
The parameters passed to a command file or piped script.
params[0] will be the command file, or `-` for a pipe.
params[1..N] will be the params listed after the command file, if any.
Will be empty for normal interactive usage.
eg. final myFinalScript.fnl param1 param2
*/
:doc Sys#params

/*
Launch an external program with argv args; wait for exit.
Captures merged stdout/stderr. timeoutS is whole seconds; 0 = no deadline.
Timeout returns err "timeout" and any partial output.
Nonzero exit is still modal true (check code).
Requires SR_System.
( program:s args:StringList timeoutS:i -- out:s code:i true | out:s err:s false )
*/
:doc Sys#exec?
