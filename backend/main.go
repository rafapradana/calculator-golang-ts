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

// Tokenizer

type TokenType int

const (
	NUMBER TokenType = iota
	PLUS
	MINUS
	MULTIPLY
	DIVIDE
	LPAREN
	RPAREN
)

type Token struct {
	Type  TokenType
	Value float64
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
			return 0, fmt.Errorf("invalid expression")
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
				return nil, fmt.Errorf("invalid number")
			}
			tokens = append(tokens, Token{Type: NUMBER, Value: num})
			continue
		}

		switch ch {
		case '+':
			tokens = append(tokens, Token{Type: PLUS})
		case '-':
			if len(tokens) == 0 || tokens[len(tokens)-1].Type != NUMBER {
				i++
				num, err := readNumber("-")
				if err != nil {
					return nil, fmt.Errorf("invalid number")
				}
				tokens = append(tokens, Token{Type: NUMBER, Value: num})
				continue
			}
			tokens = append(tokens, Token{Type: MINUS})
		case '*':
			tokens = append(tokens, Token{Type: MULTIPLY})
		case '/':
			tokens = append(tokens, Token{Type: DIVIDE})
		case '(':
			tokens = append(tokens, Token{Type: LPAREN})
		case ')':
			tokens = append(tokens, Token{Type: RPAREN})
		default:
			return nil, fmt.Errorf("invalid character")
		}
		i++
	}

	return tokens, nil
}

// Parser with operator precedence (precedence climbing)

type Parser struct {
	tokens []Token
	pos    int
}

func precedence(op TokenType) int {
	switch op {
	case PLUS, MINUS:
		return 1
	case MULTIPLY, DIVIDE:
		return 2
	}
	return 0
}

func evaluate(expr string) (float64, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	if len(tokens) == 0 {
		return 0, fmt.Errorf("empty expression")
	}

	p := &Parser{tokens: tokens}
	result := p.parseBinary(0)

	if p.pos < len(tokens) {
		return 0, fmt.Errorf("invalid expression")
	}

	return result, nil
}

func (p *Parser) parseBinary(minPrec int) float64 {
	left := p.parseFactor()

	for p.pos < len(p.tokens) {
		op := p.tokens[p.pos].Type
		prec := precedence(op)
		if prec < minPrec || prec == 0 {
			break
		}
		p.pos++
		right := p.parseBinary(prec + 1)
		switch op {
		case PLUS:
			left += right
		case MINUS:
			left -= right
		case MULTIPLY:
			left *= right
		case DIVIDE:
			if right != 0 {
				left /= right
			}
		}
	}

	return left
}

func (p *Parser) parseFactor() float64 {
	if p.pos >= len(p.tokens) {
		return 0
	}

	token := p.tokens[p.pos]

	if token.Type == NUMBER {
		p.pos++
		return token.Value
	}

	if token.Type == LPAREN {
		p.pos++
		result := p.parseBinary(0)
		if p.pos < len(p.tokens) && p.tokens[p.pos].Type == RPAREN {
			p.pos++
		}
		return result
	}

	return 0
}
