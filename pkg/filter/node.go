package filter

// Trie樹節點
type node struct {
	isEnd    bool
	children map[rune]*node
}

func newNode() *node {
	return &node{
		children: make(map[rune]*node, 0),
	}
}
