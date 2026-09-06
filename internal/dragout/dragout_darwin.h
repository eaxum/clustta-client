#pragma once

#include <stdint.h>
#include <stddef.h>

enum ClusttaDragStatus {
    ClusttaDragDropped,
    ClusttaDragCancelled,
    ClusttaDragInvalidWindow,
    ClusttaDragInvalidFile,
    ClusttaDragBusy,
    ClusttaDragFailed
};

void clustta_begin_drag(void *window, const char *paths, size_t count, uintptr_t callback);
void clustta_drag_completed(uintptr_t callback, int status);
