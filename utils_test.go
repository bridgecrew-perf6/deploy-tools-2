package main

import "testing"

func TestGetPidByProcessName(t *testing.T) {
	GetPidByProcessName("electron")
}
