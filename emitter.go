package joystick

import "time"

type PadEvent struct {
	State State
}

func NewEmitter(id int, rate time.Duration) (chan PadEvent, chan error) {
	errc := make(chan error)
	pevc := make(chan PadEvent)

	go func() {
		joy, err := Open(id)
		if err != nil {
			errc <- err
			return
		}

		lastPressTime := time.Now().Add(-1 * rate)
		lastPress := []ButtonID{}

		lb := false
		for {
			st, err := joy.Read()
			if err != nil {
				errc <- err
				joy.Close()
				return
			}

			prb := len(st.Pressed())

			if !lb && prb == 0 {
				lb = true
			}

			if lb && prb == 0 {
				continue
			}

			if lb && prb != 0 {
				lb = false
			}

			cp := st.Pressed()
			if testEqBtn(cp, lastPress) {
				if time.Since(lastPressTime) >= rate {
					pevc <- PadEvent{st}
					lastPressTime = time.Now()
				}
			} else {
				lastPress = cp
				lastPressTime = time.Now()
				pevc <- PadEvent{st}
			}
		}
	}()

	return pevc, errc
}

func testEqBtn(a, b []ButtonID) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
