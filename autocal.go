package struct_calc

import (
	"errors"
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"reflect"
	"strings"
)

// AutoCalByTag calculate the fields of the provided pointer value by expression in tagName
// support filed type int64 and float64, float64 values  retains 2 digits after the decimal point by default
func AutoCalByTag(value interface{}, tagName string) error {
	vo := reflect.ValueOf(value)
	if vo.Kind() != reflect.Pointer {
		return errors.New("value's kind is not pointer")
	}
	type Detail struct {
		FiledValue reflect.Value
		Expr       *govaluate.EvaluableExpression
		FiledName  string
		Round      int
	}
	var (
		dependents      = map[string]map[string]struct{}{}
		paraNameToFiled = make(map[string]*Detail)
		parameters      = make(map[string]any)
	)
	vo = vo.Elem()
	var rangeFields func(value reflect.Value) error
	rangeFields = func(value reflect.Value) error {
		if value.Kind() == reflect.Struct {
			for i := 0; i < value.NumField(); i++ {
				field := value.Field(i)
				if field.Kind() == reflect.Struct {
					err := rangeFields(field)
					if err != nil {
						return err
					}
					continue
				}
				tag := value.Type().Field(i).Tag
				fieldName := value.Type().Field(i).Name
				if tagValue := tag.Get(tagName); tagValue != "" {
					ss := strings.Split(tagValue, "=")
					if len(ss) == 0 || len(ss) > 2 {
						return fmt.Errorf("invalid tag expr for field:%s", value.Type().Field(i).Name)
					}
					if len(ss) == 1 {
						if parameters[ss[0]] != nil {
							return fmt.Errorf("duplicate define symple:%s", ss[0])
						}
						parameters[ss[0]] = field.Interface()
						continue
					}
					param, expr := ss[0], ss[1]
					expression, err := govaluate.NewEvaluableExpression(expr)
					if err != nil {
						return fmt.Errorf("NewEvaluableExpression err:%v by filed:%s", err, expression)
					}

					for _, v := range expression.Vars() {
						if v != param {
							if dependents[param] == nil {
								dependents[param] = make(map[string]struct{})
							}
							dependents[param][v] = struct{}{}
						}
					}
					var round = 2
					if get := tag.Get("round"); get != "" {
						round = cast.ToInt(get)
					}
					if len(dependents[param]) == 0 {
						ret, err := expression.Evaluate(map[string]interface{}{param: field.Interface()})
						if err != nil {
							return fmt.Errorf("evaluate filedName:%s for expression:%s err:%v", fieldName, expression, err)
						}
						err = setField(field, ret, round)
						if err != nil {
							return fmt.Errorf("fieldValue of filedName:%s can't be set", fieldName)
						}
						if parameters[param] != nil {
							return fmt.Errorf("duplicate define symple:%s", param)
						}
						parameters[param] = ret
					}
					paraNameToFiled[param] = &Detail{
						FiledValue: field,
						FiledName:  fieldName,
						Expr:       expression,
						Round:      round,
					}
				}
			}
		}
		return nil
	}
	err := rangeFields(vo)
	if err != nil {
		return err
	}
	var vis = make(map[string]int)
	var dfs func(string) error
	dfs = func(node string) error {
		if v, ok := vis[node]; ok && v == 0 {
			return fmt.Errorf("node:%s has circle", node)
		}
		if vis[node] == 1 {
			return nil
		}
		vis[node] = 0
		for k := range dependents[node] {
			err := dfs(k)
			if err != nil {
				return err
			}
		}
		vis[node] = 1
		return nil
	}
	for node := range dependents {
		if err := dfs(node); err != nil {
			return err
		}
	}
	for knownParam := range parameters {
		for _, v := range dependents {
			delete(v, knownParam)
		}
	}
	var findInDegreeZero = true
	for len(dependents) > 0 && findInDegreeZero {
		findInDegreeZero = false
		for k, v := range dependents {
			if len(v) == 0 {
				findInDegreeZero = true
				detail := paraNameToFiled[k]
				ret, err := detail.Expr.Evaluate(parameters)
				if err != nil {
					return fmt.Errorf("evaluate expr %s err:%v", detail.Expr.String(), err)
				}
				err = setField(detail.FiledValue, ret, detail.Round)
				if err != nil {
					return fmt.Errorf("evaluate expr %s err:%v", detail.Expr.String(), err)
				}
				if parameters[k] != nil {
					return fmt.Errorf("duplicate define symple:%s", k)
				}
				parameters[k] = ret
				delete(dependents, k)
				for _, v2 := range dependents {
					delete(v2, k)
				}
			}
		}
	}
	if len(dependents) > 0 {
		sb := strings.Builder{}
		for k, vv := range dependents {
			var values []string
			for k := range vv {
				values = append(values, k)
			}
			sb.WriteString(fmt.Sprintf("%s dependents on values:%v, but not find values's defines\n", k, values))
		}
		return errors.New(sb.String())
	}
	return nil
}

func setField(field reflect.Value, ret interface{}, round int) error {
	if !field.CanSet() {
		return fmt.Errorf("value can't be set")
	}
	var fKind, rKind = field.Kind(), reflect.ValueOf(ret).Kind()
	if rKind == reflect.Float64 {
		ret = decimal.NewFromFloat(cast.ToFloat64(ret)).Round(int32(round)).InexactFloat64()
	}
	if fKind != rKind {
		switch fKind {
		case reflect.Int64:
			ret = cast.ToInt64(ret)
		case reflect.Float64:
			ret = cast.ToFloat64(ret)
		default:
			return fmt.Errorf("not support auto convert ret kind:%s to kind:%s", rKind, fKind)
		}
	}
	field.Set(reflect.ValueOf(ret))
	return nil
}
