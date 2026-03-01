package loglint

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "loglint",
	Doc:  "checks log messages",
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
			if !ok {
				return true
			}

			if !isSupportedLogger(fnName, pkgPath) {
				return true
			}

			if len(call.Args) == 0 {
				return true
			}

			msgExpr := call.Args[0]

			if containsSensitive(pass, msgExpr) {
				pass.Reportf(msgExpr.Pos(), "log message may contain sensitive data")
			}

			msg, ok := extractStringLiteral(msgExpr)
			if !ok {
				return true
			}

			if !startsWithLowercase(msg) {
				pass.Reportf(msgExpr.Pos(), "log message must start with a lowercase letter")
			}

			if !isEnglishOnly(msg) {
				pass.Reportf(msgExpr.Pos(), "log message must be in English only")
			}

			if containsSpecialChars(msg) {
				pass.Reportf(msgExpr.Pos(), "log message must not contain special symbols or emoji")
			}

			return true
		})
	}

	return nil, nil
}

func getCalledFunction(pass *analysis.Pass, call *ast.CallExpr) (string, string, bool) {

	switch fun := call.Fun.(type) {

	case *ast.SelectorExpr:
		obj := pass.TypesInfo.Uses[fun.Sel]
		if obj == nil {
			return "", "", false
		}

		fn, ok := obj.(*types.Func)
		if !ok {
			return "", "", false
		}

		pkg := fn.Pkg()
		if pkg == nil {
			return fn.Name(), "", true
		}

		return fn.Name(), pkg.Path(), true
	}

	return "", "", false
}

func isSupportedLogger(fnName, pkgPath string) bool {

	if pkgPath == "log/slog" {
		return fnName == "Debug" ||
			fnName == "Info" ||
			fnName == "Warn" ||
			fnName == "Error"
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
	return r >= 'a' && r <= 'z'
}

func isEnglishOnly(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == ' ' {
			continue
		}
		return false
	}
	return true
}

func containsSpecialChars(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == ' ' {
			continue
		}
		return true
	}
	return false
}

var sensitiveKeywords = []string{
	"password",
	"token",
	"api_key",
	"apikey",
	"secret",
}

func containsSensitive(pass *analysis.Pass, expr ast.Expr) bool {

	if s, ok := extractStringLiteral(expr); ok {
		low := strings.ToLower(s)
		for _, k := range sensitiveKeywords {
			if strings.Contains(low, k) {
				return true
			}
		}
	}

	found := false

	ast.Inspect(expr, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}

		name := strings.ToLower(id.Name)

		for _, k := range sensitiveKeywords {
			k = strings.ReplaceAll(k, "_", "")
			if strings.Contains(name, k) {
				found = true
				return false
			}
		}

		return true
	})

	return found
}
