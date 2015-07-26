package joystick

type State struct {
	AxisData []int
	Buttons  uint32
}

type Joystick interface {
	AxisCount() int
	ButtonCount() int
	Name() string
	Read() (State, error)
	Close()
}
