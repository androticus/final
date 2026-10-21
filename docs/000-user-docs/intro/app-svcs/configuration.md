# Application Configuration

Final configuration is based on JSON.

Final provides Session configuration via a DynamicObject, that can be initialized with JSON.

## Final API

 ```
 DynamicObject Session.config via sess.config
 ```
## Embedder Initialization Options

Once the Session has been created, the following call can be made to set Session.config from a JSON string. Note that this call invokes a Send on the VM, in order to execute Final code. JSON parsing and rendering in Final is all done with Final code, there is no native code involved.

```go
sess := NewSession | NewLockSession

err := sess.InitConfig(jsonString)
```

## Final Initialization Options

Session provides an api for initializing sess.config with a JSON string. The api is set up this way so it works for both Embedder and Final initialization:

```
// 1. put your JSON on session.configJson
`{ "paramA" : 1, "paramB" : 2 }` sess.configJson!

// 2. Call the init
sess.initConfig^
```
This replaces the default empty sess.config object with one initialized from the JSON string supplied.
This call should be made during your application startup, before any code tries to access the config members. Successful initialization also clears sess.configJson

## Using sess.config

Session.config via sess.config is a standard DynamicObject. It will take the shape of the JSON string that was used to initialize it. The various properties of the config can be accessed using standard compound object syntax. For example, using the schema from the initialization done above:
```forth
: print-config
  ( sess.config.paramA@ sess.config.paramB@ )
  "paramA: %d  paramB: %d\n" fmt $
;
print-config

paramA: 1  paramB: 2
```

You can also store nested fields on the live config object without parking it:

```forth
3 sess.config.paramA!
```
