package main

import (
	"fmt"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func main() {
	dmp := diffmatchpatch.New()
	base := "Line 1: Hello\nLine 2: World\nLine 3: Goodbye\n"
	alice := "Line 1: Hello Alice\nLine 2: World\nLine 3: Goodbye\n"
	bob := "Line 1: Hello\nLine 2: World\nLine 3: Goodbye Bob\n"

	pBob := dmp.PatchMake(base, bob)

	// Alice changes applied first
	merged := alice

	// Apply Bob patch on top of Alice changes
	final, results := dmp.PatchApply(pBob, merged)

	fmt.Printf("Results: %v\n", results)
	fmt.Printf("Final:\n%s", final)
}
