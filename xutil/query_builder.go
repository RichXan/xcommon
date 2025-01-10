package xutil

import (
	"reflect"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

// Operator 查询操作符
type Operator string

const (
	OpEq      Operator = "="
	OpIn      Operator = "IN"
	OpLike    Operator = "LIKE"
	OpGt      Operator = ">"
	OpLt      Operator = "<"
	OpGte     Operator = ">="
	OpLte     Operator = "<="
	OpNotEq   Operator = "!="
	OpDateGt  Operator = "DATE >"  // 晚于某日期
	OpDateLt  Operator = "DATE <"  // 早于某日期
	OpDateGte Operator = "DATE >=" // 晚于或等于某日期
	OpDateLte Operator = "DATE <=" // 早于或等于某日期
	OpDateEq  Operator = "DATE ="  // 等于某日期
	OpDateNe  Operator = "DATE !=" // 不等于某日期
)

// QueryOption 查询选项
type QueryOption struct {
	Operator    Operator
	Value       interface{}
	NoSnakeCase bool // 是否禁用下划线转换，默认false表示使用下划线形式
}

// BuildQueryByModel 通用的查询构建器，根据模型的非零值构建查询条件
func BuildQueryByModel[T any](db *gorm.DB, model *T, options map[string]QueryOption) *gorm.DB {
	query := db.Model(model)

	// 首先处理 options 中的查询条件
	if len(options) > 0 {
		for fieldName, opt := range options {
			// 跳过空值
			if isEmptyValue(opt.Value) {
				continue
			}

			// 获取结构体字段对应的数据库列名
			var columnName string
			if t := reflect.TypeOf(model).Elem(); t.Kind() == reflect.Struct {
				if field, found := t.FieldByName(fieldName); found {
					columnName = getColumnName(field.Tag.Get("gorm"), fieldName)
				} else {
					// 如果在结构体中找不到该字段，根据 NoSnakeCase 选项决定是否转换为下划线形式
					if opt.NoSnakeCase {
						columnName = fieldName
					} else {
						columnName = toSnakeCase(fieldName)
					}
				}
			} else {
				if opt.NoSnakeCase {
					columnName = fieldName
				} else {
					columnName = toSnakeCase(fieldName)
				}
			}

			query = buildQueryWithOperator(query, columnName, opt.Operator, opt.Value)
		}
	}

	// 然后处理 model 中的非零值字段
	val := reflect.ValueOf(model).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 跳过未导出的字段
		if !field.CanInterface() {
			continue
		}

		// 获取gorm标签
		gormTag := fieldType.Tag.Get("gorm")
		if gormTag == "-" { // 跳过标记为 "-" 的字段
			continue
		}

		// 获取列名
		columnName := getColumnName(gormTag, fieldType.Name)

		// 如果字段名已经在 options 中处理过，则跳过
		if _, exists := options[fieldType.Name]; exists {
			continue
		}

		// 处理非零值字段
		if !isZeroValue(field) {
			if isReservedWord(columnName) {
				columnName = "`" + columnName + "`"
			}
			query = query.Where(columnName+" = ?", field.Interface())
		}
	}

	return query
}

// buildQueryWithOperator 根据操作符构建查询条件
func buildQueryWithOperator(query *gorm.DB, column string, op Operator, value interface{}) *gorm.DB {
	if isReservedWord(column) {
		column = "`" + column + "`"
	}

	switch op {
	case OpIn:
		return query.Where(column+" IN ?", value)
	case OpLike:
		return query.Where(column+" LIKE ?", value)
	case OpGt:
		return query.Where(column+" > ?", value)
	case OpLt:
		return query.Where(column+" < ?", value)
	case OpGte:
		return query.Where(column+" >= ?", value)
	case OpLte:
		return query.Where(column+" <= ?", value)
	case OpNotEq:
		return query.Where(column+" != ?", value)
	case OpDateGt:
		return query.Where("DATE("+column+") > DATE(?)", value)
	case OpDateLt:
		return query.Where("DATE("+column+") < DATE(?)", value)
	case OpDateGte:
		return query.Where("DATE("+column+") >= DATE(?)", value)
	case OpDateLte:
		return query.Where("DATE("+column+") <= DATE(?)", value)
	case OpDateEq:
		return query.Where("DATE("+column+") = DATE(?)", value)
	case OpDateNe:
		return query.Where("DATE("+column+") != DATE(?)", value)
	default: // OpEq
		return query.Where(column+" = ?", value)
	}
}

// isZeroValue 检查字段是否为零值
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Struct:
		// 处理decimal类型
		if v.Type().String() == "decimal.Decimal" {
			return v.MethodByName("String").Call(nil)[0].String() == "0"
		}
		return v.IsZero()
	case reflect.Ptr:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

// getColumnName 从gorm标签或字段名获取数据库列名
func getColumnName(gormTag string, fieldName string) string {
	// 首先尝试从gorm标签获取列名
	if gormTag != "" {
		tags := strings.Split(gormTag, ";")
		for _, tag := range tags {
			if strings.HasPrefix(tag, "column:") {
				return strings.TrimPrefix(tag, "column:")
			}
		}
	}

	// 如果没有column标签，则将驼峰命名转换为下划线命名
	return toSnakeCase(fieldName)
}

// 将蛇形命名转换为驼峰命名
func toCamelCase(s string) string {
	words := strings.Split(s, "_")
	caser := cases.Title(language.English)
	for i, word := range words {
		words[i] = caser.String(word)
	}
	return strings.Join(words, "")
}

// toSnakeCase 将驼峰命名转换为下划线命名
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(unicode.ToLower(r))
		}
	}
	return result.String()
}

// 保留字列表
var reservedWords = map[string]bool{
	"type":   true,
	"order":  true,
	"select": true,
	"where":  true,
	"from":   true,
}

// isReservedWord 检查是否是SQL保留字
func isReservedWord(name string) bool {
	return reservedWords[name]
}

// contains 检查字符串是否在切片中
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// isEmptyValue 检查值是否为空
func isEmptyValue(v interface{}) bool {
	if v == nil {
		return true
	}

	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.String:
		return value.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return value.Len() == 0
	case reflect.Ptr:
		if value.IsNil() {
			return true
		}
		return isEmptyValue(value.Elem().Interface())
	case reflect.Interface:
		if value.IsNil() {
			return true
		}
		return isEmptyValue(value.Elem().Interface())
	case reflect.Struct:
		// 处理decimal类型
		if value.Type().String() == "decimal.Decimal" {
			return value.MethodByName("String").Call(nil)[0].String() == "0"
		}
		return value.IsZero()
	default:
		return value.IsZero()
	}
}
