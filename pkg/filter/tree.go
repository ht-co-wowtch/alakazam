package filter

// Trie樹，又叫字典樹、前缀樹
type trie struct {
	root *node
}

func newTrie() *trie {
	return &trie{
		root: newNode(),
	}
}

// 新增敏感詞
// 如果新增'abc'，之後又新增'abcd'，則會更新原先的'abc' => 'abcd'
func (tree *trie) add(word string) {
	var current = tree.root
	var runes = []rune(word)
	for position := 0; position < len(runes); position++ {
		r := runes[position]
		if next, ok := current.children[r]; ok {
			current = next
		} else {
			newNode := newNode()
			current.children[r] = newNode
			current = newNode
		}
		if position == len(runes)-1 {
			current.isEnd = true
		}
	}
}

// 替換敏感詞
func (tree *trie) Replace(text string, character rune) string {
	position := tree.findPosition(text)
	if len(position) == 0 {
		return text
	}

	runes := []rune(text)
	for _, p := range position {
		runes[p] = character
	}
	return string(runes)
}

// 1. 替換敏感詞
// 2. 命中哪些敏感詞
func (tree *trie) ReplaceSensitive(text string, character rune) (string, []string) {
	var (
		parent  = tree.root
		current *node
		left    = 0
		runes   = []rune(text)
		found   bool
		matches []string
	)

	for position := 0; position < len(runes); position++ {
		current, found = parent.children[runes[position]]
		if !found {
			parent = tree.root
			position = left
			left++
			continue
		}
		if current.isEnd && left <= position {
			if len(matches) == 0 {
				matches = make([]string, 0, 5)
			}
			matches = append(matches, string(runes[left:position+1]))
			for i := left; i <= position; i++ {
				runes[i] = character
			}
		}
		parent = current
	}
	return string(runes), matches
}

// 找出敏感詞的位置
// 敏感詞 => 'abc'
// 過濾詞 => '12abc45'
// 敏感詞的位置 => 2,3,4
func (tree *trie) findPosition(text string) []int {
	var (
		parent    = tree.root
		current   *node
		left      = 0
		runes     = []rune(text)
		found     bool
		positions []int
	)

	for position := 0; position < len(runes); position++ {
		current, found = parent.children[runes[position]]
		if !found {
			parent = tree.root
			position = left
			left++
			continue
		}
		if current.isEnd && left <= position {
			if len(positions) == 0 {
				positions = make([]int, 0, 5)
			}
			for i := left; i <= position; i++ {
				positions = append(positions, i)
			}
		}
		parent = current
	}
	return positions
}

// 找出各個敏感詞的位置
// 敏感詞 => 'abc'
// 敏感詞 => '5'
// 過濾詞 => '12abc45'
// 第一組敏感詞的位置 => 2,3,4
// 第二組敏感詞的位置 => 6
func (tree *trie) findSensitivePosition(text string) [][]int {
	var (
		parent    = tree.root
		current   *node
		left      = 0
		runes     = []rune(text)
		found     bool
		positions [][]int
	)

	for position := 0; position < len(runes); position++ {
		current, found = parent.children[runes[position]]
		if !found {
			parent = tree.root
			position = left
			left++
			continue
		}
		if current.isEnd && left <= position {
			l := position - left
			if l == 0 {
				l = 1
			}
			p := make([]int, 0, l)
			for i := left; i <= position; i++ {
				p = append(p, i)
			}
			positions = append(positions, p)
		}
		parent = current
	}
	return positions
}
