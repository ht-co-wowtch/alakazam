package sensitive

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFilter_Add(t *testing.T) {
	filter := New()
	filter.Add("test")
	assert.Equal(t, 1, len(filter.trie.root.children))
}

func TestFilter_Adds(t *testing.T) {
	filter := New()
	filter.Adds([]string{"A", "B"})
	assert.Equal(t, 2, len(filter.trie.root.children))
}

func TestFilter(t *testing.T) {
	filter := New()
	filter.Add("test")
	str := filter.Filter("12evtest64")
	assert.Equal(t, "12ev****64", str)
}

func TestFilterMatch(t *testing.T) {
	filter := New()
	filter.Add("test")
	str, isFilter := filter.FilterMatch("12evtest64")

	assert.True(t, isFilter)
	assert.Equal(t, "12ev****64", str)
}

func TestFilterFindSensitive(t *testing.T) {
	filter := New()
	filter.Add("test")
	filter.Add("abc")
	filter.Add("2")
	str, isFilter, sensitive := filter.FilterFindSensitive("12evtest64")

	assert.True(t, isFilter)
	assert.Equal(t, "1*ev****64", str)
	assert.Equal(t, []string{"2", "test"}, sensitive)
}

func TestFilterReplace(t *testing.T) {
	filter := New()
	filter.Add("test")
	str := filter.Replace("12evtest64", '@')

	assert.Equal(t, "12ev@@@@64", str)
}

func TestFilterReplaceMatch(t *testing.T) {
	filter := New()
	filter.Add("test")
	str, isFilter := filter.ReplaceMatch("12evtest64", '@')

	assert.True(t, isFilter)
	assert.Equal(t, "12ev@@@@64", str)
}

func TestFilterReplaceFindSensitive(t *testing.T) {
	filter := New()
	filter.Add("test")
	filter.Add("abc")
	filter.Add("2")
	str, isFilter, sensitive := filter.ReplaceFindSensitive("12evtest64", '@')

	assert.True(t, isFilter)
	assert.Equal(t, "1@ev@@@@64", str)
	assert.Equal(t, []string{"2", "test"}, sensitive)
}
