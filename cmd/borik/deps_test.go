package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestBorikDoesNotDependOnImagick(t *testing.T) {
	output, err := exec.Command("go", "list", "-deps", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("listing cmd/borik dependencies: %v\n%s", err, output)
	}

	for dep := range strings.FieldsSeq(string(output)) {
		if strings.Contains(dep, "gopkg.in/gographics/imagick") {
			t.Fatalf("cmd/borik must not depend on imagick, found %q", dep)
		}
	}
}
