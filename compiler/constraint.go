package compiler

import (
	"go/build/constraint"
	"reflect"
)

func containsExpr(expr, contains constraint.Expr) bool {
	switch x := expr.(type) {
	case *constraint.AndExpr:
		return containsExpr(x.X, contains) || containsExpr(x.Y, contains)
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
