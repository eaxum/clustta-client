//go:build windows && cgo

// Copyright (c) 2026 Eaxum.
// Adapted from drag-rs; see NOTICE and LICENSE.drag-rs.
#define COBJMACROS
#include "dragout_windows.h"
#include <windows.h>
#include <shlobj.h>
#include <stdlib.h>

typedef struct {
    IDropSource base;
    LONG references;
    HWND window;
} FileDropSource;

static ULONG STDMETHODCALLTYPE sourceAddRef(IDropSource *self) {
    return InterlockedIncrement(&((FileDropSource *)self)->references);
}

static ULONG STDMETHODCALLTYPE sourceRelease(IDropSource *self) {
    ULONG remaining = InterlockedDecrement(&((FileDropSource *)self)->references);
    if (!remaining) free(self);
    return remaining;
}

static HRESULT STDMETHODCALLTYPE sourceQueryInterface(IDropSource *self, REFIID iid, void **object) {
    if (!object) return E_POINTER;
    *object = NULL;
    if (!IsEqualIID(iid, &IID_IUnknown) && !IsEqualIID(iid, &IID_IDropSource)) return E_NOINTERFACE;
    *object = self;
    sourceAddRef(self);
    return S_OK;
}

static HRESULT STDMETHODCALLTYPE sourceContinue(IDropSource *self, BOOL escape, DWORD keys) {
    if (escape || !IsWindow(((FileDropSource *)self)->window)) return DRAGDROP_S_CANCEL;
    return keys & MK_LBUTTON ? S_OK : DRAGDROP_S_DROP;
}

static HRESULT STDMETHODCALLTYPE sourceFeedback(IDropSource *self, DWORD effect) {
    return DRAGDROP_S_USEDEFAULTCURSORS;
}

static IDropSourceVtbl sourceMethods = {
    sourceQueryInterface, sourceAddRef, sourceRelease, sourceContinue, sourceFeedback
};

static HRESULT fileDataObject(const uint16_t *paths, size_t count, IDataObject **data) {
    if (!paths || !count || count > UINT_MAX) return E_INVALIDARG;
    PIDLIST_ABSOLUTE *items = calloc(count, sizeof(*items));
    if (!items) return E_OUTOFMEMORY;
    const wchar_t *path = (const wchar_t *)paths;
    HRESULT result = S_OK;
    for (size_t index = 0; index < count; ++index) {
        result = SHParseDisplayName(path, NULL, &items[index], 0, NULL);
        if (FAILED(result)) break;
        path += wcslen(path) + 1;
    }
    IShellItemArray *array = NULL;
    if (SUCCEEDED(result)) {
        result = SHCreateShellItemArrayFromIDLists((UINT)count, (PCIDLIST_ABSOLUTE *)items, &array);
    }
    if (SUCCEEDED(result)) {
        result = IShellItemArray_BindToHandler(array, NULL, &BHID_DataObject, &IID_IDataObject, (void **)data);
        IShellItemArray_Release(array);
    }
    for (size_t index = 0; index < count; ++index) CoTaskMemFree(items[index]);
    free(items);
    return result;
}

int32_t clustta_drag_files(uintptr_t handle, const uint16_t *paths, size_t count) {
    HWND window = (HWND)handle;
    if (!IsWindow(window) || GetForegroundWindow() != window || GetAsyncKeyState(VK_LBUTTON) >= 0) {
        return DRAGDROP_S_CANCEL;
    }
    HRESULT result = OleInitialize(NULL);
    if (FAILED(result)) return result;
    IDataObject *data = NULL;
    result = fileDataObject(paths, count, &data);
    if (SUCCEEDED(result)) {
        FileDropSource *source = calloc(1, sizeof(*source));
        if (!source) {
            result = E_OUTOFMEMORY;
        } else {
            source->base.lpVtbl = &sourceMethods;
            source->references = 1;
            source->window = window;
            DWORD effect = DROPEFFECT_NONE;
            if (GetAsyncKeyState(VK_LBUTTON) >= 0) {
                result = DRAGDROP_S_CANCEL;
            } else {
                result = DoDragDrop(data, &source->base, DROPEFFECT_COPY, &effect);
                if (result == DRAGDROP_S_DROP && effect != DROPEFFECT_COPY) result = DRAGDROP_S_CANCEL;
            }
            sourceRelease(&source->base);
        }
        IDataObject_Release(data);
    }
    OleUninitialize();
    return result;
}

int32_t clustta_probe_drag_files(const uint16_t *paths, size_t count) {
    HRESULT result = OleInitialize(NULL);
    if (FAILED(result)) return result;
    IDataObject *data = NULL;
    result = fileDataObject(paths, count, &data);
    if (SUCCEEDED(result)) {
        FORMATETC format = {CF_HDROP, NULL, DVASPECT_CONTENT, -1, TYMED_HGLOBAL};
        STGMEDIUM medium = {0};
        result = IDataObject_GetData(data, &format, &medium);
        if (SUCCEEDED(result)) {
            if (DragQueryFileW((HDROP)medium.hGlobal, UINT_MAX, NULL, 0) != count) result = E_FAIL;
            ReleaseStgMedium(&medium);
        }
        IDataObject_Release(data);
    }
    OleUninitialize();
    return result;
}
