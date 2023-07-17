package struct_calc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAutoCalByTag(t *testing.T) {
	type TestStruct struct {
		A int64   `expr:"a"`
		B int64   `expr:"b=a*a"`
		C int64   `expr:"c=a+b+1"`
		D int64   `expr:"d=a"`
		E float64 `expr:"e=a/c"`
		F float64 `expr:"f=a/c" round:"5"`
	}

	var ts = TestStruct{
		A: 1,
		B: 0,
		C: 0,
		D: 0,
		E: 0,
	}

	err := AutoCalByTag(&ts, "expr")
	assert.Nil(t, err)
	assert.Equal(t, int64(1), ts.B)
	assert.Equal(t, int64(3), ts.C)
	assert.Equal(t, int64(1), ts.D)
	assert.Equal(t, 0.33, ts.E)
	assert.Equal(t, 0.33333, ts.F)
}
