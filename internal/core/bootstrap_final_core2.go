package core

const bootstrapTranche2Manifest = `
// Early extras (coercion wrappers, …)
internal/final/core/0020-extras/0022-type-coercion.fnl
internal/final/core/0020-extras/0026-string.fnl

// Json rendering
internal/final/core/0040-json/0040-render.fnl

// Text parsing helpers - reusable by other parsers
internal/final/core/0040-json/0042-lex.fnl

// Json parse budgets (idx consts, cap table, tracker factory)
internal/final/core/0040-json/0044-parse-limits.fnl

// Json parsing
internal/final/core/0040-json/0046-parse.fnl

// Http subsystem
internal/final/core/0060-http/0062-http-json.fnl
internal/final/core/0060-http/0068-http_util.fnl

// Application Services
internal/final/core/0080-app-svcs/0082-config.fnl
internal/final/core/0080-app-svcs/0084-logging.fnl
internal/final/core/0080-app-svcs/0086-events.fnl

// Vec/Mat subsystem
internal/final/core/0100-vec-mat/0102-vec-mat-proto.fnl

// Utilities and Database subsystem
internal/final/core/0400-utils/0402-words.fnl
internal/final/core/0400-utils/0404-db-utils.fnl

// Modules - 05xx

// Modules — leaf worker words (0-19)
internal/final/core/0500-modules/0502-compile-module.fnl

// Modules - Orchestration (80-99)


// Final Word and Type documentation
internal/final/core/1000-doc/1000-types.fnl
internal/final/core/1000-doc/1002-basic-ops.fnl
internal/final/core/1000-doc/1003-transcendentals.fnl
internal/final/core/1000-doc/1004-control-flow.fnl
internal/final/core/1000-doc/1006-lists.fnl
internal/final/core/1000-doc/1007-sets.fnl
internal/final/core/1000-doc/1008-maps-objects.fnl
internal/final/core/1000-doc/1009-json-etc.fnl
internal/final/core/1000-doc/1010-datetime-interval.fnl
internal/final/core/1000-doc/1011-streams.fnl
internal/final/core/1000-doc/1012-net-http.fnl
internal/final/core/1000-doc/1013-regex.fnl
internal/final/core/1000-doc/1014-table.fnl
internal/final/core/1000-doc/1015-file-system.fnl
internal/final/core/1000-doc/1016-threading.fnl
internal/final/core/1000-doc/1017-vecs-mats.fnl
internal/final/core/1000-doc/1019-misc.fnl
internal/final/core/1000-doc/1020-db.fnl
internal/final/core/1000-doc/1021-crypto.fnl
internal/final/core/1000-doc/1022-os.fnl
internal/final/core/1000-doc/1023-modules.fnl
internal/final/core/1000-doc/1024-app-svcs.fnl
internal/final/core/1000-doc/1025-draw.fnl
internal/final/core/1000-doc/1026-image.fnl

// Help and documentation functionality
internal/final/core/1000-doc/1042-type-introspect.fnl
internal/final/core/1000-doc/1044-help.fnl
internal/final/core/1000-doc/1046-list.fnl
`
