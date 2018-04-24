// +build darwin

package joystick

//#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
/*
#include <IOKit/hid/IOHIDLib.h>
extern void removeCallback(void* ctx, IOReturn res, void *sender);
extern IOHIDManagerRef openHIDManager();
extern void closeHIDManager(IOHIDManagerRef manager);
extern void addHIDElement(void *value, void *parameter);
#define kCFRunLoopMode CFSTR("go-joystick")
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"
)

type joystickManager struct {
	ref        C.IOHIDManagerRef
	devices    map[int]*joystickImpl
	deviceCnt  int
	deviceUsed int
}

var mgr *joystickManager
var mgrMutex sync.Mutex

func openManager() *joystickManager {
	if mgr == nil {
		mgr = &joystickManager{
			ref:        C.IOHIDManagerRef(0),
			devices:    make(map[int]*joystickImpl),
			deviceCnt:  0,
			deviceUsed: 0,
		}
		mgr.ref = C.openHIDManager()
		if mgr.ref == (C.IOHIDManagerRef)(0) {
			return nil
		}
	}
	return mgr
}

//export addCallback
func addCallback(ctx unsafe.Pointer, res C.IOReturn, sender unsafe.Pointer, device C.IOHIDDeviceRef) {
	if res != C.kIOReturnSuccess {
		return
	}
	if mgr.searchFromDeviceRef(device) != nil {
		return
	}
	id := mgr.deviceCnt
	mgr.deviceCnt++
	impl := &joystickImpl{
		id:  id,
		ref: device,
	}
	mgr.devices[id] = impl
	C.IOHIDDeviceRegisterRemovalCallback(device, C.IOHIDCallback(C.removeCallback), unsafe.Pointer(impl))
	C.IOHIDDeviceScheduleWithRunLoop(device, C.CFRunLoopGetCurrent(), C.kCFRunLoopMode)
	elems := C.IOHIDDeviceCopyMatchingElements(device, C.CFDictionaryRef(0), C.kIOHIDOptionsTypeNone)
	impl.addElements(elems)
}

func (js *joystickImpl) addElements(elems C.CFArrayRef) {
	max := int(C.CFArrayGetCount(elems))
	for i := 0; i < max; i++ {
		elem := C.IOHIDElementRef(C.CFArrayGetValueAtIndex(elems, C.long(i)))
		//typeID := C.CFGetTypeID(C.CFTypeRef(elem))
		usagePage := C.IOHIDElementGetUsagePage(elem)
		usage := C.IOHIDElementGetUsage(elem)
		switch C.IOHIDElementGetType(elem) {

		case C.kIOHIDElementTypeInput_Misc:
			fallthrough
		case C.kIOHIDElementTypeInput_Button:
			fallthrough
		case C.kIOHIDElementTypeInput_Axis:

			switch usagePage {
			case C.kHIDPage_GenericDesktop:
				switch usage {
				case C.kHIDUsage_GD_X:
					fallthrough
				case C.kHIDUsage_GD_Y:
					fallthrough
				case C.kHIDUsage_GD_Z:
					fallthrough
				case C.kHIDUsage_GD_Rx:
					fallthrough
				case C.kHIDUsage_GD_Ry:
					fallthrough
				case C.kHIDUsage_GD_Rz:
					fallthrough
				case C.kHIDUsage_GD_Slider:
					fallthrough
				case C.kHIDUsage_GD_Dial:
					fallthrough
				case C.kHIDUsage_GD_Wheel:
					if js.contains(elem) {
						continue
					}
					js.axes = append(js.axes, &joystickAxis{
						ref:    elem,
						min:    int(C.IOHIDElementGetLogicalMin(elem)),
						max:    int(C.IOHIDElementGetLogicalMax(elem)),
						center: -1,
					})
					js.state.AxisData = append(js.state.AxisData, 0)
				}
			case C.kHIDPage_Button:
				if js.contains(elem) {
					continue
				}
				js.buttons = append(js.buttons, &joystickButton{
					ref: elem,
				})
			}
		case C.kIOHIDElementTypeCollection:
			if children := C.IOHIDElementGetChildren(elem); children != C.CFArrayRef(0) {
				js.addElements(children)
			}
		default:
			continue /* Nothing to do */
		}
	}
}

