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
	fmt.Println("Enter one of +, -, *, /, ^, % or type 'help' for commands.")

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

		a, ok, err := readNumber(reader, "First number: ", lastResult, hasLast)
		if !ok {
			fmt.Println("Goodbye.")
			return
		}
		if err != nil {
			fmt.Println("Input error:", err)
			continue
		}

		b, ok, err := readNumber(reader, "Second number: ", lastResult, hasLast)
		if !ok {
			fmt.Println("Goodbye.")
			return
		}
		if err != nil {
			fmt.Println("Input error:", err)
			continue
		}

		result, err := calculate(op, a, b)
		if err != nil {
			fmt.Println("Calculation error:", err)
			continue
		}

		fmt.Printf("Result: %g\n", result)
		lastResult = result
		hasLast = true
		history = append(history, fmt.Sprintf("%g %s %g = %g", a, op, b, result))
	}
}

func readOperator(reader *bufio.Reader) (string, string, bool) {
	for {
		fmt.Print("Operator (+, -, *, /, ^, %) or command (help, history, clear, exit): ")
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
		case "+", "-", "*", "/", "^", "%":
			return "op", line, true
		}
		if strings.EqualFold(line, "pow") {
			return "op", "^", true
		}
		if strings.EqualFold(line, "mod") {
			return "op", "%", true
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
