package finalpi

import (
	"github.com/androticus/final/finalembed"
	"github.com/androticus/final/internal/core"
)

// init registers Pi HAL types on the global/core catalog and arms bootstrap tranche4
// so gpio.fnl compiles into sealed core before InitGlobalExecutionEnvironment (via NewSession).
func init() {
	initPiSupport() // before types / gpio.fnl; not deferred to newGpioCtrl
	registerCoreHostType[PiSupportType]("PiSupport", registerPiSupport)
	registerCoreHostType[PwmMode]("PwmMode", registerPwmMode)
	registerCoreHostType[*PwmRaw]("PwmRaw", registerPwmRaw)
	registerCoreHostType[*PwmLinear]("PwmLinear", registerPwmLinear)
	registerCoreHostType[*PwmServo]("PwmServo", registerPwmServo)
	registerCoreHostType[Peripheral]("Peripheral", registerPeripheral)
	registerCoreHostType[*PwmPeriph]("PwmPeriph", registerPwmPeriph)
	registerCoreHostType[*UartPeriph]("UartPeriph", registerUartPeriph)
	registerCoreHostType[*UartReader]("UartReader", registerUartReader)
	registerCoreHostType[*I2cPeriph]("I2cPeriph", registerI2cPeriph)
	registerCoreHostType[*SpiPeriph]("SpiPeriph", registerSpiPeriph)
	registerCoreHostType[*I2sPeriph]("I2sPeriph", registerI2sPeriph)
	registerCoreHostType[*GpioPin]("GpioPin", registerGpioPin)
	registerCoreHostType[*GpioCtrl]("GpioCtrl", registerGpioCtrl)
	core.EnableBootstrapTranche4()
}

func registerCoreHostType[T any](name string, reg func(t *finalembed.Type)) {
	t := core.NewRegisteringType(name,
		core.MakeHostTypeZero[T](),
		core.MakeHostTypeCheckCoerce[T](name),
	)
	reg(t)
	core.GetTypeRegistrar().RegisterCurrentType(t)
	core.RegisterHostType[T](t)
}
