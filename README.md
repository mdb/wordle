# wordle

A Go-based `wordle` CLI.

[Josh Wardle](https://github.com/powerlanguage) deserves credit for creating [Wordle](https://www.powerlanguage.co.uk/wordle/); this is just a command line interface implementation of his creation.

## Development

Run tests and compile `wordle` release artifacts:

```
make
```

## Improvement ideas

* https://raw.githubusercontent.com/dwyl/english-words/master/words_alpha.txt returns some tricky and uncommon words; is there a better data source from which to retrieve words?
* Could https://github.com/rivo/tview enable some UI improvements?
* It'd be nice to be able to `brew install wordle`
