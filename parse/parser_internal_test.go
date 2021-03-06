/*
 * Copyright 2018-2019 De-labtory
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package parse

import (
	"errors"
	"fmt"
	"testing"

	"github.com/DE-labtory/koa/symbol"

	"github.com/DE-labtory/koa/ast"
)

type mockTokenBuffer struct {
	buf []Token
	sp  int
}

func (m *mockTokenBuffer) Read() Token {
	ret := m.buf[m.sp]
	if m.sp+1 < len(m.buf) {
		m.sp++
	}
	return ret
}

func (m *mockTokenBuffer) Peek(n peekNumber) Token {
	return m.buf[m.sp+int(n)]
}

// setupScopeFn helps to build Scope for each test case
type setupScopeFn func() *symbol.Scope

// chkScopeFn helps to verify whether symbols are correctly saved
type chkScopeFn func(scope *symbol.Scope) bool

func defaultSetupScopeFn() *symbol.Scope {
	return symbol.NewScope()
}

func defaultChkScopeFn(_ *symbol.Scope) bool {
	return true
}

// TestParserOnly tests three things
//
// 1. "contract" keyword with its open-brace & close-brace
// 2. When there's single & multiple function inside contract
// 3. When there's statements other than function literal
//
func TestParserOnly(t *testing.T) {
	tests := []struct {
		buf         TokenBuffer
		expected    string
		expectedErr error
	}{
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Contract, Val: "contract"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expected: `
contract {
}`,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Contract, Val: "contract"},
					{Type: Lbrace, Val: "{"},
					{Type: Eof},
				},
				0,
			},
			expected: "",
			expectedErr: ExpectError{
				Token{Eof, "eof", 0, 0},
				Rbrace,
			},
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Contract, Val: "contract"},
					{Type: Lbrace, Val: "{"},
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "foo"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expected: `
contract {
func foo() void {

}
}`,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Contract, Val: "contract"},
					{Type: Lbrace, Val: "{"},
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "foo"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "bar"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expected: `
contract {
func foo() void {

}
func bar() void {

}
}`,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Contract, Val: "contract"},
					{Type: Lbrace, Val: "{"},
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "foo"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "a"},
					{Type: IntType, Val: "int"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "bar"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "a"},
					{Type: IntType, Val: "int"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expected: `
contract {
func foo(Parameter : (Identifier: a, Type: int)) void {

}
func bar(Parameter : (Identifier: a, Type: int)) void {

}
}`,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Contract, Val: "contract"},
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "1"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expected: ``,
			expectedErr: ExpectError{
				Token{IntType, "int", 0, 0},
				Rbrace,
			},
		},
	}

	for i, tt := range tests {
		stmt, err := Parse(tt.buf)

		if err != nil && err.Error() != tt.expectedErr.Error() {
			t.Errorf(`test[%d] - Wrong error returned Expected="%v", got="%v"`,
				i, tt.expectedErr, err)
			continue
		}

		if err == nil && stmt.String() != tt.expected {
			t.Errorf(`test[%d] - Wrong result returned Expected="%s", got="%s"`,
				i, tt.expected, stmt.String())
		}
	}
}

func TestCurTokenIs(t *testing.T) {
	tokens := []Token{
		{Type: Int, Val: "1"},
		{Type: Ident, Val: "ADD"},
		{Type: Plus, Val: "+"},
		{Type: Asterisk, Val: "*"},
		{Type: Lparen, Val: "("},
	}
	tokenBuf := mockTokenBuffer{tokens, 0}
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{
			tokenType: Int,
			expected:  true,
		},
		{
			tokenType: Ident,
			expected:  true,
		},
		{
			tokenType: Mod,
			expected:  false,
		},
		{
			tokenType: Rbrace,
			expected:  false,
		},
		{
			tokenType: Lparen,
			expected:  true,
		},
	}

	for i, test := range tests {
		ret := curTokenIs(&tokenBuf, test.tokenType)
		if ret != test.expected {
			t.Fatalf("test[%d] - curTokenIs() result wrong. Expected=%t, got=%t", i, test.expected, ret)
		}
		tokenBuf.Read()
	}
}

func TestNextTokenIs(t *testing.T) {
	tokens := []Token{
		{Type: Int, Val: "1"},
		{Type: Ident, Val: "ADD"},
		{Type: Plus, Val: "+"},
		{Type: Asterisk, Val: "*"},
		{Type: Lparen, Val: "("},
	}
	tokenBuf := mockTokenBuffer{tokens, 0}
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{
			tokenType: Ident,
			expected:  true,
		},
		{
			tokenType: Plus,
			expected:  true,
		},
		{
			tokenType: Minus,
			expected:  false,
		},
		{
			tokenType: Rbrace,
			expected:  false,
		},
	}

	for i, test := range tests {
		ret := nextTokenIs(&tokenBuf, test.tokenType)
		if ret != test.expected {
			t.Fatalf("test[%d] - nextTokenIs() result wrong. Expected=%t, got=%t", i, test.expected, ret)
		}
		tokenBuf.Read()
	}
}

func TestCurPrecedence(t *testing.T) {
	tokens := []Token{
		{Type: Int, Val: "1"},
		{Type: Ident, Val: "ADD"},
		{Type: Plus, Val: "+"},
		{Type: Asterisk, Val: "*"},
		{Type: Lparen, Val: "("},
	}
	tokenBuf := mockTokenBuffer{tokens, 0}
	tests := []struct {
		expected precedence
	}{
		{
			expected: LOWEST,
		},
		{
			expected: LOWEST,
		},
		{
			expected: SUM,
		},
		{
			expected: PRODUCT,
		},
		{
			expected: CALL,
		},
	}

	for i, test := range tests {
		ret := curPrecedence(&tokenBuf)
		if ret != test.expected {
			t.Fatalf("test[%d] - curPrecedence() result wrong. Expected=%d, got=%d", i, test.expected, ret)
		}
		tokenBuf.Read()
	}
}

func TestNextPrecedence(t *testing.T) {
	tokens := []Token{
		{Type: Int, Val: "1"},
		{Type: Ident, Val: "ADD"},
		{Type: Plus, Val: "+"},
		{Type: Asterisk, Val: "*"},
		{Type: Lparen, Val: "("},
	}
	tokenBuf := mockTokenBuffer{tokens, 0}
	tests := []struct {
		expected precedence
	}{
		{
			expected: LOWEST,
		},
		{
			expected: SUM,
		},
		{
			expected: PRODUCT,
		},
		{
			expected: CALL,
		},
	}

	for i, test := range tests {
		ret := nextPrecedence(&tokenBuf)
		if ret != test.expected {
			t.Fatalf("test[%d] - curPrecedence() result wrong. Expected=%d, got=%d", i, test.expected, ret)
		}
		tokenBuf.Read()
	}
}

func TestPeekNumber_IsValid(t *testing.T) {
	tests := []struct {
		n        peekNumber
		expected bool
	}{
		{
			n:        peekNumber(0),
			expected: true,
		},
		{
			n:        peekNumber(1),
			expected: true,
		},
		{
			n:        peekNumber(2),
			expected: false,
		},
		{
			n:        peekNumber(-1),
			expected: false,
		},
	}

	for i, test := range tests {
		n := test.n
		if n.isValid() != test.expected {
			t.Fatalf("test[%d] - isValid() result wrong. Expected=%t, got=%t", i, test.expected, n.isValid())
		}
	}
}

func TestExpectNext(t *testing.T) {
	tokens := []Token{
		{Type: Int, Val: "1"},
		{Type: Ident, Val: "ADD"},
		{Type: Plus, Val: "+"},
		{Type: Asterisk, Val: "*"},
		{Type: Lparen, Val: "("},
	}
	tokenBuf := mockTokenBuffer{tokens, 0}
	tests := []struct {
		token         TokenType
		expectedBool  bool
		expectedError error
	}{
		{
			token:         Int,
			expectedBool:  true,
			expectedError: nil,
		},
		{
			token:        Minus,
			expectedBool: false,
			expectedError: ExpectError{
				Token{Ident, "a", 0, 0},
				Minus,
			},
		},
		{
			token:         Plus,
			expectedBool:  true,
			expectedError: nil,
		},
		{
			token:        Rbrace,
			expectedBool: false,
			expectedError: ExpectError{
				Token{Asterisk, "*", 0, 0},
				Rbrace,
			},
		},
	}

	for i, test := range tests {
		retError := expectNext(&tokenBuf, test.token)
		if retError != nil && retError.Error() != test.expectedError.Error() {
			t.Fatalf("test[%d] - expectNext() result wrong.\n"+
				"Expected error: %s\n"+
				"got error: %s", i, test.expectedError.Error(), retError.Error())
		}

		if retError != nil {
			tokenBuf.Read()
		}
	}
}

func TestParseIdentifier(t *testing.T) {
	tests := []struct {
		buf          TokenBuffer
		setupScope   setupScopeFn
		expected     ast.Expression
		expectedErrs error
	}{
		{
			buf: &mockTokenBuffer{
				[]Token{
					{
						Int,
						"1",
						24,
						12,
					},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   nil,
			expectedErrs: ExpectError{
				Token{Int, "1", 24, 12},
				Ident,
			},
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{
						Ident,
						"ADD",
						125,
						225,
					},
				},
				0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("ADD", &symbol.Integer{Name: &ast.Identifier{Name: "ADD"}})
				return scope
			},
			expected: &ast.Identifier{
				Name: "ADD",
			},
			expectedErrs: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{
						Ident,
						"a",
						125,
						225,
					},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   nil,
			expectedErrs: NotExistSymError{
				Source: Token{
					Ident,
					"a",
					125,
					225,
				},
			},
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{
						Plus,
						"+",
						422,
						12,
					},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   nil,
			expectedErrs: ExpectError{
				Token{Plus, "+", 422, 12},
				Ident,
			},
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{
						Asterisk,
						"*",
						12,
						123,
					},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   nil,
			expectedErrs: ExpectError{
				Token{Asterisk, "*", 12, 123},
				Ident,
			},
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{
						Lparen,
						"(",
						5,
						876,
					},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   nil,
			expectedErrs: ExpectError{
				Token{Lparen, "(", 5, 876},
				Ident,
			},
		},
	}

	for i, test := range tests {
		scope = test.setupScope()
		exp, err := parseIdentifier(test.buf)

		if err != nil && err.Error() != test.expectedErrs.Error() {
			t.Fatalf("test[%d] - wrong error. Expected=%s, got=%s", i, test.expectedErrs, err)
		}

		switch exp {
		case nil:
			if test.expected != nil {
				t.Fatalf("test[%d] - wrong result. Expected=%s, got=%s", i, test.expected.String(), exp.String())
			}
		case &ast.Identifier{Name: exp.String()}:
			if exp.String() != exp.String() {
				t.Fatalf("test[%d] - wrong result. Expected=%s, got=%s", i, test.expected.String(), exp.String())
			}
		}
	}
}

func TestParseIntegerLiteral(t *testing.T) {
	tokens := []Token{
		{Type: Int, Val: "12"},
		{Type: Int, Val: "55"},
		{Type: Int, Val: "a"},
		{Type: String, Val: "abcdefg"},
		{Type: Int, Val: "-13"},
	}
	tokenBuf := mockTokenBuffer{tokens, 0}
	tests := []struct {
		expected    *ast.IntegerLiteral
		expectedErr error
	}{
		{
			expected:    &ast.IntegerLiteral{Value: 12},
			expectedErr: nil,
		},
		{
			expected:    &ast.IntegerLiteral{Value: 55},
			expectedErr: nil,
		},
		{
			expected:    nil,
			expectedErr: errors.New(`strconv.ParseInt: parsing "a": invalid syntax`),
		},
		{
			expected: nil,
			expectedErr: ExpectError{
				Token{Type: String},
				Int,
			},
		},
		{
			expected:    &ast.IntegerLiteral{Value: -13},
			expectedErr: nil,
		},
	}

	for i, test := range tests {
		// For debugging
		tokenBuf.sp = i
		exp, err := parseIntegerLiteral(&tokenBuf)
		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - TestParseIntegerLiteral() wrong error. Expected=%s, got=%s",
				i, test.expectedErr, err.Error())
		}

		if exp != nil && exp.String() != test.expected.String() {
			t.Fatalf("test[%d] - TestParseIntegerLiteral() wrong error. Expected=%s, got=%s",
				i, test.expectedErr, err.Error())
		}
	}
}

func TestParseBooleanLiteral(t *testing.T) {
	tokens := []Token{
		{Type: True, Val: "true"},
		{Type: False, Val: "false"},
		{Type: True, Val: "azzx"},
		{Type: String, Val: "abcdefg"},
	}
	tokenBuf := mockTokenBuffer{tokens, 0}
	tests := []struct {
		expected    *ast.BooleanLiteral
		expectedErr error
	}{
		{
			expected:    &ast.BooleanLiteral{true},
			expectedErr: nil,
		},
		{
			expected:    &ast.BooleanLiteral{false},
			expectedErr: nil,
		},
		{
			expected:    nil,
			expectedErr: errors.New(`strconv.ParseBool: parsing "azzx": invalid syntax`),
		},
		{
			expected: nil,
			expectedErr: ExpectError{
				Token{Type: String},
				BoolType,
			},
		},
	}

	for i, test := range tests {
		exp, err := parseBooleanLiteral(&tokenBuf)

		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf(`test[%d] - TestParseBooleanLiteral() wrong error. Expected="%s", got="%s"`,
				i, test.expectedErr.Error(), err.Error())
		}

		lit, ok := exp.(*ast.BooleanLiteral)
		if err == nil && !ok {
			t.Fatalf("test[%d] - TestParseBooleanLiteral() returned expression is not *ast.BooleanLiteral", i)
		}

		if err == nil && lit.String() != test.expected.String() {
			t.Fatalf(`test[%d] - TestParseBooleanLiteral() wrong result. Expected="%s", got="%s"`,
				i, test.expected, lit.String())
		}
	}
}

func TestParseStringLiteral(t *testing.T) {
	tokens := []Token{
		{Type: String, Val: "hello"},
		{Type: String, Val: "hihi"},
		{Type: Int, Val: "3"},
		{Type: String, Val: "koa zzang"},
	}
	tokenBuf := mockTokenBuffer{tokens, 0}
	tests := []struct {
		expected    *ast.StringLiteral
		expectedErr error
	}{
		{
			expected:    &ast.StringLiteral{Value: "hello"},
			expectedErr: nil,
		},
		{
			expected:    &ast.StringLiteral{Value: "hihi"},
			expectedErr: nil,
		},
		{
			expected: nil,
			expectedErr: ExpectError{
				Token{Type: Int},
				String,
			},
		},
		{
			expected:    &ast.StringLiteral{Value: "koa zzang"},
			expectedErr: nil,
		},
	}

	for i, test := range tests {
		// For debbuging
		tokenBuf.sp = i
		exp, err := parseStringLiteral(&tokenBuf)

		switch err != nil {
		case true:
			if err.Error() != test.expectedErr.Error() {
				t.Fatalf("test[%d] - TestParseStringLiteral() wrong error. Expected=%s, got=%s",
					i, test.expectedErr, err.Error())
			}
		}

		switch exp != nil {
		case true:
			if exp.String() != test.expected.String() {
				t.Fatalf("test[%d] - TestParseStringLiteral() wrong result. Expected=%s, got=%s",
					i, test.expected, exp.String())
			}
		}
	}
}

func TestParseFunctionLiteral(t *testing.T) {
	initParseFnMap()

	tests := []struct {
		buf          TokenBuffer
		setupScope   setupScopeFn
		expectedExpr string
		expectedErr  error
	}{
		{
			&mockTokenBuffer{
				[]Token{
					// func example (a int, b string) {}
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "example"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "a"},
					{Type: IntType, Val: "int"},
					{Type: Comma, Val: ","},
					{Type: Ident, Val: "b"},
					{Type: StringType, Val: "string"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			defaultSetupScopeFn,
			"func example(Parameter : (Identifier: a, Type: int), Parameter : (Identifier: b, Type: string)) void {\n\n}",
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					// func name (a int, b string) {
					//	int c = 5
					// }
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "name"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "a"},
					{Type: IntType, Val: "int"},
					{Type: Comma, Val: ","},
					{Type: Ident, Val: "b"},
					{Type: StringType, Val: "string"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "c"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "5"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			defaultSetupScopeFn,
			"func name(Parameter : (Identifier: a, Type: int), Parameter : (Identifier: b, Type: string)) void {\nint c = 5\n}",
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					// error case
					{Type: Lbrace, Val: "{"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			defaultSetupScopeFn,
			"",
			ExpectError{
				Token{Lbrace, "{", 0, 0},
				Function,
			},
		},
		{
			&mockTokenBuffer{
				[]Token{
					// func example () string {}
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "example"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: StringType, Val: "string"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			defaultSetupScopeFn,
			"func example() string {\n\n}",
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					// func example () string {}
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "example"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: StringType, Val: "string"},
					{Type: Lbrace, Val: "{"},
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Else, Val: "else"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					//{Type: Semicolon, Val: "\n"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			defaultSetupScopeFn,
			`func example() string {
if ( true ) {  } else {  }
}`,
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					// func example () string {}
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "example"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: StringType, Val: "string"},
					{Type: Lbrace, Val: "{"},
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Else, Val: "else"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Else, Val: "else"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			defaultSetupScopeFn,
			`func example() string {
if ( true ) {  } else {  }
if ( true ) {  } else {  }
}`,
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					// func example () invalid {}
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "example"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: Illegal, Val: "invalid"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			defaultSetupScopeFn,
			"",
			Error{
				Token{Type: Illegal},
				"invalid function return type",
			},
		},
		{
			&mockTokenBuffer{
				[]Token{
					// func example () invalid {}
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "example"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: Int, Val: "1"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			defaultSetupScopeFn,
			"",
			Error{
				Token{Type: Int},
				"invalid function return type",
			},
		},
		{
			&mockTokenBuffer{
				[]Token{
					// func example (a int, b string) {}
					{Type: Function, Val: "func"},
					{Type: Ident, Val: "example"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "a"},
					{Type: IntType, Val: "int"},
					{Type: Comma, Val: ","},
					{Type: Ident, Val: "b"},
					{Type: StringType, Val: "string"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("example", &symbol.Function{Name: "example"})
				return scope
			},
			"",
			DupSymError{Token{Type: Ident, Val: "example"}},
		},
	}

	for i, test := range tests {
		scope = test.setupScope()

		exp, err := parseFunctionLiteral(test.buf)

		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - TestParseFunctionLiteral() wrong error\n"+
				"Expected: %s\n"+
				"got: %s", i, test.expectedErr.Error(), err.Error())
		}

		if exp != nil && exp.String() != test.expectedExpr {
			t.Fatalf("test[%d] - TestParseFunctionLiteral wrong result\n"+
				"Expected: %s\n"+
				"got: %s", i, test.expectedExpr, exp.String())
		}
	}
}

func TestParseFunctionParameter(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		buf         TokenBuffer
		setupScope  setupScopeFn
		expected    []string
		expectedErr error
	}{
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "a"},
					{Type: IntType, Val: "int"},
					{Type: Comma, Val: ","},
					{Type: Ident, Val: "b"},
					{Type: StringType, Val: "string"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected: []string{
				"Parameter : (Identifier: a, Type: int)",
				"Parameter : (Identifier: b, Type: string)",
			},
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "arg"},
					{Type: BoolType, Val: "bool"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected: []string{
				"Parameter : (Identifier: arg, Type: bool)",
			},
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "arg"},
					{Type: IntType, Val: "int"},
					{Type: Rbrace, Val: "}"},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   nil,
			expectedErr: ExpectError{
				Token{Rbrace, "}", 0, 0},
				Rparen,
			},
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "arg"},
					{Type: IntType, Val: "int"},
					{Type: Comma, Val: ","},
					{Type: Ident, Val: "arg"},
					{Type: IntType, Val: "int"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			setupScope:  defaultSetupScopeFn,
			expected:    nil,
			expectedErr: DupSymError{Token{Type: Ident, Val: "arg"}},
		},
	}

	for i, test := range tests {
		scope = test.setupScope()
		identifiers, err := parseFunctionParameterList(test.buf)
		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - TestParseFunctionParameter() wrong error.\n"+
				"Expected: %s\n"+
				"got: %s", i, test.expectedErr.Error(), err.Error())
		} else {
			for j, identifier := range identifiers {
				if identifier.String() != test.expected[j] {
					t.Fatalf("test[%d-%d] - TestParseFunctionParameter() failed.\n"+
						"Expected: %s\n"+
						"got: %s", i, j, test.expected[j], identifier)
				}
			}
		}
	}
}

func TestMakePrefixExpression(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		buf         TokenBuffer
		setupScope  setupScopeFn
		expected    string
		expectedErr error
	}{
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "1"},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   "(-1)",
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Minus, Val: "-"},
					{Type: Ident, Val: "a"},
				},
				0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				return scope
			},
			expected: "(-a)",
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Bang, Val: "!"},
					{Type: True, Val: "true"},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   "(!true)",
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Bang, Val: "!"},
					{Type: Bang, Val: "!"},
					{Type: True, Val: "false"},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   "(!(!false))",
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Bang, Val: "!"},
					{Type: Minus, Val: "-"},
					{Type: Ident, Val: "foo"},
				},
				0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("foo", &symbol.Integer{Name: &ast.Identifier{Name: "foo"}})
				return scope
			},
			expected: "(!(-foo))",
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Minus, Val: "-"},
					{Type: Bang, Val: "!"},
					{Type: Ident, Val: "foo"},
				},
				0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("foo", &symbol.Integer{Name: &ast.Identifier{Name: "foo"}})
				return scope
			},
			expected: "(-(!foo))",
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Minus, Val: "-"},
					{Type: True, Val: "true"},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expectedErr: PrefixError{
				Token{Type: Minus, Val: "-"},
				&ast.BooleanLiteral{Value: true},
			},
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Bang, Val: "!"},
					{Type: String, Val: "hello"},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expectedErr: PrefixError{
				Token{Type: Bang, Val: "!"},
				&ast.StringLiteral{Value: "hello"},
			},
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Bang, Val: "!"},
					{Type: Int, Val: "3"},
				},
				0,
			},
			expectedErr: PrefixError{
				Token{Type: Bang, Val: "!"},
				&ast.IntegerLiteral{Value: 3},
			},
			setupScope: defaultSetupScopeFn,
		},
	}

	for i, tt := range tests {
		scope = tt.setupScope()
		exp, err := makePrefixExpression(tt.buf)

		if err != nil && err.Error() != tt.expectedErr.Error() {
			t.Errorf(`test[%d] - Wrong error returned Expected="%v", got="%v"`,
				i, tt.expectedErr, err)
			continue
		}

		if err == nil && exp.String() != tt.expected {
			t.Errorf(`test[%d] - Wrong result returned Expected="%s", got="%s"`,
				i, tt.expected, exp.String())
		}
	}
}

func TestMakeInfixExpression(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		prefix      ast.IntegerLiteral
		buf         TokenBuffer
		expected    string
		expectedErr error
	}{
		{
			prefix: ast.IntegerLiteral{
				Value: 1,
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "3"},
					{Type: Eof, Val: ""},
				},
				0,
			},
			expected:    "(1 + (2 * 3))",
			expectedErr: nil,
		},
		{
			prefix: ast.IntegerLiteral{
				Value: 121,
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "242"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "312"},
					{Type: Eof, Val: ""},
				},
				0,
			},
			expected:    "((121 * 242) + 312)",
			expectedErr: nil,
		},
		{
			prefix: ast.IntegerLiteral{
				Value: -10,
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "15"},
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "55"},
					{Type: Eof, Val: ""},
				},
				0,
			},
			expected:    "((-10 - 15) - 55)",
			expectedErr: nil,
		},
		{
			prefix: ast.IntegerLiteral{
				Value: 1,
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "2"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "3"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "4"},
					{Type: Eof, Val: ""},
				},
				0,
			},
			expected:    "((1 - (2 * 3)) + 4)",
			expectedErr: nil,
		},
		{
			prefix: ast.IntegerLiteral{
				Value: 1,
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Plus, Val: "+"},
					{Type: Plus, Val: "+"},
					{Type: Eof, Val: ""},
				},
				0,
			},
			expected: "",
			expectedErr: Error{
				Token{Type: Plus},
				"prefix parse function not defined",
			},
		},
	}

	// Expected value is
	//      +
	//     / \
	//    1   *
	//       / \
	//      2  3
	// result String() : 1+(2*3)

	for i, test := range tests {
		exp, err := makeInfixExpression(test.buf, &test.prefix, LOWEST)

		if err != nil && test.expectedErr.Error() != err.Error() {
			t.Fatalf("test[%d] - TestMakeInfixExpression() wrong error. Expected=%s, got=%s",
				i, test.expected, err.Error())
		}
		if err == nil && test.expected != exp.String() {
			t.Fatalf("test[%d] - TestMakeInfixExpression() wrong result. Expected=%s, got=%s",
				i, test.expected, exp.String())
		}
	}
}

func TestParseInfixExpression(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		buf         TokenBuffer
		left        ast.IntegerLiteral
		expected    string
		expectedErr error
	}{
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Int, Val: "1"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "3"},
					{Type: Eof, Val: ""},
				},
				1,
			},
			left: ast.IntegerLiteral{
				Value: 1,
			},
			expected:    "(1 + (2 * 3))",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Int, Val: "121"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "242"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "312"},
					{Type: Eof, Val: ""},
				},
				1,
			},
			left: ast.IntegerLiteral{
				Value: 121,
			},
			expected:    "(121 * 242)",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Int, Val: "-10"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "15"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "55"},
					{Type: Eof, Val: ""},
				},
				1,
			},
			left: ast.IntegerLiteral{
				Value: -10,
			},
			expected:    "(-10 * 15)",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Int, Val: "1"},
					{Type: Plus, Val: "+"},
					{Type: Plus, Val: "+"},
					{Type: Eof, Val: ""},
				},
				1,
			},
			left: ast.IntegerLiteral{
				Value: 1,
			},
			expected: "",
			expectedErr: Error{
				Token{Type: Plus},
				"prefix parse function not defined",
			},
		},
	}

	for i, test := range tests {
		exp, err := parseInfixExpression(test.buf, &test.left)

		if err != nil && test.expectedErr.Error() != err.Error() {
			t.Fatalf("test[%d] - TestMakeInfixExpression() wrong error. Expected=%s, got=%s",
				i, test.expectedErr.Error(), err.Error())
		}

		if err == nil && test.expected != exp.String() {
			t.Fatalf("test[%d] - TestMakeInfixExpression() wrong result. Expected=%s, got=%s",
				i, test.expected, exp.String())
		}
	}
}

func TestParseGroupedExpression(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		buf         TokenBuffer
		setupScope  setupScopeFn
		expected    string
		expectedErr error
	}{
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Int, Val: "2"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "1"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
				},
				0,
			},
			setupScope:  defaultSetupScopeFn,
			expected:    "(2 + 1)",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "a"},
					{Type: Plus, Val: "+"},
					{Type: Lparen, Val: "("},
					{Type: Int, Val: "1"},
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "2"},
					{Type: Rparen, Val: ")"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
				},
				0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				return scope
			},
			expected:    "(a + (1 - 2))",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "a"},
					{Type: Plus, Val: "+"},
					{Type: Lparen, Val: "("},
					{Type: Int, Val: "1"},
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "2"},
					{Type: Rparen, Val: ")"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "3"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
				},
				0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				return scope
			},
			expected:    "((a + (1 - 2)) + 3)",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Int, Val: "2"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "1"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   "",
			expectedErr: ExpectError{
				Token{Rbrace, "{", 0, 0},
				Rparen,
			},
		},
	}

	for i, test := range tests {
		scope = test.setupScope()
		exp, err := parseGroupedExpression(test.buf)

		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - TestParseGroupedExpression() wrong error.\n"+
				"Expected=%s,\n"+
				"got=%s",
				i, test.expectedErr.Error(), err.Error())
		}

		if exp != nil && exp.String() != test.expected {
			t.Fatalf("test[%d] - TestParseGroupedExpression() wrong answer.\n"+
				"Expected=%s,\n"+
				"got=%s",
				i, test.expected, exp.String())
		}
	}
}

func TestParseReturnStatement(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		buf         TokenBuffer
		expected    string
		expectedErr error
	}{
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Return, Val: "return"},
					{Type: True, Val: "true"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			expected:    "return true",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Return, Val: "return"},
					{Type: Int, Val: "1"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			expected:    "return (1 + 2)",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Return, Val: "return"},
					{Type: Int, Val: "1"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "3"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			expected:    "return (1 + (2 * 3))",
			expectedErr: nil,
		},
		// test return void
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Return, Val: "return"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			expected:    "return",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "1"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			expected: "",
			expectedErr: ExpectError{
				Token{IntType, "int", 0, 0},
				Return,
			},
		},
	}

	for i, test := range tests {
		exp, err := parseReturnStatement(test.buf)

		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - TestParseReturnStatement() wrong error.\n"+
				"Expected=%s,\n"+
				"got=%s",
				i, test.expectedErr.Error(), err.Error())
		}

		if exp != nil && exp.String() != test.expected {
			t.Fatalf("test[%d] - TestParseReturnStatement() wrong result.\n"+
				"Expected=%s,\n"+
				"got=%s",
				i, test.expected, exp.String())
		}
	}
}

func TestParsePrefixExpression(t *testing.T) {
	tests := []struct {
		tokenBuffer      TokenBuffer
		expectedOperator string
		expectedRight    string
	}{
		{
			&mockTokenBuffer{
				buf: []Token{
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "1"},
					{Type: Eof}},
				sp: 0,
			},
			"-", "1",
		},
		{
			&mockTokenBuffer{
				buf: []Token{
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "3333"},
					{Type: Eof}},
				sp: 0,
			},
			"-", "3333",
		},
		{
			&mockTokenBuffer{
				buf: []Token{
					{Type: Bang, Val: "!"},
					{Type: True, Val: "true"},
					{Type: Eof}},
				sp: 0,
			},
			"!", "true",
		},
		{
			&mockTokenBuffer{
				buf: []Token{
					{Type: Bang, Val: "!"},
					{Type: False, Val: "false"},
					{Type: Eof}},
				sp: 0,
			},
			"!", "false",
		},
	}

	prefixParseFnMap[Int] = parseIntegerLiteral
	prefixParseFnMap[True] = parseBooleanLiteral
	prefixParseFnMap[False] = parseBooleanLiteral

	for i, tt := range tests {
		exp, err := parsePrefixExpression(tt.tokenBuffer)
		if err != nil {
			t.Errorf(`tests[%d] - Returned error is "%s"`,
				i, err)
		}

		expression, ok := exp.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("exp is not *ast.PrefixExpression. got=%T", exp)
		}

		if expression.Operator.String() != tt.expectedOperator {
			t.Errorf("tests[%d] - Type is not %s but got %s",
				i, tt.expectedOperator, expression.Operator.String())
		}

		if expression.Right.String() != tt.expectedRight {
			t.Errorf("tests[%d] - Name is not %s but got %s",
				i, tt.expectedRight, expression.Right.String())
		}
	}
}

func TestParseCallExpression(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		setupScope  setupScopeFn
		buf         TokenBuffer
		function    ast.Expression
		expected    string
		expectedErr error
	}{
		{
			setupScope: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Int, Val: "1"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			function:    &ast.Identifier{Name: "add"},
			expected:    `function add( (1 + 2) )`,
			expectedErr: nil,
		},
		{
			setupScope: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: String, Val: "a"},
					{Type: Comma, Val: ","},
					{Type: String, Val: "b"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "5"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			function:    &ast.Identifier{Name: "testFunc"},
			expected:    `function testFunc( a, b, 5 )`,
			expectedErr: nil,
		},
		{
			setupScope: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: String, Val: "a"},
					{Type: Comma, Val: ","},
					{Type: Asterisk, Val: "*"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "5"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			function: &ast.Identifier{Name: "errorFunc"},
			expected: "",
			expectedErr: Error{
				Token{Type: Asterisk},
				"prefix parse function not defined",
			},
		},
		{
			setupScope: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: String, Val: "a"},
					{Type: Plus, Val: "+"},
					{Type: String, Val: "b"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "5"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "3"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			function:    &ast.Identifier{Name: "complexFunc"},
			expected:    `function complexFunc( (a + b), (5 * 3) )`,
			expectedErr: nil,
		},
		{
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("add", &symbol.Integer{Name: &ast.Identifier{Name: "add"}})
				return scope
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "add"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			function:    &ast.Identifier{Name: "complexFunc"},
			expected:    `function complexFunc( function add(  ) )`,
			expectedErr: nil,
		},
		{
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("add", &symbol.Integer{Name: &ast.Identifier{Name: "add"}})
				return scope
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "add"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "1"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			function:    &ast.Identifier{Name: "complexFunc"},
			expected:    `function complexFunc( function add(  ), 1 )`,
			expectedErr: nil,
		},
		// test empty arguments function call
		{
			setupScope: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			function:    &ast.Identifier{Name: "complexFunc"},
			expected:    `function complexFunc(  )`,
			expectedErr: nil,
		},
	}

	for i, test := range tests {
		scope = test.setupScope()

		exp, err := parseCallExpression(test.buf, test.function)

		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - parseCallExpression() wrong error. Expected=%s, got=%s",
				i, test.expectedErr.Error(), err.Error())
		}
		if exp != nil && exp.String() != test.expected {
			t.Fatalf("test[%d] - parseCallExpression() wrong answer. Expected=%s, got=%s",
				i, test.expected, exp.String())
		}
	}
}

func TestParseCallArguments(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		buf         TokenBuffer
		setupScope  setupScopeFn
		expected    string
		expectedErr error
	}{
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			setupScope:  defaultSetupScopeFn,
			expected:    "function testFunction(  )",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Int, Val: "1"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			setupScope:  defaultSetupScopeFn,
			expected:    "function testFunction( 1 )",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: String, Val: "a"},
					{Type: Comma, Val: ","},
					{Type: String, Val: "b"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "5"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				scope.Set("b", &symbol.Integer{Name: &ast.Identifier{Name: "b"}})
				return scope
			},
			expected:    `function testFunction( a, b, 5 )`,
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: String, Val: "a"},
					{Type: Plus, Val: "+"},
					{Type: String, Val: "b"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "5"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "3"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				scope.Set("b", &symbol.Integer{Name: &ast.Identifier{Name: "b"}})
				return scope
			},
			expected:    `function testFunction( (a + b), (5 * 3) )`,
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Asterisk, Val: "*"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			setupScope: defaultSetupScopeFn,
			expected:   "",
			expectedErr: Error{
				Token{Type: Asterisk},
				"prefix parse function not defined",
			},
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: String, Val: "a"},
					{Type: Comma, Val: ","},
					{Type: Asterisk, Val: "*"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "5"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				return scope
			},
			expected: "",
			expectedErr: Error{
				Token{Type: Asterisk},
				"prefix parse function not defined",
			},
		},
		{
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "add"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "1"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("add", &symbol.Integer{Name: &ast.Identifier{Name: "add"}})
				return scope
			},
			expected:    `function testFunction( function add(  ), 1 )`,
			expectedErr: nil,
		},
	}

	for i, test := range tests {
		scope = test.setupScope()
		exp, err := parseCallArguments(test.buf)

		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - TestParseCallArguments() wrong error. Expected=%s, got=%s",
				i, test.expectedErr.Error(), err.Error())
		}

		mockFn := &ast.CallExpression{
			Function: &ast.Identifier{Name: "testFunction"},
		}
		mockFn.Arguments = exp
		if exp != nil && mockFn.String() != test.expected {
			t.Fatalf("test[%d] - TestParseCallArguments() wrong error. Expected=%s, got=%s",
				i, test.expected, mockFn.String())
		}
	}
}

func TestParseAssignStatement(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		setupScopeFn
		tokenBuffer           TokenBuffer
		expectedDataStructure string
		expectedIdent         string
		expectedVal           string
		expectedErr           error
		chkScopeFn
	}{
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				buf: []Token{
					{Type: StringType, Val: "string"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "hello"},
					{Type: Semicolon},
					{Type: Eof},
				},
				sp: 0,
			},
			"string",
			"a",
			"hello",
			nil,
			func(scope *symbol.Scope) bool {
				sym := scope.Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.StringSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				return true
			},
		},
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				buf: []Token{
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "myInt"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "1"},
					{Type: Semicolon},
					{Type: Eof},
				},
				sp: 0,
			},
			"int",
			"myInt",
			"1",
			nil,
			func(scope *symbol.Scope) bool {
				sym := scope.Get("myInt")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "myInt" {
					return false
				}

				return true
			},
		},
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				buf: []Token{
					{Type: BoolType, Val: "bool"},
					{Type: Ident, Val: "ddd"},
					{Type: Assign, Val: "="},
					{Type: True, Val: "true"},
					{Type: Semicolon},
					{Type: Eof},
				},
				sp: 0,
			},
			"bool",
			"ddd",
			"true",
			nil,
			func(scope *symbol.Scope) bool {
				sym := scope.Get("ddd")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.BooleanSymbol {
					return false
				}

				if sym.String() != "ddd" {
					return false
				}

				return true
			},
		},
		{
			// type mismatch tc - int ddd2 = "iam_string"
			defaultSetupScopeFn,
			&mockTokenBuffer{
				buf: []Token{
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "ddd2"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "iam_string"},
					{Type: Semicolon},
					{Type: Eof},
				},
				sp: 0,
			},
			"int",
			"ddd2",
			"iam_string",
			nil,
			func(scope *symbol.Scope) bool {
				sym := scope.Get("ddd2")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "ddd2" {
					return false
				}

				return true
			},
		},
		{
			// type mismatch tc - bool foo = "iam_string"
			defaultSetupScopeFn,
			&mockTokenBuffer{
				buf: []Token{
					{Type: BoolType, Val: "bool"},
					{Type: Ident, Val: "foo"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "iam_string"},
					{Type: Semicolon},
					{Type: Eof},
				},
				sp: 0,
			},
			"bool",
			"foo",
			"iam_string",
			nil,
			func(scope *symbol.Scope) bool {
				sym := scope.Get("foo")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.BooleanSymbol {
					return false
				}

				if sym.String() != "foo" {
					return false
				}

				return true
			},
		},
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				buf: []Token{
					{Type: BoolType, Val: "bool"},
					{Type: String, Val: "foo"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "iam_string"},
					{Type: Semicolon},
					{Type: Eof},
				},
				sp: 0,
			},
			"bool",
			"foo",
			`"iam_string"`,
			ExpectError{
				Token{Type: String},
				Ident,
			},
			func(scope *symbol.Scope) bool {
				sym := scope.Get("foo")
				if sym != nil {
					return false
				}

				return true
			},
		},
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				buf: []Token{
					{Type: BoolType, Val: "bool"},
					{Type: Ident, Val: "foo"},
					{Type: String, Val: "iam_string"},
					{Type: Semicolon},
					{Type: Eof},
				},
				sp: 0,
			},
			"bool",
			"foo",
			"iam_string",
			ExpectError{
				Token{Type: String},
				Assign,
			},
			func(scope *symbol.Scope) bool {
				sym := scope.Get("foo")
				if sym == nil {
					return false
				}

				return true
			},
		},
		// test when identifier already exist
		{
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("ddd", &symbol.String{Name: &ast.Identifier{Name: "ddd"}})
				return scope
			},
			&mockTokenBuffer{
				buf: []Token{
					{Type: BoolType, Val: "bool"},
					{Type: Ident, Val: "ddd"},
					{Type: Assign, Val: "="},
					{Type: True, Val: "true"},
					{Type: Semicolon},
					{Type: Eof},
				},
				sp: 0,
			},
			"bool",
			"ddd",
			"true",
			DupSymError{Token{Type: Ident, Val: "ddd"}},
			func(scope *symbol.Scope) bool {
				sym := scope.Get("ddd")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.StringSymbol {
					return false
				}

				if sym.String() != "ddd" {
					return false
				}

				return true
			},
		},
	}

	for i, tt := range tests {
		// setup
		scope = tt.setupScopeFn()

		// exercise
		exp, err := parseAssignStatement(tt.tokenBuffer)

		// verify
		if err != nil && err.Error() != tt.expectedErr.Error() {
			t.Errorf(`tests[%d] - Returned err is not "%s", but got "%s"`,
				i, tt.expectedErr.Error(), err.Error())
		}

		if err == nil && exp.Type.String() != tt.expectedDataStructure {
			t.Errorf("tests[%d] - Type is not %s but got %s",
				i, tt.expectedDataStructure, exp.Type.String())
		}

		if err == nil && exp.Variable.String() != tt.expectedIdent {
			t.Errorf("tests[%d] - Variable is not %s but got %s",
				i, tt.expectedIdent, exp.Variable.String())
		}

		if err == nil && exp.Value.String() != tt.expectedVal {
			t.Errorf("tests[%d] - Name is not %s but got %s",
				i, tt.expectedVal, exp.Value.String())
		}

		if !tt.chkScopeFn(scope) {
			t.Errorf("test[%d] - updateScopeSymbol updates scope incorrectly", i)
		}
	}
}

func TestParseReassignStatement(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		buf         TokenBuffer
		setupScope  setupScopeFn
		expected    string
		expectedErr error
	}{
		{
			buf: &mockTokenBuffer{
				buf: []Token{
					{
						Type: Ident,
						Val:  "a",
					},
					{
						Type: Assign,
						Val:  "=",
					},
					{
						Type: Int,
						Val:  "1",
					},
				},
				sp: 0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				return scope
			},
			expected:    "a = 1",
			expectedErr: nil,
		},
		{
			buf: &mockTokenBuffer{
				buf: []Token{
					{
						Type: Ident,
						Val:  "b",
					},
					{
						Type: Assign,
						Val:  "=",
					},
					{
						Type: Int,
						Val:  "1",
					},
				},
				sp: 0,
			},
			setupScope: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				return scope
			},
			expected: "",
			expectedErr: NotExistSymError{Token{
				Type: Ident,
				Val:  "b",
			}},
		},
	}

	for i, test := range tests {
		scope = test.setupScope()
		stmt, err := parseReassignStatement(test.buf)
		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - parseReassignStatement() returns wrong error.\n"+
				"Expected=%s\n"+
				"got=%s", i, test.expectedErr.Error(), err.Error())
		}

		if stmt != nil && stmt.String() != test.expected {
			t.Fatalf("test[%d] - parseReassignStatement() returns wrong result.\n"+
				"Expected=%s\n"+
				"got=%s", i, test.expected, stmt.String())
		}
	}
}

// TestParseExpression tests strings which combine prefix and
// infix expression
func TestParseExpression(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		buf         TokenBuffer
		setupScope  setupScopeFn
		expected    string
		expectedErr error
	}{
		{
			&mockTokenBuffer{
				[]Token{
					{Type: Minus, Val: "-"},
					{Type: Ident, Val: "a"},
					{Type: Asterisk, Val: "*"},
					{Type: Ident, Val: "b"},
					{Type: Eof},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				scope.Set("b", &symbol.Integer{Name: &ast.Identifier{Name: "b"}})
				return scope
			},
			"((-a) * b)",
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					{Type: Bang, Val: "!"},
					{Type: Minus, Val: "-"},
					{Type: Ident, Val: "b"},
					{Type: Eof},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("b", &symbol.Integer{Name: &ast.Identifier{Name: "b"}})
				return scope
			},
			"(!(-b))",
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "33"},
					{Type: Slash, Val: "/"},
					{Type: Int, Val: "67"},
					{Type: Plus, Val: "+"},
					{Type: Ident, Val: "a"},
					{Type: Eof},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				return scope
			},
			"(((-33) / 67) + a)",
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					{Type: Int, Val: "33"},
					{Type: Mod, Val: "%"},
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "67"},
					{Type: Plus, Val: "+"},
					{Type: Ident, Val: "a"},
					{Type: Asterisk, Val: "*"},
					{Type: Ident, Val: "c"},
					{Type: Eof},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				scope.Set("c", &symbol.Integer{Name: &ast.Identifier{Name: "c"}})
				return scope
			},
			"((33 % (-67)) + (a * c))",
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					{Type: Int, Val: "33"},
					{Type: Mod, Val: "%"},
					{Type: Lparen, Val: "("},
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "67"},
					{Type: Plus, Val: "+"},
					{Type: Ident, Val: "a"},
					{Type: Rparen, Val: ")"},
					{Type: Asterisk, Val: "*"},
					{Type: Ident, Val: "c"},
					{Type: Eof},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				scope.Set("c", &symbol.Integer{Name: &ast.Identifier{Name: "c"}})
				return scope
			},
			"((33 % ((-67) + a)) * c)",
			nil},
		{
			&mockTokenBuffer{
				[]Token{
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "33"},
					{Type: Slash, Val: "/"},
					{Type: Int, Val: "67"},
					{Type: LT, Val: "<"},
					{Type: Ident, Val: "a"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "67"},
					{Type: Eof},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				return scope
			},
			"(((-33) / 67) < (a * 67))",
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					{Type: Minus, Val: "-"},
					{Type: Int, Val: "33"},
					{Type: Slash, Val: "/"},
					{Type: Int, Val: "67"},
					{Type: GTE, Val: ">="},
					{Type: Ident, Val: "a"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "67"},
					{Type: Mod, Val: "%"},
					{Type: Ident, Val: "z"},
					{Type: Eof},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				scope.Set("z", &symbol.Integer{Name: &ast.Identifier{Name: "z"}})
				return scope
			},
			"(((-33) / 67) >= (a + (67 % z)))",
			nil},
		{
			&mockTokenBuffer{
				[]Token{
					{Type: Int, Val: "33"},
					{Type: EQ, Val: "=="},
					{Type: Int, Val: "3"},
					{Type: Land, Val: "&&"},
					{Type: Int, Val: "44"},
					{Type: EQ, Val: "=="},
					{Type: Int, Val: "67"},
					{Type: Eof},
				},
				0,
			},
			defaultSetupScopeFn,
			"((33 == 3) && (44 == 67))",
			nil,
		},
		{
			&mockTokenBuffer{
				[]Token{
					{Type: True, Val: "true"},
					{Type: Lor, Val: "||"},
					{Type: False, Val: "false"},
					{Type: Land, Val: "&&"},
					{Type: True, Val: "true"},
					{Type: Eof},
				},
				0,
			},
			defaultSetupScopeFn,
			"(true || (false && true))",
			nil,
		},
	}

	for i, test := range tests {
		scope = test.setupScope()
		exp, err := parseExpression(test.buf, LOWEST)

		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - parseExpression() with wrong error. Expected=%s, got=%s",
				i, test.expectedErr.Error(), err)
		}

		if err == nil && exp.String() != test.expected {
			t.Fatalf("test[%d] - parseExpression() with wrong expression. Expected=%s, got=%s",
				i, test.expected, exp.String())
		}
	}
}

func TestParseIfStatement(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		setupScopeFn
		buf         TokenBuffer
		expected    string
		expectedErr error
		chkScopeFn
	}{
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				[]Token{
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "0"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			"if ( true ) { int a = 0 }",
			nil,
			func(scope *symbol.Scope) bool {
				sym := scope.GetInner()[0].Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				return true
			},
		},
		{
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				return scope
			},
			&mockTokenBuffer{
				[]Token{
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "a"},
					{Type: EQ, Val: "=="},
					{Type: Int, Val: "5"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "b"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "1"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Else, Val: "else"},
					{Type: Lbrace, Val: "{"},
					{Type: StringType, Val: "string"},
					{Type: Ident, Val: "b"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "example"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			`if ( (a == 5) ) { int b = 1 } else { string b = example }`,
			nil,
			func(scope *symbol.Scope) bool {
				sym := scope.GetInner()[0].Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				sym = scope.GetInner()[1].Get("b")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.StringSymbol {
					return false
				}

				if sym.String() != "b" {
					return false
				}

				return true
			},
		},
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				[]Token{
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "5"},
					{Type: Eof},
				},
				0,
			},
			"",
			ExpectError{
				Token{Type: IntType},
				If,
			},
			defaultChkScopeFn,
		},
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				[]Token{
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rparen, Val: ")"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "0"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			"",
			ExpectError{
				Token{Type: IntType},
				Lbrace,
			},
			defaultChkScopeFn,
		},
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				[]Token{
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rbrace, Val: "}"},
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "0"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			"",
			ExpectError{
				Token{Type: Rbrace},
				Rparen,
			},
			defaultChkScopeFn,
		},
	}

	for i, test := range tests {
		// setup
		scope = test.setupScopeFn()

		// exercise
		stmt, err := parseIfStatement(test.buf)

		// verify
		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - TestParseIfStatement() wrong error. Expected=%s got=%s",
				i, test.expectedErr.Error(), err.Error())
		}

		if stmt != nil && stmt.String() != test.expected {
			t.Fatalf("test[%d] - TestParseIfStatement() wrong result. Expected=%s, got=%s",
				i, test.expected, stmt.String())
		}

		if !test.chkScopeFn(scope) {
			t.Fatalf("test[%d] - updateScopeSymbol updates scope incorrectly", i)
		}
	}
}

func TestParseBlockStatement(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		setupScopeFn
		buf         TokenBuffer
		expected    string
		expectedErr error
		chkScopeFn
	}{
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				[]Token{
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "0"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
				},
				0,
			},
			"int a = 0",
			nil,
			func(scope *symbol.Scope) bool {
				sym := scope.GetInner()[0].Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				return true
			},
		},
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				[]Token{
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "0"},
					{Type: Semicolon, Val: "\n"},
					{Type: StringType, Val: "string"},
					{Type: Ident, Val: "b"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "abc"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
				},
				0,
			},
			`int a = 0
string b = abc`,
			nil,
			func(scope *symbol.Scope) bool {
				sym := scope.GetInner()[0].Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				sym = scope.GetInner()[0].Get("b")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.StringSymbol {
					return false
				}

				if sym.String() != "b" {
					return false
				}

				return true
			},
		},
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				[]Token{
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "0"},
					{Type: Semicolon, Val: "\n"},
					{Type: StringType, Val: "string"},
					{Type: Ident, Val: "b"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "abc"},
					{Type: Semicolon, Val: "\n"},
					{Type: BoolType, Val: "bool"},
					{Type: Ident, Val: "c"},
					{Type: Assign, Val: "="},
					{Type: True, Val: "true"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
				},
				0,
			},
			`int a = 0
string b = abc
bool c = true`,
			nil,
			func(scope *symbol.Scope) bool {
				sym := scope.GetInner()[0].Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				sym = scope.GetInner()[0].Get("b")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.StringSymbol {
					return false
				}

				if sym.String() != "b" {
					return false
				}

				sym = scope.GetInner()[0].Get("c")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.BooleanSymbol {
					return false
				}

				if sym.String() != "c" {
					return false
				}

				return true
			},
		},
		{
			defaultSetupScopeFn,
			&mockTokenBuffer{
				[]Token{
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "0"},
					{Type: Semicolon, Val: "\n"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "1"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
				},
				0,
			},
			"int a = 0\na = 1",
			nil,
			defaultChkScopeFn,
		},
	}

	for i, test := range tests {
		// setup
		scope = test.setupScopeFn()

		// exercise
		exp, err := parseBlockStatement(test.buf)

		// verify
		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - TestParseBlockStatement() wrong error. Expected=%s, got=%s",
				i, test.expectedErr.Error(), err.Error())
		}

		if exp != nil && exp.String() != test.expected {
			t.Fatalf("test[%d] - TestParseBlockStatement() wrong result. Expected=%s, got=%s",
				i, test.expected, exp.String())
		}

		if !test.chkScopeFn(scope) {
			t.Fatalf("test[%d] - updateScopeSymbol updates scope incorrectly", i)
		}
	}
}

func TestParseStatement(t *testing.T) {
	initParseFnMap()
	tests := []struct {
		setupScopeFn
		buf          TokenBuffer
		expectedErr  error
		expectedStmt string
		chkScopeFn
	}{
		// tests for IntType
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "1"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: "int a = 1",
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "1"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: "int a = (1 + 2)",
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "1"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "3"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: "int a = (1 + (2 * 3))",
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "1"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `int a = 1`,
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				return true
			},
		},

		// tests for StringType
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: StringType, Val: "string"},
					{Type: Ident, Val: "abb"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "do not merge, rebase!"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `string abb = do not merge, rebase!`,
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.Get("abb")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.StringSymbol {
					return false
				}

				if sym.String() != "abb" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: StringType, Val: "string"},
					{Type: Ident, Val: "abb"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "hello,*+"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `string abb = hello,*+`,
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.Get("abb")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.StringSymbol {
					return false
				}

				if sym.String() != "abb" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: StringType, Val: "string"},
					{Type: Ident, Val: "abb"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "1"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `string abb = 1`,
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.Get("abb")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.StringSymbol {
					return false
				}

				if sym.String() != "abb" {
					return false
				}

				return true
			},
		},

		// tests for BoolType
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: BoolType, Val: "bool"},
					{Type: Ident, Val: "asdf"},
					{Type: Assign, Val: "="},
					{Type: True, Val: "true"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `bool asdf = true`,
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.Get("asdf")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.BooleanSymbol {
					return false
				}

				if sym.String() != "asdf" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: BoolType, Val: "bool"},
					{Type: Ident, Val: "asdf"},
					{Type: Assign, Val: "="},
					{Type: False, Val: "false"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `bool asdf = false`,
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.Get("asdf")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.BooleanSymbol {
					return false
				}

				if sym.String() != "asdf" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: BoolType, Val: "bool"},
					{Type: Ident, Val: "asdf"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "1"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `bool asdf = 1`,
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.Get("asdf")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.BooleanSymbol {
					return false
				}

				if sym.String() != "asdf" {
					return false
				}

				return true
			},
		},

		// tests for If statement
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `if ( true ) {  }`,
			chkScopeFn:   defaultChkScopeFn,
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: Int, Val: "1"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: EQ, Val: "=="},
					{Type: Int, Val: "3"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `if ( ((1 + 2) == 3) ) {  }`,
			chkScopeFn:   defaultChkScopeFn,
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: Int, Val: "1"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: EQ, Val: "=="},
					{Type: Int, Val: "3"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr: ExpectError{
				Token{Type: Int},
				Ident,
			},
			expectedStmt: ``,
			chkScopeFn:   defaultChkScopeFn,
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "2"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `if ( true ) { int a = 2 }`,
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.GetInner()[0].Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "2"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Else, Val: "else"},
					{Type: Lbrace, Val: "{"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `if ( true ) { int a = 2 } else {  }`,
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.GetInner()[0].Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: If, Val: "if"},
					{Type: Lparen, Val: "("},
					{Type: True, Val: "true"},
					{Type: Rparen, Val: ")"},
					{Type: Lbrace, Val: "{"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "a"},
					{Type: Assign, Val: "="},
					{Type: Int, Val: "2"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Else, Val: "else"},
					{Type: Lbrace, Val: "{"},
					{Type: StringType, Val: "string"},
					{Type: Ident, Val: "b"},
					{Type: Assign, Val: "="},
					{Type: String, Val: "hello"},
					{Type: Semicolon, Val: "\n"},
					{Type: Rbrace, Val: "}"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `if ( true ) { int a = 2 } else { string b = hello }`,
			chkScopeFn: func(scope *symbol.Scope) bool {
				sym := scope.GetInner()[0].Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				sym = scope.GetInner()[1].Get("b")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.StringSymbol {
					return false
				}

				if sym.String() != "b" {
					return false
				}

				return true
			},
		},

		// tests for Return statement
		{
			setupScopeFn: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("asdf", &symbol.Integer{Name: &ast.Identifier{Name: "asdf"}})
				return scope
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Return, Val: "return"},
					{Type: Ident, Val: "asdf"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `return asdf`,
			chkScopeFn:   defaultChkScopeFn,
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Return, Val: "return"},
					{Type: Lparen, Val: "("},
					{Type: Int, Val: "1"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "3"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `return (1 + (2 * 3))`,
			chkScopeFn:   defaultChkScopeFn,
		},
		// test call expression
		{
			setupScopeFn: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("add", &symbol.Function{Name: "add"})
				return scope
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "add"},
					{Type: Lparen, Val: "("},
					{Type: Int, Val: "1"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "2"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `function add( 1, 2 )`,
			chkScopeFn:   defaultChkScopeFn,
		},
		{
			setupScopeFn: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("add", &symbol.Function{Name: "add"})
				scope.Set("foo", &symbol.Function{Name: "foo"})
				return scope
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "add"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "foo"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "2"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `function add( function foo(  ), 2 )`,
			chkScopeFn:   defaultChkScopeFn,
		},
		{
			setupScopeFn: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("add", &symbol.Function{Name: "add"})
				return scope
			},
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Return, Val: "return"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "add"},
					{Type: Lparen, Val: "("},
					{Type: Int, Val: "1"},
					{Type: Comma, Val: ","},
					{Type: Int, Val: "2"},
					{Type: Rparen, Val: ")"},
					{Type: Plus, Val: "+"},
					{Type: Int, Val: "2"},
					{Type: Asterisk, Val: "*"},
					{Type: Int, Val: "3"},
					{Type: Rparen, Val: ")"},
					{Type: Semicolon, Val: "\n"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			expectedErr:  nil,
			expectedStmt: `return (function add( 1, 2 ) + (2 * 3))`,
			chkScopeFn:   defaultChkScopeFn,
		},

		// tests for Default
		{
			setupScopeFn: defaultSetupScopeFn,
			buf: &mockTokenBuffer{
				[]Token{
					{Type: Int, Val: "1"},
					{Type: Eof, Val: "eof"},
				},
				0,
			},
			expectedErr: ExpectError{
				Token{Type: Int},
				Ident,
			},
			expectedStmt: ``,
			chkScopeFn:   defaultChkScopeFn,
		},
	}

	for i, test := range tests {
		// setup
		scope = test.setupScopeFn()

		// exercise
		stmt, err := parseStatement(test.buf)

		// verify
		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Errorf(`test[%d] - parseStatement wrong error. Expected="%v", got="%v"`,
				i, test.expectedErr, err)
			continue
		}

		if err == nil && stmt.String() != test.expectedStmt {
			t.Errorf(`test[%d] - parseStatement wrong result. Expected="%s", got="%s"`,
				i, test.expectedStmt, stmt.String())
		}

		if !test.chkScopeFn(scope) {
			t.Errorf("test[%d] - updateScopeSymbol updates scope incorrectly", i)
		}
	}
}

func TestParseExpressionStatement(t *testing.T) {
	initParseFnMap()

	tests := []struct {
		buf          TokenBuffer
		setupScope   setupScopeFn
		expectedStmt string
		expectedErr  error
	}{
		{
			// add()
			&mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "add"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("add", &symbol.Function{Name: "add"})
				return scope
			},
			"function add(  )",
			nil,
		},
		{
			// read(x int)
			&mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "read"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "x"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("read", &symbol.Function{Name: "read"})
				scope.Set("x", &symbol.Integer{Name: &ast.Identifier{Name: "x"}})
				return scope
			},
			"function read( x )",
			nil,
		},
		{
			// testFunction(a int, b string)
			&mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "testFunction"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "a"},
					{Type: Comma, Val: ","},
					{Type: Ident, Val: "b"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("testFunction", &symbol.Function{Name: "testFunction"})
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				scope.Set("b", &symbol.Integer{Name: &ast.Identifier{Name: "b"}})
				return scope
			},
			"function testFunction( a, b )",
			nil,
		},
		{
			// testFunction(a int b string) <= error case
			&mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "testFunction"},
					{Type: Lparen, Val: "("},
					{Type: Ident, Val: "a"},
					{Type: IntType, Val: "int"},
					{Type: Ident, Val: "b"},
					{Type: IntType, Val: "string"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("testFunction", &symbol.Function{Name: "testFunction"})
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				scope.Set("b", &symbol.Integer{Name: &ast.Identifier{Name: "b"}})
				return scope
			},
			"",
			ExpectError{
				Token{IntType, "int", 0, 0},
				Rparen,
			},
		},
		{
			// 1() <= error case
			&mockTokenBuffer{
				[]Token{
					{Type: Int, Val: "1"},
					{Type: Lparen, Val: "("},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			defaultSetupScopeFn,
			"",
			ExpectError{
				Token{Type: Int},
				Ident,
			},
		},
		{
			// add) <= error case
			&mockTokenBuffer{
				[]Token{
					{Type: Ident, Val: "add"},
					{Type: Rparen, Val: ")"},
				},
				0,
			},
			func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("add", &symbol.Function{Name: "add"})
				return scope
			},
			"",
			ExpectError{
				Token{Rparen, "}", 0, 0},
				Lparen,
			},
		},
	}

	for i, test := range tests {
		scope = test.setupScope()
		stmt, err := parseExpressionStatement(test.buf)
		if stmt != nil && stmt.String() != test.expectedStmt {
			t.Fatalf("test[%d] - TestParseFunctionStatement wrong answer.\n"+
				"Expected= %s\n"+
				"got= %s", i, test.expectedStmt, stmt.String())
		}

		if err != nil && err.Error() != test.expectedErr.Error() {
			t.Fatalf("test[%d] - TestParseFunctionStatement wrong error.\n"+
				"Expected= %s\n"+
				"got= %s", i, test.expectedErr.Error(), err.Error())
		}
	}
}

func TestEnterLeaveScope(t *testing.T) {
	// scope is global variable which defined in parser.go
	scope = symbol.NewScope()
	scope.Set("foo", &symbol.String{Name: &ast.Identifier{Name: "foo"}})

	enterScope()

	scope.Set("bar", &symbol.String{Name: &ast.Identifier{Name: "bar"}})

	if scope.Get("foo") == nil {
		t.Errorf("scope should have foo symbol, because we're in the inner scope")
	}

	leaveScope()

	// test whether inner exist
	inner := scope.GetInner()
	if len(inner) != 1 {
		t.Errorf("scope should have 1 inner scope, but have %d", len(inner))
	}

	if inner[0].Get("bar") == nil {
		t.Errorf("scope should have bar symbol, because we're in the inner scope")
	}

	scope.Set("baz", &symbol.String{Name: &ast.Identifier{Name: "baz"}})

	if scope.Get("bar") != nil {
		t.Errorf("scope should NOT have \"bar\" symbol, because we're in the outer scope")
	}
	if scope.Get("foo") == nil {
		t.Errorf("scope should have \"foo\" symbol, because we're in the outer scope")
	}
	if scope.Get("baz") == nil {
		t.Errorf("scope should have \"baz\" symbol, because we're in the outer scope")
	}
}

func TestUpdateScopeSymbol(t *testing.T) {
	tests := []struct {
		setupScopeFn
		ident       Token
		keyword     Token
		expectedErr error
		chkScope    func(scope *symbol.Scope) bool
	}{
		// test when scope already have identifier symbol
		{
			setupScopeFn: func() *symbol.Scope {
				scope := symbol.NewScope()
				scope.Set("a", &symbol.Integer{Name: &ast.Identifier{Name: "a"}})
				return scope
			},
			ident:       Token{Type: Ident, Val: "a"},
			keyword:     Token{Type: StringType, Val: "string"},
			expectedErr: DupSymError{Token{Type: Ident, Val: "a"}},
			chkScope: func(scope *symbol.Scope) bool {
				return true
			},
		},
		// test when keyword token is IntType, BoolType, StringType, Function, Or Not one of those
		{
			setupScopeFn: defaultSetupScopeFn,
			ident:        Token{Type: Ident, Val: "a"},
			keyword:      Token{Type: IntType, Val: "int"},
			expectedErr:  nil,
			chkScope: func(scope *symbol.Scope) bool {
				sym := scope.Get("a")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.IntegerSymbol {
					return false
				}

				if sym.String() != "a" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			ident:        Token{Type: Ident, Val: "b"},
			keyword:      Token{Type: BoolType, Val: "bool"},
			expectedErr:  nil,
			chkScope: func(scope *symbol.Scope) bool {
				sym := scope.Get("b")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.BooleanSymbol {
					return false
				}

				if sym.String() != "b" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			ident:        Token{Type: Ident, Val: "c"},
			keyword:      Token{Type: StringType, Val: "string"},
			expectedErr:  nil,
			chkScope: func(scope *symbol.Scope) bool {
				sym := scope.Get("c")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.StringSymbol {
					return false
				}

				if sym.String() != "c" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			ident:        Token{Type: Ident, Val: "d"},
			keyword:      Token{Type: Function, Val: "func"},
			expectedErr:  nil,
			chkScope: func(scope *symbol.Scope) bool {
				sym := scope.Get("d")
				if sym == nil {
					return false
				}

				if sym.Type() != symbol.FunctionSymbol {
					return false
				}

				if sym.String() != "d" {
					return false
				}

				return true
			},
		},
		{
			setupScopeFn: defaultSetupScopeFn,
			ident:        Token{Type: Ident, Val: "e"},
			keyword:      Token{Type: Illegal, Val: "illegal"},
			expectedErr: Error{
				Token{Type: Illegal, Val: "illegal"},
				fmt.Sprintf("unexpected symbol type [%v]", TokenTypeMap[Illegal]),
			},
			chkScope: func(scope *symbol.Scope) bool {
				return true
			},
		},
	}

	for i, tt := range tests {
		// setup
		scope = tt.setupScopeFn()

		// exercise
		err := updateScopeSymbol(tt.ident, tt.keyword)

		// verify
		if err != nil && err.Error() != tt.expectedErr.Error() {
			t.Errorf("test[%d] - updateScopeSymbol returns unexpected error, Expected=\"%s\", got=\"%s\"",
				i, err.Error(), tt.expectedErr.Error())
		}
		// verify
		if ok := tt.chkScope(scope); !ok {
			t.Errorf("test[%d] - updateScopeSymbol updates scope incorrectly", i)
		}
	}
}
