package joystick

type JoystickInfo struct {
	AxisData []int
	Buttons  uint32
}

type Joystick interface {
	AxisCount() int
	ButtonCount() int
	Name() string
	Read() JoystickInfo
	Close()
}
