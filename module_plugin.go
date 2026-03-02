package main

import (
	"loglint/pkg/loglint"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

type plugin struct {
	cfg *loglint.Config
}

func init() {
	register.Plugin("loglint", newPlugin)
}

func newPlugin(conf any) (register.LinterPlugin, error) {
	cfg, err := loglint.ConfigFromAny(conf)
	if err != nil {
		return nil, err
	}

	return &plugin{cfg: cfg}, nil
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{loglint.NewAnalyzer(p.cfg)}, nil
}

func (p *plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func main() {}
