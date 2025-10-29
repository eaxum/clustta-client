//go:build windows

package system_thumbnail

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"syscall"
	"unsafe"
)

// Windows COM structures and interfaces
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type SIZE struct {
	cx int32
	cy int32
}

type BITMAPINFOHEADER struct {
	biSize          uint32
	biWidth         int32
	biHeight        int32
	biPlanes        uint16
	biBitCount      uint16
	biCompression   uint32
	biSizeImage     uint32
	biXPelsPerMeter int32
	biYPelsPerMeter int32
	biClrUsed       uint32
	biClrImportant  uint32
}

type BITMAPINFO struct {
	Header BITMAPINFOHEADER
	Colors [1]uint32
}

// BITMAP structure for GetObject
type BITMAP struct {
	bmType       int32
	bmWidth      int32
	bmHeight     int32
	bmWidthBytes int32
	bmPlanes     uint16
	bmBitsPixel  uint16
	bmBits       uintptr
}

// Windows API constants
const (
	CLSCTX_INPROC_SERVER = 0x1
	S_OK                 = 0
	DIB_RGB_COLORS       = 0
	BI_RGB               = 0

	// SIIGBF flags for IShellItemImageFactory::GetImage
	SIIGBF_RESIZETOFIT    = 0x00
	SIIGBF_BIGGERSIZEOK   = 0x01
	SIIGBF_MEMORYONLY     = 0x02
	SIIGBF_ICONONLY       = 0x04
	SIIGBF_THUMBNAILONLY  = 0x08
	SIIGBF_INCACHEONLY    = 0x10
	SIIGBF_CROPTOSQUARE   = 0x20
	SIIGBF_WIDETHUMBNAILS = 0x40
	SIIGBF_ICONBACKGROUND = 0x80
	SIIGBF_SCALEUP        = 0x100
)

// IShellItem GUID
var (
	IID_IShellItem = GUID{
		Data1: 0x43826d1e,
		Data2: 0xe718,
		Data3: 0x42ee,
		Data4: [8]byte{0xbc, 0x55, 0xa1, 0xe2, 0x61, 0xc3, 0x7b, 0xfe},
	}

	IID_IShellItemImageFactory = GUID{
		Data1: 0xbcc18b79,
		Data2: 0xba16,
		Data3: 0x442f,
		Data4: [8]byte{0x80, 0xc4, 0x8a, 0x59, 0xc3, 0x0c, 0x46, 0x3b},
	}
)

var (
	shell32                         = syscall.NewLazyDLL("shell32.dll")
	user32                          = syscall.NewLazyDLL("user32.dll")
	gdi32                           = syscall.NewLazyDLL("gdi32.dll")
	ole32                           = syscall.NewLazyDLL("ole32.dll")
	procSHCreateItemFromParsingName = shell32.NewProc("SHCreateItemFromParsingName")
	procCoInitializeEx              = ole32.NewProc("CoInitializeEx")
	procCoUninitialize              = ole32.NewProc("CoUninitialize")
	procGetDC                       = user32.NewProc("GetDC")
	procCreateCompatibleDC          = gdi32.NewProc("CreateCompatibleDC")
	procGetDIBits                   = gdi32.NewProc("GetDIBits")
	procDeleteDC                    = gdi32.NewProc("DeleteDC")
	procReleaseDC                   = user32.NewProc("ReleaseDC")
	procDeleteObject                = gdi32.NewProc("DeleteObject")
	procGetDesktopWindow            = user32.NewProc("GetDesktopWindow")
	procGetDpiForWindow             = user32.NewProc("GetDpiForWindow")
	procGetObjectW                  = gdi32.NewProc("GetObjectW")
)

// ThumbnailOptions configures thumbnail generation behavior
type ThumbnailOptions uint32

const (
	ThumbnailDefault         ThumbnailOptions = 0
	ThumbnailOnlyIfCached    ThumbnailOptions = 1 << 0 // Only return if already cached
	ThumbnailIconFallback    ThumbnailOptions = 1 << 1 // Fall back to icon if thumbnail fails
	ThumbnailUseCurrentScale ThumbnailOptions = 1 << 2 // Apply DPI scaling
	ThumbnailHighQuality     ThumbnailOptions = 1 << 3 // Prefer quality over speed
)

