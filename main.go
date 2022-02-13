package main

import (
	"bufio"
	"embed"
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
	maxGuesses     int    = 6
	wordLength     int    = 5
	emptyGuessChar string = "*"
)

type evaluation int

const (
	absent evaluation = iota
	present
	correct
)

var (
	// passed in at build time
	version string

	// Embed the words directory in the compiled binary.
	//go:embed words
	words embed.FS
)

// example: {"currentStreak":1,"maxStreak":1,"guesses":{"1":1,"2":0,"3":0,"4":0,"5":0,"6":0,"fail":0},"winPercentage":100,"gamesPlayed":1,"gamesWon":1,"averageGuesses":1}
type statistics struct {
	currentStreak  int
	maxStreak      int
	guesses        map[string]int
	winsPercentage int
	gamesPlayed    int
	gamesWon       int
	averageGuesses int
}

type wordle struct {
	in  io.Reader
	out io.Writer

	// a slice of guesses
	// example: []string{"BEACH", "", "", "", "", ""}
	guesses [maxGuesses]string

	// a slice of slices, representing each guess's evaluated chars
	evaluations [maxGuesses][wordLength]evaluation

	// the current row
	rowIndex int

	// the solution word
	solution string

	// example: IN_PROGRESS, FAIL
	// TODO: what are the other possible values?
	gameStatus string

	// example: 1644580347374
	lastPlayedTS time.Time

	// example: 1644580347374
	lastCompletedTS time.Time

	hardMode bool
}

func (w *wordle) displaySolution() {
	for _, char := range w.solution {
		w.displayGreenTile(char)
	}

	w.write("\n")
}

func (w *wordle) displayGrid() {
	for i, guess := range w.guesses {
		for j, guessLetter := range guess {
			switch w.evaluations[i][j] {
			case correct:
				w.displayGreenTile(guessLetter)
			case present:
				w.displayYellowTile(guessLetter)
			case absent:
				w.displayGrayTile(guessLetter)
			}
		}

		w.write("\n")
	}
}

func (w *wordle) displayGreenTile(char rune) {
	w.write("\033[42m\033[1;30m")
	w.displayOnTile(char)
}

func (w *wordle) displayYellowTile(char rune) {
	w.write("\033[43m\033[1;30m")
	w.displayOnTile(char)
}

func (w *wordle) displayGrayTile(char rune) {
	w.write("\033[40m\033[1;37m")
	w.displayOnTile(char)
}

func (w *wordle) displayOnTile(char rune) {
	w.write(fmt.Sprintf(" %c ", char))
	w.write("\033[m\033[m")
}

func (w *wordle) evaluateGuess(guess string) [wordLength]evaluation {
	evaluation := [wordLength]evaluation{}

	for i := 0; i < wordLength; i++ {
		evaluation[i] = absent
	}

	for j, guessLetter := range guess {
		for k, letter := range w.solution {
			if guessLetter == letter {
				if j == k {
					evaluation[j] = correct
					break
				}

				evaluation[j] = present
			}
		}
	}

	return evaluation
}

func (w *wordle) write(str string) {
	w.out.Write([]byte(str))
}

func (w *wordle) run() {
	reader := bufio.NewScanner(w.in)
	solution := w.solution

	w.write(fmt.Sprintf("Version: \t%s\n", version))
	w.write("Info: \t\thttps://github.com/mdb/wordle\n")
	w.write("About: \t\tA CLI adaptation of Josh Wardle's Wordle (https://powerlanguage.co.uk/wordle/)\n\n")
	w.write(fmt.Sprintf("Guess a %v-letter word within %v guesses...\n", wordLength, maxGuesses))

	for w.rowIndex = 0; w.rowIndex < maxGuesses; w.rowIndex++ {
		w.write(fmt.Sprintf("\nGuess (%v/%v): ", w.rowIndex+1, maxGuesses))

		reader.Scan()
		guess := strings.ToUpper(reader.Text())

		if guess == "STOP" {
			break
		}

		if len(guess) != len(solution) {
			w.write(fmt.Sprintf("%s is not a %v-letter word. Try again...\n", guess, wordLength))
			w.rowIndex--
		}

		if len(guess) == len(solution) {
			w.guesses[w.rowIndex] = guess
			w.evaluations[w.rowIndex] = w.evaluateGuess(guess)
			w.displayGrid()
		}

		if guess == solution {
			break
		}

		if w.rowIndex == maxGuesses-1 {
			fmt.Println()
			w.displaySolution()
			os.Exit(1)
		}
	}
}

func newWordle(word string, in io.Reader, out io.Writer) *wordle {
	w := &wordle{
		in:       in,
		out:      out,
		solution: word,
	}
	emptyGuess := ""
	emptyGuessEvaluation := [wordLength]evaluation{}

	for i := 0; i < wordLength; i++ {
		emptyGuess = emptyGuess + emptyGuessChar
		emptyGuessEvaluation[i] = absent
	}

	// By seeding w with dummy guesses and dummy evaluations,
	// displayGrid displays remaining rows with each grid rendering.
	for i := 0; i < maxGuesses; i++ {
		w.evaluations[i] = emptyGuessEvaluation
		w.guesses[i] = emptyGuess
	}

	return w
}

func getWordFromFile() string {
	data, err := words.ReadFile("words/words.txt")
	if err != nil {
		log.Fatalln(err)
	}

	today := time.Now().UTC()
	startDay := time.Date(2021, time.Month(6), 19, 0, 0, 0, 0, time.UTC)
	daysSinceStart := int(today.Sub(startDay).Hours() / 24)

	return strings.ToUpper(strings.Split(string(data), ",")[daysSinceStart])
}

func getWordFromURL() string {
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
	word := getWordFromFile()
	f := os.Stdin
	defer f.Close()

	w := newWordle(word, os.Stdin, f)
	w.run()
}
