package compiler

import "golang.org/x/tools/go/packages"

// flattenPackages walks the package import graph to produce the set of all
// packages in the graph. The order of packages is currently unstable.
func flattenPackages(pp []*packages.Package) (flattened []*packages.Package) {
	seen := map[*packages.Package]struct{}{}
	for _, p := range pp {
		flattenPackages0(seen, p)
	}
	flattened = make([]*packages.Package, len(seen))
	i := 0
	for p := range seen {
		flattened[i] = p
		i++
	}
	return
}

func flattenPackages0(seen map[*packages.Package]struct{}, p *packages.Package) {
	if _, ok := seen[p]; ok {
		return
	}
	seen[p] = struct{}{}
	for _, child := range p.Imports {
		flattenPackages0(seen, child)
	}
}
