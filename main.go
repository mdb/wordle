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

// example: {"boardState":["BEACH","UNDER","","","",""],"evaluations":[["absent","present","absent","present","absent"],["correct","absent","absent","correct","correct"],null,null,null,null],"rowIndex":2,"solution":"ulcer","gameStatus":"IN_PROGRESS","lastPlayedTs":1644580347374,"lastCompletedTs":null,"restoringFromLocalStorage":null,"hardMode":false}
type gameState struct {
	// a slice of guesses
	// example: []string{"BEACH", "", "", "", "", ""}
	boardState [maxGuesses]string

	// a slice of slices, representing each guess's evaluated chars
	// example: []string{[]string{"correct", "present", "absent", "absent", "absent"}, []string{}, []string{}, []string{}, []string{}, []string{}}
	evaluations [maxGuesses][wordLength]string

	// the current row
	rowIndex int

	// the solution word
	solution string

	// example: IN_PROGRESS
	// TODO: what are the other possible values?
	gameStatus string

	// example: 1644580347374
	lastPlayedTS time.Time

	// example: 1644580347374
	lastCompletedTS time.Time

	hardMode bool
}

type wordle struct {
	state   *gameState
	guesses []map[string][wordLength]tileColor
	in      io.Reader
	out     io.Writer
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

func (w *wordle) displayGrid() {
	for i, guess := range w.state.boardState {
		for j, guessLetter := range guess {
			switch w.state.evaluations[i][j] {
			case "correct":
				w.write("\033[42m\033[1;30m")
			case "present":
				w.write("\033[43m\033[1;30m")
			case "absent":
				w.write("\033[40m\033[1;37m")
			}

			w.write(fmt.Sprintf(" %c ", guessLetter))
			w.write("\033[m\033[m")
		}

		w.write("\n")
	}
}

func (w *wordle) getLetterTileColors(guess string) [wordLength]tileColor {
	colors := [wordLength]tileColor{}

	for i := range colors {
		colors[i] = gray
	}

	for j, guessLetter := range guess {
		for k, letter := range w.state.solution {
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
	for i := 0; i < wordLength; i++ {
		emptyGuessChars = append(emptyGuessChars, "*")
	}

	emptyGuess := strings.Join(emptyGuessChars, "")
	emptyTileColors := w.getLetterTileColors(emptyGuess)
	emptyRowCount := maxGuesses - guessCount - 1

	for i := 0; i < emptyRowCount; i++ {
		w.displayRow(emptyGuess, emptyTileColors)
	}
}

func (w *wordle) evaluateGuess(guess string) [wordLength]string {
	evaluation := [wordLength]string{}

	for i := 0; i < wordLength; i++ {
		evaluation[i] = "absent"
	}

	for j, guessLetter := range guess {
		for k, letter := range w.state.solution {
			if guessLetter == letter {
				if j == k {
					evaluation[j] = "correct"
					break
				}

				evaluation[j] = "present"
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
	solution := w.state.solution

	w.write(fmt.Sprintf("Version: \t%s\n", version))
	w.write("Info: \t\thttps://github.com/mdb/wordle\n")
	w.write("About: \t\tA CLI adaptation of Josh Wardle's Wordle (https://powerlanguage.co.uk/wordle/)\n\n")
	w.write(fmt.Sprintf("Guess a %v-letter word within %v guesses...\n", wordLength, maxGuesses))

	for w.state.rowIndex = 0; w.state.rowIndex < maxGuesses; w.state.rowIndex++ {
		w.write(fmt.Sprintf("\nGuess (%v/%v): ", len(w.guesses)+1, maxGuesses))

		reader.Scan()
		guess := strings.ToUpper(reader.Text())

		if guess == "STOP" {
			break
		}

		if len(guess) != len(solution) {
			w.write(fmt.Sprintf("%s is not a %v-letter word. Try again...\n", guess, wordLength))
			w.state.rowIndex--
		}

		if len(guess) == len(solution) {
			w.state.boardState[w.state.rowIndex] = guess
			w.state.evaluations[w.state.rowIndex] = w.evaluateGuess(guess)
			w.displayGrid()
		}

		if guess == solution {
			break
		}

		if w.state.rowIndex == maxGuesses-1 {
			fmt.Println()
			w.displayRow(solution, w.getLetterTileColors(solution))
			os.Exit(1)
		}
	}
}

func newWordle(word string, in io.Reader, out io.Writer) *wordle {
	w := &wordle{
		in:  in,
		out: out,
		state: &gameState{
			solution: word,
		},
	}
	emptyGuessChar := "*"
	emptyGuess := ""
	emptyGuessEvaluation := [wordLength]string{}

	for i := 0; i < wordLength; i++ {
		emptyGuess = emptyGuess + emptyGuessChar
		fmt.Println(emptyGuess)
		emptyGuessEvaluation[i] = "absent"
	}

	for i := 0; i < maxGuesses; i++ {
		w.state.evaluations[i] = emptyGuessEvaluation
		w.state.boardState[i] = emptyGuess
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
