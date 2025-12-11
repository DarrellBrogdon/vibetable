package formula

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Evaluator handles formula evaluation
type Evaluator struct{}

// NewEvaluator creates a new formula evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// FieldResolver is a function that resolves field references to values
type FieldResolver func(fieldRef string) (interface{}, error)

// Evaluate evaluates a formula expression with the given field resolver
func (e *Evaluator) Evaluate(expression string, resolver FieldResolver) (interface{}, error) {
	// First, resolve all field references
	resolved, err := e.resolveFields(expression, resolver)
	if err != nil {
		return nil, err
	}

	// Then evaluate the expression
	return e.evaluateExpression(resolved)
}

// resolveFields replaces {field} references with actual values
func (e *Evaluator) resolveFields(expression string, resolver FieldResolver) (string, error) {
	// Match {field_name} or {field_id} patterns
	re := regexp.MustCompile(`\{([^}]+)\}`)

	var lastErr error
	result := re.ReplaceAllStringFunc(expression, func(match string) string {
		fieldRef := match[1 : len(match)-1] // Remove { and }
		value, err := resolver(fieldRef)
		if err != nil {
			lastErr = err
			return match
		}
		return formatValue(value)
	})

	if lastErr != nil {
		return "", lastErr
	}
	return result, nil
}

// formatValue converts a value to its string representation for formula evaluation
func formatValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(val, `"`, `\"`))
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case int, int64:
		return fmt.Sprintf("%d", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case []interface{}:
		// For arrays (like multi-select), join with comma
		parts := make([]string, len(val))
		for i, item := range val {
			parts[i] = fmt.Sprintf("%v", item)
		}
		return fmt.Sprintf(`"%s"`, strings.Join(parts, ", "))
	default:
		return fmt.Sprintf("%v", val)
	}
}

