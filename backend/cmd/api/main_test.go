package main

import "testing"

func TestGreeting(t *testing.T) {
	got := greeting("Nexpulse")
	want := "Nexpulse is running!"
	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
