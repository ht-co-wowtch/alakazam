package sensitive

import (
	"bufio"
	"io"
)

type Filter struct {
	trie             *trie
	defaultCharacter rune
}

func New() *Filter {
	return &Filter{
		trie:             newTrie(),
		defaultCharacter: '*',
	}
}

func (f *Filter) Add(word string) {
	f.trie.add(word)
}

func (f *Filter) Adds(words []string) {
	for _, word := range words {
		f.trie.add(word)
	}
}

func (f *Filter) Load(rd io.Reader) error {
	buf := bufio.NewReader(rd)
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		f.trie.add(string(line))
	}
	return nil
}

func (f *Filter) Filter(text string) string {
	return f.trie.Replace(text, f.defaultCharacter)
}

func (f *Filter) FilterMatch(text string) (string, bool) {
	str := f.trie.Replace(text, f.defaultCharacter)
	return str, !(text == str)
}

func (f *Filter) FilterFindSensitive(text string) (string, bool, []string) {
	str, sensitive := f.trie.ReplaceSensitive(text, f.defaultCharacter)
	return str, !(text == str), sensitive
}

func (f *Filter) Replace(text string, character rune) string {
	return f.trie.Replace(text, character)
}

func (f *Filter) ReplaceMatch(text string, character rune) (string, bool) {
	str := f.trie.Replace(text, character)
	return str, !(text == str)
}

func (f *Filter) ReplaceFindSensitive(text string, character rune) (string, bool, []string) {
	str, sensitive := f.trie.ReplaceSensitive(text, character)
	return str, !(text == str), sensitive
}
