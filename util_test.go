package adb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsWhitespaceYes(t *testing.T) {
	assert.True(t, containsWhitespace("hello world"))
}

func TestContainsWhitespaceNo(t *testing.T) {
	assert.False(t, containsWhitespace("hello"))
}

func TestIsBlankWhenEmpty(t *testing.T) {
	assert.True(t, isBlank(""))
}

func TestIsBlankWhenJustWhitespace(t *testing.T) {
	assert.True(t, isBlank(" \t"))
}

func TestIsBlankNo(t *testing.T) {
	assert.False(t, isBlank("     h   "))
}