//export removeCallback
func removeCallback(self unsafe.Pointer, res C.IOReturn, sender unsafe.Pointer) {
	if res != C.kIOReturnSuccess {
		return
	}
	impl := (*joystickImpl)(self)
	impl.ref = C.IOHIDDeviceRef(0)
	impl.removed = true
}

func (mgr *joystickManager) searchFromDeviceRef(ref C.IOHIDDeviceRef) *joystickImpl {
	for _, impl := range mgr.devices {
		if impl.ref == ref {
			return impl
		}
	}
	return nil
}

func (mgr *joystickManager) Close() {
	C.closeHIDManager(mgr.ref)
}

// -- elem

type joystickAxis struct {
	ref    C.IOHIDElementRef
	min    int
	max    int
	center int
}

type joystickButton struct {
	ref C.IOHIDElementRef
}

// -- impl

type joystickImpl struct {
	id      int
	ref     C.IOHIDDeviceRef
	removed bool
	axes    []*joystickAxis
	buttons []*joystickButton
	state   State
}

func Open(id int) (Joystick, error) {
	mgrMutex.Lock()
	defer mgrMutex.Unlock()
	mgr := openManager()
	if mgr == nil {
		return nil, fmt.Errorf("Could not open joystick manager")
	}
	js := mgr.devices[id]
	if js == nil {
		return nil, fmt.Errorf("Device not found")
	}
	mgr.deviceUsed++
	return js, nil
}

func (js *joystickImpl) AxisCount() int {
	return len(js.axes)
}

func (js *joystickImpl) ButtonCount() int {
	return len(js.buttons)
}

func (js *joystickImpl) Name() string {
	return ""
}

func (js *joystickImpl) Read() (State, error) {
	for idx, axe := range js.axes {
		var valueRef C.IOHIDValueRef
		if C.IOHIDDeviceGetValue(js.ref, axe.ref, &valueRef) != C.kIOReturnSuccess {
			continue
		}
		min := -32767
		max := 32768
		value := int(C.IOHIDValueGetIntegerValue(valueRef))
		if axe.center < 0 {
			axe.center = value
			js.state.AxisData[idx] = 0
		} else if value < axe.center {
			js.state.AxisData[idx] = int(float64(value-axe.min)*float64(0-min)/float64(axe.center-axe.min)) + min
		} else {
			js.state.AxisData[idx] = int(float64(value-axe.center)*float64(max-0)/float64(axe.max-axe.center)) + 0
		}
	}
	buttons := uint32(0)
	for idx, btn := range js.buttons {
		var valueRef C.IOHIDValueRef
		if C.IOHIDDeviceGetValue(js.ref, btn.ref, &valueRef) != C.kIOReturnSuccess {
			continue
		}
		if int(C.IOHIDValueGetIntegerValue(valueRef)) > 0 {
			buttons |= uint32(1) << uint(idx)
		}
	}
	js.state.Buttons = buttons
	return js.state, nil
}

func (js *joystickImpl) Close() {
	mgrMutex.Lock()
	defer mgrMutex.Unlock()
	mgr := openManager()
	if mgr == nil {
		return
	}
	mgr.deviceUsed--
	if mgr.deviceUsed == 0 {
		mgr.Close()
		mgr = nil
	}
}

func (js *joystickImpl) contains(ref C.IOHIDElementRef) bool {
	for _, elem := range js.axes {
		if elem.ref == ref {
			return true
		}
	}
	for _, elem := range js.buttons {
		if elem.ref == ref {
			return true
		}
	}
	return false
}
