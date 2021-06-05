package minlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemovePrefix(t *testing.T) {
	assert.Equal(t, "abc", mkLongestPrefixCutter("")("abc"))
	assert.Equal(t, "bc", mkLongestPrefixCutter("a")("abc"))
	assert.Equal(t, "", mkLongestPrefixCutter("abc")("a"))
	assert.Equal(t, "bc", mkLongestPrefixCutter("aa")("abc"))
}
