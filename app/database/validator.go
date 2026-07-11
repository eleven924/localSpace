package database

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// Validator 数据验证器
type Validator struct {
	errors []ValidationError
}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{
		errors: make([]ValidationError, 0),
	}
}

// AddError 添加验证错误
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors 检查是否有错误
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors 获取所有错误
func (v *Validator) GetErrors() []ValidationError {
	return v.errors
}

// Error 实现 error 接口
func (v *Validator) Error() string {
	if len(v.errors) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range v.errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateString 验证字符串
func (v *Validator) ValidateString(field, value string, minLength, maxLength int, required bool) {
	if required && value == "" {
		v.AddError(field, "is required")
		return
	}

	if !required && value == "" {
		return
	}

	if len(value) < minLength {
		v.AddError(field, fmt.Sprintf("must be at least %d characters long", minLength))
	}

	if len(value) > maxLength {
		v.AddError(field, fmt.Sprintf("must not exceed %d characters", maxLength))
	}
}

// ValidateEmail 验证邮箱格式
func (v *Validator) ValidateEmail(field, value string, required bool) {
	if required && value == "" {
		v.AddError(field, "is required")
		return
	}

	if !required && value == "" {
		return
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		v.AddError(field, "must be a valid email address")
	}
}

// ValidateInt 验证整数
func (v *Validator) ValidateInt(field string, value int, min, max int, required bool) {
	if required && value == 0 {
		v.AddError(field, "is required")
		return
	}

	if !required && value == 0 {
		return
	}

	if value < min {
		v.AddError(field, fmt.Sprintf("must be at least %d", min))
	}

	if value > max {
		v.AddError(field, fmt.Sprintf("must not exceed %d", max))
	}
}

// ValidateFilePath 验证文件路径
func (v *Validator) ValidateFilePath(field, value string, required bool) {
	if required && value == "" {
		v.AddError(field, "is required")
		return
	}

	if !required && value == "" {
		return
	}

	// 检查路径是否包含非法字符
	if strings.ContainsAny(value, "<>\"|?*") {
		v.AddError(field, "contains invalid characters")
	}

	// 检查是否以文件名结尾（不是目录）
	if strings.HasSuffix(value, "/") || strings.HasSuffix(value, "\\") {
		v.AddError(field, "must be a file path, not a directory")
	}
}

// ValidateFileSize 验证文件大小
func (v *Validator) ValidateFileSize(field string, value int64, maxSize int64) {
	if value <= 0 {
		v.AddError(field, "must be greater than 0")
	}

	if maxSize > 0 && value > maxSize {
		v.AddError(field, fmt.Sprintf("must not exceed %d bytes", maxSize))
	}
}

// ValidateFileType 验证文件类型
func (v *Validator) ValidateFileType(field, value string, allowedTypes []string, required bool) {
	if required && value == "" {
		v.AddError(field, "is required")
		return
	}

	if !required && value == "" {
		return
	}

	valid := false
	for _, allowedType := range allowedTypes {
		if value == allowedType {
			valid = true
			break
		}
	}

	if !valid {
		v.AddError(field, fmt.Sprintf("must be one of: %s", strings.Join(allowedTypes, ", ")))
	}
}

// ValidateTags 验证标签
func (v *Validator) ValidateTags(field string, tags []string, maxTags int) {
	if len(tags) == 0 {
		return
	}

	if len(tags) > maxTags {
		v.AddError(field, fmt.Sprintf("must not exceed %d tags", maxTags))
	}

	for i, tag := range tags {
		if strings.TrimSpace(tag) == "" {
			v.AddError(field, fmt.Sprintf("tag %d cannot be empty", i+1))
		}

		if len(tag) > 50 {
			v.AddError(field, fmt.Sprintf("tag %d must not exceed 50 characters", i+1))
		}

		// 检查标签是否包含特殊字符
		if strings.ContainsAny(tag, "<>\"|?*") {
			v.AddError(field, fmt.Sprintf("tag %d contains invalid characters", i+1))
		}
	}
}

// ValidateJSONString 验证 JSON 字符串
func (v *Validator) ValidateJSONString(field, value string, required bool) {
	if required && value == "" {
		v.AddError(field, "is required")
		return
	}

	if !required && value == "" {
		return
	}

	// 简单的 JSON 格式检查
	if !strings.HasPrefix(value, "{") && !strings.HasPrefix(value, "[") {
		v.AddError(field, "must be valid JSON")
	}

	// 检查是否是有效的 JSON 格式（括号匹配）
	if strings.HasPrefix(value, "{") && !strings.HasSuffix(value, "}") {
		v.AddError(field, "invalid JSON format")
	}

	if strings.HasPrefix(value, "[") && !strings.HasSuffix(value, "]") {
		v.AddError(field, "invalid JSON format")
	}
}

// ValidateConfigKey 验证配置键
func (v *Validator) ValidateConfigKey(field, value string, required bool) {
	if required && value == "" {
		v.AddError(field, "is required")
		return
	}

	if !required && value == "" {
		return
	}

	// 配置键只允许字母、数字、下划线和点
	configKeyRegex := regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
	if !configKeyRegex.MatchString(value) {
		v.AddError(field, "can only contain letters, numbers, underscores and dots")
	}
}

// ValidateStoragePath 验证存储路径
func (v *Validator) ValidateStoragePath(field, value string, required bool) {
	if required && value == "" {
		v.AddError(field, "is required")
		return
	}

	if !required && value == "" {
		return
	}

	// 路径不能包含非法字符
	if strings.ContainsAny(value, "<>\"|?*") {
		v.AddError(field, "contains invalid characters")
	}

	// 路径不能以 / 或 \ 结尾
	if strings.HasSuffix(value, "/") || strings.HasSuffix(value, "\\") {
		v.AddError(field, "must not end with a path separator")
	}

	// 检查相对路径尝试（如 ../ 或 ./）
	if strings.Contains(value, "../") {
		v.AddError(field, "must not contain parent directory references")
	}
}