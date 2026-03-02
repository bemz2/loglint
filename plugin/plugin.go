package main

import (
	"loglint/pkg/loglint"

	"golang.org/x/tools/go/analysis"
)

// New is the golangci-lint Go plugin entrypoint.
func New(_ any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{loglint.Analyzer}, nil
}

func main() {}
