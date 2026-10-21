package core

// tranche4Enabled is armed by EnableBootstrapTranche4 (finalpi package init)
// before InitGlobalExecutionEnvironment, so RunBootstrapInit compiles core4.
var tranche4Enabled = false

// EnableBootstrapTranche4 arms core4 (Pi HAL Final sources) for InitGlobalExecutionEnvironment.
// Call from finalpi init() before that Init runs. No-op if already enabled.
func EnableBootstrapTranche4() {
	tranche4Enabled = true
}

const bootstrapTranche4Manifest = `
// Pi HAL foundation (compiled only when tranche4 is enabled / finalpi linked)
internal/final/pi/gpio.fnl
internal/final/pi/utils.fnl
`
