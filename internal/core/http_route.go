package core

import "net/http"

// HttpRoute binds one ServeMux pattern to a pre-spawned Final handler's request Pipe.
type HttpRoute struct {
	index       int
	name        string // spawn word name (diagnostics / errors)
	pattern     string // Go ServeMux pattern, e.g. "GET /api/items/{id}"
	requestPipe *Pipe
	goHandler   http.Handler
	spawnXt     XT
	paramNames  []string // {name} wildcards extracted from pattern
}

func (r *HttpRoute) FinalType() *Type {
	return httpRouteType
}

func registerHttpRouteMembers(t *Type) {
	t.registerProperty("index", "Int",
		func(inst Any) Any { return Int(inst.(*HttpRoute).index) },
		nil,
	)
	t.registerProperty("name", "String",
		func(inst Any) Any { return String(inst.(*HttpRoute).name) },
		func(inst Any, value Any) { inst.(*HttpRoute).name = string(value.(String)) },
	)
	t.registerProperty("pattern", "String",
		func(inst Any) Any { return String(inst.(*HttpRoute).pattern) },
		nil,
	)
	t.registerProperty("requestPipe", "Pipe",
		func(inst Any) Any { return inst.(*HttpRoute).requestPipe },
		nil,
	)
	t.registerProperty("spawnXt", "Int",
		func(inst Any) Any { return inst.(*HttpRoute).spawnXt },
		func(inst Any, value Any) { inst.(*HttpRoute).spawnXt = value.(XT) },
	)
}

func CreateRegisterHttpRouteType() {
	t := &Type{
		name:            "HttpRoute",
		getZero:         ReferenceTypeZeroFn,
		checkCoerceType: MakePointerTypeCheckCoerce[HttpRoute]("HttpRoute"),
		restrict:        SR_HttpListen,
	}
	registerHttpRouteMembers(t)
	GetTypeRegistrar().RegisterCurrentType(t)
	httpRouteType = t
}

var httpRouteType *Type

func init() {
	CreateRegisterHttpRouteType()
}

var _ Typed = (*HttpRoute)(nil)

// serveMuxPathParamNames returns {name} wildcard names in pattern order (excludes {name...}).
func serveMuxPathParamNames(pattern string) []string {
	var names []string
	for i := 0; i < len(pattern); i++ {
		if pattern[i] != '{' {
			continue
		}
		j := i + 1
		for j < len(pattern) && pattern[j] != '}' {
			j++
		}
		if j >= len(pattern) {
			break
		}
		name := pattern[i+1 : j]
		if name == "" || name == "$" {
			i = j
			continue
		}
		if len(name) >= 3 && name[len(name)-3:] == "..." {
			name = name[:len(name)-3]
		}
		if name != "" {
			names = append(names, name)
		}
		i = j
	}
	return names
}
