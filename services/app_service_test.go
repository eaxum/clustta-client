package services

import "testing"

func TestDispatchDeepLinkRejectsOtherSchemes(t *testing.T) {
	if err := DispatchDeepLink("https://clustta.com"); err == nil {
		t.Fatal("expected a non-Clustta URL to be rejected")
	}
}

func TestDispatchDeepLinkQueuesColdStart(t *testing.T) {
	service := &AppService{}
	service.GetPendingDeepLink()
	t.Cleanup(func() {
		service.GetPendingDeepLink()
	})

	deepLink := "clustta://open?studio=Studio&project=project-id"
	if err := DispatchDeepLink(deepLink); err != nil {
		t.Fatal(err)
	}
	if pending := service.GetPendingDeepLink(); pending != deepLink {
		t.Fatalf("expected %q, got %q", deepLink, pending)
	}
	if pending := service.GetPendingDeepLink(); pending != "" {
		t.Fatalf("expected the pending link to be consumed, got %q", pending)
	}
}
