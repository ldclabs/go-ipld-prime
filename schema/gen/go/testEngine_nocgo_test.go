// +build !cgo,!skipgenbehavtests

// Confession:
// This build tag specification is NOT sufficient nor necessarily correct --
// it's a vague approximation of what's present in the stdlib 'plugin' package.
// It's also not at all a sure thing that cgo will actually *work* just
// because a build tag hasn't explicitly stated that it *mayn't* -- cgo can
// and will fail for environmental reasons at the point the compiler uses it.
//
// Ideally, there'd be a way to *ask* the plugin package if it's going to
// work or not before we try to use it; unfortunately, at the time of writing,
// it does not appear there is such an ability.
//
// If you run afoul of these build tags somehow (e.g., building plugins isn't
// possible in your environment for some reason), use the 'skipgenbehavtests'
// build tag to right yourself.  That's what it's there for.

package gengo

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
)

func buildGennedCode(t *testing.T, prefix string, pkgName string) {
	// Emit a small file with a 'main' method.
	//  'go build' doesn't like it we're in a package called "main" and there isn't one
	//   (and at the same time, plugins demand that they be in a package called 'main',
	//    so 'pkgName' in practice is almost always "main").
	//  I dunno, friend.  I didn't write the rules.
	if pkgName == "main" {
		withFile(prefix+"/main.go", func(w io.Writer) {
			fmt.Fprintf(w, "package %s\n\n", pkgName)
			fmt.Fprintf(w, "func main() {}\n")
		})
	}

	// Invoke 'go build' -- nothing fancy.
	cmd := exec.Command("go", "build", "./_test/"+prefix)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Fatalf("genned code failed to compile: %s", err)
	}
}

func runBehavioralTests(t *testing.T, _ string, _ behavioralTests) {
	t.Run("bhvtest=skip", func(t *testing.T) {
		t.Skip("behavioral tests for generated code skipped: cgo is required for these tests")
	})
}
