package loglint

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "loglint",
	Doc:  "checks slog and zap log messages against style rules",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			fnName, pkgPath, ok := getCalledFunction(pass, call)
			if !ok || !isSupportedLogger(fnName, pkgPath) {
				return true
			}

			if len(call.Args) == 0 {
				return true
			}

			// RULE 4 — Sensitive data (анализируем весь вызов)
			if containsSensitiveCall(call) {
				pass.Reportf(call.Pos(), "log message may contain sensitive data")
			}

			// RULE 1–3 — только если первый аргумент строка
			msg, ok := extractStringLiteral(call.Args[0])
			if !ok {
				return true
			}

			if !startsWithLowercase(msg) {
				pass.Reportf(call.Args[0].Pos(), "log message must start with a lowercase letter")
			}

			if !isEnglishOnly(msg) {
				pass.Reportf(call.Args[0].Pos(), "log message must be in English only")
			}

			if containsSpecialChars(msg) {
				pass.Reportf(call.Args[0].Pos(), "log message must not contain special symbols or emoji")
			}

			return true
		})
	}
	return nil, nil
}

func getCalledFunction(pass *analysis.Pass, call *ast.CallExpr) (string, string, bool) {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		obj := pass.TypesInfo.Uses[sel.Sel]
		if obj == nil {
			return "", "", false
		}
		fn, ok := obj.(*types.Func)
		if !ok {
			return "", "", false
		}
		if fn.Pkg() == nil {
			return fn.Name(), "", true
		}
		return fn.Name(), fn.Pkg().Path(), true
	}
	return "", "", false
}

func isSupportedLogger(fnName, pkgPath string) bool {
	if pkgPath == "log/slog" {
		switch fnName {
		case "Debug", "Info", "Warn", "Error":
			return true
		}
	}

	if pkgPath == "go.uber.org/zap" {
		switch fnName {
		case "Debug", "Info", "Warn", "Error",
			"Debugf", "Infof", "Warnf", "Errorf",
			"Debugw", "Infow", "Warnw", "Errorw":
			return true
		}
	}

	return false
}

func extractStringLiteral(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return s, true
}

func startsWithLowercase(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}

	r := []rune(s)[0]

	if !unicode.IsLetter(r) {
		return true
	}

	return r >= 'a' && r <= 'z'
}

func isEnglishOnly(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z') {
				return false
			}
		}
	}
	return true
}

func containsSpecialChars(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) ||
			unicode.IsDigit(r) ||
			r == ' ' ||
			r == '_' ||
			r == '-' ||
			r == '%' {
			continue
		}
		return true
	}
	return false
}

var sensitiveIdentifiers = []string{
	"password",
	"passwd",
	"pwd",
	"apikey",
	"secret",
	"token",
}

func containsSensitiveCall(call *ast.CallExpr) bool {
	for _, arg := range call.Args {
		found := false
		ast.Inspect(arg, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				name := strings.ToLower(id.Name)
				for _, k := range sensitiveIdentifiers {
					if strings.Contains(name, k) {
						found = true
						return false
					}
				}
			}
			return true
		})
		if found {
			return true
		}
	}

	// structured logging key-value pairs
	for i := 1; i+1 < len(call.Args); i += 2 {
		key, ok := extractStringLiteral(call.Args[i])
		if !ok {
			continue
		}
		key = strings.ToLower(key)
		for _, k := range sensitiveIdentifiers {
			if key == k {
				return true
			}
		}
	}

	return false
}
