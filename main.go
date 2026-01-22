package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Calculator")
	fmt.Println("Enter one of +, -, *, / or type 'exit' to quit.")

	for {
		op, ok := readOperator(reader)
		if !ok {
			fmt.Println("Goodbye.")
			return
		}

		a, err := readNumber(reader, "First number: ")
		if err != nil {
			fmt.Println("Input error:", err)
			continue
		}

		b, err := readNumber(reader, "Second number: ")
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
	}
}

func readOperator(reader *bufio.Reader) (string, bool) {
	for {
		fmt.Print("Operator (+, -, *, /) or exit: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", false
		}
		op := strings.TrimSpace(line)
		if strings.EqualFold(op, "exit") {
			return "", false
		}
		if op == "+" || op == "-" || op == "*" || op == "/" {
			return op, true
		}
		fmt.Println("Please enter a valid operator or 'exit'.")
	}
}

func readNumber(reader *bufio.Reader, prompt string) (float64, error) {
	fmt.Print(prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return 0, errors.New("empty input")
	}
	value, err := strconv.ParseFloat(line, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
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
	default:
		return 0, errors.New("unknown operator")
	}
}
