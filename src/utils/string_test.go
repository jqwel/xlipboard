package utils

import (
	"math/rand"
	"testing"
)

func TestRandStringBytes(t *testing.T) {
	rs := RandStringBytes(rand.Intn(80-64+1) + 64)
	t.Log(rs)
}
