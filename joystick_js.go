//go:build js && !linux && !windows && !darwin

package joystick

import (
	"fmt"
	"syscall/js"
)

type gamePad struct {
	id int
}

func (g *gamePad) AxisCount() (count int) {
	gamepad, err := getGamepad(g.id)
	if err != nil {
		return 0
	}
	defer (func() {
		if r := recover(); r != nil {
			_, ok := r.(js.Error)
			if ok {
				count = 0
			}
		}
	})()
	return gamepad.Get("axes").Length()
}

func (g *gamePad) ButtonCount() (count int) {
	gamepad, err := getGamepad(g.id)
	if err != nil {
		return 0
	}
	defer (func() {
		if r := recover(); r != nil {
			count = 0
		}
	})()
	return gamepad.Get("buttons").Length()
}

func (g *gamePad) Close() {}

func (g *gamePad) Name() (name string) {
	gamepad, err := getGamepad(g.id)
	if err != nil {
		return ""
	}
	defer (func() {
		if r := recover(); r != nil {
			name = ""
		}
	})()
	return gamepad.Get("id").String()
}

func (g *gamePad) Read() (State, error) {
	gamepad, err := getGamepad(g.id)
	if err != nil {
		return State{}, err
	}
	defer (func() {
		if r := recover(); r != nil {
			jsErr, ok := r.(js.Error)
			if ok {
				err = jsErr
			}
		}
	})()
	axes := gamepad.Get("axes")
	buttons := gamepad.Get("buttons")

	state := State{
		AxisData: make([]int, axes.Length()),
		Buttons:  0,
	}

	for i := 0; i < axes.Length(); i++ {
		state.AxisData[i] = scaleAxis(axes.Index(i).Float())
	}

	for i := 0; i < buttons.Length(); i++ {
		if buttons.Index(i).Get("pressed").Bool() {
			state.Buttons |= 1 << i
		}
	}

	return state, nil
}

func Open(id int) (Joystick, error) {
	_, err := getGamepad(id)
	if err != nil {
		return nil, err
	}
	return &gamePad{id: id}, nil
}

func getGamepad(id int) (gp js.Value, err error) {
	defer (func() {
		if r := recover(); r != nil {
			jsErr, ok := r.(js.Error)
			if ok {
				err = jsErr
			}
		}
	})()
	navigator := js.Global().Get("navigator")
	if !navigator.Truthy() {
		return js.Value{}, fmt.Errorf("navigator not found")
	}
	gamepads := navigator.Call("getGamepads")
	if !gamepads.Truthy() {
		return js.Value{}, fmt.Errorf("gamepads not found")
	}
	gamepad := gamepads.Index(id)
	if !gamepad.Truthy() {
		return js.Value{}, fmt.Errorf("gamepad not found")
	}
	return gamepad, nil
}

func scaleAxis(value float64) int {
	return int(-32767 + (value+1.0)*(32768+32767)/(1.0-(-1.0)))
}
