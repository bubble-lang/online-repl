package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall/js"
)

// Token types
type TokenType int

const (
	TOKEN_NUMBER TokenType = iota
	TOKEN_STRING
	TOKEN_PLUS
	TOKEN_MINUS
	TOKEN_TIMES
	TOKEN_DIVIDED_BY
	TOKEN_ASSIGN
	TOKEN_REMEMBER
	TOKEN_AS
	TOKEN_IDENTIFIER
	TOKEN_SAY
	TOKEN_EOF
	TOKEN_HELP
)

// Token structure
type Token struct {
	Type  TokenType
	Value string
}

// Lexer
func lex(input string) []Token {
    tokens := []Token{}
    inQuotes := false
    currentToken := ""

    for i := 0; i < len(input); i++ {
        char := input[i]

        if char == '"' {
            if inQuotes {
                tokens = append(tokens, Token{Type: TOKEN_STRING, Value: currentToken})
                currentToken = ""
            }
            inQuotes = !inQuotes
        } else if inQuotes {
            currentToken += string(char)
        } else if char == ' ' {
            if currentToken != "" {
                tokens = append(tokens, tokenFromString(currentToken))
                currentToken = ""
            }
        } else if char == '=' {
            if currentToken != "" {
                tokens = append(tokens, tokenFromString(currentToken))
                currentToken = ""
            }
            tokens = append(tokens, Token{Type: TOKEN_ASSIGN, Value: "="})
        } else if char == '+' || char == '-' {  // Corrected this line
            if currentToken != "" {
                tokens = append(tokens, tokenFromString(currentToken))
                currentToken = ""
            }
            tokens = append(tokens, tokenFromString(string(char)))
        } else {
            currentToken += string(char)
        }
    }

    if currentToken != "" {
        tokens = append(tokens, tokenFromString(currentToken))
    }

    tokens = append(tokens, Token{Type: TOKEN_EOF, Value: ""})
    return tokens
}

func tokenFromString(s string) Token {
	switch s {
	case "+":
		return Token{Type: TOKEN_PLUS, Value: "plus"}
	case "-":
		return Token{Type: TOKEN_MINUS, Value: "minus"}
	case "say":
		return Token{Type: TOKEN_SAY, Value: s}
	case "plus":
		return Token{Type: TOKEN_PLUS, Value: s}
	case "minus":
		return Token{Type: TOKEN_MINUS, Value: s}
	case "times":
		return Token{Type: TOKEN_TIMES, Value: s}
	case "*":
		return Token{Type: TOKEN_TIMES, Value: s}
	case "over":
		return Token{Type: TOKEN_DIVIDED_BY, Value: s}
	case "/":
		return Token{Type: TOKEN_DIVIDED_BY, Value: s}
	case "by":
		return Token{Type: TOKEN_IDENTIFIER, Value: s}
	case "remember":
		return Token{Type: TOKEN_REMEMBER, Value: s}
	case "as":
		return Token{Type: TOKEN_AS, Value: s}
	default:
		if _, err := strconv.ParseFloat(s, 64); err == nil {
			return Token{Type: TOKEN_NUMBER, Value: s}
		}
		return Token{Type: TOKEN_IDENTIFIER, Value: s}
	}
}

// AST Node interface
type Node interface {
	TokenLiteral() string
}

// Expression interface
type Expression interface {
	Node
	expressionNode()
}

// Statement interface
type Statement interface {
	Node
	statementNode()
}

// NumberLiteral structure
type NumberLiteral struct {
	Token Token
	Value float64
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Value }

// StringLiteral structure
type StringLiteral struct {
	Token Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Value }

// Identifier structure
type Identifier struct {
	Token Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Value }

// InfixExpression structure
type InfixExpression struct {
	Token    Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Value }

// AssignmentStatement structure
type AssignmentStatement struct {
	Token Token
	Name  string
	Value Expression
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Value }

// SayStatement structure
type SayStatement struct {
	Token  Token
	Values []Expression
}

func (ss *SayStatement) statementNode()       {}
func (ss *SayStatement) TokenLiteral() string { return ss.Token.Value }

// Parser
func parse(tokens []Token) []Statement {
	statements := []Statement{}
	i := 0

	for i < len(tokens)-1 {
		if tokens[i].Type == TOKEN_IDENTIFIER && i+1 < len(tokens) && tokens[i+1].Type == TOKEN_ASSIGN {
			stmt := parseAssignStatement(tokens, &i)
			statements = append(statements, stmt)
		} else if tokens[i].Type == TOKEN_REMEMBER {
			stmt := parseRememberStatement(tokens, &i)
			statements = append(statements, stmt)
		} else if tokens[i].Type == TOKEN_SAY {
			stmt := parseSayStatement(tokens, &i)
			statements = append(statements, stmt)
		} else {
			expr := parseExpression(tokens, &i)
			statements = append(statements, &AssignmentStatement{Name: "", Value: expr})
		}
	}

	return statements
}

func parseAssignStatement(tokens []Token, i *int) *AssignmentStatement {
	stmt := &AssignmentStatement{Token: tokens[*i]}
	stmt.Name = tokens[*i].Value
	*i += 2 // skip identifier and '='
	stmt.Value = parseExpression(tokens, i)
	return stmt
}

