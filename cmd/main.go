package main

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/go-seatbelt/seatbelt"

	"github.com/alecthomas/chroma"
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
				"highlight": highlight,
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

// highlight adds syntax highlighting to the given string.
//
// TODO Cache highlight results so that we don't have to precompile them.
func highlight(lang, s string) template.HTML {
	lexer := lexers.Get(lang)

	style := registerSeatbeltStyle()

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
	highlighted = strings.TrimPrefix(highlighted, "<html>\n<body style=\"\">\n")
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

func registerSeatbeltStyle() *chroma.Style {
	const (
		black = "#0f172a"
		blue  = "#0284c7"
		pink  = "#ec4899"
	)
	return styles.Register(chroma.MustNewStyle(
		"seatbelt",
		chroma.StyleEntries{
			chroma.Error: "#bf616a",
			// chroma.Background:            "#ffffff bg:#ffffff",
			chroma.Keyword:               blue,
			chroma.KeywordPseudo:         blue,
			chroma.KeywordType:           blue,
			chroma.Name:                  black,
			chroma.NameAttribute:         blue,
			chroma.NameBuiltin:           black,
			chroma.NameClass:             blue,
			chroma.NameConstant:          blue,
			chroma.NameDecorator:         "#d08770",
			chroma.NameEntity:            "#d08770",
			chroma.NameException:         "#bf616a",
			chroma.NameFunction:          blue,
			chroma.NameLabel:             blue,
			chroma.NameNamespace:         blue,
			chroma.NameTag:               black,
			chroma.NameVariable:          black,
			chroma.LiteralString:         pink,
			chroma.LiteralStringDoc:      "#616e87",
			chroma.LiteralStringEscape:   "#ebcb8b",
			chroma.LiteralStringInterpol: pink,
			chroma.LiteralStringOther:    pink,
			chroma.LiteralStringRegex:    "#ebcb8b",
			chroma.LiteralStringSymbol:   pink,
			chroma.LiteralNumber:         pink,
			chroma.Operator:              black,
			chroma.OperatorWord:          black,
			chroma.Punctuation:           black,
			chroma.Comment:               "italic #616e87",
			chroma.CommentPreproc:        black,
			chroma.GenericDeleted:        "#bf616a",
			chroma.GenericEmph:           "italic",
			chroma.GenericError:          "#bf616a",
			chroma.GenericHeading:        "bold #0284c7",
			chroma.GenericInserted:       pink,
			chroma.GenericOutput:         black,
			chroma.GenericPrompt:         "bold #4c566a",
			chroma.GenericStrong:         "bold",
			chroma.GenericSubheading:     "bold #0284c7",
			chroma.GenericTraceback:      "#bf616a",
			chroma.TextWhitespace:        black,
		},
	))
}
