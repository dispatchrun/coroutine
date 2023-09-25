package compiler

import (
	"go/ast"
	"go/build/constraint"
	"reflect"
	"slices"
)

func containsExpr(expr, contains constraint.Expr) bool {
	switch x := expr.(type) {
	case *constraint.AndExpr:
		return containsExpr(x.X, contains) || containsExpr(x.Y, contains)
	case *constraint.OrExpr:
		return containsExpr(x.X, contains) && containsExpr(x.Y, contains)
	default:
		return reflect.DeepEqual(expr, contains)
	}
}

func withBuildTag(expr constraint.Expr, buildTag *constraint.TagExpr) constraint.Expr {
	if buildTag == nil || containsExpr(expr, buildTag) {
		return expr
	} else if expr == nil {
		return buildTag
	} else {
		return &constraint.AndExpr{X: expr, Y: buildTag}
	}
}

func withoutBuildTag(expr constraint.Expr, buildTag *constraint.TagExpr) constraint.Expr {
	notBuildTag := &constraint.NotExpr{X: buildTag}
	if buildTag == nil || containsExpr(expr, notBuildTag) {
		return expr
	} else if expr == nil {
		return notBuildTag
	} else {
		return &constraint.AndExpr{X: expr, Y: notBuildTag}
	}
}

func parseBuildTags(file *ast.File) (constraint.Expr, error) {
	groups := commentGroupsOf(file)

	for _, group := range groups {
		for _, c := range group.List {
			if constraint.IsGoBuild(c.Text) {
				return constraint.Parse(c.Text)
			}
		}
	}

	var plusBuildLines constraint.Expr
	for _, group := range groups {
		for _, c := range group.List {
			if constraint.IsPlusBuild(c.Text) {
				x, err := constraint.Parse(c.Text)
				if err != nil {
					return nil, err
				}
				if plusBuildLines == nil {
					plusBuildLines = x
				} else {
					plusBuildLines = &constraint.AndExpr{X: plusBuildLines, Y: x}
				}
			}
		}
	}

	return plusBuildLines, nil
}

func stripBuildTagsOf(file *ast.File, path string) {
	for _, group := range commentGroupsOf(file) {
		group.List = slices.DeleteFunc(group.List, func(c *ast.Comment) bool {
			return constraint.IsGoBuild(c.Text) || constraint.IsPlusBuild(c.Text)
		})
	}
}
