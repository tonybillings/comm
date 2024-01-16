package rfcomm

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"tonysoft.com/comm/pkg/comerr"
)

func MacStringToByteArray(address string) ([6]byte, error) {
	addrArr := strings.Split(address, ":")
	var result [6]byte

	for i, tmp := range addrArr {
		addrOct, err := strconv.ParseUint(tmp, 16, 8)
		if err == nil {
			result[len(result)-1-i] = byte(addrOct)
		} else {
			return result, fmt.Errorf("%w : %v", comerr.ErrParseMacAddress, err)
		}
	}

	return result, nil
}

func ByteArrayToMacString(mac [6]uint8) (string, error) {
	str := ""
	for _, b := range mac {
		str += fmt.Sprintf("%x:", b)
	}

	return string(([]byte(str))[:len(str)-1]), nil
}

// RestartBluetooth Note that this requires root privileges!
func RestartBluetooth() {
	// Using rfkill to block/unblock (reset) the adapter should not be necessary,
	// but it helps ensure it will be freed up and available for reuse.
	cmd := exec.Command("rfkill", "block", "bluetooth")
	_ = cmd.Run()
	time.Sleep(time.Second)
	cmd = exec.Command("rfkill", "unblock", "bluetooth")
	_ = cmd.Run()
	time.Sleep(time.Second)
	cmd = exec.Command("service", "bluetooth", "restart")
	_ = cmd.Run()
	time.Sleep(time.Second)
}

// BecomeDiscoverable Note that this requires root privileges!
func BecomeDiscoverable() {
	cmd := exec.Command("bluetoothctl", "discoverable", "on")
	_ = cmd.Run()
}
