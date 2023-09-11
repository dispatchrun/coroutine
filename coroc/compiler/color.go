package compiler

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

type functionColors map[*ssa.Function]*types.Signature

// colorFunctions walks the call graph, coloring functions that yield (or may
// yield) by their yield type. It's an error if a function has more than one
// yield type.
func colorFunctions(cg *callgraph.Graph, yieldInstances functionColors) (functionColors, error) {
	colors := map[*ssa.Function]*types.Signature{}
	for yieldInstance, color := range yieldInstances {
		for _, edge := range cg.Nodes[yieldInstance].In {
			caller := edge.Caller.Func
			if err := colorFunctions0(cg, colors, caller, color); err != nil {
				return nil, err
			}
		}
	}
	return colors, nil
}

func colorFunctions0(cg *callgraph.Graph, colors functionColors, fn *ssa.Function, color *types.Signature) error {
	existing, ok := colors[fn]
	if ok {
		if !types.Identical(existing, color) {
			return fmt.Errorf("function %s has more than one color (%v + %v)", fn, existing, color)
		}
		return nil // already walked
	}
	colors[fn] = color
	for _, edge := range cg.Nodes[fn].In {
		if err := colorFunctions0(cg, colors, edge.Caller.Func, color); err != nil {
			return err
		}
	}
	return nil
}
