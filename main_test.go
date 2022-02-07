package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		word   string
		input  string
		output string
	}{{
		word:   "BEACH",
		input:  "see",
		output: "SEE is not a 5-letter word. Try again...\n",
	}, {
		word:  "BEACH",
		input: "beach",
		output: strings.Join([]string{
			"\033[42m\033[1;30m B \033[m\033[m",
			"\033[42m\033[1;30m E \033[m\033[m",
			"\033[42m\033[1;30m A \033[m\033[m",
			"\033[42m\033[1;30m C \033[m\033[m",
			"\033[42m\033[1;30m H \033[m\033[m",
			"\n",
		}, ""),
	}, {
		word:  "BEATS",
		input: "burst",
		output: strings.Join([]string{
			"\033[42m\033[1;30m B \033[m\033[m",
			"\033[40m\033[1;37m U \033[m\033[m",
			"\033[40m\033[1;37m R \033[m\033[m",
			"\033[43m\033[1;30m S \033[m\033[m",
			"\033[43m\033[1;30m T \033[m\033[m",
			"\n",
		}, ""),
	}, {
		word:  "BOOTY",
		input: "raise",
		output: strings.Join([]string{
			"\033[40m\033[1;37m R \033[m\033[m",
			"\033[40m\033[1;37m A \033[m\033[m",
			"\033[40m\033[1;37m I \033[m\033[m",
			"\033[40m\033[1;37m S \033[m\033[m",
			"\033[40m\033[1;37m E \033[m\033[m",
			"\n",
		}, ""),
	}}

	for _, test := range tests {
		t.Run(fmt.Sprintf("when '%s' is provided as input", test.input), func(t *testing.T) {
			var command, result bytes.Buffer
			defer result.Reset()

			fmt.Fprintf(&command, fmt.Sprintf("%s\n", test.input))
			fmt.Fprintf(&command, "stop\n")

			run(test.word, &command, &result)

			got := result.String()
			if !strings.Contains(got, test.output) {
				t.Errorf("expected '%s' to produce output '%s'; got '%s'", test.input, test.output, got)
			}
		})
	}
}
