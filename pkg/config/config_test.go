package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadConfigYml(t *testing.T) {
	testCase := []struct {
		path string
		tag  string
	}{
		{path: "./testdata/config.yml", tag: ""},
		{path: "./testdata/config_tag.yml", tag: "tag"},
	}

	for _, v := range testCase {
		c, err := ReadTag(v.path, v.tag)
		if err != nil {
			t.Fatal(err)
		}
		t.Run("read config ", func(t *testing.T) {
			assert.Equal(t, "Hello", c.Get("test"))
		})
	}
}
