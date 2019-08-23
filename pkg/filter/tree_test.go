package filter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAdd(t *testing.T) {
	trie := newTrie()

	word := "test"
	trie.add(word)

	parent := trie.root
	runes := []rune(word)
	isEnd := []bool{false, false, false, true}
	for i, r := range runes {
		ch, ok := parent.children[r]
		if !ok {
			t.Fatalf("found not for rune => [%d]", r)
		}

		if isEnd[i] != ch.isEnd {
			t.Fatalf("node is end for rune => [%d]", r)
		}

		parent = ch
	}
}

func TestAdds(t *testing.T) {
	trie := newTrie()

	word1 := "test"
	trie.add(word1)

	word2 := "abc"
	trie.add(word2)

	parent := trie.root
	runes := []rune(word2)
	isEnd := []bool{false, false, true}
	for i, r := range runes {
		ch, ok := parent.children[r]
		if !ok {
			t.Fatalf("found not for rune => [%d]", r)
		}

		if isEnd[i] != ch.isEnd {
			t.Fatalf("node is end for rune => [%d]", r)
		}

		parent = ch
	}

	_, ok := trie.root.children[rune(word1[0])]

	assert.True(t, ok)
	assert.Equal(t, 2, len(trie.root.children))
}

func TestAddReplace(t *testing.T) {
	trie := newTrie()

	trie.add("test")
	trie.add("abc")
	trie.add("abcd")

	parent := trie.root
	runes := []rune("abcd")
	for _, r := range runes {
		ch, ok := parent.children[r]
		if !ok {
			t.Fatalf("found not for rune => [%d]", r)
		}
		parent = ch
	}

	assert.Equal(t, 2, len(trie.root.children))
}

func TestReplace(t *testing.T) {
	trie := newTrie()
	trie.add("測試")

	s := trie.Replace("123測de測試", '*')

	assert.Equal(t, "123測de**", s)
}

func TestReplaces(t *testing.T) {
	trie := newTrie()
	trie.add("test")
	trie.add("測試")

	s := trie.Replace("123測de測試", '*')

	assert.Equal(t, "123測de**", s)
}

func TestReplaceMatch(t *testing.T) {
	trie := newTrie()
	trie.add("123")
	trie.add("測試")
	trie.add("abc")

	s := trie.Replace("123測de測試", '*')

	assert.Equal(t, "***測de**", s)
}

func TestReplaceSensitive(t *testing.T) {
	trie := newTrie()
	trie.add("123")
	trie.add("測試")
	trie.add("a")

	s, sensitive := trie.ReplaceSensitive("123測de測試1a", '*')

	assert.Equal(t, []string{"123", "測試", "a"}, sensitive)
	assert.Equal(t, "***測de**1*", s)
}

func TestFindSensitivePosition(t *testing.T) {
	trie := newTrie()
	trie.add("123")
	trie.add("v")
	trie.add("r")

	position := trie.findSensitivePosition("r3123fv")

	assert.Equal(t, [][]int{
		[]int{0},
		[]int{2, 3, 4},
		[]int{6},
	}, position)
}

func TestFindPosition(t *testing.T) {
	trie := newTrie()
	trie.add("123")
	trie.add("v")
	trie.add("r")

	position := trie.findPosition("r3123fv")

	assert.Equal(t, []int{0, 2, 3, 4, 6}, position)
}

// BenchmarkFindPosition-4   	10000000	       125 ns/op
func BenchmarkFindPosition(b *testing.B) {
	trie := newTrie()
	trie.add("123")
	trie.add("v")
	trie.add("r")
	for i := 0; i < b.N; i++ {
		trie.findPosition("r3123fv")
	}
}

// BenchmarkReplace-4   	10000000	       216 ns/op
func BenchmarkReplace(b *testing.B) {
	trie := newTrie()
	trie.add("123")
	trie.add("v")
	trie.add("r")
	for i := 0; i < b.N; i++ {
		trie.Replace("r3123fv", '*')
	}
}

// BenchmarkReplaceSensitive-4   	 5000000	       361 ns/op
func BenchmarkReplaceSensitive(b *testing.B) {
	trie := newTrie()
	trie.add("123")
	trie.add("v")
	trie.add("r")
	for i := 0; i < b.N; i++ {
		trie.ReplaceSensitive("r3123fv", '*')
	}
}
