// Package clock provides a tiny time abstraction.
//
// Production code should depend on the Clocker interface instead of calling
// time.Now() directly. This makes business logic easier to test because you can
// swap in a fake clock that returns a deterministic time.
package clock
