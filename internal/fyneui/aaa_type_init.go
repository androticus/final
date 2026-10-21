package fyneui

// Mixin catalog types must register before other fyneui init() functions
// ApplyMixin them (Go inits files in lexical name order).
func init() {
	CreateRegisterCanvasType()
	CreateRegisterDisableableType()
}
