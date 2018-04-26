#include <IOKit/hid/IOHIDLib.h>

// golang callback
extern void addCallback(void* self, IOReturn res, void *sender, IOHIDDeviceRef ioHIDDeviceObject);
extern void removeCallback(void* self, IOReturn res, void *sender);

static CFDictionaryRef createMatcherElement(const UInt32 page, const UInt32 usage, int *okay) {
	CFDictionaryRef dict = 0;
	CFNumberRef pageNumRef = CFNumberCreate(kCFAllocatorDefault, kCFNumberIntType, &page);
	CFNumberRef usageNumRef = CFNumberCreate(kCFAllocatorDefault, kCFNumberIntType, &usage);
	const void *keys[2] = { (void *) CFSTR(kIOHIDDeviceUsagePageKey), (void *) CFSTR(kIOHIDDeviceUsageKey) };
	const void *vals[2] = { (void *) pageNumRef, (void *) usageNumRef };

	if (pageNumRef && usageNumRef) {
		dict = CFDictionaryCreate(kCFAllocatorDefault, keys, vals, 2, &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	}

	if (pageNumRef) {
		CFRelease(pageNumRef);
	}
	if (usageNumRef) {
		CFRelease(usageNumRef);
	}

	if (!dict) {
		*okay = 0;
	}

	return dict;
}

static CFArrayRef createMatcher() {
	int okay = 1;
	const void *vals[] = {
		(void *) createMatcherElement(kHIDPage_GenericDesktop, kHIDUsage_GD_Joystick, &okay),
		(void *) createMatcherElement(kHIDPage_GenericDesktop, kHIDUsage_GD_GamePad, &okay),
		(void *) createMatcherElement(kHIDPage_GenericDesktop, kHIDUsage_GD_MultiAxisController, &okay),
	};
	CFArrayRef matcher = okay ? CFArrayCreate(kCFAllocatorDefault, vals, 3, &kCFTypeArrayCallBacks) : 0;
	for (size_t i = 0; i < 3; i++) {
		if (vals[i]) {
			CFRelease((CFTypeRef) vals[i]);
		}
	}
	return matcher;
}

#define kCFRunLoopMode CFSTR("go-joystick")

IOHIDManagerRef openHIDManager() {
	CFRunLoopRef runloop = CFRunLoopGetCurrent();
	IOHIDManagerRef manager = IOHIDManagerCreate(kCFAllocatorDefault, kIOHIDOptionsTypeNone);
	if(!manager) {
		return 0;
	}
	if(kIOReturnSuccess != IOHIDManagerOpen(manager, kIOHIDOptionsTypeNone)) {
		CFRelease(manager);
		return 0;
	}
	CFArrayRef matcher = createMatcher();
	IOHIDManagerSetDeviceMatchingMultiple(manager, matcher);
	IOHIDManagerRegisterDeviceMatchingCallback(manager, addCallback, 0);
	IOHIDManagerScheduleWithRunLoop(manager, runloop, kCFRunLoopMode);
	while (CFRunLoopRunInMode(kCFRunLoopMode, 0, TRUE) == kCFRunLoopRunHandledSource) {

	}
	CFRelease(matcher);
	return manager;
}

void closeHIDManager(IOHIDManagerRef manager) {
	CFRunLoopRef runloop = CFRunLoopGetCurrent();
	CFRunLoopStop(runloop);
	IOHIDManagerUnscheduleFromRunLoop(manager, runloop, kCFRunLoopMode);
	CFRelease(manager);
}

