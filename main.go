package main

import (
	"bufio"
	"fmt"
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
	guesses = []map[string][wordLength]tileColor{}
)

func main() {
	word := getWord()
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(fmt.Sprintf("Guess a %v-letter word...", wordLength))

	for guessCount := 0; guessCount < maxGuesses; guessCount++ {
		guess, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalln(err)
		}

		guess = strings.ToUpper(guess[:len(guess)-1])

		if len(guess) != len(word) {
			fmt.Println(fmt.Sprintf("%s is not a a %v-letter word. Try again...", guess, wordLength))
			guessCount--
		}

		if len(guess) == len(word) {
			displayWordleGrid(guess, word)
		}

		if guess == word {
			break
		}

		if guessCount == maxGuesses-1 {
			fmt.Println()
			displayWordleRow(word, getLetterTileColors(word, word))
			os.Exit(1)
		}
	}
}

func getWord() string {
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

func displayWordleGrid(guess string, word string) {
	tileColors := getLetterTileColors(guess, word)
	guesses = append(guesses, map[string][wordLength]tileColor{guess: tileColors})

	for _, guess := range guesses {
		for g, colorVect := range guess {
			displayWordleRow(g, colorVect)
		}
	}
}

func displayWordleRow(word string, colors [wordLength]tileColor) {
	for i, c := range word {
		switch colors[i] {
		case green:
			fmt.Print("\033[42m\033[1;30m")
		case yellow:
			fmt.Print("\033[43m\033[1;30m")
		case gray:
			fmt.Print("\033[40m\033[1;37m")
		}

		fmt.Printf(" %c ", c)
		fmt.Print("\033[m\033[m")
	}

	fmt.Println()
}
