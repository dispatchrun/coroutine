package compiler

import (
	"fmt"
	"go/types"
	"strings"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/ssa"
)

// colorFunctions walks the call graph, coloring functions that yield (or may
// yield) by their yield type. It's an error if a function has more than one
// yield type.
func (c *compiler) colorFunctions(cg *callgraph.Graph, yieldInstances functionColors) (functionColors, error) {
	colors := map[*ssa.Function]*types.Signature{}
	for yieldInstance, color := range yieldInstances {
		if c.debugColors {
			fmt.Println("[color] scanning root", yieldInstance, "with color:", color)
		}
		if err := c.colorFunctions0(cg, colors, yieldInstance, color, 1); err != nil {
			return nil, err
		}
		if c.debugColors {
			fmt.Println("[color]")
		}
	}
	return colors, nil
}

type functionColors map[*ssa.Function]*types.Signature

func (c *compiler) colorFunctions0(cg *callgraph.Graph, colors functionColors, fn *ssa.Function, color *types.Signature, depth int) error {
	var prevCaller *ssa.Function
	for _, edge := range cg.Nodes[fn].In {
		caller := edge.Caller.Func
		if caller == prevCaller {
			continue
		}
		if err := c.colorFunctions1(cg, colors, edge.Caller.Func, color, depth); err != nil {
			return err
		}
		prevCaller = caller
	}
	return nil
}

func (c *compiler) colorFunctions1(cg *callgraph.Graph, colors functionColors, fn *ssa.Function, color *types.Signature, depth int) error {
	if c.debugColors {
		fmt.Println("[color] ", strings.Repeat("  ", depth-1), "<~", fn)
	}
	if origin := fn.Origin(); origin != nil && origin.Pkg != nil {
		// Don't follow edges into and through the coroutine package.
		if pkgPath := origin.Pkg.Pkg.Path(); pkgPath == coroutinePackage {
			return nil
		}
	}

	existing, ok := colors[fn]
	if ok {
		if !types.Identical(existing, color) {
			return fmt.Errorf("function %s has more than one color (%v + %v)", fn, existing, color)
		}
		return nil // already walked
	}

	colors[fn] = color
	return c.colorFunctions0(cg, colors, fn, color, depth+1)
}
