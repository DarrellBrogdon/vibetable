package formula

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEvaluator(t *testing.T) {
	evaluator := NewEvaluator()
	assert.NotNil(t, evaluator)
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"string", "hello", `"hello"`},
		{"string with quotes", `say "hello"`, `"say \"hello\""`},
		{"integer float", 42.0, "42"},
		{"decimal float", 3.14, "3.14"},
		{"int", 42, "42"},
		{"int64", int64(42), "42"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"array", []interface{}{"a", "b"}, `"a, b"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"integer float", 42.0, "42"},
		{"decimal float", 3.14, "3.14"},
		{"other", struct{}{}, "{}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"nil", nil, 0},
		{"float64", 3.14, 3.14},
		{"int", 42, 42.0},
		{"int64", int64(42), 42.0},
		{"string number", "3.14", 3.14},
		{"string invalid", "abc", 0},
		{"bool true", true, 1},
		{"bool false", false, 0},
		{"other", struct{}{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"nil", nil, false},
		{"bool true", true, true},
		{"bool false", false, false},
		{"string non-empty", "hello", true},
		{"string empty", "", false},
		{"string false", "false", false},
		{"string FALSE", "FALSE", false},
		{"float non-zero", 3.14, true},
		{"float zero", 0.0, false},
		{"other", struct{}{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toBool(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsBlank(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"whitespace string", "   ", true},
		{"non-empty string", "hello", false},
		{"empty array", []interface{}{}, true},
		{"non-empty array", []interface{}{"a"}, false},
		{"number", 42.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBlank(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToDate(t *testing.T) {
	t.Run("returns zero time for nil", func(t *testing.T) {
		result := toDate(nil)
		assert.True(t, result.IsZero())
	})

	t.Run("returns zero time for empty string", func(t *testing.T) {
		result := toDate("")
		assert.True(t, result.IsZero())
	})

	t.Run("parses ISO date", func(t *testing.T) {
		result := toDate("2024-01-15")
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, time.January, result.Month())
		assert.Equal(t, 15, result.Day())
	})

	t.Run("parses RFC3339 date", func(t *testing.T) {
		result := toDate("2024-01-15T10:30:00Z")
		assert.Equal(t, 2024, result.Year())
		assert.Equal(t, 10, result.Hour())
	})

	t.Run("returns zero time for invalid date", func(t *testing.T) {
		result := toDate("not-a-date")
		assert.True(t, result.IsZero())
	})
}

// String function tests

func TestFuncConcat(t *testing.T) {
	e := NewEvaluator()

	result, err := e.funcConcat([]interface{}{"Hello", " ", "World"})
	require.NoError(t, err)
	assert.Equal(t, "Hello World", result)
}

func TestFuncUpper(t *testing.T) {
	e := NewEvaluator()

	t.Run("converts to uppercase", func(t *testing.T) {
		result, err := e.funcUpper([]interface{}{"hello"})
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcUpper([]interface{}{"a", "b"})
		assert.Error(t, err)
	})
}

func TestFuncLower(t *testing.T) {
	e := NewEvaluator()

	t.Run("converts to lowercase", func(t *testing.T) {
		result, err := e.funcLower([]interface{}{"HELLO"})
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcLower([]interface{}{"a", "b"})
		assert.Error(t, err)
	})
}

func TestFuncLen(t *testing.T) {
	e := NewEvaluator()

	t.Run("returns string length", func(t *testing.T) {
		result, err := e.funcLen([]interface{}{"hello"})
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcLen([]interface{}{"a", "b"})
		assert.Error(t, err)
	})
}

func TestFuncTrim(t *testing.T) {
	e := NewEvaluator()

	t.Run("trims whitespace", func(t *testing.T) {
		result, err := e.funcTrim([]interface{}{"  hello  "})
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcTrim([]interface{}{"a", "b"})
		assert.Error(t, err)
	})
}

func TestFuncLeft(t *testing.T) {
	e := NewEvaluator()

	t.Run("returns left characters", func(t *testing.T) {
		result, err := e.funcLeft([]interface{}{"hello", 3.0})
		require.NoError(t, err)
		assert.Equal(t, "hel", result)
	})

	t.Run("handles n larger than string", func(t *testing.T) {
		result, err := e.funcLeft([]interface{}{"hi", 10.0})
		require.NoError(t, err)
		assert.Equal(t, "hi", result)
	})

	t.Run("handles negative n", func(t *testing.T) {
		result, err := e.funcLeft([]interface{}{"hello", -1.0})
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestFuncRight(t *testing.T) {
	e := NewEvaluator()

	t.Run("returns right characters", func(t *testing.T) {
		result, err := e.funcRight([]interface{}{"hello", 3.0})
		require.NoError(t, err)
		assert.Equal(t, "llo", result)
	})

	t.Run("handles n larger than string", func(t *testing.T) {
		result, err := e.funcRight([]interface{}{"hi", 10.0})
		require.NoError(t, err)
		assert.Equal(t, "hi", result)
	})
}

func TestFuncMid(t *testing.T) {
	e := NewEvaluator()

	t.Run("returns middle characters", func(t *testing.T) {
		result, err := e.funcMid([]interface{}{"hello", 2.0, 3.0})
		require.NoError(t, err)
		assert.Equal(t, "ell", result)
	})

	t.Run("handles start beyond string", func(t *testing.T) {
		result, err := e.funcMid([]interface{}{"hi", 10.0, 2.0})
		require.NoError(t, err)
		assert.Equal(t, "", result)
	})
}

func TestFuncSubstitute(t *testing.T) {
	e := NewEvaluator()

	t.Run("replaces substring", func(t *testing.T) {
		result, err := e.funcSubstitute([]interface{}{"hello world", "world", "there"})
		require.NoError(t, err)
		assert.Equal(t, "hello there", result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcSubstitute([]interface{}{"a", "b"})
		assert.Error(t, err)
	})
}

// Numeric function tests

func TestFuncSum(t *testing.T) {
	e := NewEvaluator()

	result, err := e.funcSum([]interface{}{1.0, 2.0, 3.0})
	require.NoError(t, err)
	assert.Equal(t, 6.0, result)
}

func TestFuncAverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("calculates average", func(t *testing.T) {
		result, err := e.funcAverage([]interface{}{2.0, 4.0, 6.0})
		require.NoError(t, err)
		assert.Equal(t, 4.0, result)
	})

	t.Run("returns 0 for empty args", func(t *testing.T) {
		result, err := e.funcAverage([]interface{}{})
		require.NoError(t, err)
		assert.Equal(t, 0.0, result)
	})
}

func TestFuncMin(t *testing.T) {
	e := NewEvaluator()

	t.Run("finds minimum", func(t *testing.T) {
		result, err := e.funcMin([]interface{}{3.0, 1.0, 2.0})
		require.NoError(t, err)
		assert.Equal(t, 1.0, result)
	})

	t.Run("returns 0 for empty args", func(t *testing.T) {
		result, err := e.funcMin([]interface{}{})
		require.NoError(t, err)
		assert.Equal(t, 0.0, result)
	})
}

func TestFuncMax(t *testing.T) {
	e := NewEvaluator()

	t.Run("finds maximum", func(t *testing.T) {
		result, err := e.funcMax([]interface{}{1.0, 3.0, 2.0})
		require.NoError(t, err)
		assert.Equal(t, 3.0, result)
	})

	t.Run("returns 0 for empty args", func(t *testing.T) {
		result, err := e.funcMax([]interface{}{})
		require.NoError(t, err)
		assert.Equal(t, 0.0, result)
	})
}

func TestFuncRound(t *testing.T) {
	e := NewEvaluator()

	t.Run("rounds to nearest integer by default", func(t *testing.T) {
		result, err := e.funcRound([]interface{}{3.7})
		require.NoError(t, err)
		assert.Equal(t, 4.0, result)
	})

	t.Run("rounds to specified precision", func(t *testing.T) {
		result, err := e.funcRound([]interface{}{3.14159, 2.0})
		require.NoError(t, err)
		assert.Equal(t, 3.14, result)
	})

	t.Run("returns error for no args", func(t *testing.T) {
		_, err := e.funcRound([]interface{}{})
		assert.Error(t, err)
	})
}

func TestFuncFloor(t *testing.T) {
	e := NewEvaluator()

	t.Run("floors number", func(t *testing.T) {
		result, err := e.funcFloor([]interface{}{3.7})
		require.NoError(t, err)
		assert.Equal(t, 3.0, result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcFloor([]interface{}{1.0, 2.0})
		assert.Error(t, err)
	})
}

func TestFuncCeiling(t *testing.T) {
	e := NewEvaluator()

	t.Run("ceils number", func(t *testing.T) {
		result, err := e.funcCeiling([]interface{}{3.2})
		require.NoError(t, err)
		assert.Equal(t, 4.0, result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcCeiling([]interface{}{1.0, 2.0})
		assert.Error(t, err)
	})
}

func TestFuncAbs(t *testing.T) {
	e := NewEvaluator()

	t.Run("returns absolute value", func(t *testing.T) {
		result, err := e.funcAbs([]interface{}{-5.0})
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcAbs([]interface{}{1.0, 2.0})
		assert.Error(t, err)
	})
}

// Logic function tests

func TestFuncIf(t *testing.T) {
	e := NewEvaluator()

	t.Run("returns then value when true", func(t *testing.T) {
		result, err := e.funcIf([]interface{}{true, "yes", "no"})
		require.NoError(t, err)
		assert.Equal(t, "yes", result)
	})

	t.Run("returns else value when false", func(t *testing.T) {
		result, err := e.funcIf([]interface{}{false, "yes", "no"})
		require.NoError(t, err)
		assert.Equal(t, "no", result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcIf([]interface{}{true, "yes"})
		assert.Error(t, err)
	})
}

func TestFuncAnd(t *testing.T) {
	e := NewEvaluator()

	t.Run("returns true when all true", func(t *testing.T) {
		result, err := e.funcAnd([]interface{}{true, true, true})
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("returns false when any false", func(t *testing.T) {
		result, err := e.funcAnd([]interface{}{true, false, true})
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("returns true for empty args", func(t *testing.T) {
		result, err := e.funcAnd([]interface{}{})
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})
}

func TestFuncOr(t *testing.T) {
	e := NewEvaluator()

	t.Run("returns true when any true", func(t *testing.T) {
		result, err := e.funcOr([]interface{}{false, true, false})
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("returns false when all false", func(t *testing.T) {
		result, err := e.funcOr([]interface{}{false, false, false})
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("returns false for empty args", func(t *testing.T) {
		result, err := e.funcOr([]interface{}{})
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})
}

func TestFuncNot(t *testing.T) {
	e := NewEvaluator()

	t.Run("negates true", func(t *testing.T) {
		result, err := e.funcNot([]interface{}{true})
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("negates false", func(t *testing.T) {
		result, err := e.funcNot([]interface{}{false})
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcNot([]interface{}{true, false})
		assert.Error(t, err)
	})
}

func TestFuncIsBlank(t *testing.T) {
	e := NewEvaluator()

	t.Run("returns true for blank", func(t *testing.T) {
		result, err := e.funcIsBlank([]interface{}{""})
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("returns false for non-blank", func(t *testing.T) {
		result, err := e.funcIsBlank([]interface{}{"hello"})
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("returns error for wrong arg count", func(t *testing.T) {
		_, err := e.funcIsBlank([]interface{}{"a", "b"})
		assert.Error(t, err)
	})
}

// Date function tests

func TestFuncToday(t *testing.T) {
	e := NewEvaluator()

	result, err := e.funcToday([]interface{}{})
	require.NoError(t, err)

	today := time.Now().Format("2006-01-02")
	assert.Equal(t, today, result)
}

func TestFuncNow(t *testing.T) {
	e := NewEvaluator()

	result, err := e.funcNow([]interface{}{})
	require.NoError(t, err)

	resultStr := result.(string)
	_, err = time.Parse(time.RFC3339, resultStr)
	assert.NoError(t, err)
}

func TestFuncYear(t *testing.T) {
	e := NewEvaluator()

	t.Run("extracts year", func(t *testing.T) {
		result, err := e.funcYear([]interface{}{"2024-01-15"})
		require.NoError(t, err)
		assert.Equal(t, 2024.0, result)
	})

	t.Run("returns 0 for invalid date", func(t *testing.T) {
		result, err := e.funcYear([]interface{}{"not-a-date"})
		require.NoError(t, err)
		assert.Equal(t, 0.0, result)
	})
}

func TestFuncMonth(t *testing.T) {
	e := NewEvaluator()

	t.Run("extracts month", func(t *testing.T) {
		result, err := e.funcMonth([]interface{}{"2024-06-15"})
		require.NoError(t, err)
		assert.Equal(t, 6.0, result)
	})
}

func TestFuncDay(t *testing.T) {
	e := NewEvaluator()

	t.Run("extracts day", func(t *testing.T) {
		result, err := e.funcDay([]interface{}{"2024-01-25"})
		require.NoError(t, err)
		assert.Equal(t, 25.0, result)
	})
}

// Arithmetic evaluation tests

func TestEvaluateArithmetic(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name     string
		expr     string
		expected float64
	}{
		{"simple number", "42", 42.0},
		{"addition", "1 + 2", 3.0},
		{"subtraction", "5 - 3", 2.0},
		{"multiplication", "3 * 4", 12.0},
		{"division", "10 / 2", 5.0},
		{"combined", "2 + 3 * 4", 14.0},
		{"with parens", "(2 + 3) * 4", 20.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.evaluateArithmetic(tt.expr)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("division by zero", func(t *testing.T) {
		_, err := e.evaluateArithmetic("10 / 0")
		assert.Error(t, err)
	})
}

// Full expression evaluation tests

func TestEvaluateExpression(t *testing.T) {
	e := NewEvaluator()

	t.Run("evaluates string literal", func(t *testing.T) {
		result, err := e.evaluateExpression(`"hello"`)
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("evaluates number", func(t *testing.T) {
		result, err := e.evaluateExpression("42")
		require.NoError(t, err)
		assert.Equal(t, 42.0, result)
	})

	t.Run("evaluates true", func(t *testing.T) {
		result, err := e.evaluateExpression("true")
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("evaluates false", func(t *testing.T) {
		result, err := e.evaluateExpression("false")
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})
}

func TestEvaluate(t *testing.T) {
	e := NewEvaluator()

	t.Run("resolves field references", func(t *testing.T) {
		resolver := func(fieldRef string) (interface{}, error) {
			if fieldRef == "name" {
				return "John", nil
			}
			return nil, fmt.Errorf("unknown field: %s", fieldRef)
		}

		result, err := e.Evaluate(`UPPER({name})`, resolver)
		require.NoError(t, err)
		assert.Equal(t, "JOHN", result)
	})

	t.Run("returns error for unresolved field", func(t *testing.T) {
		resolver := func(fieldRef string) (interface{}, error) {
			return nil, fmt.Errorf("unknown field: %s", fieldRef)
		}

		_, err := e.Evaluate(`UPPER({unknown})`, resolver)
		assert.Error(t, err)
	})
}

func TestCallFunction(t *testing.T) {
	e := NewEvaluator()

	t.Run("returns error for unknown function", func(t *testing.T) {
		_, err := e.callFunction("UNKNOWN", []interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown function")
	})

	// Test string functions through callFunction
	t.Run("CONCAT", func(t *testing.T) {
		result, err := e.callFunction("CONCAT", []interface{}{"hello", " ", "world"})
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)
	})

	t.Run("UPPER", func(t *testing.T) {
		result, err := e.callFunction("UPPER", []interface{}{"hello"})
		require.NoError(t, err)
		assert.Equal(t, "HELLO", result)
	})

	t.Run("LOWER", func(t *testing.T) {
		result, err := e.callFunction("LOWER", []interface{}{"HELLO"})
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("LEN", func(t *testing.T) {
		result, err := e.callFunction("LEN", []interface{}{"hello"})
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)
	})

	t.Run("TRIM", func(t *testing.T) {
		result, err := e.callFunction("TRIM", []interface{}{"  hello  "})
		require.NoError(t, err)
		assert.Equal(t, "hello", result)
	})

	t.Run("LEFT", func(t *testing.T) {
		result, err := e.callFunction("LEFT", []interface{}{"hello", 3.0})
		require.NoError(t, err)
		assert.Equal(t, "hel", result)
	})

	t.Run("RIGHT", func(t *testing.T) {
		result, err := e.callFunction("RIGHT", []interface{}{"hello", 3.0})
		require.NoError(t, err)
		assert.Equal(t, "llo", result)
	})

	t.Run("MID", func(t *testing.T) {
		result, err := e.callFunction("MID", []interface{}{"hello", 2.0, 3.0})
		require.NoError(t, err)
		assert.Equal(t, "ell", result)
	})

	t.Run("SUBSTITUTE", func(t *testing.T) {
		result, err := e.callFunction("SUBSTITUTE", []interface{}{"hello world", "world", "there"})
		require.NoError(t, err)
		assert.Equal(t, "hello there", result)
	})

	// Test numeric functions through callFunction
	t.Run("SUM", func(t *testing.T) {
		result, err := e.callFunction("SUM", []interface{}{1.0, 2.0, 3.0})
		require.NoError(t, err)
		assert.Equal(t, 6.0, result)
	})

	t.Run("AVERAGE", func(t *testing.T) {
		result, err := e.callFunction("AVERAGE", []interface{}{1.0, 2.0, 3.0})
		require.NoError(t, err)
		assert.Equal(t, 2.0, result)
	})

	t.Run("AVG alias", func(t *testing.T) {
		result, err := e.callFunction("AVG", []interface{}{10.0, 20.0})
		require.NoError(t, err)
		assert.Equal(t, 15.0, result)
	})

	t.Run("MIN", func(t *testing.T) {
		result, err := e.callFunction("MIN", []interface{}{5.0, 2.0, 8.0})
		require.NoError(t, err)
		assert.Equal(t, 2.0, result)
	})

	t.Run("MAX", func(t *testing.T) {
		result, err := e.callFunction("MAX", []interface{}{5.0, 2.0, 8.0})
		require.NoError(t, err)
		assert.Equal(t, 8.0, result)
	})

	t.Run("ROUND", func(t *testing.T) {
		result, err := e.callFunction("ROUND", []interface{}{3.456, 2.0})
		require.NoError(t, err)
		assert.Equal(t, 3.46, result)
	})

	t.Run("FLOOR", func(t *testing.T) {
		result, err := e.callFunction("FLOOR", []interface{}{3.7})
		require.NoError(t, err)
		assert.Equal(t, 3.0, result)
	})

	t.Run("CEILING", func(t *testing.T) {
		result, err := e.callFunction("CEILING", []interface{}{3.2})
		require.NoError(t, err)
		assert.Equal(t, 4.0, result)
	})

	t.Run("CEIL alias", func(t *testing.T) {
		result, err := e.callFunction("CEIL", []interface{}{3.1})
		require.NoError(t, err)
		assert.Equal(t, 4.0, result)
	})

	t.Run("ABS", func(t *testing.T) {
		result, err := e.callFunction("ABS", []interface{}{-5.0})
		require.NoError(t, err)
		assert.Equal(t, 5.0, result)
	})

	// Test logic functions through callFunction
	t.Run("IF", func(t *testing.T) {
		result, err := e.callFunction("IF", []interface{}{true, "yes", "no"})
		require.NoError(t, err)
		assert.Equal(t, "yes", result)
	})

	t.Run("AND", func(t *testing.T) {
		result, err := e.callFunction("AND", []interface{}{true, true})
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("OR", func(t *testing.T) {
		result, err := e.callFunction("OR", []interface{}{false, true})
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("NOT", func(t *testing.T) {
		result, err := e.callFunction("NOT", []interface{}{false})
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	t.Run("ISBLANK", func(t *testing.T) {
		result, err := e.callFunction("ISBLANK", []interface{}{""})
		require.NoError(t, err)
		assert.Equal(t, true, result)
	})

	// Test date functions through callFunction
	t.Run("TODAY", func(t *testing.T) {
		result, err := e.callFunction("TODAY", []interface{}{})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("NOW", func(t *testing.T) {
		result, err := e.callFunction("NOW", []interface{}{})
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("YEAR", func(t *testing.T) {
		result, err := e.callFunction("YEAR", []interface{}{"2024-06-15"})
		require.NoError(t, err)
		assert.Equal(t, float64(2024), result)
	})

	t.Run("MONTH", func(t *testing.T) {
		result, err := e.callFunction("MONTH", []interface{}{"2024-06-15"})
		require.NoError(t, err)
		assert.Equal(t, float64(6), result)
	})

	t.Run("DAY", func(t *testing.T) {
		result, err := e.callFunction("DAY", []interface{}{"2024-06-15"})
		require.NoError(t, err)
		assert.Equal(t, float64(15), result)
	})
}

func TestParseArguments(t *testing.T) {
	e := NewEvaluator()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"simple", "a, b, c", []string{"a", "b", "c"}},
		{"with quotes", `"a, b", c`, []string{`"a, b"`, "c"}},
		{"with nested parens", "SUM(1, 2), 3", []string{"SUM(1, 2)", "3"}},
		{"empty", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := e.parseArguments(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
