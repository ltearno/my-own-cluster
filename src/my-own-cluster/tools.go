package main

import (
	"fmt"
	"regexp"
	"unsafe"
)

var isLetter = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

func printStack(sp unsafe.Pointer, count int) {
	lineLength := 8

	chars := ""
	for i := 0; i < count; i++ {
		if i%lineLength == 0 {
			fmt.Printf("%08x: ", unsafe.Pointer(uintptr(sp)+uintptr(i)))
			chars = ""
		}

		b := *(*byte)(unsafe.Pointer(uintptr(sp) + uintptr(i)))
		fmt.Printf("%02x ", b)
		if isLetter(string(b)) {
			chars = chars + string(b)
		} else {
			chars = chars + "."
		}

		if (i+1)%lineLength == 0 {
			fmt.Printf(" %s\n", chars)
		}
	}

	fmt.Println()
}

func min(a int, b int) int {
	if a <= b {
		return a
	}

	return b
}
