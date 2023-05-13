package templates

import "github.com/tylermmorton/torque/www/docsite/model"

// TODO(tmpl) change after binder utility refactor
//go:generate tmplbind

// ArticleView is the dot context of the article view
//
//tmpl:bind article.tmpl.html --watch
type ArticleView struct {
	Article *model.Article
}
