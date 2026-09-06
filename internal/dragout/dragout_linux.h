#pragma once

#include <stddef.h>
#include <stdint.h>

enum ClusttaLinuxDragStatus {
    ClusttaLinuxDragDropped,
    ClusttaLinuxDragCancelled,
    ClusttaLinuxDragInvalidWindow,
    ClusttaLinuxDragBusy,
    ClusttaLinuxDragFailed
};

void clustta_linux_begin_drag(void *window, const char *uris, size_t count, uintptr_t callback);
void clustta_linux_drag_completed(uintptr_t callback, int status);
