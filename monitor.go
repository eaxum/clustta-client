//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c -mmacosx-version-min=15.0
#cgo LDFLAGS: -framework Cocoa -framework Foundation

#import <Cocoa/Cocoa.h>
#import <Foundation/Foundation.h>

// Forward declarations for Go callbacks
extern void goFullscreenWillEnter();
extern void goFullscreenDidExit();

@interface FullscreenObserver : NSObject
@property (nonatomic, strong) NSWindow *targetWindow;
- (instancetype)initWithWindow:(NSWindow *)window;
- (void)startObserving;
- (void)stopObserving;
- (void)windowWillEnterFullscreen:(NSNotification *)notification;
- (void)windowDidExitFullscreen:(NSNotification *)notification;
@end

@implementation FullscreenObserver

- (instancetype)initWithWindow:(NSWindow *)window {
    self = [super init];
    if (self) {
        self.targetWindow = window;
    }
    return self;
}

- (void)startObserving {
    NSNotificationCenter *center = [NSNotificationCenter defaultCenter];

    [center addObserver:self
               selector:@selector(windowWillEnterFullscreen:)
                   name:NSWindowWillEnterFullScreenNotification
                 object:self.targetWindow];

    [center addObserver:self
               selector:@selector(windowDidExitFullscreen:)
                   name:NSWindowDidExitFullScreenNotification
                 object:self.targetWindow];
}

- (void)stopObserving {
    NSNotificationCenter *center = [NSNotificationCenter defaultCenter];
    [center removeObserver:self name:NSWindowWillEnterFullScreenNotification object:self.targetWindow];
    [center removeObserver:self name:NSWindowDidExitFullScreenNotification object:self.targetWindow];
}

- (void)windowWillEnterFullscreen:(NSNotification *)notification {
    goFullscreenWillEnter();
}

- (void)windowDidExitFullscreen:(NSNotification *)notification {
    goFullscreenDidExit();
}

@end

// Global observer instance
static FullscreenObserver *observer = nil;

// C functions to be called from Go
void startFullscreenMonitoringC(void *windowPtr) {
    if (observer != nil) {
        [observer stopObserving];
        observer = nil;
    }

    NSWindow *window = (__bridge NSWindow *)windowPtr;
    if (window) {
        observer = [[FullscreenObserver alloc] initWithWindow:window];
        [observer startObserving];
    }
}

void stopFullscreenMonitoringC(void) {
    if (observer != nil) {
        [observer stopObserving];
        observer = nil;
    }
}

// Helper function to get a window by searching through all windows
void* findWindowByTitleC(const char* title) {
    NSString *targetTitle = [NSString stringWithUTF8String:title];
    NSArray *windows = [[NSApplication sharedApplication] windows];

    for (NSWindow *window in windows) {
        if ([[window title] isEqualToString:targetTitle]) {
            return (__bridge void *)window;
        }
    }
    return NULL;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// StartFullscreenMonitoring starts monitoring fullscreen events for a window with the given title
func StartFullscreenMonitoring(windowTitle string) error {
	titleCStr := C.CString(windowTitle)
	defer C.free(unsafe.Pointer(titleCStr))

	windowPtr := C.findWindowByTitleC(titleCStr)
	if windowPtr == nil {
		return errors.New("window not found")
	}

	C.startFullscreenMonitoringC(windowPtr)
	return nil
}

// StopFullscreenMonitoring stops monitoring fullscreen events
func StopFullscreenMonitoring() {
	C.stopFullscreenMonitoringC()
}
