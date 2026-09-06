//go:build linux && cgo

// Copyright (c) 2026 Eaxum.
// Adapted from drag-rs; see NOTICE and LICENSE.drag-rs.
#include <gtk/gtk.h>
#include <string.h>
#include "dragout_linux.h"

static const gint ClusttaLeftMouseButton = 1;
static const gint ClusttaCurrentPointerPosition = -1;
static const guint ClusttaURITarget = 0;

typedef struct {
    GtkWidget *window;
    GdkDragContext *context;
    gchar **uris;
    uintptr_t callback;
    gboolean failed;
    gboolean finishing;
    int result;
} ClusttaExportDrag;

static ClusttaExportDrag *activeDrag;

static gboolean finishDrag(gpointer data) {
    ClusttaExportDrag *drag = data;
    uintptr_t callback = drag->callback;
    int result = drag->result;
    g_signal_handlers_disconnect_by_data(drag->window, drag);
    g_clear_object(&drag->context);
    g_object_unref(drag->window);
    g_strfreev(drag->uris);
    activeDrag = NULL;
    g_free(drag);
    clustta_linux_drag_completed(callback, result);
    return G_SOURCE_REMOVE;
}

static void completeDrag(ClusttaExportDrag *drag, int result) {
    if (drag->finishing) return;
    drag->finishing = TRUE;
    drag->result = result;
    // Cleanup after GTK's signal emission, including synchronous startup failures.
    g_idle_add(finishDrag, drag);
}

static void dragDataGet(GtkWidget *widget, GdkDragContext *context,
    GtkSelectionData *selection, guint info, guint time, gpointer data) {
    ClusttaExportDrag *drag = data;
    if (context == drag->context && info == ClusttaURITarget) {
        gtk_selection_data_set_uris(selection, drag->uris);
    }
}

static gboolean dragFailed(GtkWidget *widget, GdkDragContext *context,
    GtkDragResult result, gpointer data) {
    ClusttaExportDrag *drag = data;
    if (drag->context && context != drag->context) return FALSE;
    drag->failed = TRUE;
    return TRUE;
}

static void dragEnded(GtkWidget *widget, GdkDragContext *context, gpointer data) {
    ClusttaExportDrag *drag = data;
    if (drag->context && context != drag->context) return;
    gboolean copied = !drag->failed && gdk_drag_context_get_selected_action(context) == GDK_ACTION_COPY;
    completeDrag(drag, copied ? ClusttaLinuxDragDropped : ClusttaLinuxDragCancelled);
}

static void windowDestroyed(GtkWidget *widget, gpointer data) {
    ClusttaExportDrag *drag = data;
    if (drag->finishing) return;
    drag->failed = TRUE;
    if (drag->context) gtk_drag_cancel(drag->context);
}

static gboolean validWindow(GtkWidget *window) {
    if (!window || !g_main_context_is_owner(g_main_context_default())) return FALSE;
    GList *windows = gtk_window_list_toplevels();
    gboolean found = g_list_find(windows, window) != NULL;
    g_list_free(windows);
    return found && gtk_widget_get_mapped(window) && gtk_window_is_active(GTK_WINDOW(window));
}

void clustta_linux_begin_drag(void *windowPointer, const char *uris, size_t count, uintptr_t callback) {
    GtkWidget *window = windowPointer;
    if (!validWindow(window)) {
        clustta_linux_drag_completed(callback, ClusttaLinuxDragInvalidWindow);
        return;
    }
    if (activeDrag) {
        clustta_linux_drag_completed(callback, ClusttaLinuxDragBusy);
        return;
    }
    GdkSeat *seat = gdk_display_get_default_seat(gtk_widget_get_display(window));
    GdkDevice *pointer = seat ? gdk_seat_get_pointer(seat) : NULL;
    GdkModifierType modifiers = 0;
    if (pointer) {
        gdk_window_get_device_position(gtk_widget_get_window(window), pointer, NULL, NULL, &modifiers);
    }
    if (!pointer || !(modifiers & GDK_BUTTON1_MASK)) {
        clustta_linux_drag_completed(callback, ClusttaLinuxDragCancelled);
        return;
    }
    ClusttaExportDrag *drag = g_new0(ClusttaExportDrag, 1);
    drag->window = g_object_ref(window);
    drag->callback = callback;
    drag->uris = g_new0(gchar *, count + 1);
    for (size_t index = 0; index < count; index++) {
        drag->uris[index] = g_strdup(uris);
        uris += strlen(uris) + 1;
    }
    activeDrag = drag;
    g_signal_connect(window, "drag-data-get", G_CALLBACK(dragDataGet), drag);
    g_signal_connect(window, "drag-failed", G_CALLBACK(dragFailed), drag);
    g_signal_connect(window, "drag-end", G_CALLBACK(dragEnded), drag);
    g_signal_connect(window, "destroy", G_CALLBACK(windowDestroyed), drag);

    GtkTargetList *targets = gtk_target_list_new(NULL, 0);
    gtk_target_list_add_uri_targets(targets, ClusttaURITarget);
    // The webview binding is asynchronous, so no original GDK event is available.
    GdkDragContext *context = gtk_drag_begin_with_coordinates(window, targets, GDK_ACTION_COPY,
        ClusttaLeftMouseButton, NULL, ClusttaCurrentPointerPosition, ClusttaCurrentPointerPosition);
    gtk_target_list_unref(targets);
    if (!context) {
        completeDrag(drag, ClusttaLinuxDragFailed);
        return;
    }
    drag->context = g_object_ref(context);
    if (!drag->finishing) gtk_drag_set_icon_default(context);
}
