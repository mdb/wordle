package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	maxGuesses int = 6
	wordLength int = 5
)

type tileColor int

const (
	gray tileColor = iota
	yellow
	green
)

var (
	// passed in at build time
	version string
)

type wordle struct {
	wordLength int
	maxGuesses int
	word       string
	guesses    []map[string][wordLength]tileColor
	in         io.Reader
	out        io.Writer
}

func (w *wordle) displayRow(word string, colors [wordLength]tileColor) {
	for i, c := range word {
		switch colors[i] {
		case green:
			w.write("\033[42m\033[1;30m")
		case yellow:
			w.write("\033[43m\033[1;30m")
		case gray:
			w.write("\033[40m\033[1;37m")
		}

		w.write(fmt.Sprintf(" %c ", c))
		w.write("\033[m\033[m")
	}

	w.write("\n")
}

func (w *wordle) displayGrid(guess string, guessCount int) {
	tileColors := w.getLetterTileColors(guess)
	w.guesses = append(w.guesses, map[string][wordLength]tileColor{guess: tileColors})

	for _, guess := range w.guesses {
		for g, colors := range guess {
			w.displayRow(g, colors)
		}
	}

	w.displayEmptyRows(guessCount)
}

func (w *wordle) getLetterTileColors(guess string) [wordLength]tileColor {
	colors := [wordLength]tileColor{}

	for i := range colors {
		colors[i] = gray
	}

	for j, guessLetter := range guess {
		for k, letter := range w.word {
			if guessLetter == letter {
				if j == k {
					colors[j] = green
					break
				}

				colors[j] = yellow
			}
		}
	}

	return colors
}

func (w *wordle) displayEmptyRows(guessCount int) {
	emptyGuessChars := []string{}
	for i := 0; i < w.wordLength; i++ {
		emptyGuessChars = append(emptyGuessChars, "*")
	}

	emptyGuess := strings.Join(emptyGuessChars, "")
	emptyTileColors := w.getLetterTileColors(emptyGuess)
	emptyRowCount := w.maxGuesses - guessCount - 1

	for i := 0; i < emptyRowCount; i++ {
		w.displayRow(emptyGuess, emptyTileColors)
	}
}

func (w *wordle) write(str string) {
	w.out.Write([]byte(str))
}

func (w *wordle) run() {
	reader := bufio.NewScanner(w.in)

	w.write(fmt.Sprintf("Version: \t%s\n", version))
	w.write("Info: \t\thttps://github.com/mdb/wordle\n")
	w.write("About: \t\tA CLI adaptation of Josh Wardle's Wordle (https://powerlanguage.co.uk/wordle/)\n\n")
	w.write(fmt.Sprintf("Guess a %v-letter word within %v guesses...\n", w.wordLength, w.maxGuesses))

	for guessCount := 0; guessCount < w.maxGuesses; guessCount++ {
		w.write(fmt.Sprintf("\nGuess (%v/%v): ", len(w.guesses)+1, maxGuesses))

		reader.Scan()
		guess := strings.ToUpper(reader.Text())

		if guess == "STOP" {
			break
		}

		if len(guess) != len(w.word) {
			w.write(fmt.Sprintf("%s is not a %v-letter word. Try again...\n", guess, wordLength))
			guessCount--
		}

		if len(guess) == len(w.word) {
			w.displayGrid(guess, guessCount)
		}

		if guess == w.word {
			break
		}

		if guessCount == maxGuesses-1 {
			fmt.Println()
			w.displayRow(w.word, w.getLetterTileColors(w.word))
			os.Exit(1)
		}
	}
}

func newWordle(word string, in io.Reader, out io.Writer) *wordle {
	return &wordle{
		wordLength: wordLength,
		maxGuesses: maxGuesses,
		word:       word,
		in:         in,
		out:        out,
	}
}

func getWord() string {
	// NOTE: this list inludes many uncommon and seemingly not-English words. Is there a better data source?
	res, err := http.Get("https://raw.githubusercontent.com/dwyl/english-words/master/words_alpha.txt")
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}

	words := strings.Split(string(body), "\r\n")
	candidates := []string{}
	for _, word := range words {
		if len(word) == wordLength {
			candidates = append(candidates, strings.ToUpper(word))
		}
	}

	rand.Seed(time.Now().Unix())

	return candidates[rand.Intn(len(candidates))]
}

func main() {
	word := getWord()
	f := os.Stdin
	defer f.Close()

	w := newWordle(word, os.Stdin, f)
	w.run()
}