// getDpiScale returns the system DPI scale factor
func getDpiScale() float64 {
	// Get desktop window handle
	hWnd, _, _ := procGetDesktopWindow.Call()

	// Try GetDpiForWindow (Windows 10 1607 and later)
	if dpi, _, _ := procGetDpiForWindow.Call(hWnd); dpi != 0 {
		return float64(dpi) / 96.0
	}

	return 1.0 // Default to 100% if method fails
}

// getScaledSize returns the appropriate thumbnail size based on DPI
func getScaledSize(baseSize int, useDpiScaling bool) int {
	if !useDpiScaling {
		return baseSize
	}
	scale := getDpiScale()
	return int(float64(baseSize) * scale)
}

// IShellItem interface methods (simplified vtable)
type IShellItem struct {
	vtbl *IShellItemVtbl
}

type IShellItemVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	// Other methods omitted for brevity
}

// IShellItemImageFactory interface methods
type IShellItemImageFactory struct {
	vtbl *IShellItemImageFactoryVtbl
}

type IShellItemImageFactoryVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	GetImage       uintptr
}

// Release releases a COM interface
func (obj *IShellItem) Release() {
	if obj != nil && obj.vtbl != nil {
		syscall.SyscallN(obj.vtbl.Release, uintptr(unsafe.Pointer(obj)))
	}
}

// Release releases the image factory interface
func (obj *IShellItemImageFactory) Release() {
	if obj != nil && obj.vtbl != nil {
		syscall.SyscallN(obj.vtbl.Release, uintptr(unsafe.Pointer(obj)))
	}
}

// GetImage gets a thumbnail image from IShellItemImageFactory
func (obj *IShellItemImageFactory) GetImage(size SIZE, flags uint32) (uintptr, error) {
	var hBitmap uintptr
	ret, _, _ := syscall.SyscallN(
		obj.vtbl.GetImage,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(&size)),
		uintptr(flags),
		uintptr(unsafe.Pointer(&hBitmap)),
	)

	if ret != S_OK {
		return 0, fmt.Errorf("GetImage failed with HRESULT: 0x%X", ret)
	}

	return hBitmap, nil
}

// createShellItemFromPath creates an IShellItem from a file path
func createShellItemFromPath(path string) (*IShellItem, error) {
	// Initialize COM
	procCoInitializeEx.Call(0, 0x2) // COINIT_APARTMENTTHREADED

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to UTF16: %v", err)
	}

	var shellItem *IShellItem
	ret, _, _ := procSHCreateItemFromParsingName.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		uintptr(unsafe.Pointer(&IID_IShellItem)),
		uintptr(unsafe.Pointer(&shellItem)),
	)

	if ret != S_OK {
		return nil, fmt.Errorf("SHCreateItemFromParsingName failed with HRESULT: 0x%X", ret)
	}

	return shellItem, nil
}

// queryInterfaceImageFactory queries for IShellItemImageFactory interface
func queryInterfaceImageFactory(shellItem *IShellItem) (*IShellItemImageFactory, error) {
	var imageFactory *IShellItemImageFactory
	ret, _, _ := syscall.SyscallN(
		shellItem.vtbl.QueryInterface,
		uintptr(unsafe.Pointer(shellItem)),
		uintptr(unsafe.Pointer(&IID_IShellItemImageFactory)),
		uintptr(unsafe.Pointer(&imageFactory)),
	)

	if ret != S_OK {
		return nil, fmt.Errorf("QueryInterface for IShellItemImageFactory failed: 0x%X", ret)
	}

	return imageFactory, nil
}

// getBitmapDimensions retrieves the actual width and height of an HBITMAP
func getBitmapDimensions(hBitmap uintptr) (width, height int, err error) {
	var bmp BITMAP
	ret, _, _ := procGetObjectW.Call(
		hBitmap,
		uintptr(unsafe.Sizeof(bmp)),
		uintptr(unsafe.Pointer(&bmp)),
	)

	if ret == 0 {
		return 0, 0, fmt.Errorf("GetObject failed")
	}

	return int(bmp.bmWidth), int(bmp.bmHeight), nil
}

