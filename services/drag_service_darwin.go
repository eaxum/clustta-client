//go:build darwin

package services

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

// ClусttaDragSource implements NSDraggingSource protocol
@interface ClusttaDragSource : NSObject <NSDraggingSource>
@end

@implementation ClusttaDragSource

- (NSDragOperation)draggingSession:(NSDraggingSession *)session
    sourceOperationMaskForDraggingContext:(NSDraggingContext)context {
    return NSDragOperationCopy | NSDragOperationMove | NSDragOperationLink;
}

@end

static ClusttaDragSource *dragSource = nil;

int startDragWithFiles(const char** paths, int count) {
    if (count == 0) return 0;

    __block int result = 0;

    // Must run on main thread for UI operations
    dispatch_sync(dispatch_get_main_queue(), ^{
        @autoreleasepool {
            // Initialize drag source if needed
            if (dragSource == nil) {
                dragSource = [[ClusttaDragSource alloc] init];
            }

            // Get the key window (Wails window)
            NSWindow *window = [[NSApplication sharedApplication] keyWindow];
            if (!window) {
                result = -1;
                return;
            }

            NSView *view = [window contentView];
            if (!view) {
                result = -2;
                return;
            }

            // Create file URLs from paths
            NSMutableArray *urls = [NSMutableArray arrayWithCapacity:count];
            for (int i = 0; i < count; i++) {
                NSString *path = [NSString stringWithUTF8String:paths[i]];
                NSURL *url = [NSURL fileURLWithPath:path];
                if (url) {
                    [urls addObject:url];
                }
            }

            if ([urls count] == 0) {
                result = -3;
                return;
            }

            // Create dragging items
            NSMutableArray *draggingItems = [NSMutableArray arrayWithCapacity:[urls count]];
            NSPoint mouseLocation = [NSEvent mouseLocation];
            NSPoint windowPoint = [window convertPointFromScreen:mouseLocation];
            NSPoint viewPoint = [view convertPoint:windowPoint fromView:nil];

            for (NSURL *url in urls) {
                NSDraggingItem *item = [[NSDraggingItem alloc] initWithPasteboardWriter:url];

                // Get file icon for drag image
                NSImage *icon = [[NSWorkspace sharedWorkspace] iconForFile:[url path]];
                [icon setSize:NSMakeSize(32, 32)];

                // Set drag frame centered on mouse
                NSRect dragFrame = NSMakeRect(viewPoint.x - 16, viewPoint.y - 16, 32, 32);
                [item setDraggingFrame:dragFrame contents:icon];

                [draggingItems addObject:item];
            }

            // Get the current event or create a synthetic one
            NSEvent *currentEvent = [NSApp currentEvent];
            if (!currentEvent || ([currentEvent type] != NSEventTypeLeftMouseDown &&
                                   [currentEvent type] != NSEventTypeLeftMouseDragged)) {
                // Create synthetic mouse down event at current location
                currentEvent = [NSEvent mouseEventWithType:NSEventTypeLeftMouseDragged
                                                  location:windowPoint
                                             modifierFlags:0
                                                 timestamp:[[NSProcessInfo processInfo] systemUptime]
                                              windowNumber:[window windowNumber]
                                                   context:nil
                                               eventNumber:0
                                                clickCount:1
                                                  pressure:1.0];
            }

            // Begin dragging session
            [view beginDraggingSessionWithItems:draggingItems
                                          event:currentEvent
                                         source:dragSource];

            result = 1;
        }
    });

    return result;
}

int isMouseButtonDown() {
    return ([NSEvent pressedMouseButtons] & 1) != 0 ? 1 : 0;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// StartNativeDrag initiates a native macOS drag-and-drop operation.
// Returns the drop effect indicator.
func (d *DragService) StartNativeDrag(filePaths []string) (int, error) {
	if len(filePaths) == 0 {
		return 0, nil
	}

	// Convert Go strings to C strings
	cPaths := make([]*C.char, len(filePaths))
	for i, p := range filePaths {
		cPaths[i] = C.CString(p)
	}
	defer func() {
		for _, p := range cPaths {
			C.free(unsafe.Pointer(p))
		}
	}()

	result := C.startDragWithFiles(&cPaths[0], C.int(len(filePaths)))

	if result < 0 {
		return 0, fmt.Errorf("failed to start native drag: error code %d", result)
	}

	return int(result), nil
}

// IsMouseButtonDown checks if the left mouse button is currently pressed.
func (d *DragService) IsMouseButtonDown() bool {
	return C.isMouseButtonDown() != 0
}
