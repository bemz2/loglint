package loglint

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = NewAnalyzer(nil)

func NewAnalyzer(override *Config) *analysis.Analyzer {
	analyzer := &analysis.Analyzer{
		Name: "loglint",
		Doc:  "checks slog and zap log messages against style rules",
	}

	analyzer.Flags.String("config", "", "path to a loglint JSON config file")
	analyzer.Run = func(pass *analysis.Pass) (any, error) {
		return run(pass, analyzer, override)
	}

	return analyzer
}

func run(pass *analysis.Pass, analyzer *analysis.Analyzer, override *Config) (any, error) {
	cfg, err := resolveConfig(analyzer, override)
	if err != nil {
		return nil, err
	}

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

			if cfg.CheckSensitiveData && containsSensitiveCall(call, cfg.SensitiveKeywords) {
				pass.Reportf(call.Pos(), "log message may contain sensitive data")
			}

			msg, ok := extractStringLiteral(call.Args[0])
			if !ok {
				return true
			}

			if cfg.CheckLowercaseStart && !startsWithLowercase(msg) {
				pass.Reportf(call.Args[0].Pos(), "log message must start with a lowercase letter")
			}

			if cfg.CheckEnglishOnly && !isEnglishOnly(msg) {
				pass.Reportf(call.Args[0].Pos(), "log message must be in English only")
			}

			if cfg.CheckSpecialChars && containsSpecialChars(msg) {
				pass.Reportf(call.Args[0].Pos(), "log message must not contain special symbols or emoji")
			}

			return true
		})
	}
	return nil, nil
}

func resolveConfig(analyzer *analysis.Analyzer, override *Config) (effectiveConfig, error) {
	cfg := defaultConfig()
	if analyzer == nil {
		return mergeConfig(cfg, override), nil
	}

	pathFlag := analyzer.Flags.Lookup("config")
	if pathFlag == nil || pathFlag.Value.String() == "" {
		return mergeConfig(cfg, override), nil
	}

	fileConfig, err := loadConfigFile(pathFlag.Value.String())
	if err != nil {
		return effectiveConfig{}, fmt.Errorf("load loglint config: %w", err)
	}

	return mergeConfig(cfg, fileConfig, override), nil
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

	if r >= 'a' && r <= 'z' {
		return true
	}

	if r >= 'A' && r <= 'Z' {
		return false
	}

	return true
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

func containsSensitiveCall(call *ast.CallExpr, sensitiveKeywords []string) bool {
	for _, arg := range call.Args {
		found := false
		ast.Inspect(arg, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				name := strings.ToLower(id.Name)
				for _, k := range sensitiveKeywords {
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
		for _, k := range sensitiveKeywords {
			if key == k {
				return true
			}
		}
	}

	return false
}
