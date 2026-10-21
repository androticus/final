# Application Services

Final provides basic application services often needed in embedding scenarios, or when building an application using Final. Most of these services present a single usage API in Final, but can be configured in two complementary but compatible ways:
1. By the embedder, when intitializing a Session;
2. In Final, during Session startup.

Final provides the following Application Services, as of this release:
* [Configuration](configuration.md): a configuration object built from a JSON string, accessible as Session.config via sess.config in Final code,
* [Logging](logging.md): a Logging interface, with a pluggable LogProvider, accessible as Session.logger via sess.logger in Final code
* [Cmdline Scripting and OS Piping](cmd-piping.md)
