package sse

import "testing"

func TestHubPublishReachesSubscriber(t *testing.T) {
	h := NewHub()
	ch, unsubscribe := h.Subscribe("task-1")

	h.Publish("task-1", "done")
	if got := <-ch; got != "done" {
		t.Fatalf("got %q, want %q", got, "done")
	}

	h.Publish("other-task", "done")
	select {
	case got := <-ch:
		t.Fatalf("got %q for another task's update", got)
	default:
	}

	unsubscribe()
	if len(h.subs) != 0 {
		t.Fatalf("subs not cleaned up: %v", h.subs)
	}
}
