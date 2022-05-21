package utils

import "fmt"

func ExampleResolveSemver_prerelease() {
	vs, err := ResolveSemver("v1.2.3-beta.1")
	fmt.Println(err)
	fmt.Println(vs)
	// Output:
	// <nil>
	// [1.2.3-beta.1]
}
func ExampleResolveSemver_patch() {
	vs, err := ResolveSemver("v1.2.3")
	fmt.Println(err)
	fmt.Println(vs)
	// Output:
	// <nil>
	// [1.2.3 1.2.0 1.0.0]
}
func ExampleResolveSemver_minor() {
	vs, err := ResolveSemver("v1.2")
	fmt.Println(err)
	fmt.Println(vs)
	// Output:
	// <nil>
	// [1.2.0 1.0.0]
}
func ExampleResolveSemver_major() {
	vs, err := ResolveSemver("v1")
	fmt.Println(err)
	fmt.Println(vs)
	// Output:
	// <nil>
	// [1.0.0]
}
