package services

import (
	"context"
	"testing"
)

func TestDragOutRequiresDesktopWindow(t *testing.T) {
	service := &DragOutService{}
	if _, err := service.StartDrag(context.Background(), DragOutRequest{}); err == nil {
		t.Fatal("drag without a calling desktop window was accepted")
	}
	if service.active.Load() {
		t.Fatal("failed request left drag state active")
	}
}
