//go:build linux

package finalpi

import (
	"os"
	"testing"
)

// Hardware smoke mirroring modules/pi/drivers/mpu9250 register protocol.
func TestMpu9250WhoAmIAndRawSample(t *testing.T) {
	if _, err := os.Stat("/dev/i2c-1"); err != nil {
		t.Skip("no /dev/i2c-1")
	}
	c := GpioCtrlInst()
	c.Reset()
	defer c.Reset()
	if err := c.i2c1.SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	const (
		addr       = 0x68
		regWhoAmI  = 0x75
		expectID   = 0x71
		regPwrMgmt = 0x6B
		regAccel   = 0x3B
		regGyro    = 0x43
	)
	// wake
	if err := c.i2c1.Write(addr, []byte{regPwrMgmt, 0}); err != nil {
		t.Fatalf("wake: %v (is MPU-9250 on i2c1 @0x68?)", err)
	}
	id, err := c.i2c1.Tx(addr, []byte{regWhoAmI}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 1 || id[0] != expectID {
		t.Fatalf("WHO_AM_I=%v want 0x%02x", id, expectID)
	}
	accel, err := c.i2c1.Tx(addr, []byte{regAccel}, 6)
	if err != nil {
		t.Fatal(err)
	}
	if len(accel) != 6 {
		t.Fatalf("accel len=%d", len(accel))
	}
	gyro, err := c.i2c1.Tx(addr, []byte{regGyro}, 6)
	if err != nil {
		t.Fatal(err)
	}
	if len(gyro) != 6 {
		t.Fatalf("gyro len=%d", len(gyro))
	}
	t.Logf("WHO_AM_I=0x%02x accel=%v gyro=%v", id[0], accel, gyro)
}
