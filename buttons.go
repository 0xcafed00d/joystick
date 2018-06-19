package joystick

type ButtonID uint32

//go:generate stringer -type=ButtonID
const (
	A_Button ButtonID = iota
	B_Button
	X_Button
	Y_Button
	LeftBumper
	RightBumper
	BackButton
	StartButton
	LeftStick
	RightStick
	DPadUp
	DPadDown
	DPadLeft
	DPadRight
)

// Pressed returns all pressed buttons on this State
func (s State) Pressed() []ButtonID {
	flags := []ButtonID{}

	for i := ButtonID(0); i < 32; i++ {
		if s.IsPressed(i) {
			flags = append(flags, ButtonID(i))
		}
	}

	return flags
}

// Pressed returns whether or not a certain button ID is pressed.
func (s State) IsPressed(b ButtonID) bool {
	if b >= DPadUp && b <= DPadRight {
		if b == DPadUp {
			return s.AxisData[DPad_Y] < 0
		}

		if b == DPadDown {
			return s.AxisData[DPad_Y] > 0
		}

		if b == DPadLeft {
			return s.AxisData[DPad_X] < 0
		}

		if b == DPadRight {
			return s.AxisData[DPad_X] > 0
		}
	}

	return s.Buttons&(1<<b) != 0
}

func (s State) GetAxis(a AxisID) int {
	if a == RightTrigger {
		e := 0 - s.AxisData[Trigger]
		if e < 0 {
			return 0
		}
		return e
	}

	if a == LeftTrigger {
		e := s.AxisData[Trigger]
		if e < 0 {
			return 0
		}
		return e
	}

	return s.AxisData[a]
}

type AxisID int

const (
	LeftStick_X AxisID = iota
	LeftStick_Y
	Trigger
	RightStick_X
	RightStick_Y
	DPad_X
	DPad_Y
	RightTrigger
	LeftTrigger
)
