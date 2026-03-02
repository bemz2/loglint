package main

import (
	"loglint/pkg/loglint"

	"golang.org/x/tools/go/analysis"
)

// New is the golangci-lint Go plugin entrypoint.
func New(conf any) ([]*analysis.Analyzer, error) {
	cfg, err := loglint.ConfigFromAny(conf)
	if err != nil {
		return nil, err
	}

	return []*analysis.Analyzer{loglint.NewAnalyzer(cfg)}, nil
}

func main() {}