func parseRememberStatement(tokens []Token, i *int) *AssignmentStatement {
	stmt := &AssignmentStatement{Token: tokens[*i]}
	*i++ // skip "remember"

	if *i < len(tokens) && tokens[*i].Type == TOKEN_IDENTIFIER {
		stmt.Name = tokens[*i].Value
		*i++

		if *i < len(tokens) && tokens[*i].Type == TOKEN_AS {
			*i++ // skip "as"
			stmt.Value = parseExpression(tokens, i)
		}
	}

	return stmt
}

func parseSayStatement(tokens []Token, i *int) *SayStatement {
	stmt := &SayStatement{Token: tokens[*i], Values: []Expression{}}
	*i++ // skip "say"

	for *i < len(tokens)-1 {
		expr := parseExpression(tokens, i)
		stmt.Values = append(stmt.Values, expr)
	}

	return stmt
}

func parseExpression(tokens []Token, i *int) Expression {
    expr := parseTerm(tokens, i)

    for *i < len(tokens)-1 && (tokens[*i].Type == TOKEN_PLUS || tokens[*i].Type == TOKEN_MINUS || 
        tokens[*i].Value == "plus" || tokens[*i].Value == "minus") {
        operator := tokens[*i].Value
        *i++
        right := parseTerm(tokens, i)
        expr = &InfixExpression{Token: tokens[*i-1], Left: expr, Operator: operator, Right: right}
    }

    return expr
}

func parseTerm(tokens []Token, i *int) Expression {
	expr := parseFactor(tokens, i)

	for *i < len(tokens)-1 && (tokens[*i].Type == TOKEN_TIMES || tokens[*i].Type == TOKEN_DIVIDED_BY) {
		operator := tokens[*i].Value
		*i++
		right := parseFactor(tokens, i)
		expr = &InfixExpression{Token: tokens[*i-1], Left: expr, Operator: operator, Right: right}
	}

	return expr
}

func parseFactor(tokens []Token, i *int) Expression {
	if tokens[*i].Type == TOKEN_NUMBER {
		value, _ := strconv.ParseFloat(tokens[*i].Value, 64)
		expr := &NumberLiteral{Token: tokens[*i], Value: value}
		*i++
		return expr
	} else if tokens[*i].Type == TOKEN_STRING {
		expr := &StringLiteral{Token: tokens[*i], Value: tokens[*i].Value}
		*i++
		return expr
	} else if tokens[*i].Type == TOKEN_IDENTIFIER {
		expr := &Identifier{Token: tokens[*i], Value: tokens[*i].Value}
		*i++
		return expr
	}
	// Handle other cases or errors
	return nil
}

// Interpreter
var memory = make(map[string]interface{})

func interpret(statements []Statement) interface{} {
	var result interface{}
	for _, statement := range statements {
		switch stmt := statement.(type) {
		case *AssignmentStatement:
			value := evalExpression(stmt.Value)
			if stmt.Name != "" {
				memory[stmt.Name] = value
			}
			result = value
		case *SayStatement:
			var output strings.Builder
			for _, expr := range stmt.Values {
				value := evalExpression(expr)
				output.WriteString(fmt.Sprintf("%v", value))
			}
			result = output.String()
		default:
			result = evalExpression(statement.(*AssignmentStatement).Value)
		}
	}
	return result
}

func evalExpression(exp Expression) interface{} {
	switch e := exp.(type) {
	case *NumberLiteral:
		return e.Value
	case *StringLiteral:
		return e.Value
	case *Identifier:
		if value, ok := memory[e.Value]; ok {
			return value
		}
		return fmt.Sprintf("<%s>", e.Value) // Return unspecified identifiers in angle brackets
	case *InfixExpression:
		left := evalExpression(e.Left)
		right := evalExpression(e.Right)

		// Handle arithmetic operations
		if l, ok := left.(float64); ok {
			if r, ok := right.(float64); ok {
				switch e.Operator {
				case "plus", "+":
					return l + r
				case "minus", "-":
					return l - r
				case "times", "*":
					return l * r
				case "over":
					if r != 0 {
						return l / r
					}
					return "Error: Division by zero"
				}
			}
		}

		// Handle string concatenation
		return fmt.Sprintf("%v%v", toString(left), toString(right))
	}
	return nil
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// REPL
func startREPL() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		if line == "exit" {
			return
		}

		if line == "help" {
			fmt.Println()
			fmt.Println("  Bubble commands:")
			fmt.Println("    exit - Exit the Bubble REPL")
			fmt.Println("    help - Display this help message")
			fmt.Println("    <expression> - Evaluate the expression")
			fmt.Println("    remember <variable> as <expression> - Assign a value to a variable")
			fmt.Println("    say <expression> - Print the value of an expression")
			fmt.Println()
			continue
		}



		tokens := lex(line)
		statements := parse(tokens)
		result := interpret(statements)
		fmt.Println(" ", result)
		fmt.Println(" ")
	}
}

func main() {
	fmt.Println()
	fmt.Println()
	fmt.Println(`  BUBBLE

  Welcome to the Bubble programming language!

  Type help for a list of commands. To exit, type exit.
  
  `)
	fmt.Println()
	

	startREPL()

	obj := js.Global().Get("Object").New()

    // Set the "evaluateBubble" property of the object to the evaluateBubble function
    obj.Set("evaluateBubble", js.FuncOf(evaluateBubble))

    // Set the object as a property of the global object
    js.Global().Set("bubbleREPL", obj)

    // Keep the program running
    <-make(chan struct{})
}

func evaluateBubble(this js.Value, args []js.Value) interface{} {
    input := args[0].String()
    tokens := lex(input)
    statements := parse(tokens)
    result := interpret(statements)
    return result
}