package compiler

import "go/ast"

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
