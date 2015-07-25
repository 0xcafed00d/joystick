package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/simulatedsimian/joystick"
	"time"
)

func printAt(x, y int, s string) {
	for _, r := range s {
		termbox.SetCell(x, y, r, termbox.ColorDefault, termbox.ColorDefault)
		x++
	}
}

func readJoystick(js joystick.Joystick) error {
	jinfo := js.Read()

	printAt(1, 5, "Buttons:")
	for button := 0; button < js.ButtonCount(); button++ {
		if jinfo.Buttons&(1<<uint32(button)) != 0 {
			printAt(10+button, 5, "X")
		} else {
			printAt(10+button, 5, ".")
		}
	}

	for axis := 0; axis < js.AxisCount(); axis++ {
		printAt(1, axis+7, fmt.Sprintf("Axis %2d Value: %7d", axis, jinfo.AxisData[axis]))
	}

	return nil
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	js, jserr := joystick.Open(0)

	if jserr != nil {
		fmt.Println(jserr)
		return
	}

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 40)

	for doQuit := false; !doQuit; {
		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				if ev.Ch == 'q' {
					doQuit = true
				}
			}
			if ev.Type == termbox.EventResize {
				termbox.Flush()
			}

		case <-ticker.C:
			printAt(1, 0, "-- Press 'q' to Exit --")
			printAt(1, 1, fmt.Sprintf("Joystick Name: %s", js.Name()))
			printAt(1, 2, fmt.Sprintf("   Axis Count: %d", js.AxisCount()))
			printAt(1, 3, fmt.Sprintf(" Button Count: %d", js.ButtonCount()))
			readJoystick(js)
			termbox.Flush()
		}
	}
}