// hBitmapToPNG converts a Windows HBITMAP to PNG bytes
func hBitmapToPNG(hBitmap uintptr, width, height int) ([]byte, error) {
	// Get the device context for the screen
	hDC, _, _ := procGetDC.Call(0)
	if hDC == 0 {
		return nil, fmt.Errorf("failed to get DC")
	}
	defer procReleaseDC.Call(0, hDC)

	// Create a compatible DC
	hMemDC, _, _ := procCreateCompatibleDC.Call(hDC)
	if hMemDC == 0 {
		return nil, fmt.Errorf("failed to create compatible DC")
	}
	defer procDeleteDC.Call(hMemDC)

	// Prepare BITMAPINFO
	bi := BITMAPINFO{}
	bi.Header.biSize = uint32(unsafe.Sizeof(bi.Header))
	bi.Header.biWidth = int32(width)
	bi.Header.biHeight = -int32(height) // Negative for top-down DIB
	bi.Header.biPlanes = 1
	bi.Header.biBitCount = 32
	bi.Header.biCompression = BI_RGB

	// Allocate buffer for pixel data
	bufferSize := width * height * 4
	pixelData := make([]byte, bufferSize)

	// Get bitmap bits
	ret, _, _ := procGetDIBits.Call(
		hMemDC,
		hBitmap,
		0,
		uintptr(height),
		uintptr(unsafe.Pointer(&pixelData[0])),
		uintptr(unsafe.Pointer(&bi)),
		DIB_RGB_COLORS,
	)

	if ret == 0 {
		return nil, fmt.Errorf("GetDIBits failed")
	}

	// Create RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Convert BGRA to RGBA
	for i := 0; i < len(pixelData); i += 4 {
		img.Pix[i] = pixelData[i+2]   // R
		img.Pix[i+1] = pixelData[i+1] // G
		img.Pix[i+2] = pixelData[i]   // B
		img.Pix[i+3] = pixelData[i+3] // A
	}

	// Encode to PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %v", err)
	}

	return buf.Bytes(), nil
}

// GetOSThumbnail generates a thumbnail for the specified file using Windows Shell APIs
// Always fetches the largest cached thumbnail size (512px) for maximum quality
func GetOSThumbnail(filePath string, size int, options ThumbnailOptions) ([]byte, error) {
	// Always use 512px - the largest Windows cached thumbnail size
	// This provides maximum quality and lets CSS handle sizing/cropping
	actualSize := 512

	// Create shell item from path
	shellItem, err := createShellItemFromPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create shell item: %v", err)
	}
	defer shellItem.Release()
	defer procCoUninitialize.Call()

	// Query for IShellItemImageFactory
	imageFactory, err := queryInterfaceImageFactory(shellItem)
	if err != nil {
		return nil, fmt.Errorf("failed to get image factory: %v", err)
	}
	defer imageFactory.Release()

	// Determine flags
	flags := uint32(SIIGBF_RESIZETOFIT)

	if (options & ThumbnailOnlyIfCached) != 0 {
		flags |= SIIGBF_INCACHEONLY | SIIGBF_MEMORYONLY
	}

	if (options & ThumbnailIconFallback) == 0 {
		flags |= SIIGBF_THUMBNAILONLY
	}

	if (options & ThumbnailHighQuality) != 0 {
		flags |= SIIGBF_BIGGERSIZEOK
	}

	// Get the thumbnail with max dimension of 512px
	// Windows will preserve aspect ratio with SIIGBF_RESIZETOFIT
	sizeStruct := SIZE{cx: int32(actualSize), cy: int32(actualSize)}
	hBitmap, err := imageFactory.GetImage(sizeStruct, flags)
	if err != nil {
		return nil, fmt.Errorf("failed to get thumbnail: %v", err)
	}
	defer procDeleteObject.Call(hBitmap)

	// Get the actual dimensions of the returned bitmap (preserves aspect ratio)
	width, height, err := getBitmapDimensions(hBitmap)
	if err != nil {
		return nil, fmt.Errorf("failed to get bitmap dimensions: %v", err)
	}

	// Convert HBITMAP to PNG using actual dimensions
	pngBytes, err := hBitmapToPNG(hBitmap, width, height)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bitmap to PNG: %v", err)
	}

	return pngBytes, nil
}

// GetCachedThumbnail attempts to get a cached thumbnail without generating a new one
func GetCachedThumbnail(filePath string, size int) ([]byte, error) {
	return GetOSThumbnail(filePath, size, ThumbnailOnlyIfCached|ThumbnailUseCurrentScale)
}
