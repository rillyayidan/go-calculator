package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

type Calculator struct {
	reader     *bufio.Reader
	history    []string
	results    []float64
	lastResult float64
	hasLast    bool
	useDegrees bool
	precision  int
}

func main() {
	calc := NewCalculator()
	calc.Run()
}

/* =======================
   Core Application
======================= */

func NewCalculator() *Calculator {
	return &Calculator{
		reader:    bufio.NewReader(os.Stdin),
		precision: -1,
	}
}

func (c *Calculator) Run() {
	fmt.Println("=== Advanced CLI Calculator ===")
	fmt.Println("Type 'help' to see available commands.")

	for {
		action, op, ok := c.readOperator()
		if !ok {
			fmt.Println("Goodbye.")
			return
		}

		if c.handleCommand(action) {
			continue
		}

		numbers, ok, err := c.readNumbers(op)
		if !ok {
			fmt.Println("Goodbye.")
			return
		}
		if err != nil {
			fmt.Println("Input error:", err)
			continue
		}

		result, err := calculateMany(op, numbers, c.useDegrees)
		if err != nil {
			fmt.Println("Calculation error:", err)
			continue
		}

		c.lastResult = result
		c.hasLast = true

		expr := formatExpression(numbers, op)
		c.history = append(c.history, fmt.Sprintf("%s = %s", expr, c.formatResult(result)))
		c.results = append(c.results, result)

		fmt.Printf("Result -> %s\n", c.formatResult(result))
	}
}

/* =======================
   Operator & Commands
======================= */

func (c *Calculator) readOperator() (string, string, bool) {
	for {
		fmt.Print("\nOperator (+, -, *, /, ^, %, sin, cos, tan, sqrt, log)\nCommand (help, history, degrees, radians, export, clear, exit)\n> ")
		line, ok, err := readLine(c.reader)
		if err != nil || !ok {
			return "", "", false
		}

		line = strings.ToLower(line)

		if line == "exit" {
			return "exit", "", false
		}

		if isCommand(line) {
			return line, "", true
		}

		if mapped := normalizeOperator(line); mapped != "" {
			return "op", mapped, true
		}

		fmt.Println("Invalid operator or command.")
	}
}

func isCommand(cmd string) bool {
	switch cmd {
	case "help", "history", "degrees", "radians", "export", "clear", "mode", "precision", "stats", "ops":
		return true
	default:
		return false
	}
}

func normalizeOperator(op string) string {
	switch op {
	case "+", "-", "*", "/", "^", "%", "sin", "cos", "tan", "sqrt", "log":
		return op
	case "pow":
		return "^"
	case "mod":
		return "%"
	case "ln":
		return "log"
	default:
		return ""
	}
}

func (c *Calculator) handleCommand(cmd string) bool {
	switch cmd {
	case "help":
		printHelp(c.useDegrees)
	case "history":
		printHistory(c.history)
	case "degrees":
		c.useDegrees = true
		fmt.Println("Trig mode set to degrees.")
	case "radians":
		c.useDegrees = false
		fmt.Println("Trig mode set to radians.")
	case "export":
		if err := exportHistory(c.reader, c.history); err != nil {
			fmt.Println("Export error:", err)
		}
	case "mode":
		if c.useDegrees {
			fmt.Println("Mode: degrees")
		} else {
			fmt.Println("Mode: radians")
		}
		if c.hasLast {
			fmt.Printf("Last result: %g\n", c.lastResult)
		} else {
			fmt.Println("Last result: none")
		}
	case "precision":
		if err := c.setPrecision(); err != nil {
			fmt.Println("Precision error:", err)
		}
	case "stats":
		c.printStats()
	case "ops":
		printOperators()
	case "clear":
		c.history = nil
		c.results = nil
		c.hasLast = false
		fmt.Println("Memory cleared.")
	default:
		return false
	}
	return true
}

/* =======================
   Input Handling
======================= */

func (c *Calculator) readNumbers(op string) ([]float64, bool, error) {
	fmt.Println("Enter numbers (blank line to finish). You can enter multiple numbers per line.")
	if c.hasLast {
		fmt.Println("Tip: type 'ans' to reuse last result.")
	}

	var numbers []float64
	for {
		fmt.Printf("Numbers (next %d+): ", len(numbers)+1)
		line, ok, err := readLine(c.reader)
		if err != nil || !ok {
			return nil, false, err
		}
		if line == "" {
			break
		}
		values, err := parseNumbersLine(line, c.lastResult, c.hasLast)
		if err != nil {
			return nil, true, err
		}
		numbers = append(numbers, values...)
	}

	return validateOperandCount(op, numbers)
}

func parseNumbersLine(line string, last float64, hasLast bool) ([]float64, error) {
	fields := strings.FieldsFunc(line, func(r rune) bool {
		return r == ' ' || r == ',' || r == '\t'
	})
	if len(fields) == 0 {
		return nil, errors.New("empty input")
	}

	values := make([]float64, 0, len(fields))
	for _, field := range fields {
		token := strings.ToLower(field)
		if token == "ans" || token == "last" {
			if !hasLast {
				return nil, errors.New("no previous result")
			}
			values = append(values, last)
			continue
		}
		if token == "pi" {
			values = append(values, math.Pi)
			continue
		}
		if token == "e" {
			values = append(values, math.E)
			continue
		}

		val, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return nil, errors.New("invalid number")
		}
		values = append(values, val)
	}
	return values, nil
}

func validateOperandCount(op string, nums []float64) ([]float64, bool, error) {
	if isUnaryOperator(op) {
		if len(nums) != 1 {
			return nil, true, errors.New("this operator requires exactly one operand")
		}
		return nums, true, nil
	}

	if len(nums) < 2 {
		return nil, true, errors.New("at least two operands required")
	}

	if (op == "-" || op == "/" || op == "%" || op == "^") && len(nums) != 2 {
		return nil, true, errors.New("this operator supports exactly two operands")
	}

	return nums, true, nil
}

