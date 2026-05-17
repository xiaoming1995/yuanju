package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_AllowsRateThenDenies(t *testing.T) {
	l := New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow("user-1") {
			t.Fatalf("call %d should be allowed", i+1)
		}
	}
	if l.Allow("user-1") {
		t.Fatal("4th call should be denied")
	}
}

func TestLimiter_PerKeyIsolated(t *testing.T) {
	l := New(1, time.Minute)
	if !l.Allow("user-a") {
		t.Fatal("user-a first call should be allowed")
	}
	if !l.Allow("user-b") {
		t.Fatal("user-b first call should be allowed even when user-a is over limit")
	}
	if l.Allow("user-a") {
		t.Fatal("user-a second call should be denied")
	}
}

func TestLimiter_WindowResets(t *testing.T) {
	l := New(2, 50*time.Millisecond)
	l.Allow("user-1")
	l.Allow("user-1")
	if l.Allow("user-1") {
		t.Fatal("3rd call within window should be denied")
	}
	time.Sleep(70 * time.Millisecond)
	if !l.Allow("user-1") {
		t.Fatal("after window reset, call should be allowed")
	}
}
