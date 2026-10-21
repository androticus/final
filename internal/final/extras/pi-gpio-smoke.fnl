// Pi GPIO soft smoke — GpioCtrl foundation (no chip I/O). Safe on any host with finalpi.

: run-pi-gpio-soft-smoke
  GpioCtrl.inst.reset^

  gpio12 .bcm@ 12 <> if
    "SMOKE_FAIL pi-gpio: gpio12.bcm\n" $$
    1 0 /
  then
  pwm0 .bcm@ 12 <> if
    "SMOKE_FAIL pi-gpio: pwm0.bcm\n" $$
    1 0 /
  then

  gpio4 >ts
  GpioPin.out t:.direction!
  true t:.value!
  true t:.enabled!
  t:.value@ 0= if
    "SMOKE_FAIL pi-gpio: gpio4.value\n" $$
    1 0 /
  then
  ts>

  GpioCtrl.inst.reset^
  "SMOKE_OK pi-gpio-soft\n" $
;
run-pi-gpio-soft-smoke
