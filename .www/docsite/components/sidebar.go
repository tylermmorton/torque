package components

type SidebarLink struct {
	Label  string
	Href   string
	Active bool
}

type SidebarSection struct {
	Title string
	Links []SidebarLink
}

type Sidebar struct {
	Sections []SidebarSection
}

func NewSidebar(currentPath string) Sidebar {
	type linkDef struct {
		label string
		slug  string
	}
	sections := []struct {
		title string
		links []linkDef
	}{
		{
			title: "Introduction",
			links: []linkDef{
				{"Getting Started", "getting-started"},
				{"Component", "component"},
			},
		},
		{
			title: "Core Concepts",
			links: []linkDef{
				{"Router", "router"},
				{"Middleware", "middleware"},
				{"Layouts", "layout"},
				{"Outlets", "outlet"},
				{"Context Provider", "context-provider"},
			},
		},
		{
			title: "Templates",
			links: []linkDef{
				{"Template API", "template-api"},
				{"Template Provider", "template-provider"},
				{"Stylesheet Provider", "stylesheet-provider"},
				{"Page Layout", "page-layout"},
			},
		},
		{
			title: "Data",
			links: []linkDef{
				{"Loader", "loader"},
				{"Path Params", "path-params"},
				{"Query Params", "query-params"},
				{"Forms", "forms"},
			},
		},
		{
			title: "Advanced",
			links: []linkDef{
				{"Renderer", "renderer"},
				{"Error Handling", "errors"},
				{"Composability", "composability"},
			},
		},
	}

	sidebar := Sidebar{}
	for _, sec := range sections {
		section := SidebarSection{Title: sec.title}
		for _, l := range sec.links {
			href := "/docs/" + l.slug
			section.Links = append(section.Links, SidebarLink{
				Label:  l.label,
				Href:   href,
				Active: currentPath == href,
			})
		}
		sidebar.Sections = append(sidebar.Sections, section)
	}
	return sidebar
}
