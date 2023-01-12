package main

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/go-seatbelt/seatbelt"

	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

func main() {
	app := seatbelt.New(seatbelt.Option{
		TemplateDir: "templates",
		Reload:      os.Getenv("ENV") != "production",
		Funcs: func(w http.ResponseWriter, r *http.Request) template.FuncMap {
			return template.FuncMap{
				"CurrentPageClasses": func(path, active, inactive string) string {
					if strings.Contains(r.URL.Path, path) {
						return active
					}
					return inactive
				},
			}
		},
	})

	app.Get("/", func(c *seatbelt.Context) error {
		data := make(map[string]interface{})
		for k, v := range samples {
			data[k] = v
		}
		name := c.Session.Get("name")
		data["Name"] = name
		return c.Render("index", data)
	})
	app.Post("/", func(c *seatbelt.Context) error {
		c.Session.Set("name", c.FormValue("name"))
		return c.Redirect("/")
	})
	app.Get("/guide", func(c *seatbelt.Context) error {
		return c.Render("guide", nil)
	})
	app.Get("/api", func(c *seatbelt.Context) error {
		return c.Render("api", nil)
	})
	app.Start(":3000")
}

func highlight(lang, s string) template.HTML {
	lexer := lexers.Get(lang)

	style := styles.Get("nord")
	if style == nil {
		style = styles.Fallback
	}

	formatter := html.New(
		html.Standalone(true),
		html.PreventSurroundingPre(true),
	)

	iterator, err := lexer.Tokenise(nil, s)
	if err != nil {
		return template.HTML(err.Error())
	}

	buf := &bytes.Buffer{}
	if err := formatter.Format(buf, style, iterator); err != nil {
		return template.HTML(err.Error())
	}

	highlighted := buf.String()
	highlighted = strings.TrimPrefix(highlighted, "<html>\n<body style=\"color:#d8dee9;background-color:#2e3440\">")
	highlighted = strings.TrimSuffix(highlighted, "</body>\n</html>\n")
	highlighted = strings.TrimSpace(highlighted)

	return template.HTML(highlighted)
}

var samples = map[string]interface{}{
	"QuickstartGo": highlight("go", `package main

import "github.com/go-seatbelt/seatbelt"

func main() {
    app := seatbelt.New()
    app.Get("/", func(c *seatbelt.Context) error {
    	return c.String(200, "Hello, world!")
	})
    app.Start(":3000")
}`),

	"RenderGo": highlight("go", `package main

import "github.com/go-seatbelt/seatbelt"

func main() {
    app := seatbelt.New()
    app.Get("/", func(c *seatbelt.Context) error {
    	return c.Render("index", map[string]any{
			"Message": "Hello, world!",
		})
	})
    app.Start(":3000")
}`),

	"RenderHTML": highlight("go html template", `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Hello, world!</title>
</head>
<body>
  <h1>{{ .Message }}</h1>
</body>
</html>`),

	"SessionGo": highlight("go", `package main

import "github.com/go-seatbelt/seatbelt"

func main() {
	app := seatbelt.New()
	app.Get("/", func(c *seatbelt.Context) error {
		return c.Render("index", map[string]any{
			"Name": c.Session.Get("name"),
		})
	})
	app.Post("/", func(c *seatbelt.Context) error {
		c.Session.Set("name", c.FormValue("name"))
		return c.Redirect("/")
	})
	app.Start(":3000")
}`),

	"SessionHTML": highlight("go html template", `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Sessions</title>
</head>
<body>
  <h1>Current session value: {{ .Name }}</h1>
  <form method="POST" action="/">
    {{ csrf }}
    <label>Set new session value:</label>
    <input type="text" name="name"/>
    <input type="submit" value="Submit"/>
  </form>
</body>
</html>`),
}
