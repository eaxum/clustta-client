package handlers

import (
	"testing"
	"time"
)

func TestStartJobReusesIdempotencyKey(t *testing.T) {
	resetJobsForTest()

	first := startJob("checkpoint", "project:asset:key", false, func(func(string, int)) (any, error) {
		return map[string]string{"status": "ok"}, nil
	})
	second := startJob("checkpoint", "project:asset:key", false, func(func(string, int)) (any, error) {
		t.Fatal("duplicate operation should not run")
		return nil, nil
	})

	if first.ID != second.ID {
		t.Fatalf("expected duplicate request to reuse job %q, got %q", first.ID, second.ID)
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		job, err := getJob(first.ID)
		if err != nil {
			t.Fatal(err)
		}
		if job.Status == jobSucceeded {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("job did not complete")
}

func TestFailedJobMarksChangedSourceRetryable(t *testing.T) {
	resetJobsForTest()
	job := startJob("checkpoint", "changed-source", false, func(func(string, int)) (any, error) {
		return nil, errSourceFileChanged
	})

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		current, err := getJob(job.ID)
		if err != nil {
			t.Fatal(err)
		}
		if current.Status != jobFailed {
			time.Sleep(time.Millisecond)
			continue
		}
		if current.ErrorCode != "source_file_changed" || !current.Retryable {
			t.Fatalf("unexpected job error %#v", current)
		}
		return
	}
	t.Fatal("job did not fail")
}

func TestCancelRejectsNonCancellableJob(t *testing.T) {
	resetJobsForTest()
	job := startJob("checkpoint", "", false, func(func(string, int)) (any, error) {
		return nil, nil
	})

	if _, err := cancelJob(job.ID); err != errNotCancelable {
		t.Fatalf("expected errNotCancelable, got %v", err)
	}
}

func resetJobsForTest() {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	jobs = map[string]*bridgeJob{}
	idempotentJobs = map[string]string{}
}
