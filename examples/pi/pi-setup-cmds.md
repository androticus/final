# Raspberry Pi setup commands to enable hardware

Edit `/boot/firmware/config.txt` (reboot after changes). Prefer header defaults (`i2c` / `spi` / `uart` / `pwm0`/`pwm1`). See [Getting Ready](../../docs/000-user-docs/raspi/getting-ready.md) and [Z2 Lite smoke](../../docs/000-user-docs/raspi/z2-lite-smoke.md).

## Pi 5

Groups (log out / in after):

```bash
sudo usermod -aG gpio,i2c,spi,dialout "$USER"
groups   # confirm
```

Append to `/boot/firmware/config.txt` (defaults + optional Pi 5–only alts):

```bash
sudo tee -a /boot/firmware/config.txt >/dev/null <<'EOF'
# Final — header defaults (i2c / spi)
dtparam=i2c_arm=on
#dtparam=i2s=on
dtparam=spi=on

# Final — header PWM on BCM 12/13 (pwm0/pwm1); func=4 = Alt0 (required on 12/13)
dtoverlay=pwm-2chan,pin=12,pin2=13,func=4,func2=4

# Final — header UART
dtoverlay=uart0-pi5
# optional Pi5-only alt for loopback tests:
# dtoverlay=uart2-pi5

# I2C alts — Pi 5 only; only include ones you use
# dtoverlay=i2c2-pi5
# dtoverlay=i2c3-pi5

# SPI alts — Pi 5 only (spi1-Ncs / spi5-1cs|2cs)
# dtoverlay=spi1-1cs
# dtoverlay=spi5-1cs
EOF
sudo reboot
```

Uncomment only the alt overlays you need. Leave `i2s` commented unless you use an I2S hat; if you enable it, set `i2s.reserved` true in Final so reset does not clobber those pins.

## Pi Zero 2W

Groups (log out / in after):

```bash
sudo usermod -aG gpio,i2c,spi,dialout "$USER"
groups   # confirm
```

Append to `/boot/firmware/config.txt` (defaults only — no `*-pi5` alt overlays):

```bash
sudo tee -a /boot/firmware/config.txt >/dev/null <<'EOF'
# Final — header defaults (i2c / spi)
dtparam=i2c_arm=on
dtparam=spi=on

# Final — header PWM on BCM 12/13 (pwm0/pwm1); func=4 = Alt0 (required on 12/13)
dtoverlay=pwm-2chan,pin=12,pin2=13,func=4,func2=4

# Final — header UART (Zero 2)
enable_uart=1
EOF
sudo reboot
```

Final on Zero 2 defaults `uart` to `/dev/ttyS0`. Only set `FINAL_UART0` if you remapped the header UART (unusual).
