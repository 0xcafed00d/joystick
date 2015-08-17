// Package joystick implements a Polled API to read the state of an attached joystick.
// currently Windows & Linux are supported.
package joystick

// State holds the current state of the joystick
type State struct {
	AxisData []int  // Value of each axis as an integer in the range -32767 to 32768
	Buttons  uint32 // The state of each button as a bit in a 32 bit integer. 1 = pressed, 0 = not pressed
}

type Joystick interface {
	AxisCount() int
	ButtonCount() int
	Name() string
	Read() (State, error)
	Close()
}
