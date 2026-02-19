//go:build windows

package services

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	ole32                = syscall.NewLazyDLL("ole32.dll")
	shell32              = syscall.NewLazyDLL("shell32.dll")
	procOleInitialize    = ole32.NewProc("OleInitialize")
	procDoDragDrop       = ole32.NewProc("DoDragDrop")
	procReleaseStgMedium = ole32.NewProc("ReleaseStgMedium")
)

const (
	DROPEFFECT_NONE = 0
	DROPEFFECT_COPY = 1
	DROPEFFECT_MOVE = 2
	DROPEFFECT_LINK = 4

	CF_HDROP = 15

	TYMED_HGLOBAL = 1

	DVASPECT_CONTENT = 1

	S_OK              = 0
	E_NOTIMPL         = 0x80004001
	E_NOINTERFACE     = 0x80004002
	DRAGDROP_S_DROP   = 0x00040100
	DRAGDROP_S_CANCEL = 0x00040101
)

// GUID structure for COM interfaces.
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var (
	IID_IUnknown    = GUID{0x00000000, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	IID_IDataObject = GUID{0x0000010E, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	IID_IDropSource = GUID{0x00000121, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
)

// FORMATETC describes data format for clipboard/drag-drop.
type FORMATETC struct {
	CfFormat uint16
	Ptd      uintptr
	DwAspect uint32
	Lindex   int32
	Tymed    uint32
}

// STGMEDIUM describes storage medium for data transfer.
type STGMEDIUM struct {
	Tymed          uint32
	Data           uintptr
	PUnkForRelease uintptr
}

// DROPFILES structure for CF_HDROP format.
type DROPFILES struct {
	PFiles uint32
	Pt     struct{ X, Y int32 }
	FNC    int32
	FWide  int32
}

// IDataObjectVtbl is the virtual function table for IDataObject.
type IDataObjectVtbl struct {
	QueryInterface        uintptr
	AddRef                uintptr
	Release               uintptr
	GetData               uintptr
	GetDataHere           uintptr
	QueryGetData          uintptr
	GetCanonicalFormatEtc uintptr
	SetData               uintptr
	EnumFormatEtc         uintptr
	DAdvise               uintptr
	DUnadvise             uintptr
	EnumDAdvise           uintptr
}

// IDropSourceVtbl is the virtual function table for IDropSource.
type IDropSourceVtbl struct {
	QueryInterface    uintptr
	AddRef            uintptr
	Release           uintptr
	QueryContinueDrag uintptr
	GiveFeedback      uintptr
}

// FileDataObject implements IDataObject for file drag operations.
type FileDataObject struct {
	vtbl      *IDataObjectVtbl
	refCount  int32
	filePaths []string
	hGlobal   uintptr
}

// FileDropSource implements IDropSource for drag operations.
type FileDropSource struct {
	vtbl     *IDropSourceVtbl
	refCount int32
}

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procGlobalAlloc  = kernel32.NewProc("GlobalAlloc")
	procGlobalFree   = kernel32.NewProc("GlobalFree")
	procGlobalLock   = kernel32.NewProc("GlobalLock")
	procGlobalUnlock = kernel32.NewProc("GlobalUnlock")
	user32           = syscall.NewLazyDLL("user32.dll")
	procGetKeyState  = user32.NewProc("GetKeyState")
)

const (
	GMEM_MOVEABLE = 0x0002
	GMEM_ZEROINIT = 0x0040
	GHND          = GMEM_MOVEABLE | GMEM_ZEROINIT
	VK_LBUTTON    = 0x01
	VK_ESCAPE     = 0x1B
)

// createDropFilesBuffer creates a CF_HDROP format buffer from file paths.
func createDropFilesBuffer(paths []string) (uintptr, error) {
	// Calculate total size: DROPFILES struct + null-terminated wide strings + final null
	var totalSize int = int(unsafe.Sizeof(DROPFILES{}))
	for _, path := range paths {
		totalSize += (len(path) + 1) * 2 // UTF-16 + null terminator
	}
	totalSize += 2 // Final double-null terminator

	// Allocate global memory
	hGlobal, _, err := procGlobalAlloc.Call(GHND, uintptr(totalSize))
	if hGlobal == 0 {
		return 0, fmt.Errorf("GlobalAlloc failed: %v", err)
	}

	// Lock memory and get pointer
	ptr, _, err := procGlobalLock.Call(hGlobal)
	if ptr == 0 {
		procGlobalFree.Call(hGlobal)
		return 0, fmt.Errorf("GlobalLock failed: %v", err)
	}

	// Fill DROPFILES structure
	dropFiles := (*DROPFILES)(unsafe.Pointer(ptr))
	dropFiles.PFiles = uint32(unsafe.Sizeof(DROPFILES{}))
	dropFiles.FWide = 1 // Unicode paths

	// Copy file paths as null-terminated wide strings
	offset := unsafe.Sizeof(DROPFILES{})
	for _, path := range paths {
		pathPtr, _ := syscall.UTF16FromString(path)
		dst := unsafe.Pointer(ptr + offset)
		for i, c := range pathPtr {
			*(*uint16)(unsafe.Pointer(uintptr(dst) + uintptr(i*2))) = c
		}
		offset += uintptr((len(pathPtr)) * 2)
	}

	procGlobalUnlock.Call(hGlobal)
	return hGlobal, nil
}

// DataObject COM method implementations
func dataObjectQueryInterface(this uintptr, riid *GUID, ppvObject *uintptr) uintptr {
	if guidEqual(riid, &IID_IUnknown) || guidEqual(riid, &IID_IDataObject) {
		*ppvObject = this
		dataObjectAddRef(this)
		return S_OK
	}
	*ppvObject = 0
	return E_NOINTERFACE
}

func dataObjectAddRef(this uintptr) uintptr {
	obj := (*FileDataObject)(unsafe.Pointer(this))
	obj.refCount++
	return uintptr(obj.refCount)
}

func dataObjectRelease(this uintptr) uintptr {
	obj := (*FileDataObject)(unsafe.Pointer(this))
	obj.refCount--
	if obj.refCount == 0 {
		if obj.hGlobal != 0 {
			procGlobalFree.Call(obj.hGlobal)
		}
		return 0
	}
	return uintptr(obj.refCount)
}

func dataObjectGetData(this uintptr, pformatetc *FORMATETC, pmedium *STGMEDIUM) uintptr {
	obj := (*FileDataObject)(unsafe.Pointer(this))
	fmt.Printf("[DragService] GetData called: format=%d, tymed=%d\n", pformatetc.CfFormat, pformatetc.Tymed)

	if pformatetc.CfFormat != CF_HDROP {
		fmt.Printf("[DragService] GetData: wrong format, expected CF_HDROP (%d)\n", CF_HDROP)
		return 0x80040064 // DV_E_FORMATETC
	}
	if pformatetc.Tymed&TYMED_HGLOBAL == 0 {
		fmt.Println("[DragService] GetData: wrong tymed")
		return 0x80040069 // DV_E_TYMED
	}

	// Create the drop files buffer if not already created
	if obj.hGlobal == 0 {
		fmt.Println("[DragService] GetData: creating drop files buffer...")
		hGlobal, err := createDropFilesBuffer(obj.filePaths)
		if err != nil {
			fmt.Printf("[DragService] GetData: failed to create buffer: %v\n", err)
			return 0x80004005 // E_FAIL
		}
		obj.hGlobal = hGlobal
		fmt.Printf("[DragService] GetData: created buffer at 0x%X\n", hGlobal)
	}

	pmedium.Tymed = TYMED_HGLOBAL
	pmedium.Data = obj.hGlobal
	pmedium.PUnkForRelease = 0

	fmt.Println("[DragService] GetData: returning S_OK")
	return S_OK
}

func dataObjectQueryGetData(this uintptr, pformatetc *FORMATETC) uintptr {
	if pformatetc.CfFormat == CF_HDROP && pformatetc.Tymed&TYMED_HGLOBAL != 0 {
		return S_OK
	}
	return 0x80040064 // DV_E_FORMATETC
}

// DropSource COM method implementations
func dropSourceQueryInterface(this uintptr, riid *GUID, ppvObject *uintptr) uintptr {
	if guidEqual(riid, &IID_IUnknown) || guidEqual(riid, &IID_IDropSource) {
		*ppvObject = this
		dropSourceAddRef(this)
		return S_OK
	}
	*ppvObject = 0
	return E_NOINTERFACE
}

func dropSourceAddRef(this uintptr) uintptr {
	obj := (*FileDropSource)(unsafe.Pointer(this))
	obj.refCount++
	return uintptr(obj.refCount)
}

func dropSourceRelease(this uintptr) uintptr {
	obj := (*FileDropSource)(unsafe.Pointer(this))
	obj.refCount--
	return uintptr(obj.refCount)
}

func dropSourceQueryContinueDrag(this uintptr, fEscapePressed int32, grfKeyState uint32) uintptr {
	fmt.Printf("[DragService] QueryContinueDrag called: escape=%d, keyState=0x%X\n", fEscapePressed, grfKeyState)

	// Check if escape was pressed
	if fEscapePressed != 0 {
		fmt.Println("[DragService] Escape pressed, canceling drag")
		return DRAGDROP_S_CANCEL
	}

	// Check if mouse button was released (MK_LBUTTON = 0x0001)
	if grfKeyState&0x0001 == 0 {
		fmt.Println("[DragService] Mouse released, completing drop")
		return DRAGDROP_S_DROP
	}

	return S_OK
}

func dropSourceGiveFeedback(this uintptr, dwEffect uint32) uintptr {
	// DRAGDROP_S_USEDEFAULTCURSORS = 0x00040102
	return 0x00040102
}

func guidEqual(g1, g2 *GUID) bool {
	return g1.Data1 == g2.Data1 && g1.Data2 == g2.Data2 && g1.Data3 == g2.Data3 && g1.Data4 == g2.Data4
}

// Global callback references to prevent garbage collection
var (
	dataObjectVtblInstance *IDataObjectVtbl
	dropSourceVtblInstance *IDropSourceVtbl
)

func init() {
	// Initialize vtables with callback functions
	dataObjectVtblInstance = &IDataObjectVtbl{
		QueryInterface:        syscall.NewCallback(dataObjectQueryInterface),
		AddRef:                syscall.NewCallback(dataObjectAddRef),
		Release:               syscall.NewCallback(dataObjectRelease),
		GetData:               syscall.NewCallback(dataObjectGetData),
		GetDataHere:           syscall.NewCallback(func(uintptr, *FORMATETC, *STGMEDIUM) uintptr { return E_NOTIMPL }),
		QueryGetData:          syscall.NewCallback(dataObjectQueryGetData),
		GetCanonicalFormatEtc: syscall.NewCallback(func(uintptr, *FORMATETC, *FORMATETC) uintptr { return E_NOTIMPL }),
		SetData:               syscall.NewCallback(func(uintptr, *FORMATETC, *STGMEDIUM, int32) uintptr { return E_NOTIMPL }),
		EnumFormatEtc:         syscall.NewCallback(func(uintptr, uint32, *uintptr) uintptr { return E_NOTIMPL }),
		DAdvise:               syscall.NewCallback(func(uintptr, *FORMATETC, uint32, uintptr, *uint32) uintptr { return E_NOTIMPL }),
		DUnadvise:             syscall.NewCallback(func(uintptr, uint32) uintptr { return E_NOTIMPL }),
		EnumDAdvise:           syscall.NewCallback(func(uintptr, *uintptr) uintptr { return E_NOTIMPL }),
	}

	dropSourceVtblInstance = &IDropSourceVtbl{
		QueryInterface:    syscall.NewCallback(dropSourceQueryInterface),
		AddRef:            syscall.NewCallback(dropSourceAddRef),
		Release:           syscall.NewCallback(dropSourceRelease),
		QueryContinueDrag: syscall.NewCallback(dropSourceQueryContinueDrag),
		GiveFeedback:      syscall.NewCallback(dropSourceGiveFeedback),
	}
}

// StartNativeDrag initiates a native Windows drag-and-drop operation.
// This function blocks until the user completes or cancels the drag.
// Returns the drop effect (copy, move, link, or none).
func (d *DragService) StartNativeDrag(filePaths []string) (int, error) {
	if len(filePaths) == 0 {
		return DROPEFFECT_NONE, nil
	}

	fmt.Printf("[DragService] StartNativeDrag called with %d files: %v\n", len(filePaths), filePaths)

	// Check if mouse button is down
	mouseDown := d.IsMouseButtonDown()
	fmt.Printf("[DragService] Mouse button down: %v\n", mouseDown)

	// Initialize OLE (safe to call multiple times)
	hr, _, err := procOleInitialize.Call(0)
	fmt.Printf("[DragService] OleInitialize returned: 0x%X, err: %v\n", hr, err)

	// Create IDataObject
	dataObject := &FileDataObject{
		vtbl:      dataObjectVtblInstance,
		refCount:  1,
		filePaths: filePaths,
	}
	fmt.Printf("[DragService] Created dataObject at %p, vtbl at %p\n", dataObject, dataObject.vtbl)

	// Create IDropSource
	dropSource := &FileDropSource{
		vtbl:     dropSourceVtblInstance,
		refCount: 1,
	}
	fmt.Printf("[DragService] Created dropSource at %p, vtbl at %p\n", dropSource, dropSource.vtbl)

	// Call DoDragDrop - this blocks until drag completes
	var dwEffect uint32
	fmt.Printf("[DragService] Calling DoDragDrop...\n")
	hr, _, err = procDoDragDrop.Call(
		uintptr(unsafe.Pointer(dataObject)),
		uintptr(unsafe.Pointer(dropSource)),
		DROPEFFECT_COPY|DROPEFFECT_MOVE|DROPEFFECT_LINK,
		uintptr(unsafe.Pointer(&dwEffect)),
	)
	fmt.Printf("[DragService] DoDragDrop returned: hr=0x%X, effect=%d, err=%v\n", hr, dwEffect, err)

	// Clean up
	dataObjectRelease(uintptr(unsafe.Pointer(dataObject)))
	dropSourceRelease(uintptr(unsafe.Pointer(dropSource)))

	if hr != S_OK && hr != DRAGDROP_S_DROP && hr != DRAGDROP_S_CANCEL {
		return DROPEFFECT_NONE, fmt.Errorf("DoDragDrop failed with HRESULT: 0x%X", hr)
	}

	return int(dwEffect), nil
}

// IsMouseButtonDown checks if the left mouse button is currently pressed.
// Used by frontend to verify drag state before initiating native drag.
func (d *DragService) IsMouseButtonDown() bool {
	ret, _, _ := procGetKeyState.Call(VK_LBUTTON)
	// High bit set means key is down
	return int16(ret) < 0
}
