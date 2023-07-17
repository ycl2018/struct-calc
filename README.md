# struct-calc
automatically calculate the fields of the struct by the expression on the Tag

## English
- install `go get github.com/ycl2018/struct-calc`
- Given the initial value in the struct, it supports calculating the values ​​of other fields through expressions in the tag, and supports field types int64 and float64. The float64 value retains 2 decimal places by default, and you can specify the reserved decimal places by adding round to the tag
- Package based on [govaluate](https://github.com/Knetic/govaluate), supported expressions please refer to govaluate library
- struct structures can be nested

## 中文说明
- 安装 `go get github.com/ycl2018/struct-calc`
- 给定struct中的字段初始值，支持通过tag中的表达式计算其余字段的值，支持字段类型int64和float64，float64值默认保留小数点后2位，可以通过tag中添加round来指定保留的小数位
- 基于[govaluate](https://github.com/Knetic/govaluate)的封装, 支持的表达式请参考govaluate库
- struct结构体可以嵌套 

## example
```go
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
		A: 1
	}

	err := AutoCalByTag(&ts, "expr")
	assert.Nil(t, err)
	assert.Equal(t, int64(1), ts.B)
	assert.Equal(t, int64(3), ts.C)
	assert.Equal(t, int64(1), ts.D)
	assert.Equal(t, 0.33, ts.E)
	assert.Equal(t, 0.33333, ts.F)
}
```