package atomany

import "errors"

// 定义原子操作相关的错误
var (
	// ErrUninitialized 表示原子值未初始化
	ErrUninitialized = errors.New("atomany: value not initialized")

	// ErrTypeMismatch 表示类型不匹配
	ErrTypeMismatch = errors.New("atomany: type mismatch")

	// ErrNilValue 表示不允许的 nil 值
	ErrNilValue = errors.New("atomany: nil value not allowed")

	// ErrInconsistentType 表示类型不一致
	ErrInconsistentType = errors.New("atomany: store of inconsistently typed value")
)

// IsUninitialized 检查错误是否为未初始化错误
func IsUninitialized(err error) bool {
	return errors.Is(err, ErrUninitialized)
}

// IsTypeMismatch 检查错误是否为类型不匹配错误
func IsTypeMismatch(err error) bool {
	return errors.Is(err, ErrTypeMismatch)
}

// IsNilValue 检查错误是否为 nil 值错误
func IsNilValue(err error) bool {
	return errors.Is(err, ErrNilValue)
}

// IsInconsistentType 检查错误是否为类型不一致错误
func IsInconsistentType(err error) bool {
	return errors.Is(err, ErrInconsistentType)
}
