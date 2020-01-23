# joystick
Go Joystick API

[![GoDoc](https://godoc.org/github.com/0xcafed00d/joystick?status.svg)](https://godoc.org/github.com/0xcafed00d/joystick) [![Build Status](https://travis-ci.org/0xcafed00d/joystick.svg)](https://travis-ci.org/0xcafed00d/joystick)

Package joystick implements a Polled API to read the state of an attached joystick.
Windows, Linux & OSX are supported.
Package requires no external dependencies to be installed.

Mac OSX code developed by:  https://github.com/ledyba

## Installation:
```bash
$ go get github.com/0xcafed00d/joystick/...
```
## Sample Program 
```bash
$ go install github.com/0xcafed00d/joystick/joysticktest
$ joysticktest 0
```
Displays the state of the specified joystick
## Example:
```go
import "github.com/0xcafed00d/joystick"
```
```go
js, err := joystick.Open(jsid)
if err != nil {
  panic(err)
}

fmt.Printf("Joystick Name: %s", js.Name())
fmt.Printf("   Axis Count: %d", js.AxisCount())
fmt.Printf(" Button Count: %d", js.ButtonCount())

state, err := joystick.Read()
if err != nil {
  panic(err)
}

fmt.Printf("Axis Data: %v", state.AxisData)
js.Close()
```
