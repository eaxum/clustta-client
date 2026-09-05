#pragma once
#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif
int32_t clustta_drag_files(uintptr_t window, const uint16_t *paths, size_t count);
int32_t clustta_probe_drag_files(const uint16_t *paths, size_t count);
#ifdef __cplusplus
}
#endif
