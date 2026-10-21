package core

import "fmt"

// CreateRegisterParserType registers the Parser sealed type.
func CreateRegisterParserType() {
	t := &Type{
		name: "Parser",
		getZero: func() any {
			return (*Parser)(nil)
		},
		checkCoerceType: func(value any) any {
			p, ok := value.(*Parser)
			if !ok {
				panic(fmt.Errorf("Parser.checkCoerceType: value %T is not *Parser", value))
			}
			return p
		},
	}
	t.registerProperty("lastComment", "String",
		func(inst Any) Any { return String(inst.(*Parser).LastComment()) },
		nil,
	)
	t.registerNativeMethod("getClearLastComment", func(f *RunFrame, ths *TypeObjPair) {
		f.RP.DataStack.Push(String(ths.Obj.(*Parser).GetClearLastComment()))
	})
	GetTypeRegistrar().RegisterCurrentType(t)
}

// CreateRegisterTokenTypeType registers TokenType as a synthetic sealed type whose class
// properties expose lexer flag constants (name → Int value).
func CreateRegisterTokenTypeType() {
	t := &Type{
		name: "TokenType",
		getZero: func() any {
			return TokenType(0)
		},
		checkCoerceType: func(value any) any {
			switch x := value.(type) {
			case TokenType:
				return x
			case Int:
				return TokenType(x)
			default:
				panic(fmt.Errorf("TokenType.checkCoerceType: value %T is not TokenType", value))
			}
		},
	}
	registerTokenTypeClassConstants(t)
	GetTypeRegistrar().RegisterCurrentType(t)
}

func registerTokenTypeClassConstants(t *Type) {
	for _, c := range []struct {
		name string
		val  TokenType
	}{
		{"TokenEOF", TokenEOF},
		{"TokenIdent", TokenIdent},
		{"TokenCompoundIdent", TokenCompoundIdent},
		{"TokenInt", TokenInt},
		{"TokenUint", TokenUint},
		{"TokenFloat", TokenFloat},
		{"TokenBool", TokenBool},
		{"TokenString", TokenString},
		{"TokenRawString", TokenRawString},
		{"TokenRune", TokenRune},
		{"TokenLiteral", TokenLiteral},
		{"TokenNull", TokenNull},
		{"TokenComment", TokenComment},
		{"TokenNewline", TokenNewline},
	} {
		v := Int(c.val)
		name := c.name
		t.registerClassProperty(name, "Int", func(_ Any) Any {
			return v
		}, nil)
	}
}

func init() {
	CreateRegisterParserType()
	CreateRegisterTokenTypeType()
}
