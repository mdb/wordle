package main

import (
	"bytes"
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{{
		input:  "see",
		output: "SEE is not a 5-letter word. Try again...\n",
	}}

	for _, test := range tests {
		t.Run(fmt.Sprintf("when '%s' is provided as input", test.input), func(t *testing.T) {
			var command, result bytes.Buffer

			fmt.Fprintf(&command, fmt.Sprintf("%s\n", test.input))
			fmt.Fprintf(&command, "stop\n")

			run(&command, &result)

			got := result.String()
			if got != test.output {
				t.Errorf("expected '%s' to produce output '%s'; got '%s'", test.input, test.output, got)
			}

		})
	}
}
