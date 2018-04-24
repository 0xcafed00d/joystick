// +build !linux,!windows,!darwin

package joystick

import (
	"errors"
)

// Open opens the Joystick for reading, with the supplied id
//
// Under linux the id is used to construct the joystick device name:
//   for example: id 0 will open device: "/dev/input/js0"
// Under Windows the id is the actual numeric id of the joystick
//
// If successful, a Joystick interface is returned which can be used to
// read the state of the joystick, else an error is returned
func Open(id int) (Joystick, error) {
	return nil, errors.New("Joystick API unsupported on this platform")
}