// evaluateExpression evaluates a formula expression
func (e *Evaluator) evaluateExpression(expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)

	// Check for function calls
	if funcResult, handled, err := e.tryEvaluateFunction(expr); handled {
		return funcResult, err
	}

	// Check for simple string literal
	if strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`) {
		return strings.ReplaceAll(expr[1:len(expr)-1], `\"`, `"`), nil
	}

	// Check for number
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return num, nil
	}

	// Check for boolean
	if expr == "true" {
		return true, nil
	}
	if expr == "false" {
		return false, nil
	}

	// Try to evaluate as arithmetic expression
	if result, err := e.evaluateArithmetic(expr); err == nil {
		return result, nil
	}

	// Return as string
	return expr, nil
}

// tryEvaluateFunction tries to evaluate a function call
func (e *Evaluator) tryEvaluateFunction(expr string) (interface{}, bool, error) {
	// Match FUNCTION_NAME(args)
	re := regexp.MustCompile(`^([A-Z_]+)\((.*)\)$`)
	matches := re.FindStringSubmatch(expr)
	if matches == nil {
		return nil, false, nil
	}

	funcName := matches[1]
	argsStr := matches[2]

	// Parse arguments
	args, err := e.parseArguments(argsStr)
	if err != nil {
		return nil, true, err
	}

	// Evaluate each argument
	evalArgs := make([]interface{}, len(args))
	for i, arg := range args {
		evalArgs[i], err = e.evaluateExpression(arg)
		if err != nil {
			return nil, true, err
		}
	}

	// Call the function
	result, err := e.callFunction(funcName, evalArgs)
	return result, true, err
}

// parseArguments parses comma-separated arguments, respecting quotes and nested parens
func (e *Evaluator) parseArguments(argsStr string) ([]string, error) {
	var args []string
	var current strings.Builder
	depth := 0
	inQuotes := false
	escaped := false

	for _, ch := range argsStr {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}

		switch ch {
		case '\\':
			escaped = true
			current.WriteRune(ch)
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(ch)
		case '(':
			if !inQuotes {
				depth++
			}
			current.WriteRune(ch)
		case ')':
			if !inQuotes {
				depth--
			}
			current.WriteRune(ch)
		case ',':
			if !inQuotes && depth == 0 {
				args = append(args, strings.TrimSpace(current.String()))
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 || len(args) > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}

	return args, nil
}

// callFunction calls a formula function by name
func (e *Evaluator) callFunction(name string, args []interface{}) (interface{}, error) {
	switch name {
	// String functions
	case "CONCAT":
		return e.funcConcat(args)
	case "UPPER":
		return e.funcUpper(args)
	case "LOWER":
		return e.funcLower(args)
	case "LEN":
		return e.funcLen(args)
	case "TRIM":
		return e.funcTrim(args)
	case "LEFT":
		return e.funcLeft(args)
	case "RIGHT":
		return e.funcRight(args)
	case "MID":
		return e.funcMid(args)
	case "SUBSTITUTE":
		return e.funcSubstitute(args)

	// Numeric functions
	case "SUM":
		return e.funcSum(args)
	case "AVERAGE", "AVG":
		return e.funcAverage(args)
	case "MIN":
		return e.funcMin(args)
	case "MAX":
		return e.funcMax(args)
	case "ROUND":
		return e.funcRound(args)
	case "FLOOR":
		return e.funcFloor(args)
	case "CEILING", "CEIL":
		return e.funcCeiling(args)
	case "ABS":
		return e.funcAbs(args)

	// Logic functions
	case "IF":
		return e.funcIf(args)
	case "AND":
		return e.funcAnd(args)
	case "OR":
		return e.funcOr(args)
	case "NOT":
		return e.funcNot(args)
	case "ISBLANK":
		return e.funcIsBlank(args)

	// Date functions
	case "TODAY":
		return e.funcToday(args)
	case "NOW":
		return e.funcNow(args)
	case "YEAR":
		return e.funcYear(args)
	case "MONTH":
		return e.funcMonth(args)
	case "DAY":
		return e.funcDay(args)

	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}

// String functions

func (e *Evaluator) funcConcat(args []interface{}) (interface{}, error) {
	var result strings.Builder
	for _, arg := range args {
		result.WriteString(toString(arg))
	}
	return result.String(), nil
}

func (e *Evaluator) funcUpper(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("UPPER requires exactly 1 argument")
	}
	return strings.ToUpper(toString(args[0])), nil
}

func (e *Evaluator) funcLower(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("LOWER requires exactly 1 argument")
	}
	return strings.ToLower(toString(args[0])), nil
}

func (e *Evaluator) funcLen(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("LEN requires exactly 1 argument")
	}
	return float64(len(toString(args[0]))), nil
}

func (e *Evaluator) funcTrim(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("TRIM requires exactly 1 argument")
	}
	return strings.TrimSpace(toString(args[0])), nil
}

func (e *Evaluator) funcLeft(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("LEFT requires exactly 2 arguments")
	}
	str := toString(args[0])
	n := int(toNumber(args[1]))
	if n < 0 {
		n = 0
	}
	if n > len(str) {
		n = len(str)
	}
	return str[:n], nil
}

func (e *Evaluator) funcRight(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("RIGHT requires exactly 2 arguments")
	}
	str := toString(args[0])
	n := int(toNumber(args[1]))
	if n < 0 {
		n = 0
	}
	if n > len(str) {
		n = len(str)
	}
	return str[len(str)-n:], nil
}

func (e *Evaluator) funcMid(args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("MID requires exactly 3 arguments")
	}
	str := toString(args[0])
	start := int(toNumber(args[1])) - 1 // 1-indexed
	length := int(toNumber(args[2]))
	if start < 0 {
		start = 0
	}
	if start > len(str) {
		return "", nil
	}
	end := start + length
	if end > len(str) {
		end = len(str)
	}
	return str[start:end], nil
}

func (e *Evaluator) funcSubstitute(args []interface{}) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("SUBSTITUTE requires at least 3 arguments")
	}
	str := toString(args[0])
	old := toString(args[1])
	new := toString(args[2])
	return strings.ReplaceAll(str, old, new), nil
}

// Numeric functions

func (e *Evaluator) funcSum(args []interface{}) (interface{}, error) {
	sum := 0.0
	for _, arg := range args {
		sum += toNumber(arg)
	}
	return sum, nil
}

func (e *Evaluator) funcAverage(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0.0, nil
	}
	sum := 0.0
	for _, arg := range args {
		sum += toNumber(arg)
	}
	return sum / float64(len(args)), nil
}

func (e *Evaluator) funcMin(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0.0, nil
	}
	min := toNumber(args[0])
	for _, arg := range args[1:] {
		if v := toNumber(arg); v < min {
			min = v
		}
	}
	return min, nil
}

func (e *Evaluator) funcMax(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return 0.0, nil
	}
	max := toNumber(args[0])
	for _, arg := range args[1:] {
		if v := toNumber(arg); v > max {
			max = v
		}
	}
	return max, nil
}

func (e *Evaluator) funcRound(args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("ROUND requires at least 1 argument")
	}
	num := toNumber(args[0])
	precision := 0
	if len(args) > 1 {
		precision = int(toNumber(args[1]))
	}
	mult := math.Pow(10, float64(precision))
	return math.Round(num*mult) / mult, nil
}

func (e *Evaluator) funcFloor(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("FLOOR requires exactly 1 argument")
	}
	return math.Floor(toNumber(args[0])), nil
}

func (e *Evaluator) funcCeiling(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("CEILING requires exactly 1 argument")
	}
	return math.Ceil(toNumber(args[0])), nil
}

func (e *Evaluator) funcAbs(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ABS requires exactly 1 argument")
	}
	return math.Abs(toNumber(args[0])), nil
}

// Logic functions

func (e *Evaluator) funcIf(args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("IF requires exactly 3 arguments")
	}
	if toBool(args[0]) {
		return args[1], nil
	}
	return args[2], nil
}

func (e *Evaluator) funcAnd(args []interface{}) (interface{}, error) {
	for _, arg := range args {
		if !toBool(arg) {
			return false, nil
		}
	}
	return true, nil
}

func (e *Evaluator) funcOr(args []interface{}) (interface{}, error) {
	for _, arg := range args {
		if toBool(arg) {
			return true, nil
		}
	}
	return false, nil
}

func (e *Evaluator) funcNot(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("NOT requires exactly 1 argument")
	}
	return !toBool(args[0]), nil
}

func (e *Evaluator) funcIsBlank(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ISBLANK requires exactly 1 argument")
	}
	return isBlank(args[0]), nil
}

// Date functions

func (e *Evaluator) funcToday(args []interface{}) (interface{}, error) {
	return time.Now().Format("2006-01-02"), nil
}

func (e *Evaluator) funcNow(args []interface{}) (interface{}, error) {
	return time.Now().Format(time.RFC3339), nil
}

func (e *Evaluator) funcYear(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("YEAR requires exactly 1 argument")
	}
	t := toDate(args[0])
	if t.IsZero() {
		return 0.0, nil
	}
	return float64(t.Year()), nil
}

func (e *Evaluator) funcMonth(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("MONTH requires exactly 1 argument")
	}
	t := toDate(args[0])
	if t.IsZero() {
		return 0.0, nil
	}
	return float64(t.Month()), nil
}

func (e *Evaluator) funcDay(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("DAY requires exactly 1 argument")
	}
	t := toDate(args[0])
	if t.IsZero() {
		return 0.0, nil
	}
	return float64(t.Day()), nil
}

// Arithmetic evaluation

func (e *Evaluator) evaluateArithmetic(expr string) (float64, error) {
	// Simple arithmetic parser for +, -, *, /
	// This is a basic implementation - for complex math, consider using a proper parser

	expr = strings.TrimSpace(expr)

	// Try to parse as a simple number
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return num, nil
	}

	// Look for + or - (lowest precedence)
	depth := 0
	for i := len(expr) - 1; i >= 0; i-- {
		ch := expr[i]
		if ch == ')' {
			depth++
		} else if ch == '(' {
			depth--
		} else if depth == 0 && (ch == '+' || ch == '-') && i > 0 {
			left, err := e.evaluateArithmetic(expr[:i])
			if err != nil {
				return 0, err
			}
			right, err := e.evaluateArithmetic(expr[i+1:])
			if err != nil {
				return 0, err
			}
			if ch == '+' {
				return left + right, nil
			}
			return left - right, nil
		}
	}

	// Look for * or / (higher precedence)
	for i := len(expr) - 1; i >= 0; i-- {
		ch := expr[i]
		if ch == ')' {
			depth++
		} else if ch == '(' {
			depth--
		} else if depth == 0 && (ch == '*' || ch == '/') {
			left, err := e.evaluateArithmetic(expr[:i])
			if err != nil {
				return 0, err
			}
			right, err := e.evaluateArithmetic(expr[i+1:])
			if err != nil {
				return 0, err
			}
			if ch == '*' {
				return left * right, nil
			}
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		}
	}

	// Handle parentheses
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		return e.evaluateArithmetic(expr[1 : len(expr)-1])
	}

	return 0, fmt.Errorf("cannot evaluate arithmetic expression: %s", expr)
}

// Helper functions

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func toNumber(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			return num
		}
		return 0
	case bool:
		if val {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val != "" && strings.ToLower(val) != "false"
	case float64:
		return val != 0
	case int, int64:
		return val != 0
	default:
		return true
	}
}

func isBlank(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case []interface{}:
		return len(val) == 0
	default:
		return false
	}
}

func toDate(v interface{}) time.Time {
	if v == nil {
		return time.Time{}
	}
	str := toString(v)
	if str == "" {
		return time.Time{}
	}

	// Try common date formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		"01/02/2006",
		"02-01-2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			return t
		}
	}

	return time.Time{}
}
