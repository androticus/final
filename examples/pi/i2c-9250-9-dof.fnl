forget: reset-i2c-9250-demo
requiresPi

// install and load the mpu9250 driver module (Mpu9250 / Ak8963)
"github.com/androticus/final/modules/pi/drivers/mpu9250"
{
  "modules/pi/drivers/mpu9250" "path"
}
install

"" "github.com/androticus/final/modules/pi/drivers/mpu9250" import

`
I2C demonstration using the MPU-9250 9-DoF
accelerometer/gyro/magnetometer sensor.

One breakout, two chips on I²C: MPU at 0x68 (accel/gyro),
AK8963 mag at 0x0C after the MPU's bypass mux is opened.

Wiring (default i2c / I2C1):
  SDA GPIO2
  SCL GPIO3
  AD0 low → address 0x68

Then:
  switch to View:StdOut/StdErr
  
  i2c-9250-demo
` const: i2c-9250-demo-blurb

{
  "I2c9250Demo" "name"
  {
    Mpu9250 "mpu"
    Ak8963 "mag"
    GpioPin "btnRead"
    EventEscFn "escFn"
    Bool "stop"
  } "fields"
}
TypedObject.makeType^

// instance to prevent GC
null var: I2c9250Demo_inst

/*
Print Point3 components (host $ only shows the type name).
( p:Point3 -- )
*/
: I2c9250Demo_printP3
  >ts
  "(" $ t:.x@ $ ", " $ t:.y@ $ ", " $ t:.z@ $ ")" $
  tdrop
;

/*
Factored worker: sample and print raw accel + gyro + mag.
( -- )  o( inst:I2c9250Demo )
*/
: I2c9250Demo_read
  "accel " $ o:.mpu.readAccelRaw^ I2c9250Demo_printP3 nl
  "gyro  " $ o:.mpu.readGyroRaw^ I2c9250Demo_printP3 nl
  "mag   " $ o:.mag.readMagRaw^ I2c9250Demo_printP3 nl
;

/*
Take a new set of readings.
( -- )  o( inst:I2c9250Demo )
*/
: I2c9250Demo#read
  o:.stop nif
    I2c9250Demo_read
  then
;

/*
( ev:Event -- )  o( inst:I2c9250Demo )
*/
: I2c9250Demo#handleGpioLow
  drop
  I2c9250Demo_read
;

/*
( index:i gpio -- ) o( inst:I2c9250Demo )
*/
: I2c9250Demo_initBtn
  >ts
  GpioPin.pullUp t:.pull!
  GpioPin.in t:.direction!
  GpioPin.debounceButton t:.debounceMs!
  // we only need the low-going btn handler
  o:.handleGpioLow~ swap t:.setOnLow^
  true t:.enabled!
  tdrop
;

/*
Stops the demo (Fn keys ignored).
( ev:Event -- ) o( inst:I2cSsd1306Demo )
*/
: I2c9250Demo_doStop
  true o:.stop!
  GpioCtrl.inst.reset^
  null I2c9250Demo_inst ! // enables us to be GCed
  "Demo stopped.\n" $
;

/*
Esc stops the demo (Fn keys ignored).
( code:i -- ) o( inst:I2c9250Demo )
*/
: I2c9250Demo#handleEscFn
  ==0 if
    I2c9250Demo_doStop
  then
;

/*
Initialize bus, MPU-9250, AK8963 (via bypass), and sample button.
( -- ) o( inst:I2c9250Demo )
*/
: I2c9250Demo#init
  :s initPi
    o:.handleEscFn~ EventEscFn.new^ o:.escFn!
    o:.escFn.start^
    true i2c.enabled!
    i2c Mpu9250.onBus^ o:.mpu!
    o:.mpu.wake^
    "MPU WHO_AM_I=" $ o:.mpu.whoAmI^ $
    " (expect " $ Mpu9250.whoAmIExpect $ ")\n" $
    o:.mpu.enableMagBypass^
    i2c Ak8963.onBus^ o:.mag!
    o:.mag.startContinuous^
    "AK8963 WIA=" $ o:.mag.whoAmI^ $
    " (expect " $ Ak8963.whoAmIExpect $ ")\n" $
    gpio26 dup o:.btnRead!
    0 swap I2c9250Demo_initBtn
  ;s

  GpioCtrl.inst.reset^
  initPi
  "MPU-9250 + AK8963 ready; press gpio26 btn to sample.\n" $
  "Press Esc to stop demo.\n" $
;


: i2c-9250-demo
  I2c9250Demo.new^ dup >ts
  I2c9250Demo_inst !
  t:.init^
  tdrop
;

i2c-9250-demo-blurb $ nl
