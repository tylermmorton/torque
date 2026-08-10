package simple

type Button struct {
	Label string
}

func (*Button) Template() string {
	return `<button>{{.Label}}</button>`
}
