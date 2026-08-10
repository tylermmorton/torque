package simple

type Card struct {
	Title  string
	Action Button `template:"card-action"`
}

func (*Card) Template() string {
	return `<div class="card"><h2>{{.Title}}</h2>{{template "card-action" .Action}}</div>`
}

func (*Card) StyleSheet() string {
	return `.card { border: 1px solid #ccc; padding: 16px; }`
}
