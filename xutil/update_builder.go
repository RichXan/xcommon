package xutil

import (
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// UpdateOption 更新选项
type UpdateOption struct {
	Fields    []string               // 指定要更新的字段，为空则更新所有非零值字段
	Condition map[string]QueryOption // 更新条件
}

// BuildUpdateByModel 通用的更新构建器
func BuildUpdateByModel[T any](db *gorm.DB, model *T, opt *UpdateOption) *gorm.DB {
	query := db.Model(model)
	updates := make(map[string]interface{})

	val := reflect.ValueOf(model).Elem()
	typ := val.Type()

	// 处理所有字段，包括嵌入式结构体
	processFields(val, typ, "", updates, opt)

	// 添加更新条件
	if opt != nil && len(opt.Condition) > 0 {
		for fieldName, queryOpt := range opt.Condition {
			// 跳过空值条件
			if isEmptyValue(queryOpt.Value) {
				continue
			}

			columnName := extractColumnName("", fieldName)
			query = buildQueryWithOperator(query, columnName, queryOpt.Operator, queryOpt.Value)
		}
	}

	return query.Updates(updates)
}

// processFields 处理结构体字段，包括嵌入式结构体
func processFields(val reflect.Value, typ reflect.Type, prefix string, updates map[string]interface{}, opt *UpdateOption) {
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 跳过未导出的字段
		if !field.CanInterface() {
			continue
		}

		// 如果是嵌入式结构体，递归处理
		if fieldType.Anonymous {
			if field.Kind() == reflect.Struct {
				processFields(field, field.Type(), prefix, updates, opt)
			}
			continue
		}

		// 获取gorm标签
		gormTag := fieldType.Tag.Get("gorm")
		if gormTag == "-" { // 跳过标记为 "-" 的字段
			continue
		}

		// 获取列名
		columnName := extractColumnName(gormTag, fieldType.Name)

		// 如果指定了要更新的字段，则只更新指定字段
		if opt != nil && len(opt.Fields) > 0 {
			if !contains(opt.Fields, fieldType.Name) {
				continue
			}
			updates[columnName] = field.Interface()
			continue
		}

		// 否则更新所有非零值字段
		if !isZeroValue(field) {
			updates[columnName] = field.Interface()
		}
	}
}

// extractColumnName 从gorm标签或字段名获取列名
func extractColumnName(gormTag string, fieldName string) string {
	if gormTag == "" {
		return fieldName
	}

	tags := strings.Split(gormTag, ";")
	for _, tag := range tags {
		if strings.HasPrefix(tag, "column:") {
			return strings.TrimPrefix(tag, "column:")
		}
	}

	return fieldName
}
