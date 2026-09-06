//go:build darwin && cgo

// Copyright (c) 2026 Eaxum.
// Adapted from drag-rs; see NOTICE and LICENSE.drag-rs.
#import <Cocoa/Cocoa.h>
#include <string.h>
#include "dragout_darwin.h"

static const CGFloat ClusttaDragIconSize = 32.0;
static const NSUInteger ClusttaLeftMouseButton = 1;

@interface ClusttaExportDragSource : NSObject <NSDraggingSource>
@property(nonatomic, assign) uintptr_t callback;
- (void)finish:(int)status;
@end

// AppKit does not own the source; keep it alive until the completion callback.
static ClusttaExportDragSource *activeSource;

@implementation ClusttaExportDragSource
- (NSDragOperation)draggingSession:(NSDraggingSession *)session
    sourceOperationMaskForDraggingContext:(NSDraggingContext)context {
    return NSDragOperationCopy;
}

- (BOOL)ignoreModifierKeysForDraggingSession:(NSDraggingSession *)session {
    return YES;
}

- (void)draggingSession:(NSDraggingSession *)session endedAtPoint:(NSPoint)point
    operation:(NSDragOperation)operation {
    [self finish:(operation & NSDragOperationCopy) ? ClusttaDragDropped : ClusttaDragCancelled];
}

- (void)finish:(int)status {
    if (!self.callback) return;
    uintptr_t callback = self.callback;
    self.callback = 0;
    activeSource = nil;
    clustta_drag_completed(callback, status);
    [self autorelease];
}
@end

static NSArray *draggingItems(const char *paths, size_t count, NSPoint location) {
    NSMutableArray *items = [NSMutableArray arrayWithCapacity:count];
    const char *path = paths;
    for (size_t index = 0; index < count; index++) {
        NSString *filePath = [NSString stringWithUTF8String:path];
        path += strlen(path) + 1;
        BOOL directory = NO;
        if (!filePath || ![[NSFileManager defaultManager] fileExistsAtPath:filePath isDirectory:&directory]
            || directory || ![[NSFileManager defaultManager] isReadableFileAtPath:filePath]) return nil;
        NSURL *url = [NSURL fileURLWithPath:filePath isDirectory:NO];
        NSDraggingItem *item = [[[NSDraggingItem alloc] initWithPasteboardWriter:url] autorelease];
        NSImage *icon = [[NSWorkspace sharedWorkspace] iconForFile:filePath];
        NSRect frame = NSMakeRect(location.x - ClusttaDragIconSize / 2,
            location.y - ClusttaDragIconSize / 2, ClusttaDragIconSize, ClusttaDragIconSize);
        [item setDraggingFrame:frame contents:icon];
        [items addObject:item];
    }
    return items;
}

void clustta_begin_drag(void *windowPointer, const char *paths, size_t count, uintptr_t callback) {
    @autoreleasepool {
        NSWindow *window = (NSWindow *)windowPointer;
        if (![NSThread isMainThread] || !window || ![[NSApp windows] containsObject:window]
            || !window.isKeyWindow || !window.contentView) {
            clustta_drag_completed(callback, ClusttaDragInvalidWindow);
            return;
        }
        if (activeSource) {
            clustta_drag_completed(callback, ClusttaDragBusy);
            return;
        }
        if (!([NSEvent pressedMouseButtons] & ClusttaLeftMouseButton)) {
            clustta_drag_completed(callback, ClusttaDragCancelled);
            return;
        }
        ClusttaExportDragSource *source = [[ClusttaExportDragSource alloc] init];
        source.callback = callback;
        activeSource = source;
        @try {
            NSView *view = window.contentView;
            NSPoint windowLocation = window.mouseLocationOutsideOfEventStream;
            NSPoint viewLocation = [view convertPoint:windowLocation fromView:nil];
            NSArray *items = draggingItems(paths, count, viewLocation);
            if (!items.count) {
                [source finish:ClusttaDragInvalidFile];
                return;
            }
            if (!([NSEvent pressedMouseButtons] & ClusttaLeftMouseButton)) {
                [source finish:ClusttaDragCancelled];
                return;
            }
            // The asynchronous webview binding cannot retain the original mouse-drag event.
            NSEvent *event = [NSEvent mouseEventWithType:NSEventTypeLeftMouseDragged
                location:windowLocation modifierFlags:[NSEvent modifierFlags]
                timestamp:[NSProcessInfo processInfo].systemUptime windowNumber:window.windowNumber
                context:nil eventNumber:0 clickCount:1 pressure:1.0];
            NSDraggingSession *session = [view beginDraggingSessionWithItems:items event:event source:source];
            if (!session) {
                [source finish:ClusttaDragFailed];
                return;
            }
            session.animatesToStartingPositionsOnCancelOrFail = YES;
        } @catch (NSException *exception) {
            NSLog(@"Clustta export drag failed: %@", exception.reason);
            [source finish:ClusttaDragFailed];
        }
    }
}
