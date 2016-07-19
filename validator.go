package jsonapi

import (
	"fmt"
	"regexp"
	"unicode/utf8"
)

// Validator structure
type Validator struct {
	Errors
}

// Verify returns error if errors present and nil if empty
func (v Validator) Verify() error {
	if v.HasErrors() {
		return v.Errors
	}
	return nil
}

// Present validates string for presence
// 	v.Present(SomeVariable, "name")
func (v *Validator) Present(value, pointer string) {
	if value == "" {
		v.AddError(ErrorInvalidAttribute(pointer, pointer+" can't be blank"))
	}
}

// StringLength validates string min, max length. -1 for any
// 	v.StringLength(SomeVariable, "password", 6, 18) // min 6, max 18
func (v *Validator) StringLength(value, pointer string, min, max int) {
	if min > 0 && utf8.RuneCountInString(value) < min {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("min length is", min)))
	}
	if max > 0 && utf8.RuneCountInString(value) > max {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("max length is", max)))
	}
}

// Int validates int min, max. -1 for any
// 	v.Int(IntValue, "number", -1, 11)  // max 18
func (v *Validator) Int(value int, pointer string, min, max int) {
	if min > 0 && value < min {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("min value is", min)))
	}
	if max > 0 && value > max {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("max value is", max)))
	}
}

// Int64 validates int min, max. -1 for any
// 	v.Int64(Int64Value, "number", -1, 11)  // max 18
func (v *Validator) Int64(value int64, pointer string, min, max int64) {
	if min > 0 && value < min {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("min value is", min)))
	}
	if max > 0 && value > max {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("max value is", max)))
	}
}

// Uint64 validates int min, max. -1 for any
// 	v.Uint64(Uint64Value, "number", -1, 11)  // max 18
func (v *Validator) Uint64(value uint64, pointer string, min, max int) {
	if min > 0 && value < uint64(min) {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("min value is", min)))
	}
	if max > 0 && value > uint64(max) {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("max value is", max)))
	}
}

// Float32 validates int min, max. -1 for any
// 	v.Float32(Float32Value, "number", -1, 11)  // max 18
func (v *Validator) Float32(value float32, pointer string, min, max float32) {
	if min > 0 && value < min {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("min value is", min)))
	}
	if max > 0 && value > max {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("max value is", max)))
	}
}

// Float64 validates int min, max. -1 for any
// 	v.Float64(Float64Value, "number", -1, 11)  // max 18
func (v *Validator) Float64(value float64, pointer string, min, max float64) {
	if min > 0 && value < min {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("min value is", min)))
	}
	if max > 0 && value > max {
		v.AddError(ErrorInvalidAttribute(pointer, fmt.Sprint("max value is", max)))
	}
}

// Format validates string format with regex string
// 	v.Format(StringValue,"ip address", `\A(\d{1,3}\.){3}\d{1,3}\z`)
func (v *Validator) Format(value, pointer, reg string) {
	if r, _ := regexp.MatchString(reg, value); !r {
		v.AddError(ErrorInvalidAttribute(pointer, "invalid format"))
	}
}
