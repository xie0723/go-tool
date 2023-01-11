package stringmatch

import (
	"errors"
	"strings"
	"text/scanner"
)

const (
	AND = "AND"
	OR  = "OR"
)

func Calculate(str string, stackSize int, boolFunc func(string) bool) (r bool, b error) {
	defer func() {
		if err := recover(); err != nil {
			b = errors.New(err.(string))
			return
		}
	}()
	var boolStack = newStack(true, stackSize)
	var symbolStack = newStack("", stackSize)

	var s scanner.Scanner
	s.Mode = scanner.ScanIdents
	s.Init(strings.NewReader(str))
	tok := s.Scan()
	tt := s.TokenText()
	for tok != scanner.EOF {
		switch tok {
		case scanner.Ident:
			if tt == AND || tt == OR {
				if symbolStack.IsEmpty() {
					symbolStack.Push(tt)
				} else {
					symbol := symbolStack.Peek()
					if tt == ("(") {
						symbolStack.Push(tt)
					}
					if getPriority(tt) < getPriority(symbol) {
						symbolStack.Push(tt)
					} else {
						symbol := symbolStack.Pop()
						bool1 := boolStack.Pop()
						bool2 := boolStack.Pop()
						ret := getSum(bool1, bool2, symbol)
						boolStack.Push(ret)
						symbolStack.Push(tt)
					}
				}
			} else {
				boolStack.Push(boolFunc(tt))
			}
		case '(':
			symbolStack.Push("(")
		case ')':
			for {
				symbol := symbolStack.Pop()
				if symbol != "(" {
					bool1 := boolStack.Pop()
					bool2 := boolStack.Pop()
					ret := getSum(bool1, bool2, symbol)
					boolStack.Push(ret)
				} else {
					break
				}
			}
		default:
			boolStack.Push(boolFunc(tt))
		}
		tok, tt = s.Scan(), s.TokenText()
	}
	for symbolStack.Len() != 0 {
		if boolStack.Len() < 2 {
			return false, errors.New("err7 表达式错误")
		}
		symbol := symbolStack.Pop()
		data1 := boolStack.Pop()
		data2 := boolStack.Pop()
		ret := getSum(data1, data2, symbol)
		boolStack.Push(ret)
	}
	if boolStack.Len() != 1 {
		return false, errors.New("err10 表达式有误")
	}
	ret := boolStack.Peek()
	return ret, nil
}

func getSum(cond1 bool, cond2 bool, symbol string) bool {
	switch symbol {
	case AND:
		return cond1 && cond2
	case OR:
		return cond1 || cond2
	}
	return false
}

func getPriority(symbol string) int {
	switch symbol {
	case AND:
		return 3
	case OR:
		return 2
	case "(":
		return 8
	case ")":
		return 1
	default:
		return 0
	}
}
