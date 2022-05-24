package utils

import "fmt"

func ExampleResolveAndStripSemver_prerelease() {
	vs, err := ResolveAndStripSemver("v1.2.3-beta.1")
	fmt.Println(err)
	fmt.Println(vs)
	// Output:
	// <nil>
	// [1.2.3-beta.1]
}
func ExampleResolveAndStripSemver_patch() {
	vs, err := ResolveAndStripSemver("v1.2.3")
	fmt.Println(err)
	fmt.Println(vs)
	// Output:
	// <nil>
	// [1.2.3 1.2 1]
}
func ExampleResolveAndStripSemver_minor() {
	vs, err := ResolveAndStripSemver("v1.2")
	fmt.Println(err)
	fmt.Println(vs)
	// Output:
	// <nil>
	// [1.2.0 1.2 1]
}
func ExampleResolveAndStripSemver_major() {
	vs, err := ResolveAndStripSemver("v1")
	fmt.Println(err)
	fmt.Println(vs)
	// Output:
	// <nil>
	// [1.0.0 1.0 1]
}
func ExampleResolveAndStripSemver_sample() {
	vs, err := ResolveAndStripSemver("2.0.4")
	fmt.Println(err)
	fmt.Println(vs)
	// Output:
	// <nil>
	// [2.0.4 2.0 2]
}
