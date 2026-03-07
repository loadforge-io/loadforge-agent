package main

import (
	"context"
	"loadforge-agent/internal/runner"
	"loadforge-agent/internal/scenario"
)

func main() {
	parser := scenario.NewParser()
	if err := parser.ParseFile("scenario.yaml"); err != nil {
		panic(err)
	}
	sc, err := parser.GetScenario()
	if err != nil {
		panic(err)
	}

	r := runner.New(sc)
	if err := r.Run(context.Background()); err != nil {
		panic(err)
	}
}