/* =======================
   Calculation
======================= */

func calculateMany(op string, nums []float64, degrees bool) (float64, error) {
	if isUnaryOperator(op) {
		return calculateUnary(op, nums[0], degrees)
	}

	result := nums[0]
	for i := 1; i < len(nums); i++ {
		var err error
		result, err = calculateBinary(op, result, nums[i])
		if err != nil {
			return 0, err
		}
	}
	return result, nil
}

func calculateBinary(op string, a, b float64) (float64, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	case "^":
		return math.Pow(a, b), nil
	case "%":
		if b == 0 {
			return 0, errors.New("modulo by zero")
		}
		return math.Mod(a, b), nil
	default:
		return 0, errors.New("unknown operator")
	}
}

func calculateUnary(op string, v float64, degrees bool) (float64, error) {
	rad := toRadians(v, degrees)

	switch op {
	case "sin":
		return math.Sin(rad), nil
	case "cos":
		return math.Cos(rad), nil
	case "tan":
		if math.Abs(math.Cos(rad)) < 1e-12 {
			return 0, errors.New("tan undefined for this value")
		}
		return math.Tan(rad), nil
	case "sqrt":
		if v < 0 {
			return 0, errors.New("sqrt of negative number")
		}
		return math.Sqrt(v), nil
	case "log":
		if v <= 0 {
			return 0, errors.New("log domain error")
		}
		return math.Log(v), nil
	default:
		return 0, errors.New("unknown unary operator")
	}
}

func isUnaryOperator(op string) bool {
	return op == "sin" || op == "cos" || op == "tan" || op == "sqrt" || op == "log"
}

func toRadians(v float64, degrees bool) float64 {
	if degrees {
		return v * math.Pi / 180
	}
	return v
}

/* =======================
   Utilities
======================= */

func formatExpression(nums []float64, op string) string {
	if isUnaryOperator(op) {
		return fmt.Sprintf("%s(%g)", op, nums[0])
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%g", nums[0]))
	for i := 1; i < len(nums); i++ {
		b.WriteString(fmt.Sprintf(" %s %g", op, nums[i]))
	}
	return b.String()
}

func (c *Calculator) formatResult(value float64) string {
	if c.precision < 0 {
		return fmt.Sprintf("%g", value)
	}
	return fmt.Sprintf("%.*f", c.precision, value)
}

func readLine(reader *bufio.Reader) (string, bool, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			return strings.TrimSpace(line), false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(line), true, nil
}

/* =======================
   Help & History
======================= */

func printHelp(deg bool) {
	fmt.Println("\nCommands:")
	fmt.Println("  help     Show help")
	fmt.Println("  history  Show calculation history")
	fmt.Println("  degrees  Use degrees for trig")
	fmt.Println("  radians  Use radians for trig")
	fmt.Println("  mode     Show trig mode and last result")
	fmt.Println("  precision Set decimal places (auto or 0-10)")
	fmt.Println("  stats    Show count, min, max, average")
	fmt.Println("  ops      List available operators")
	fmt.Println("  export   Save history to file")
	fmt.Println("  clear    Clear memory")
	fmt.Println("  exit     Quit")

	if deg {
		fmt.Println("Trig mode: degrees")
	} else {
		fmt.Println("Trig mode: radians")
	}
	fmt.Println("Input: enter multiple numbers per line, separated by spaces or commas.")
	fmt.Println("Constants: pi, e.")
}

func printOperators() {
	fmt.Println("\nOperators:")
	fmt.Println("  Binary: +  -  *  /  ^  %")
	fmt.Println("  Unary:  sin  cos  tan  sqrt  log")
	fmt.Println("Aliases: pow -> ^, mod -> %, ln -> log")
}

func (c *Calculator) setPrecision() error {
	fmt.Print("Precision (auto or 0-10): ")
	line, ok, err := readLine(c.reader)
	if err != nil || !ok {
		return err
	}
	if line == "" || strings.EqualFold(line, "auto") {
		c.precision = -1
		fmt.Println("Precision set to auto.")
		return nil
	}
	value, err := strconv.Atoi(line)
	if err != nil {
		return errors.New("invalid precision")
	}
	if value < 0 || value > 10 {
		return errors.New("precision must be between 0 and 10")
	}
	c.precision = value
	fmt.Printf("Precision set to %d.\n", value)
	return nil
}

func (c *Calculator) printStats() {
	if len(c.results) == 0 {
		fmt.Println("No results yet.")
		return
	}

	min := c.results[0]
	max := c.results[0]
	sum := 0.0
	for _, v := range c.results {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	avg := sum / float64(len(c.results))

	fmt.Printf("Count: %d\n", len(c.results))
	fmt.Printf("Min: %s\n", c.formatResult(min))
	fmt.Printf("Max: %s\n", c.formatResult(max))
	fmt.Printf("Avg: %s\n", c.formatResult(avg))
}

func printHistory(history []string) {
	if len(history) == 0 {
		fmt.Println("No history yet.")
		return
	}
	fmt.Println("\nHistory:")
	for i, h := range history {
		fmt.Printf(" %2d) %s\n", i+1, h)
	}
}

func exportHistory(reader *bufio.Reader, history []string) error {
	if len(history) == 0 {
		return errors.New("no history to export")
	}

	fmt.Print("Export file (default history.txt): ")
	line, _, _ := readLine(reader)

	if line == "" {
		line = "history.txt"
	}

	return os.WriteFile(line, []byte(strings.Join(history, "\n")), 0644)
}
