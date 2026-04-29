package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result float64 `json:"result"`
	Error  string  `json:"error,omitempty"`
}

func main() {
	fs := http.FileServer(http.Dir("./dist"))
	http.Handle("/", fs)
	http.HandleFunc("/api/calculate", handleCalculate)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := evaluate(req.Expression)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Result: result})
}

func sendError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{Error: msg})
}

// Token represents either a number or an operator.
// Op is '+', '-', '*', '/' for operators; 0 for numbers.
type Token struct {
	IsNumber bool
	Op       byte
	Value    float64
}

func tokenize(expr string) ([]Token, error) {
	var tokens []Token
	i := 0

	readNumber := func(prefix string) (float64, error) {
		start := i
		for i < len(expr) && ((expr[i] >= '0' && expr[i] <= '9') || expr[i] == '.') {
			i++
		}
		if start == i {
			return 0, fmt.Errorf("expected number")
		}
		return strconv.ParseFloat(prefix+expr[start:i], 64)
	}

	for i < len(expr) {
		ch := expr[i]

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}

		if (ch >= '0' && ch <= '9') || ch == '.' {
			num, err := readNumber("")
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, Token{IsNumber: true, Value: num})
			continue
		}

		switch ch {
		case '+':
			tokens = append(tokens, Token{Op: '+'})
		case '-':
			if len(tokens) == 0 || !tokens[len(tokens)-1].IsNumber {
				i++
				num, err := readNumber("-")
				if err != nil {
					return nil, err
				}
				tokens = append(tokens, Token{IsNumber: true, Value: num})
				continue
			}
			tokens = append(tokens, Token{Op: '-'})
		case '*':
			tokens = append(tokens, Token{Op: '*'})
		case '/':
			tokens = append(tokens, Token{Op: '/'})
		default:
			return nil, fmt.Errorf("invalid character")
		}
		i++
	}

	// validate: expression must start and end with a number
	if len(tokens) > 0 && !tokens[0].IsNumber {
		return nil, fmt.Errorf("invalid expression")
	}
	if len(tokens) > 0 && !tokens[len(tokens)-1].IsNumber {
		return nil, fmt.Errorf("invalid expression")
	}

	return tokens, nil
}

// Flat 2-pass evaluator.
func evaluate(expr string) (float64, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	if len(tokens) == 0 {
		return 0, fmt.Errorf("empty expression")
	}

	// Pass 1: handle * and /
	var pass1 []Token
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Op == '*' || tokens[i].Op == '/' {
			left := pass1[len(pass1)-1].Value
			right := tokens[i+1].Value
			var result float64
			if tokens[i].Op == '*' {
				result = left * right
			} else {
				if right == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				result = left / right
			}
			pass1[len(pass1)-1] = Token{IsNumber: true, Value: result}
			i++ // skip the number we just consumed
		} else {
			pass1 = append(pass1, tokens[i])
		}
	}

	// Pass 2: handle + and -
	result := pass1[0].Value
	for i := 1; i < len(pass1); i += 2 {
		if pass1[i].Op == '+' {
			result += pass1[i+1].Value
		} else {
			result -= pass1[i+1].Value
		}
	}

	return result, nil
}
