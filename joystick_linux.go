// +build linux

package joystick

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"sync"
	"unsafe"
)

const (
	_JS_EVENT_BUTTON uint8 = 0x01 /* button pressed/released */
	_JS_EVENT_AXIS   uint8 = 0x02 /* joystick moved */
	_JS_EVENT_INIT   uint8 = 0x80
)

var (
	_JSIOCGAXES    = _IOR('j', 0x11, 1)  /* get number of axes */
	_JSIOCGBUTTONS = _IOR('j', 0x12, 1)  /* get number of buttons */
	_JSIOCGNAME    = func(len int) int { /* get identifier string */
		return _IOR('j', 0x13, len)
	}
)

type joystickImpl struct {
	file        *os.File
	axisCount   int
	buttonCount int
	name        string
	state       State
	mutex       sync.RWMutex
	readerr     error
}

// Open opens the Joystick for reading, with the supplied id
//
// Under linux the id is used to construct the joystick device name:
//   for example: id 0 will open device: "/dev/input/js0"
// Under Windows the id is the actual numeric id of the joystick
//
// If successful, a Joystick interface is returned which can be used to
// read the state of the joystick, else an error is returned
func Open(id int) (Joystick, error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/input/js%d", id), os.O_RDONLY, 0666)

	if err != nil {
		return nil, err
	}

	var axisCount uint8 = 0
	var buttCount uint8 = 0
	var buffer [256]byte

	ioerr := ioctl(f, _JSIOCGAXES, unsafe.Pointer(&axisCount))
	if ioerr != 0 {
		panic(ioerr)
	}

	ioerr = ioctl(f, _JSIOCGBUTTONS, unsafe.Pointer(&buttCount))
	if ioerr != 0 {
		panic(ioerr)
	}

	ioerr = ioctl(f, _JSIOCGNAME(len(buffer)-1), unsafe.Pointer(&buffer))
	if ioerr != 0 {
		panic(ioerr)
	}

	js := &joystickImpl{}
	js.axisCount = int(axisCount)
	js.buttonCount = int(buttCount)
	js.file = f
	js.name = string(buffer[:])
	js.state.AxisData = make([]int, axisCount, axisCount)

	go updateState(js)

	return js, nil
}

func updateState(js *joystickImpl) {
	var err error
	var ev event

	for err == nil {
		ev, err = js.getEvent()

		if ev.Type&_JS_EVENT_BUTTON != 0 {
			js.mutex.Lock()
			if ev.Value == 0 {
				js.state.Buttons &= ^(1 << uint(ev.Number))
			} else {
				js.state.Buttons |= 1 << ev.Number
			}
			js.mutex.Unlock()
		}

		if ev.Type&_JS_EVENT_AXIS != 0 {
			js.mutex.Lock()
			js.state.AxisData[ev.Number] = int(ev.Value)
			js.mutex.Unlock()
		}
	}
	js.mutex.Lock()
	js.readerr = err
	js.mutex.Unlock()
}

func (js *joystickImpl) AxisCount() int {
	return js.axisCount
}

func (js *joystickImpl) ButtonCount() int {
	return js.buttonCount
}

func (js *joystickImpl) Name() string {
	return js.name
}

func (js *joystickImpl) Read() (State, error) {
	js.mutex.RLock()
	state, err := js.state, js.readerr
	js.mutex.RUnlock()
	return state, err
}

func (js *joystickImpl) Close() {
	js.file.Close()
}

type event struct {
	Time   uint32 /* event timestamp in milliseconds */
	Value  int16  /* value */
	Type   uint8  /* event type */
	Number uint8  /* axis/button number */
}

func (j *event) String() string {
	var Type, Number string

	if j.Type&_JS_EVENT_INIT > 0 {
		Type = "Init "
	}
	if j.Type&_JS_EVENT_BUTTON > 0 {
		Type += "Button"
		Number = strconv.FormatUint(uint64(j.Number), 10)
	}
	if j.Type&_JS_EVENT_AXIS > 0 {
		Type = "Axis"
		Number = "Axis " + strconv.FormatUint(uint64(j.Number), 10)
	}

	return fmt.Sprintf("[Time: %v, Type: %v, Number: %v, Value: %v]", j.Time, Type, Number, j.Value)
}

func (j *joystickImpl) getEvent() (event, error) {
	var ev event

	if j.file == nil {
		panic("file is nil")
	}

	b := make([]byte, 8)
	_, err := j.file.Read(b)
	if err != nil {
		return event{}, err
	}

	data := bytes.NewReader(b)
	err = binary.Read(data, binary.LittleEndian, &ev)
	if err != nil {
		return event{}, err
	}
	return ev, nil
}
