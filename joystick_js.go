//go:build js && wasm && !linux && !windows && !darwin

package joystick

import (
	"fmt"
	"syscall/js"
)

var navigator js.Value

func init() {
	navigator = js.Global().Get("navigator")
	if !navigator.Truthy() {
		panic("Gamepad API unsupported on this platform")
	}
}

type gamePad struct {
	id int
}

func (g *gamePad) AxisCount() int {
	gamepad := navigator.Call("getGamepads").Index(g.id)
	if !gamepad.Truthy() {
		return 0
	}
	axes := gamepad.Get("axes")
	return axes.Length()
}

func (g *gamePad) ButtonCount() int {
	gamepad := navigator.Call("getGamepads").Index(g.id)
	if !gamepad.Truthy() {
		return 0
	}
	buttons := gamepad.Get("buttons")
	return buttons.Length()
}

func (g *gamePad) Close() {}

func (g *gamePad) Name() string {
	gamepad := navigator.Call("getGamepads").Index(g.id)
	if !gamepad.Truthy() {
		return ""
	}
	return gamepad.Get("id").String()
}

func (g *gamePad) Read() (State, error) {
	gamepad := navigator.Call("getGamepads").Index(g.id)
	if !gamepad.Truthy() {
		return State{}, fmt.Errorf("gamepad not found")
	}
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
		state.Buttons |= uint32(buttons.Index(i).Get("value").Int() << i)
	}

	return state, nil
}

func Open(id int) (Joystick, error) {
	gamepad := navigator.Call("getGamepads").Index(id)
	if !gamepad.Truthy() {
		return nil, fmt.Errorf("gamepad not found")
	}
	return &gamePad{id: id}, nil
}

func scaleAxis(value float64) int {
	return int(-32767 + (value+1.0)*(32768+32767)/(1.0-(-1.0)))
}
