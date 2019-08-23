package filter

import (
	"bufio"
	"io"
	"sync"
)

type Filter struct {
	trie             *trie
	defaultCharacter rune
	mu               sync.Mutex
}

func New() *Filter {
	return &Filter{
		trie:             newTrie(),
		defaultCharacter: '*',
	}
}

// 新增敏感詞
func (f *Filter) Add(word string) {
	f.mu.Lock()
	f.trie.add(word)
	f.mu.Unlock()
}

func (f *Filter) Adds(words []string) {
	f.mu.Lock()
	for _, word := range words {
		f.trie.add(word)
	}
	f.mu.Unlock()
}

func (f *Filter) Delete(words string) bool {
	f.mu.Lock()
	ok := f.trie.delete(words)
	f.mu.Unlock()
	return ok
}

func (f *Filter) Load(rd io.Reader) error {
	buf := bufio.NewReader(rd)
	f.mu.Lock()
	defer f.mu.Unlock()
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

// 過濾敏感詞
func (f *Filter) Filter(text string) string {
	return f.trie.Replace(text, f.defaultCharacter)
}

// 過濾敏感詞
// 回傳是否有敏感詞
func (f *Filter) FilterMatch(text string) (string, bool) {
	str := f.trie.Replace(text, f.defaultCharacter)
	return str, !(text == str)
}

// 過濾敏感詞
// 回傳是否有敏感詞
// 有哪些敏感詞
func (f *Filter) FilterFindSensitive(text string) (string, bool, []string) {
	str, sensitive := f.trie.ReplaceSensitive(text, f.defaultCharacter)
	return str, !(text == str), sensitive
}

// 過濾敏感詞，自訂替換詞
func (f *Filter) Replace(text string, character rune) string {
	return f.trie.Replace(text, character)
}

// 過濾敏感詞，自訂替換詞
// 回傳是否有敏感詞
func (f *Filter) ReplaceMatch(text string, character rune) (string, bool) {
	str := f.trie.Replace(text, character)
	return str, !(text == str)
}

// 過濾敏感詞，自訂替換詞
// 回傳是否有敏感詞
// 有哪些敏感詞
func (f *Filter) ReplaceFindSensitive(text string, character rune) (string, bool, []string) {
	str, sensitive := f.trie.ReplaceSensitive(text, character)
	return str, !(text == str), sensitive
}
