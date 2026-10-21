forget: reset-mpu9250
// NOTE modules generally should not have a forget TODO: remove this
requiresPi
/*
MPU-9250 + AK8963 over I2cPeriph.

One breakout, two I²C devices: accel/gyro die at 0x68, magnetometer
(AK8963) on an internal aux bus. Linux only sees /dev/i2c-N; "bypass"
and "I2C master off" are MPU register bits that mux the aux bus onto
the host SDA/SCL so the AK8963 appears at 0x0C.

Register map:
  docs/009-datasheets/MPU-9250-Register-Map.pdf
*/

{
  "Mpu9250" "name"
  {
    I2cPeriph "bus"
    Int "addr"
  } "fields"
  {
    0x68 "defaultAddr"
    0x75 "regWhoAmI"
    0x71 "whoAmIExpect"
    0x6B "regPwrMgmt1"
    0x6A "regUserCtrl"
    0x37 "regIntPinCfg"
    0x02 "bypassEn"          // INT_PIN_CFG BYPASS_EN
    0x3B "regAccelXoutH"
    0x43 "regGyroXoutH"
  } "constants"
}
TypedObject.makeType^

{
  "Ak8963" "name"
  {
    I2cPeriph "bus"
    Int "addr"
  } "fields"
  {
    0x0C "defaultAddr"
    0x00 "regWia"
    0x48 "whoAmIExpect"
    0x02 "regSt1"
    0x03 "regHxl"
    0x0A "regCntl1"
    0x00 "cntlPowerDown"
    0x16 "cntlCont16bit100Hz"  // BIT=1, MODE=0110 continuous-2
  } "constants"
}
TypedObject.makeType^

/*
Big-endian int16 from ByteList at idx (and idx+1).
( bl:ByteList idx:i -- v:i )
*/
: Mpu9250_be16At
  swap >ts                 // idx; bl→ts
  dup t:@() >ys            // hi on ys
  1 + t:@()                // lo
  ys> 8 << |
  dup 32767 > if 65536 - then
  tdrop
;

/*
Write one register.
( reg:i v:i -- ) o( mpu:Mpu9250 )
*/
: Mpu9250#writeReg
  >ys                      // v
  ByteList.new^ >xs
  x:.append^               // reg
  ys> x:.append^           // v
  o:.bus@ >ts
  o:.addr@ xs> t:.write^
  tdrop
;

/*
Read one register (write reg index, repeated-start read 1).
( reg:i -- v:i ) o( mpu:Mpu9250 )
*/
: Mpu9250#readReg
  ByteList.new^ >xs
  x:.append^
  o:.bus@ >ts
  o:.addr@ xs> 1 t:.tx^ >ys
  0 y:@()
  y:drop
  tdrop
;

/*
Burst-read n bytes starting at reg (auto-increment).
( reg:i n:i -- bl:ByteList ) o( mpu:Mpu9250 )
*/
: Mpu9250#readRegs
  >ys                      // n
  ByteList.new^ >xs
  x:.append^               // reg
  o:.bus@ >ts
  o:.addr@ xs> ys> t:.tx^
  tdrop
;

/*
( -- id:i ) o( mpu:Mpu9250 )
*/
: Mpu9250#whoAmI
  Mpu9250.regWhoAmI o:.readReg^
;

/*
Clear sleep bit (PWR_MGMT_1 ← 0).
( -- ) o( mpu:Mpu9250 )
*/
: Mpu9250#wake
  Mpu9250.regPwrMgmt1 0 o:.writeReg^
;

/*
Open the aux I²C mux so the host can talk to AK8963 at 0x0C.
Disables MPU I2C master (USER_CTRL ← 0), then sets BYPASS_EN.
These are chip registers — not Pi/Linux HAL modes.
( -- ) o( mpu:Mpu9250 )
*/
: Mpu9250#enableMagBypass
  Mpu9250.regUserCtrl 0 o:.writeReg^
  Mpu9250.regIntPinCfg Mpu9250.bypassEn o:.writeReg^
