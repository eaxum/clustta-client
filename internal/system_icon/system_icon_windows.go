package system_icon

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"syscall"
	"unsafe"
)

// Windows API structs
type SHFILEINFO struct {
	hIcon         uintptr
	iIcon         int32
	dwAttributes  uint32
	szDisplayName [260]uint16
	szTypeName    [80]uint16
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

var (
	// Existing DLL and proc declarations...
	shell32                    = syscall.NewLazyDLL("shell32.dll")
	user32                     = syscall.NewLazyDLL("user32.dll")
	gdi32                      = syscall.NewLazyDLL("gdi32.dll")
	procSHGetFileInfoW         = shell32.NewProc("SHGetFileInfoW")
	procGetDC                  = user32.NewProc("GetDC")
	procCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	procSelectObject           = gdi32.NewProc("SelectObject")
	procDrawIcon               = user32.NewProc("DrawIcon")
	procGetDIBits              = gdi32.NewProc("GetDIBits")
	procDeleteDC               = gdi32.NewProc("DeleteDC")
	procReleaseDC              = user32.NewProc("ReleaseDC")
	procDeleteObject           = gdi32.NewProc("DeleteObject")
	procDestroyIcon            = user32.NewProc("DestroyIcon")

	// New procs for DPI awareness
	procGetDpiForWindow         = user32.NewProc("GetDpiForWindow")
	procGetDesktopWindow        = user32.NewProc("GetDesktopWindow")
	shcore                      = syscall.NewLazyDLL("Shcore.dll")
	procGetScaleFactorForDevice = shcore.NewProc("GetScaleFactorForDevice")
)

const (
	// Existing constants...
	SHGFI_ICON              = 0x000000100
	SHGFI_SMALLICON         = 0x000000001
	SHGFI_LARGEICON         = 0x000000000
	SHGFI_USEFILEATTRIBUTES = 0x000000010
	FILE_ATTRIBUTE_NORMAL   = 0x00000080
	DIB_RGB_COLORS          = 0
	BI_RGB                  = 0

	// DPI constants
	MDT_EFFECTIVE_DPI = 0
	DEVICE_PRIMARY    = 0
)

// getDpiScale returns the system DPI scale factor
func getDpiScale() float64 {
	// Get desktop window handle
	hWnd, _, _ := procGetDesktopWindow.Call()

	// Try GetDpiForWindow first (Windows 10 1607 and later)
	if dpi, _, _ := procGetDpiForWindow.Call(hWnd); dpi != 0 {
		return float64(dpi) / 96.0
	}

	// Fallback to GetScaleFactorForDevice
	var factor uint32
	ret, _, _ := procGetScaleFactorForDevice.Call(DEVICE_PRIMARY, uintptr(unsafe.Pointer(&factor)))
	if ret == 0 {
		return 1.0 // Default to 100% if all methods fail
	}

	return float64(factor) / 100.0
}

// getScaledIconSize returns the appropriate icon size based on DPI
func getScaledIconSize(baseSize int) int {
	scale := getDpiScale()
	return int(float64(baseSize) * scale)
}

func GetFileExtensionIcon(extension string, large bool) (uintptr, error) {
	if extension[0] != '.' {
		extension = "." + extension
	}

	extensionPtr, err := syscall.UTF16PtrFromString(extension)
	if err != nil {
		return 0, fmt.Errorf("failed to convert extension to UTF16: %v", err)
	}

	fileInfo := SHFILEINFO{}
	flags := SHGFI_ICON | SHGFI_USEFILEATTRIBUTES
	if large {
		flags |= SHGFI_LARGEICON
	} else {
		flags |= SHGFI_SMALLICON
	}

	ret, _, err := procSHGetFileInfoW.Call(
		uintptr(unsafe.Pointer(extensionPtr)),
		FILE_ATTRIBUTE_NORMAL,
		uintptr(unsafe.Pointer(&fileInfo)),
		uintptr(unsafe.Sizeof(fileInfo)),
		uintptr(flags),
	)

	if ret == 0 {
		return 0, fmt.Errorf("failed to get file info: %v", err)
	}

	return fileInfo.hIcon, nil
}

func iconToPNG(hIcon uintptr, baseSize int) (*image.RGBA, error) {
	// Calculate DPI-aware size
	size := getScaledIconSize(baseSize)

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

	// Create a compatible bitmap with scaled dimensions
	hBitmap, _, _ := procCreateCompatibleBitmap.Call(hDC, uintptr(size), uintptr(size))
	if hBitmap == 0 {
		return nil, fmt.Errorf("failed to create compatible bitmap")
	}
	defer procDeleteObject.Call(hBitmap)

	// Select the bitmap into the compatible DC
	prevObject, _, _ := procSelectObject.Call(hMemDC, hBitmap)
	defer procSelectObject.Call(hMemDC, prevObject)

	// Draw the icon onto the bitmap
	procDrawIcon.Call(hMemDC, 0, 0, hIcon)

	// Prepare the bitmap info with scaled dimensions
	bmi := BITMAPINFO{
		Header: BITMAPINFOHEADER{
			biSize:        uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
			biWidth:       int32(size),
			biHeight:      -int32(size), // Negative height for top-down bitmap
			biPlanes:      1,
			biBitCount:    32,
			biCompression: BI_RGB,
		},
	}

	// Create the output image with scaled dimensions
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Get the bitmap bits
	ret, _, _ := procGetDIBits.Call(
		hMemDC,
		hBitmap,
		0,
		uintptr(size),
		uintptr(unsafe.Pointer(&img.Pix[0])),
		uintptr(unsafe.Pointer(&bmi)),
		DIB_RGB_COLORS,
	)

	if ret == 0 {
		return nil, fmt.Errorf("failed to get bitmap bits")
	}

	// Convert BGR to RGB and set alpha channel
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+2] = img.Pix[i+2], img.Pix[i]
	}

	return img, nil
}

func GetExtensionIcon(extension string) ([]byte, error) {
	hIcon, err := GetFileExtensionIcon(extension, true)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to get icon: %v", err)
	}
	defer procDestroyIcon.Call(hIcon)

	// Use base sizes that will be scaled according to DPI

	baseSize := 32

	img, err := iconToPNG(hIcon, baseSize)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to convert icon to image: %v", err)
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to encode image as PNG: %v", err)
	}

	return buf.Bytes(), nil
}
