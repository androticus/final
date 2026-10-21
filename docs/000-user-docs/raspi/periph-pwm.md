# Pulse Width Modulation

Pulse WIdth Modulation (PWM) is a technique that involves changing the duty dycle of a pulse stream to vary things like the average power delivered to a load, or to operate a servo.

PWM is provided by the type PwmPeriph. Final exposes two header channels, `pwm0` and `pwm1` (BCM 12 and 13). Enable them with:

```
dtoverlay=pwm-2chan,pin=12,pin2=13,func=4,func2=4
```

`func=4` / `func2=4` select Alt0 on BCM 12/13 (PWM0/PWM1). The overlay’s defaults are Alt5, which are for pins 18/19 — without `func=4`, pinmux may still show `pwm` while the header pins do not output PWM. See [Getting Ready](getting-ready.md).

Final provides three distinct operating modes for its PwmPeriphs:
* Raw -- you control the pulse interval and on-time manually, in nanoseconds;
* Linear -- used to provide a linear variation in duty cycle from 0 to 1 for the signal whose frequency you specify.
* Servo -- used for controlling hobby-type servos, that have a defined signal characteristic; you can control the servo on a -1..+1 range, where 0 is the null position; or angle from 0 to 180, where 90 is the null position

Both Linear and Servo modes ultimately render your parameters in raw form, but provide a more convenient and less error prone interface.

Here are the parameters you specify, for each mode. Note, that the modal parameters for a mode may only be set if you have selected that mode. This is to prevent conflicting parameter usage. You set/read the properties for each mode on a subobject with the mode name.

Example property accessors, by mode:

```final
  mypwm.raw.periodNs

  mpwn.linear.freq

  mypwm.servo.byRange
```

## PwmPeriph base parameters

Parameter | Values | Description
--- | --- | ---
mode | PwmPeriph.raw, linear, servo | select the mode

## Raw mode

Raw mode is always an available choice, if you want a custom control profile. This mode corresponds to the parameters accepted by the underlying Pi peripheral. The other modes simply layer a more convenient use-case interface upon these raw values.

Parameter | Values | Description
--- | --- | ---
periodNs | >0 | the overall pulse width in nanoseconds
dutyNs | < periodNs | the width of the On portion

## Linear mode

Linear mode is useful for things like controlling the brightness of LEDs or the power of a heater or other load, using a pulse whose On duty cycle will be the average power delivered into the load. The frequency will typically be on the order of 1 kHz for LEDs, but may be much less for powered loads like heaters. Your hardware and load will determine the appropriate frequency.

Parameter | Values | Description
--- | --- | ---
freq | > 0 | the frequency of the pulse in Hz, eg 1000
duty | 0.0 to 1.0 | the `on` portion of the pulse, def = 0

## Servo mode

Servo mode is specifically for the common PWM case of controlling hobby servos, linear actuators, and such. They have a very particular pulse profile, whose design goes back many years to early analog electronics and radios. Today's devices are digital, but the control system has remained the same, with occasional variations. Make sure you initialize the PwmPeriph to correspond to your device's actual spec. The peripheral will start up with the control pulse in the neutral/zero position for the actuator, usually what you want. If the device should start up at a different value, set that immediately after setting `.enabled` true.

Parameter | Values | Description
--- | --- | ---
freq | > 0 | the frequency, default is 50 (for hobby servos)
minUs | > 0 | the minimum pulse width in microseconds, def is 1000 (1 ms)
maxUs | > minUs | the maximum pulse width, def is 2000 (2 ms)
setMinMax | min, max | setter, for both at once
asAngle | 0 to 180 | sets the duty cycle in terms of angle (common for servos), def 90 
asRange | -1.0 to 1.0 | sets the duty cycle as a range around 0, def 0 (== 90 degrees)

`asAngle` and `asRange` are convenience properties, so you can set/read values in units of your preference; they both have the same effect on the pulse width.
