package models

import (
	"fmt"
	"strings"
	"time"
)

// GenerateID はユニークなIDを生成します。
func GenerateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// ParseOrderSide は文字列からOrderSideに変換します。
func ParseOrderSide(s string) (OrderSide, error) {
	switch strings.ToLower(s) {
	case "buy":
		return Buy, nil
	case "sell":
		return Sell, nil
	default:
		return Buy, fmt.Errorf("invalid order side: %s", s)
	}
}

// ParseOrderType は文字列からOrderTypeに変換します。
func ParseOrderType(s string) (OrderType, error) {
	switch strings.ToLower(s) {
	case "market":
		return Market, nil
	case "limit":
		return Limit, nil
	case "stop":
		return Stop, nil
	default:
		return Market, fmt.Errorf("invalid order type: %s", s)
	}
}

// Validator はデータバリデーションのインターフェースです。
type Validator interface {
	Validate() error
}

// ValidationError はバリデーションエラーを表します。
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
}

// ValidateStruct は構造体全体のバリデーションを行います。
func ValidateStruct(v Validator) error {
	return v.Validate()
}