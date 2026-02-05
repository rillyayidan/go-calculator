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

func main() {
	reader := bufio.NewReader(os.Stdin)
	var history []string
	var lastResult float64
	var hasLast bool

	fmt.Println("Calculator")
	fmt.Println("Enter one of +, -, *, /, ^, %, sin, cos, tan, sqrt, log or type 'help' for commands.")

	for {
		action, op, ok := readOperator(reader)
		if !ok {
			fmt.Println("Goodbye.")
			return
		}
		switch action {
		case "help":
			printHelp()
			continue
		case "history":
			printHistory(history)
			continue
		case "clear":
			history = nil
			hasLast = false
			fmt.Println("Memory cleared.")
			continue
		}

		numbers, ok, err := readNumbers(reader, op, lastResult, hasLast)
		if !ok {
			fmt.Println("Goodbye.")
			return
		}
		if err != nil {
			fmt.Println("Input error:", err)
			continue
		}

		result, err := calculateMany(op, numbers)
		if err != nil {
			fmt.Println("Calculation error:", err)
			continue
		}

		fmt.Printf("Result: %g\n", result)
		lastResult = result
		hasLast = true
		history = append(history, fmt.Sprintf("%s = %g", formatExpression(numbers, op), result))
	}
}

func readOperator(reader *bufio.Reader) (string, string, bool) {
	for {
		fmt.Print("Operator (+, -, *, /, ^, %, sin, cos, tan, sqrt, log) or command (help, history, clear, exit): ")
		line, ok, err := readLine(reader)
		if err != nil {
			return "", "", false
		}
		if !ok {
			return "", "", false
		}
		if strings.EqualFold(line, "exit") {
			return "exit", "", false
		}
		switch strings.ToLower(line) {
		case "help", "history", "clear":
			return line, "", true
		}
		switch line {
		case "+", "-", "*", "/", "^", "%", "sin", "cos", "tan", "sqrt", "log":
			return "op", line, true
		}
		if strings.EqualFold(line, "pow") {
			return "op", "^", true
		}
		if strings.EqualFold(line, "mod") {
			return "op", "%", true
		}
		if strings.EqualFold(line, "ln") {
			return "op", "log", true
		}
		fmt.Println("Please enter a valid operator or command.")
	}
}

func readNumber(reader *bufio.Reader, prompt string, lastResult float64, hasLast bool) (float64, bool, error) {
	fmt.Print(prompt)
	line, ok, err := readLine(reader)
	if err != nil {
		return 0, false, err
	}
	if !ok {
		return 0, false, nil
	}
	if line == "" {
		return 0, true, errors.New("empty input")
	}
	if strings.EqualFold(line, "ans") || strings.EqualFold(line, "last") {
		if !hasLast {
			return 0, true, errors.New("no previous result")
		}
		return lastResult, true, nil
	}
	value, err := strconv.ParseFloat(line, 64)
	if err != nil {
		return 0, true, err
	}
	return value, true, nil
}

func readNumbers(reader *bufio.Reader, op string, lastResult float64, hasLast bool) ([]float64, bool, error) {
	fmt.Println("Enter numbers one per line. Press Enter on a blank line to finish.")
	if hasLast {
		fmt.Println("Tip: type ans or last to reuse the previous result.")
	}

	var numbers []float64
	for {
		value, ok, err := readNumber(reader, fmt.Sprintf("Number %d: ", len(numbers)+1), lastResult, hasLast)
		if !ok {
			return nil, false, nil
		}
		if err != nil {
			if strings.Contains(err.Error(), "empty input") {
				break
			}
			return nil, true, err
		}
		numbers = append(numbers, value)
	}

	if len(numbers) < 2 {
		if isUnaryOperator(op) && len(numbers) == 1 {
			return numbers, true, nil
		}
		return nil, true, errors.New("enter at least two numbers")
	}

	if (op == "-" || op == "/" || op == "%" || op == "^") && len(numbers) > 2 {
		return nil, true, errors.New("this operator supports exactly two numbers")
	}
	if isUnaryOperator(op) && len(numbers) != 1 {
		return nil, true, errors.New("this operator supports exactly one number")
	}

	return numbers, true, nil
}

func calculate(op string, a, b float64) (float64, error) {
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
			return 0, errors.New("division by zero")
		}
		return math.Mod(a, b), nil
	default:
		return 0, errors.New("unknown operator")
	}
}

func calculateMany(op string, numbers []float64) (float64, error) {
	if len(numbers) < 1 {
		return 0, errors.New("need at least one number")
	}
	if isUnaryOperator(op) {
		if len(numbers) != 1 {
			return 0, errors.New("this operator supports exactly one number")
		}
		return calculateUnary(op, numbers[0])
	}
	if len(numbers) < 2 {
		return 0, errors.New("need at least two numbers")
	}
	result := numbers[0]
	for i := 1; i < len(numbers); i++ {
		value := numbers[i]
		next, err := calculate(op, result, value)
		if err != nil {
			return 0, err
		}
		result = next
	}
	return result, nil
}

func calculateUnary(op string, value float64) (float64, error) {
	switch op {
	case "sin":
		return math.Sin(value), nil
	case "cos":
		return math.Cos(value), nil
	case "tan":
		return math.Tan(value), nil
	case "sqrt":
		if value < 0 {
			return 0, errors.New("square root of negative number")
		}
		return math.Sqrt(value), nil
	case "log":
		if value <= 0 {
			return 0, errors.New("logarithm domain error")
		}
		return math.Log(value), nil
	default:
		return 0, errors.New("unknown operator")
	}
}

func isUnaryOperator(op string) bool {
	switch op {
	case "sin", "cos", "tan", "sqrt", "log":
		return true
	default:
		return false
	}
}

func formatExpression(numbers []float64, op string) string {
	if len(numbers) == 0 {
		return ""
	}
	if isUnaryOperator(op) && len(numbers) == 1 {
		return fmt.Sprintf("%s(%g)", op, numbers[0])
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%g", numbers[0]))
	for i := 1; i < len(numbers); i++ {
		b.WriteString(fmt.Sprintf(" %s %g", op, numbers[i]))
	}
	return b.String()
}

func readLine(reader *bufio.Reader) (string, bool, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) {
			if line == "" {
				return "", false, nil
			}
			return strings.TrimSpace(line), true, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(line), true, nil
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  help    Show this help")
	fmt.Println("  history Show previous calculations")
	fmt.Println("  clear   Clear history and last result")
	fmt.Println("  exit    Quit the calculator")
	fmt.Println("Notes:")
	fmt.Println("  Use ans or last as a number to reuse the previous result.")
	fmt.Println("  Enter multiple numbers (one per line). Blank line finishes input.")
	fmt.Println("  Trig functions use radians.")
	fmt.Println("  Unary functions: sin, cos, tan, sqrt, log.")
}

func printHistory(history []string) {
	if len(history) == 0 {
		fmt.Println("No history yet.")
		return
	}
	fmt.Println("History:")
	for i, entry := range history {
		fmt.Printf("  %d) %s\n", i+1, entry)
	}
}
