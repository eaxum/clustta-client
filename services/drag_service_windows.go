//go:build windows

package services

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
)

var (
	ole32               = syscall.NewLazyDLL("ole32.dll")
	procOleInitialize   = ole32.NewProc("OleInitialize")
	procOleUninitialize = ole32.NewProc("OleUninitialize")
	procDoDragDrop      = ole32.NewProc("DoDragDrop")
)

const (
	DROPEFFECT_NONE = 0
	DROPEFFECT_COPY = 1

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
	IID_IUnknown       = GUID{0x00000000, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	IID_IDataObject    = GUID{0x0000010E, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	IID_IDropSource    = GUID{0x00000121, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
	IID_IEnumFORMATETC = GUID{0x00000103, 0x0000, 0x0000, [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
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
	enums     []*FileEnumFORMATETC // prevents GC of enumerators returned to COM callers
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
	ptrVal, _, err := procGlobalLock.Call(hGlobal)
	if ptrVal == 0 {
		procGlobalFree.Call(hGlobal)
		return 0, fmt.Errorf("GlobalLock failed: %v", err)
	}
	ptr := unsafe.Pointer(ptrVal)

	// Fill DROPFILES structure
	dropFiles := (*DROPFILES)(ptr)
	dropFiles.PFiles = uint32(unsafe.Sizeof(DROPFILES{}))
	dropFiles.FWide = 1 // Unicode paths

	// Copy file paths as null-terminated wide strings
	offset := unsafe.Sizeof(DROPFILES{})
	for _, path := range paths {
		pathPtr, _ := syscall.UTF16FromString(path)
		dst := unsafe.Add(ptr, offset)
		for i, c := range pathPtr {
			*(*uint16)(unsafe.Add(dst, i*2)) = c
		}
		offset += uintptr(len(pathPtr) * 2)
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
	return uintptr(obj.refCount)
}

func dataObjectGetData(this uintptr, pformatetc *FORMATETC, pmedium *STGMEDIUM) uintptr {
	obj := (*FileDataObject)(unsafe.Pointer(this))

	if pformatetc.CfFormat != CF_HDROP {
		return 0x80040064 // DV_E_FORMATETC
	}
	if pformatetc.Tymed&TYMED_HGLOBAL == 0 {
		return 0x80040069 // DV_E_TYMED
	}

	// Allocate a fresh buffer each call — caller owns the medium
	hGlobal, err := createDropFilesBuffer(obj.filePaths)
	if err != nil {
		return 0x80004005 // E_FAIL
	}

	pmedium.Tymed = TYMED_HGLOBAL
	pmedium.Data = hGlobal
	pmedium.PUnkForRelease = 0

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
	if fEscapePressed != 0 {
		return DRAGDROP_S_CANCEL
	}
	// MK_LBUTTON = 0x0001, when not set means mouse released
	if grfKeyState&0x0001 == 0 {
		return DRAGDROP_S_DROP
	}
	return S_OK
}

func dropSourceGiveFeedback(this uintptr, dwEffect uint32) uintptr {
	return 0x00040102 // DRAGDROP_S_USEDEFAULTCURSORS
}

// IEnumFORMATETC implementation for enumerating supported clipboard formats.

// IEnumFORMATETCVtbl is the virtual function table for IEnumFORMATETC.
type IEnumFORMATETCVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Next           uintptr
	Skip           uintptr
	Reset          uintptr
	Clone          uintptr
}

// FileEnumFORMATETC enumerates the formats supported by FileDataObject.
type FileEnumFORMATETC struct {
	vtbl     *IEnumFORMATETCVtbl
	refCount int32
	index    int32
}

// Supported format: CF_HDROP delivered via HGLOBAL.
var supportedFormat = FORMATETC{
	CfFormat: CF_HDROP,
	Ptd:      0,
	DwAspect: DVASPECT_CONTENT,
	Lindex:   -1,
	Tymed:    TYMED_HGLOBAL,
}

func enumFmtQueryInterface(this uintptr, riid *GUID, ppvObject *uintptr) uintptr {
	if guidEqual(riid, &IID_IUnknown) || guidEqual(riid, &IID_IEnumFORMATETC) {
		*ppvObject = this
		enumFmtAddRef(this)
		return S_OK
	}
	*ppvObject = 0
	return E_NOINTERFACE
}

func enumFmtAddRef(this uintptr) uintptr {
	obj := (*FileEnumFORMATETC)(unsafe.Pointer(this))
	obj.refCount++
	return uintptr(obj.refCount)
}

func enumFmtRelease(this uintptr) uintptr {
	obj := (*FileEnumFORMATETC)(unsafe.Pointer(this))
	obj.refCount--
	return uintptr(obj.refCount)
}

// enumFmtNext returns the next format(s) in the enumeration.
func enumFmtNext(this uintptr, celt uint32, rgelt *FORMATETC, pceltFetched *uint32) uintptr {
	obj := (*FileEnumFORMATETC)(unsafe.Pointer(this))

	if obj.index >= 1 || celt == 0 {
		if pceltFetched != nil {
			*pceltFetched = 0
		}
		return 1 // S_FALSE
	}

	*rgelt = supportedFormat
	obj.index++

	if pceltFetched != nil {
		*pceltFetched = 1
	}
	if celt == 1 {
		return S_OK
	}
	return 1 // S_FALSE
}

// enumFmtSkip advances the enumeration by celt entries.
func enumFmtSkip(this uintptr, celt uint32) uintptr {
	obj := (*FileEnumFORMATETC)(unsafe.Pointer(this))
	obj.index += int32(celt)
	if obj.index > 1 {
		obj.index = 1
		return 1 // S_FALSE
	}
	return S_OK
}

// enumFmtReset resets the enumeration to the beginning.
func enumFmtReset(this uintptr) uintptr {
	obj := (*FileEnumFORMATETC)(unsafe.Pointer(this))
	obj.index = 0
	return S_OK
}

// enumFmtClone creates a copy of the enumerator with the same state.
func enumFmtClone(this uintptr, ppEnum *uintptr) uintptr {
	obj := (*FileEnumFORMATETC)(unsafe.Pointer(this))
	clone := &FileEnumFORMATETC{
		vtbl:     enumFmtVtblInstance,
		refCount: 1,
		index:    obj.index,
	}
	pinComObject(clone)
	*ppEnum = uintptr(unsafe.Pointer(clone))
	return S_OK
}

// dataObjectEnumFormatEtc returns an enumerator listing supported formats.
func dataObjectEnumFormatEtc(this uintptr, dwDirection uint32, ppEnumFORMATETC *uintptr) uintptr {
	obj := (*FileDataObject)(unsafe.Pointer(this))
	if dwDirection != 1 { // DATADIR_GET
		return E_NOTIMPL
	}
	enum := &FileEnumFORMATETC{
		vtbl:     enumFmtVtblInstance,
		refCount: 1,
		index:    0,
	}
	// Pin on parent object and global list to prevent GC
	obj.enums = append(obj.enums, enum)
	pinComObject(enum)
	*ppEnumFORMATETC = uintptr(unsafe.Pointer(enum))
	return S_OK
}

func guidEqual(g1, g2 *GUID) bool {
	return g1.Data1 == g2.Data1 && g1.Data2 == g2.Data2 && g1.Data3 == g2.Data3 && g1.Data4 == g2.Data4
}

// pinnedComObjects prevents Go GC from collecting COM objects whose pointers
// have been handed to Windows via uintptr. Cleared after each drag operation.
var (
	pinnedMu         sync.Mutex
	pinnedComObjects []interface{}
)

// pinComObject stores a reference to prevent garbage collection.
func pinComObject(obj interface{}) {
	pinnedMu.Lock()
	pinnedComObjects = append(pinnedComObjects, obj)
	pinnedMu.Unlock()
}

// unpinAllComObjects releases pinned references so objects can be collected.
func unpinAllComObjects() {
	pinnedMu.Lock()
	pinnedComObjects = nil
	pinnedMu.Unlock()
}

// Global callback references to prevent garbage collection
var (
	dataObjectVtblInstance *IDataObjectVtbl
	dropSourceVtblInstance *IDropSourceVtbl
	enumFmtVtblInstance    *IEnumFORMATETCVtbl
)

func init() {
	// Initialize vtables with callback functions
	enumFmtVtblInstance = &IEnumFORMATETCVtbl{
		QueryInterface: syscall.NewCallback(enumFmtQueryInterface),
		AddRef:         syscall.NewCallback(enumFmtAddRef),
		Release:        syscall.NewCallback(enumFmtRelease),
		Next:           syscall.NewCallback(enumFmtNext),
		Skip:           syscall.NewCallback(enumFmtSkip),
		Reset:          syscall.NewCallback(enumFmtReset),
		Clone:          syscall.NewCallback(enumFmtClone),
	}

	dataObjectVtblInstance = &IDataObjectVtbl{
		QueryInterface:        syscall.NewCallback(dataObjectQueryInterface),
		AddRef:                syscall.NewCallback(dataObjectAddRef),
		Release:               syscall.NewCallback(dataObjectRelease),
		GetData:               syscall.NewCallback(dataObjectGetData),
		GetDataHere:           syscall.NewCallback(func(uintptr, *FORMATETC, *STGMEDIUM) uintptr { return E_NOTIMPL }),
		QueryGetData:          syscall.NewCallback(dataObjectQueryGetData),
		GetCanonicalFormatEtc: syscall.NewCallback(func(uintptr, *FORMATETC, *FORMATETC) uintptr { return E_NOTIMPL }),
		SetData:               syscall.NewCallback(func(uintptr, *FORMATETC, *STGMEDIUM, int32) uintptr { return E_NOTIMPL }),
		EnumFormatEtc:         syscall.NewCallback(dataObjectEnumFormatEtc),
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

// normalizeFilePaths cleans and validates file paths for CF_HDROP.
func normalizeFilePaths(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = filepath.Clean(p)
		if !filepath.IsAbs(p) {
			continue
		}
		result = append(result, p)
	}
	return result
}

// StartNativeDrag initiates a native Windows drag-and-drop operation.
// This function blocks until the user completes or cancels the drag.
// Returns the drop effect (copy, move, link, or none).
func (d *DragService) StartNativeDrag(filePaths []string) (int, error) {
	normalized := normalizeFilePaths(filePaths)
	if len(normalized) == 0 {
		return DROPEFFECT_NONE, nil
	}

	// Check if mouse is still down - if released, DoDragDrop will fail immediately
	if !d.IsMouseButtonDown() {
		return DROPEFFECT_NONE, nil
	}

	// DoDragDrop must run on the main UI thread where the message pump
	// is active. Dispatch via Wails' InvokeSyncWithResultAndError.
	result, err := application.InvokeSyncWithResultAndError(func() (int, error) {
		return d.doDragDropOnMainThread(normalized)
	})

	return result, err
}

// doDragDropOnMainThread performs the actual COM drag-drop operation.
// Must be called on the main UI thread.
func (d *DragService) doDragDropOnMainThread(filePaths []string) (int, error) {
	// Initialize OLE on this thread for drag-drop support.
	// The main thread may have COM initialized but not OLE specifically.
	hr, _, _ := procOleInitialize.Call(0)
	if hr != S_OK && hr != 1 { // 1 = S_FALSE (already initialized)
		return DROPEFFECT_NONE, fmt.Errorf("OleInitialize failed with HRESULT: 0x%X", hr)
	}
	// Note: Don't call OleUninitialize here - the main thread may need OLE for other operations

	dataObject := &FileDataObject{
		vtbl:      dataObjectVtblInstance,
		refCount:  1,
		filePaths: filePaths,
	}

	dropSource := &FileDropSource{
		vtbl:     dropSourceVtblInstance,
		refCount: 1,
	}

	// Pin COM objects to prevent GC during the blocking DoDragDrop call
	pinComObject(dataObject)
	pinComObject(dropSource)

	// Call DoDragDrop — blocks until drag completes or is cancelled
	var dwEffect uint32
	hr, _, _ = procDoDragDrop.Call(
		uintptr(unsafe.Pointer(dataObject)),
		uintptr(unsafe.Pointer(dropSource)),
		DROPEFFECT_COPY, // Only allow copy, not move
		uintptr(unsafe.Pointer(&dwEffect)),
	)

	// Release pinned references now that DoDragDrop has returned
	unpinAllComObjects()
	runtime.KeepAlive(dataObject)
	runtime.KeepAlive(dropSource)

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
