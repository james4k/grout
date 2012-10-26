package grout

import (
	"testing"
)

func TestBlerg(t *testing.T) {
	Build("test", "", &Options{Verbose: true})
}
