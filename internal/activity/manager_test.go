package activity

import (
	"context"
	"errors"
	"net/url"
	"testing"
)

func TestQueueSerializesProjectsAndSkipsCancelledActions(t *testing.T) {
	m := NewManager(nil)
	a, ctxA := m.Start(context.Background(), Operation{ProjectURI: "project-a"})
	b, ctxB := m.Start(context.Background(), Operation{ProjectURI: "project-b"})
	c, ctxC := m.Start(context.Background(), Operation{ProjectURI: "project-c"})
	if err := m.Wait(ctxA, a); err != nil {
		t.Fatal(err)
	}
	items := m.List().Operations
	if items[0].Status != Running || items[1].Status != Queued || items[2].Status != Queued {
		t.Fatal(items)
	}
	if err := m.Cancel(b); err != nil {
		t.Fatal(err)
	}
	if err := m.Wait(ctxB, b); !errors.Is(err, context.Canceled) {
		t.Fatalf("queued cancellation: %v", err)
	}
	m.Dismiss("")
	if len(m.List().Operations) != 2 {
		t.Fatal("clear finished removed live queue entries")
	}
	if err := m.Cancel(a); err != nil {
		t.Fatal(err)
	}
	if m.List().Operations[1].Status != Queued {
		t.Fatal("advanced before active action finished stopping")
	}
	m.Finish(a, ctxA.Err())
	if err := m.Wait(ctxC, c); err != nil {
		t.Fatal(err)
	}
	m.Finish(c, nil)
}

func TestOperationsHaveIndependentCancellationAndTerminalProgress(t *testing.T) {
	m := NewManager(nil)
	a, ctxA := m.Start(context.Background(), Operation{Title: "A"})
	b, ctxB := m.Start(context.Background(), Operation{Title: "B"})
	if err := m.Cancel(a); err != nil {
		t.Fatal(err)
	}
	if ctxA.Err() == nil || ctxB.Err() != nil {
		t.Fatal("cancellation was not isolated")
	}
	m.Finish(a, ctxA.Err())
	m.Update(a, func(item *Operation) { item.Percentage = 50 })
	m.Update(b, func(item *Operation) { item.Percentage = 42 })
	items := m.List().Operations
	if items[0].Status != Cancelled || items[1].Percentage != 42 {
		t.Fatalf("unexpected operation states: %+v", items)
	}
	m.Dismiss("")
	if items = m.List().Operations; len(items) != 1 || items[0].ID != b {
		t.Fatal("clear finished removed a running operation")
	}
	m.Finish(b, nil)
	_, _ = m.Start(context.Background(), Operation{Title: "C"})
}

func TestFailurePreservesProgressAndOmitsRequestURL(t *testing.T) {
	m := NewManager(nil)
	id, _ := m.Start(context.Background(), Operation{Percentage: 67})
	m.Finish(id, &url.Error{Op: "Get", URL: "https://server/chunk?signature=secret", Err: errors.New("network is unreachable")})
	item := m.List().Operations[0]
	if item.Status != Failed || item.Percentage != 67 || item.Error != "network is unreachable" {
		t.Fatalf("unexpected failure: %+v", item)
	}
	id, ctx := m.Start(context.Background(), Operation{Percentage: 25})
	if err := m.Cancel(id); err != nil {
		t.Fatal(err)
	}
	m.Finish(id, ctx.Err())
	if m.List().Operations[1].Percentage != 25 {
		t.Fatal("cancellation lost progress")
	}
}
