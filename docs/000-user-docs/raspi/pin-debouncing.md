# Debouncing

Most mechanical switches and some sensors do not open or close cleanly. Instead, there is a small amount of mechanical chatter that results in short on/off spikes at the time of actuation. If you were counting the number of times a button was pushed, and went solely by the high-speed digital value of the input pin, your count could be wildly off, as it accumulated each chatter noise spike as a valid push.

To correct for this, we use `debouncing`. Final provides the option to add debouncing to any GPIO input. Note, this capability is restricted to the situation of use of the pin as a GPIO pin, **not** when the pin is part of a peripheral.

The approach Final takes, is a two-delay approach. At the onset of a change in level, Final waits a few milliseconds, and samples the pin again. If it is still at the same level, it will pass that thru to the `value` property. It then waits the full `debounceMs` delay before it will consider a change of the pin in the other direction.

Final provides some helpful named constants for debounce delays that will probably work well as follows:

GpioPin. | Delay (ms) | When to use
--- | --- | ---
debounceOff | 0 | disable debouncing, all digital inputs
debounceSensor | 5 | an electronic sensor, eg hall effect, photo interrupter
debounceButton | 15 | push buttons
debounceSwitch | 30 | slider switches

In practice, you may need to monitor the level change events for spurious multiples, and adjust the time accordingly.
