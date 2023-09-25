package compiler

import (
	"go/ast"
	"go/build/constraint"
	"slices"
	"strings"
)

func appendCommentGroup(comments []*ast.Comment, group *ast.CommentGroup) []*ast.Comment {
	if group != nil && len(group.List) > 0 {
		comments = append(comments, group.List...)
	}
	return comments
}

func appendComment(comments []*ast.Comment, text string) []*ast.Comment {
	if len(comments) > 0 {
		comments = append(comments, &ast.Comment{
			Text: "//\n",
		})
	}
	return append(comments, &ast.Comment{
		Text: text,
	})
}

func commentGroupsOf(file *ast.File) []*ast.CommentGroup {
	groups := make([]*ast.CommentGroup, 0, 1+len(file.Comments))
	groups = append(groups, file.Comments...)
	if file.Doc != nil {
		groups = append(groups, file.Doc)
	}
	return groups
}

const gobuild = "//go:build"

func parseBuildTags(file *ast.File) (constraint.Expr, error) {
	for _, group := range commentGroupsOf(file) {
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, gobuild) {
				return constraint.Parse(comment.Text)
			}
		}
	}
	return nil, nil
}

func stripBuildTagsOf(file *ast.File) {
	for _, group := range commentGroupsOf(file) {
		for i, comment := range group.List {
			if strings.HasPrefix(comment.Text, gobuild) {
				group.List = slices.Delete(group.List, i, i+1)
			}
		}
	}
}
