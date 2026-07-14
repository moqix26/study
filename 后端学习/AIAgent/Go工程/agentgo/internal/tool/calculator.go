package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode"

	"study.local/agentgo/internal/llm"
)

func CalculatorTool() RegisteredTool {
	return RegisteredTool{
		Definition: llm.ToolDefinition{
			Type:        "function",
			Name:        "calculator",
			Description: "计算只包含加减乘除和括号的算术表达式。不要用于执行代码或访问外部系统。",
			Strict:      true,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"expression": map[string]any{
						"type":        "string",
						"description": "例如 (12.5+7.5)/2，最长 128 字符",
					},
				},
				"required":             []string{"expression"},
				"additionalProperties": false,
			},
		},
		Handler: func(ctx context.Context, _ Principal, raw json.RawMessage) (any, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			var arguments struct {
				Expression string `json:"expression"`
			}
			if err := decodeStrict(raw, &arguments); err != nil {
				return nil, fmt.Errorf("calculator 参数错误: %w", err)
			}
			if len(arguments.Expression) == 0 || len(arguments.Expression) > 128 {
				return nil, errors.New("expression 长度必须在 1 到 128 之间")
			}
			value, err := evaluate(arguments.Expression)
			if err != nil {
				return nil, err
			}
			return map[string]any{"expression": arguments.Expression, "value": value}, nil
		},
	}
}

func decodeStrict(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("存在多余 JSON 内容")
		}
		return err
	}
	return nil
}

type expressionParser struct {
	input []rune
	pos   int
}

func evaluate(expression string) (float64, error) {
	parser := &expressionParser{input: []rune(expression)}
	value, err := parser.parseExpression()
	if err != nil {
		return 0, err
	}
	parser.skipSpaces()
	if parser.pos != len(parser.input) {
		return 0, fmt.Errorf("位置 %d 附近存在非法字符", parser.pos)
	}
	if math.IsInf(value, 0) || math.IsNaN(value) {
		return 0, errors.New("计算结果不是有限数")
	}
	return value, nil
}

func (p *expressionParser) parseExpression() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpaces()
		if !p.consume('+') && !p.consume('-') {
			return left, nil
		}
		operator := p.input[p.pos-1]
		right, err := p.parseTerm()
		if err != nil {
			return 0, err
		}
		if operator == '+' {
			left += right
		} else {
			left -= right
		}
	}
}

func (p *expressionParser) parseTerm() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpaces()
		if !p.consume('*') && !p.consume('/') {
			return left, nil
		}
		operator := p.input[p.pos-1]
		right, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		if operator == '*' {
			left *= right
		} else {
			if right == 0 {
				return 0, errors.New("不能除以 0")
			}
			left /= right
		}
	}
}

func (p *expressionParser) parseFactor() (float64, error) {
	p.skipSpaces()
	if p.consume('+') {
		return p.parseFactor()
	}
	if p.consume('-') {
		value, err := p.parseFactor()
		return -value, err
	}
	if p.consume('(') {
		value, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		p.skipSpaces()
		if !p.consume(')') {
			return 0, errors.New("缺少右括号")
		}
		return value, nil
	}
	return p.parseNumber()
}

func (p *expressionParser) parseNumber() (float64, error) {
	p.skipSpaces()
	start := p.pos
	dotSeen := false
	for p.pos < len(p.input) {
		r := p.input[p.pos]
		if unicode.IsDigit(r) {
			p.pos++
			continue
		}
		if r == '.' && !dotSeen {
			dotSeen = true
			p.pos++
			continue
		}
		break
	}
	if start == p.pos {
		return 0, fmt.Errorf("位置 %d 需要数字", p.pos)
	}
	return strconv.ParseFloat(string(p.input[start:p.pos]), 64)
}

func (p *expressionParser) skipSpaces() {
	for p.pos < len(p.input) && unicode.IsSpace(p.input[p.pos]) {
		p.pos++
	}
}

func (p *expressionParser) consume(expected rune) bool {
	if p.pos < len(p.input) && p.input[p.pos] == expected {
		p.pos++
		return true
	}
	return false
}
