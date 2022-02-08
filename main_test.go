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
		inputs []string
		output string
	}{{
		word:   "SEAKS",
		inputs: []string{"seaks"},
		output: "Guess a 5-letter word within 6 guesses...\n",
	}, {
		word:   "BEACH",
		inputs: []string{"see"},
		output: "SEE is not a 5-letter word. Try again...\n",
	}, {
		word:   "BEACH",
		inputs: []string{"beach"},
		output: strings.Join([]string{
			"\033[42m\033[1;30m B \033[m\033[m",
			"\033[42m\033[1;30m E \033[m\033[m",
			"\033[42m\033[1;30m A \033[m\033[m",
			"\033[42m\033[1;30m C \033[m\033[m",
			"\033[42m\033[1;30m H \033[m\033[m",
			"\n",
		}, ""),
	}, {
		word:   "BEATS",
		inputs: []string{"burst"},
		output: strings.Join([]string{
			"\033[42m\033[1;30m B \033[m\033[m",
			"\033[40m\033[1;37m U \033[m\033[m",
			"\033[40m\033[1;37m R \033[m\033[m",
			"\033[43m\033[1;30m S \033[m\033[m",
			"\033[43m\033[1;30m T \033[m\033[m",
			"\n",
		}, ""),
	}, {
		word:   "BOOTY",
		inputs: []string{"raise"},
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
		t.Run(fmt.Sprintf("the word is '%s' and '%s' is provided as input", test.word, strings.Join(test.inputs, ", ")), func(t *testing.T) {
			var command, result bytes.Buffer
			defer result.Reset()

			for _, input := range test.inputs {
				fmt.Fprintf(&command, fmt.Sprintf("%s\n", input))
			}

			fmt.Fprintf(&command, "stop\n")

			run(test.word, &command, &result)

			got := result.String()
			if !strings.Contains(got, test.output) {
				t.Errorf("expected '%s' to produce output '%s'; got '%s'", strings.Join(test.inputs, ","), test.output, got)
			}
		})
	}
}
