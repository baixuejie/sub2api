package main

import (
	"errors"
	"testing"
)

func TestShouldWaitForExit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		err  error
		want bool
	}{
		{name: "double click installation", want: true},
		{name: "successful protocol launch", args: []string{"sub2api-harness://bootstrap?server=https%3A%2F%2Fexample.com"}},
		{name: "failed protocol launch", args: []string{"sub2api-harness://bootstrap?server=https%3A%2F%2Fexample.com"}, err: errors.New("failed"), want: true},
		{name: "explicit registration", args: []string{"register-protocol"}},
		{name: "version", args: []string{"--version"}},
		{name: "invalid CLI usage", args: []string{"one", "two"}, err: errors.New("failed")},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := shouldWaitForExit(test.args, test.err); got != test.want {
				t.Fatalf("shouldWaitForExit() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNeedsInteractiveConsole(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "double click install", want: true},
		{name: "protocol launch", args: []string{"sub2api-harness://bootstrap?ticket=test"}, want: true},
		{name: "protocol launch case insensitive", args: []string{" SUB2API-HARNESS://bootstrap?ticket=test "}, want: true},
		{name: "explicit registration", args: []string{"register-protocol"}, want: false},
		{name: "version", args: []string{"--version"}, want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := needsInteractiveConsole(test.args); got != test.want {
				t.Fatalf("needsInteractiveConsole(%q) = %t, want %t", test.args, got, test.want)
			}
		})
	}
}
