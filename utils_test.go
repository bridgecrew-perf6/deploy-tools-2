package main

import (
	"fmt"
	"testing"
)

func TestGetPidByProcessName(t *testing.T) {
	name := GetPidByProcessName("java")
	fmt.Printf("%+v", name)
}
