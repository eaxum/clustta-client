//go:build darwin

package system_thumbnail

/*
#cgo CFLAGS: -x objective-c -fblocks
#cgo LDFLAGS: -framework Foundation -framework QuickLookThumbnailing -framework CoreGraphics -framework ImageIO -framework AppKit -framework UniformTypeIdentifiers
#import <Foundation/Foundation.h>
#import <QuickLookThumbnailing/QuickLookThumbnailing.h>
#import <CoreGraphics/CoreGraphics.h>
#import <ImageIO/ImageIO.h>
#import <AppKit/AppKit.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>
#import <dispatch/dispatch.h>

// GenerateThumbnail generates a thumbnail using Quick Look
// Returns PNG data and the actual dimensions of the thumbnail
CGImageRef generateThumbnail(const char* path, int size, int* outWidth, int* outHeight) {
    @autoreleasepool {
        NSString *filePath = [NSString stringWithUTF8String:path];
        NSURL *fileURL = [NSURL fileURLWithPath:filePath];

        if (!fileURL) {
            return NULL;
        }

        CGSize thumbnailSize = CGSizeMake(size, size);
        QLThumbnailGenerationRequest *request = [[QLThumbnailGenerationRequest alloc]
            initWithFileAtURL:fileURL
            size:thumbnailSize
            scale:1.0
            representationTypes:QLThumbnailGenerationRequestRepresentationTypeThumbnail];
        request.iconMode = NO;

        dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);
        __block CGImageRef thumbnail = NULL;

        [[QLThumbnailGenerator sharedGenerator] generateBestRepresentationForRequest:request
            completionHandler:^(QLThumbnailRepresentation * _Nullable representation, NSError * _Nullable error) {
                if (representation.CGImage) {
                    thumbnail = CGImageRetain(representation.CGImage);
                }
                dispatch_semaphore_signal(semaphore);
            }];

        dispatch_time_t timeout = dispatch_time(DISPATCH_TIME_NOW, 30 * NSEC_PER_SEC);
        if (dispatch_semaphore_wait(semaphore, timeout) != 0) {
            [[QLThumbnailGenerator sharedGenerator] cancelRequest:request];
            [request release];
            return NULL;
        }

        [request release];

        if (thumbnail) {
            *outWidth = (int)CGImageGetWidth(thumbnail);
            *outHeight = (int)CGImageGetHeight(thumbnail);
        }

        return thumbnail;
    }
}

// ConvertCGImageToPNG converts a CGImage to PNG data
// Returns the PNG data and its length
unsigned char* cgImageToPNG(CGImageRef image, size_t* outLength) {
    if (!image) {
        return NULL;
    }

    @autoreleasepool {
        // Create mutable data to hold PNG
        NSMutableData *pngData = [NSMutableData data];

        // Create image destination
        CGImageDestinationRef destination = CGImageDestinationCreateWithData(
            (__bridge CFMutableDataRef)pngData,
            (__bridge CFStringRef)UTTypePNG.identifier,
            1,
            NULL
        );

        if (!destination) {
            return NULL;
        }

        // Add image to destination
        CGImageDestinationAddImage(destination, image, NULL);

        // Finalize the image destination
        if (!CGImageDestinationFinalize(destination)) {
            CFRelease(destination);
            return NULL;
        }

        CFRelease(destination);

        // Copy data to C buffer
        *outLength = [pngData length];
        unsigned char *buffer = (unsigned char*)malloc(*outLength);
        if (buffer) {
            memcpy(buffer, [pngData bytes], *outLength);
        }

        return buffer;
    }
}

// GetIconThumbnail generates an icon-based thumbnail as fallback
CGImageRef getIconThumbnail(const char* path, int size, int* outWidth, int* outHeight) {
    @autoreleasepool {
        NSString *filePath = [NSString stringWithUTF8String:path];
        NSImage *icon = [[NSWorkspace sharedWorkspace] iconForFile:filePath];

        if (!icon) {
            return NULL;
        }

        // Resize icon to requested size
        NSSize targetSize = NSMakeSize(size, size);
        NSImage *resizedIcon = [[NSImage alloc] initWithSize:targetSize];

        [resizedIcon lockFocus];
        [icon drawInRect:NSMakeRect(0, 0, targetSize.width, targetSize.height)
                fromRect:NSZeroRect
               operation:NSCompositingOperationCopy
                fraction:1.0];
        [resizedIcon unlockFocus];

        // Convert NSImage to CGImage
        CGImageRef cgImage = [resizedIcon CGImageForProposedRect:NULL
                                                         context:nil
                                                           hints:nil];

        if (cgImage) {
            // Retain the image since we're returning it
            CGImageRetain(cgImage);
            *outWidth = (int)CGImageGetWidth(cgImage);
            *outHeight = (int)CGImageGetHeight(cgImage);
        }

        return cgImage;
    }
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// ThumbnailOptions configures thumbnail generation behavior
type ThumbnailOptions uint32

const (
	ThumbnailDefault         ThumbnailOptions = 0
	ThumbnailOnlyIfCached    ThumbnailOptions = 1 << 0 // Only return if already cached (not supported on macOS)
	ThumbnailIconFallback    ThumbnailOptions = 1 << 1 // Fall back to icon if thumbnail fails
	ThumbnailUseCurrentScale ThumbnailOptions = 1 << 2 // Apply DPI scaling (not applicable on macOS)
	ThumbnailHighQuality     ThumbnailOptions = 1 << 3 // Prefer quality over speed
)

// GetOSThumbnail generates a thumbnail for the specified file using macOS Quick Look APIs
// For consistency with Windows, always generates at 512px for maximum quality
func GetOSThumbnail(filePath string, size int, options ThumbnailOptions) ([]byte, error) {
	// Always use 512px for consistency with Windows implementation
	// This provides maximum quality and lets CSS handle sizing/cropping
	actualSize := 512

	cPath := C.CString(filePath)
	defer C.free(unsafe.Pointer(cPath))

	var width, height C.int

	// Generate thumbnail using Quick Look
	thumbnail := C.generateThumbnail(cPath, C.int(actualSize), &width, &height)

	// If thumbnail generation failed and icon fallback is enabled, try getting icon
	if thumbnail == C.CGImageRef(0) && (options&ThumbnailIconFallback) != 0 {
		thumbnail = C.getIconThumbnail(cPath, C.int(actualSize), &width, &height)
	}

	if thumbnail == C.CGImageRef(0) {
		return nil, fmt.Errorf("failed to generate thumbnail for: %s", filePath)
	}
	defer C.CGImageRelease(thumbnail)

	// Convert CGImage to PNG
	var pngLength C.size_t
	pngData := C.cgImageToPNG(thumbnail, &pngLength)
	if pngData == nil {
		return nil, fmt.Errorf("failed to convert thumbnail to PNG")
	}
	defer C.free(unsafe.Pointer(pngData))

	// Copy C buffer to Go slice
	goBytes := C.GoBytes(unsafe.Pointer(pngData), C.int(pngLength))

	return goBytes, nil
}

// GetCachedThumbnail attempts to get a cached thumbnail without generating a new one
// Note: macOS Quick Look doesn't provide direct cache-only access, so this generates a new thumbnail
func GetCachedThumbnail(filePath string, size int) ([]byte, error) {
	return GetOSThumbnail(filePath, size, ThumbnailDefault)
}
