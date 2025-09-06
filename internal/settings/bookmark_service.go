//go:build darwin

package settings

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation -framework CoreFoundation
#import <Foundation/Foundation.h>

// CreateSecurityScopedBookmark creates a bookmark from a file path
char* CreateSecurityScopedBookmark(const char* path, long* bookmarkDataLength) {
    NSString *nsPath = [NSString stringWithUTF8String:path];
    NSURL *url = [NSURL fileURLWithPath:nsPath];

    NSError *error = nil;
    NSData *bookmarkData = [url bookmarkDataWithOptions:NSURLBookmarkCreationWithSecurityScope
                                 includingResourceValuesForKeys:nil
                                                  relativeToURL:nil
                                                          error:&error];

    if (error != nil || bookmarkData == nil) {
        *bookmarkDataLength = -1;
        return NULL;
    }

    *bookmarkDataLength = [bookmarkData length];
    char* result = malloc(*bookmarkDataLength);
    [bookmarkData getBytes:result length:*bookmarkDataLength];

    return result;
}

// ResolveSecurityScopedBookmark resolves a bookmark back to a file path
char* ResolveSecurityScopedBookmark(const char* bookmarkData, long bookmarkDataLength) {
    NSData *data = [NSData dataWithBytes:bookmarkData length:bookmarkDataLength];

    BOOL isStale = NO;
    NSError *error = nil;
    NSURL *url = [NSURL URLByResolvingBookmarkData:data
                                          options:NSURLBookmarkResolutionWithSecurityScope
                                    relativeToURL:nil
                              bookmarkDataIsStale:&isStale
                                            error:&error];

    if (error != nil || url == nil) {
        return NULL;
    }

    // Start accessing the security-scoped resource
    [url startAccessingSecurityScopedResource];

    NSString *path = [url path];
    const char *cPath = [path UTF8String];
    char *result = malloc(strlen(cPath) + 1);
    strcpy(result, cPath);

    return result;
}

// IsBookmarkStale checks if a bookmark is stale
bool IsBookmarkStale(const char* bookmarkData, long bookmarkDataLength) {
    NSData *data = [NSData dataWithBytes:bookmarkData length:bookmarkDataLength];

    BOOL isStale = NO;
    NSError *error = nil;
    NSURL *url = [NSURL URLByResolvingBookmarkData:data
                                          options:NSURLBookmarkResolutionWithSecurityScope
                                    relativeToURL:nil
                              bookmarkDataIsStale:&isStale
                                            error:&error];

    if (error != nil || url == nil) {
        return true;
    }

    return isStale;
}

// StopAccessingSecurityScopedResource stops accessing a security-scoped resource
void StopAccessingSecurityScopedResource(const char* bookmarkData, long bookmarkDataLength) {
    NSData *data = [NSData dataWithBytes:bookmarkData length:bookmarkDataLength];

    BOOL isStale = NO;
    NSError *error = nil;
    NSURL *url = [NSURL URLByResolvingBookmarkData:data
                                          options:NSURLBookmarkResolutionWithSecurityScope
                                    relativeToURL:nil
                              bookmarkDataIsStale:&isStale
                                            error:&error];

    if (error == nil && url != nil) {
        [url stopAccessingSecurityScopedResource];
    }
}
*/
import "C"
import (
	"errors"
	"unsafe"
)

// CreateBookmarkFromPath creates a security-scoped bookmark from a file path
func CreateBookmarkFromPath(path string) ([]byte, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var bookmarkDataLength C.long
	cBookmarkData := C.CreateSecurityScopedBookmark(cPath, &bookmarkDataLength)

	if cBookmarkData == nil || bookmarkDataLength == -1 {
		return nil, errors.New("failed to create security-scoped bookmark")
	}

	defer C.free(unsafe.Pointer(cBookmarkData))

	bookmarkData := C.GoBytes(unsafe.Pointer(cBookmarkData), C.int(bookmarkDataLength))
	return bookmarkData, nil
}

// ResolveBookmark resolves bookmark data back to an accessible path
func ResolveBookmark(bookmarkData []byte) (string, error) {
	if len(bookmarkData) == 0 {
		return "", errors.New("empty bookmark data")
	}

	cBookmarkData := C.CBytes(bookmarkData)
	defer C.free(cBookmarkData)

	cPath := C.ResolveSecurityScopedBookmark((*C.char)(cBookmarkData), C.long(len(bookmarkData)))
	if cPath == nil {
		return "", errors.New("failed to resolve security-scoped bookmark")
	}

	defer C.free(unsafe.Pointer(cPath))

	path := C.GoString(cPath)
	return path, nil
}

// IsBookmarkStale checks if a bookmark is still valid
func IsBookmarkStale(bookmarkData []byte) bool {
	if len(bookmarkData) == 0 {
		return true
	}

	cBookmarkData := C.CBytes(bookmarkData)
	defer C.free(cBookmarkData)

	return bool(C.IsBookmarkStale((*C.char)(cBookmarkData), C.long(len(bookmarkData))))
}

// StopAccessingResource stops accessing a security-scoped resource
func StopAccessingResource(bookmarkData []byte) {
	if len(bookmarkData) == 0 {
		return
	}

	cBookmarkData := C.CBytes(bookmarkData)
	defer C.free(cBookmarkData)

	C.StopAccessingSecurityScopedResource((*C.char)(cBookmarkData), C.long(len(bookmarkData)))
}