;

/*
Raw accelerometer counts (not g).
( -- p:Point3 ) o( mpu:Mpu9250 )
*/
: Mpu9250#readAccelRaw
  Mpu9250.regAccelXoutH 6 o:.readRegs^ >xs
  x:peek 0 Mpu9250_be16At
  x:peek 2 Mpu9250_be16At
  x:peek 4 Mpu9250_be16At
  Point3.new^
  x:drop
;

/*
Raw gyro counts (not dps).
( -- p:Point3 ) o( mpu:Mpu9250 )
*/
: Mpu9250#readGyroRaw
  Mpu9250.regGyroXoutH 6 o:.readRegs^ >xs
  x:peek 0 Mpu9250_be16At
  x:peek 2 Mpu9250_be16At
  x:peek 4 Mpu9250_be16At
  Point3.new^
  x:drop
;

/*
Mint bound to a bus (default addr).
( bus:I2cPeriph -- mpu:Mpu9250 ) o( type:Mpu9250 )
*/
: Mpu9250##onBus
  o:.new^ >ts
  t:.bus!
  Mpu9250.defaultAddr t:.addr!
  ts>
;

// --- AK8963 magnetometer (second I²C address after bypass) ---

/*
Little-endian int16 from ByteList at idx (and idx+1).
( bl:ByteList idx:i -- v:i )
*/
: Ak8963_le16At
  swap >ts                 // idx; bl→ts
  dup t:@() >ys            // lo on ys
  1 + t:@()                // hi
  8 << ys> |
  dup 32767 > if 65536 - then
  tdrop
;

/*
Write one register.
( reg:i v:i -- ) o( mag:Ak8963 )
*/
: Ak8963#writeReg
  >ys
  ByteList.new^ >xs
  x:.append^
  ys> x:.append^
  o:.bus@ >ts
  o:.addr@ xs> t:.write^
  tdrop
;

/*
Read one register.
( reg:i -- v:i ) o( mag:Ak8963 )
*/
: Ak8963#readReg
  ByteList.new^ >xs
  x:.append^
  o:.bus@ >ts
  o:.addr@ xs> 1 t:.tx^ >ys
  0 y:@()
  y:drop
  tdrop
;

/*
Burst-read n bytes starting at reg.
( reg:i n:i -- bl:ByteList ) o( mag:Ak8963 )
*/
: Ak8963#readRegs
  >ys
  ByteList.new^ >xs
  x:.append^
  o:.bus@ >ts
  o:.addr@ xs> ys> t:.tx^
  tdrop
;

/*
( -- id:i ) o( mag:Ak8963 )
*/
: Ak8963#whoAmI
  Ak8963.regWia o:.readReg^
;

/*
Power-down then continuous 100 Hz / 16-bit.
Call only after Mpu9250.enableMagBypass.
( -- ) o( mag:Ak8963 )
*/
: Ak8963#startContinuous
  Ak8963.regCntl1 Ak8963.cntlPowerDown o:.writeReg^
  10 delayMs
  Ak8963.regCntl1 Ak8963.cntlCont16bit100Hz o:.writeReg^
  10 delayMs
;

/*
Raw mag counts (little-endian; not µT). Reads HXL..ST2 (7 bytes).
( -- p:Point3 ) o( mag:Ak8963 )
*/
: Ak8963#readMagRaw
  Ak8963.regHxl 7 o:.readRegs^ >xs
  x:peek 0 Ak8963_le16At
  x:peek 2 Ak8963_le16At
  x:peek 4 Ak8963_le16At
  Point3.new^
  x:drop
;

/*
Mint bound to a bus (default addr 0x0C). Requires bypass already on.
( bus:I2cPeriph -- mag:Ak8963 ) o( type:Ak8963 )
*/
: Ak8963##onBus
  o:.new^ >ts
  t:.bus!
  Ak8963.defaultAddr t:.addr!
  ts>
;
