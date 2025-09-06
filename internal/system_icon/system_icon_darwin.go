package system_icon

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreServices
#import <Cocoa/Cocoa.h>
#import <CoreServices/CoreServices.h>

void* getIconForExtension(const char* extension) {
    NSString* ext = [NSString stringWithUTF8String:extension];
    NSImage* icon = [[NSWorkspace sharedWorkspace] iconForFileType:ext];
    [icon retain];
    return (void*)icon;
}

void getImageDimensions(void* imagePtr, int* width, int* height) {
    NSImage* image = (NSImage*)imagePtr;
    NSSize size = [image size];
    *width = (int)size.width;
    *height = (int)size.height;
}

void copyImagePixels(void* imagePtr, unsigned char* buffer, int width, int height) {
    NSImage* image = (NSImage*)imagePtr;
    NSBitmapImageRep* bitmap = [[NSBitmapImageRep alloc]
        initWithBitmapDataPlanes:NULL
        pixelsWide:width
        pixelsHigh:height
        bitsPerSample:8
        samplesPerPixel:4
        hasAlpha:YES
        isPlanar:NO
        colorSpaceName:NSDeviceRGBColorSpace
        bytesPerRow:width * 4
        bitsPerPixel:32];

    [NSGraphicsContext saveGraphicsState];
    [NSGraphicsContext setCurrentContext:[NSGraphicsContext graphicsContextWithBitmapImageRep:bitmap]];
    [image drawInRect:NSMakeRect(0, 0, width, height)
        fromRect:NSZeroRect
        operation:NSCompositingOperationCopy
        fraction:1.0];
    [NSGraphicsContext restoreGraphicsState];

    memcpy(buffer, [bitmap bitmapData], width * height * 4);
    [bitmap release];
}

void releaseIcon(void* iconPtr) {
    if (iconPtr) {
        [(NSImage*)iconPtr release];
    }
}
*/
import "C"
import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"unsafe"
)

// GetExtensionIcon returns a PNG image of the system icon for the given file extension
func GetExtensionIcon(extension string) ([]byte, error) {
	if extension[0] != '.' {
		extension = "." + extension
	}

	// Get the icon from the system
	cExtension := C.CString(extension)
	defer C.free(unsafe.Pointer(cExtension))

	iconPtr := C.getIconForExtension(cExtension)
	if iconPtr == nil {
		return nil, fmt.Errorf("failed to get icon for extension %s", extension)
	}
	defer C.releaseIcon(iconPtr)

	// Get the icon dimensions
	var width, height C.int
	C.getImageDimensions(iconPtr, &width, &height)

	// Create a buffer for the pixel data
	bufferSize := int(width * height * 4)
	buffer := make([]byte, bufferSize)

	// Copy the pixel data
	C.copyImagePixels(iconPtr, (*C.uchar)(unsafe.Pointer(&buffer[0])), width, height)

	// Create an RGBA image from the buffer
	img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
	copy(img.Pix, buffer)

	// Encode the image as PNG
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode image as PNG: %v", err)
	}

	return buf.Bytes(), nil
}
