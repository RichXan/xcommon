package xutil

import (
	"reflect"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

// Operator 查询操作符
type Operator string

const (
	OpEq    Operator = "="
	OpIn    Operator = "IN"
	OpLike  Operator = "LIKE"
	OpGt    Operator = ">"
	OpLt    Operator = "<"
	OpGte   Operator = ">="
	OpLte   Operator = "<="
	OpNotEq Operator = "!="
)

// QueryOption 查询选项
type QueryOption struct {
	Operator Operator
	Value    interface{}
}

// BuildQueryByModel 通用的查询构建器，根据模型的非零值构建查询条件
func BuildQueryByModel[T any](db *gorm.DB, model *T, options map[string]QueryOption) *gorm.DB {
	query := db.Model(model)

	// 首先处理 options 中的查询条件
	if len(options) > 0 {
		for fieldName, opt := range options {
			columnName := getColumnName("", fieldName)
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
	if gormTag == "" {
		return toCamelCase(fieldName)
	}

	tags := strings.Split(gormTag, ";")
	for _, tag := range tags {
		if strings.HasPrefix(tag, "column:") {
			return strings.TrimPrefix(tag, "column:")
		}
	}

	return toCamelCase(fieldName)
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

// toSnakeCase 将驼峰命名转换为蛇形命名
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
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
