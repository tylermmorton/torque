package html

type LinkTag struct {
	Rel  string
	Href string
}

func (LinkTag) Template() string {
	//language=html
	return `<link rel="{{.Rel}}" href="{{.Href}}"/>`
}
