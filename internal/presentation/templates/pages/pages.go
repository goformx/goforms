package pages

import (
	"context"
	"io"

	"github.com/a-h/templ"
	"github.com/goformx/goforms/internal/presentation/templates/layouts"
	"github.com/goformx/goforms/internal/presentation/templates/shared"
)

// PagesList renders the pages list view
func PagesList(data shared.PageData) templ.Component {
	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return templ.Raw(`
			<div class="pages-list">
				<div class="header">
					<h1>Pages</h1>
					<a href="/pages/new" class="button">New Page</a>
				</div>
				<div class="grid">
					{{range .Forms}}
					<div class="card">
						<h2>{{.Title}}</h2>
						<p>{{.Description}}</p>
						<div class="actions">
							<a href="/pages/{{.ID}}" class="link">View</a>
							<a href="/pages/{{.ID}}/edit" class="link">Edit</a>
							<button onclick="deletePage('{{.ID}}')" class="link delete">Delete</button>
						</div>
					</div>
					{{end}}
				</div>
			</div>
		`).Render(ctx, w)
	})

	return layouts.Layout(data, content)
}

// PageView renders the page view
func PageView(data shared.PageData) templ.Component {
	content := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		return templ.Raw(`
			<div class="page-view">
				<div class="card">
					<h1>{{.Form.Title}}</h1>
					<p>{{.Form.Description}}</p>
					<div id="form-builder" data-schema="{{.Form.Schema}}"></div>
				</div>
			</div>
		`).Render(ctx, w)
	})

	return layouts.Layout(data, content)
}
