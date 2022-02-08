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

	usage   string = fmt.Sprintf("Guess a %v-letter word within %v guesses...\n", wordLength, maxGuesses)
	guesses        = []map[string][wordLength]tileColor{}
)

func main() {
	word := getWord()
	f := os.Stdin
	defer f.Close()

	run(word, os.Stdin, f)
}

func run(word string, in io.Reader, out io.Writer) {
	reader := bufio.NewScanner(in)

	write(fmt.Sprintf("https://github.com/mdb/wordle version %s\n\n", version), out)
	write(usage, out)

	for guessCount := 0; guessCount < maxGuesses; guessCount++ {
		reader.Scan()
		guess := strings.ToUpper(reader.Text())

		if guess == "STOP" {
			break
		}

		if len(guess) != len(word) {
			write(fmt.Sprintf("%s is not a %v-letter word. Try again...\n", guess, wordLength), out)
			guessCount--
		}

		if len(guess) == len(word) {
			displayWordleGrid(guess, word, out, guessCount)
		}

		if guess == word {
			break
		}

		if guessCount == maxGuesses-1 {
			fmt.Println()
			displayWordleRow(word, getLetterTileColors(word, word), out)
			os.Exit(1)
		}
	}
}

func write(str string, out io.Writer) {
	out.Write([]byte(str))
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

func getLetterTileColors(guess string, word string) [wordLength]tileColor {
	colors := [wordLength]tileColor{}

	for i := range colors {
		colors[i] = gray
	}

	for j, guessLetter := range guess {
		for k, letter := range word {
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

func displayWordleGrid(guess string, word string, out io.Writer, guessCount int) {
	tileColors := getLetterTileColors(guess, word)
	guesses = append(guesses, map[string][wordLength]tileColor{guess: tileColors})

	for _, guess := range guesses {
		for g, colors := range guess {
			displayWordleRow(g, colors, out)
		}
	}

	displayEmptyWordleRows(word, out, guessCount)
}

func displayWordleRow(word string, colors [wordLength]tileColor, out io.Writer) {
	for i, c := range word {
		switch colors[i] {
		case green:
			write("\033[42m\033[1;30m", out)
		case yellow:
			write("\033[43m\033[1;30m", out)
		case gray:
			write("\033[40m\033[1;37m", out)
		}

		write(fmt.Sprintf(" %c ", c), out)
		write("\033[m\033[m", out)
	}

	write("\n", out)
}

func displayEmptyWordleRows(word string, out io.Writer, guessCount int) {
	emptyGuessChars := []string{}
	for i := 0; i < wordLength; i++ {
		emptyGuessChars = append(emptyGuessChars, "*")
	}

	emptyGuess := strings.Join(emptyGuessChars, "")
	emptyTileColors := getLetterTileColors(emptyGuess, word)
	emptyRowCount := maxGuesses - guessCount - 1

	for i := 0; i < emptyRowCount; i++ {
		displayWordleRow(emptyGuess, emptyTileColors, out)
	}
}
