package main

import (
	"bytes"
	"hash/fnv"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-seatbelt/seatbelt"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"golang.org/x/exp/slog"
)

// hl is the global highlighter.
var hl = &highlighter{
	pool: &sync.Pool{
		New: func() any {
			return &bytes.Buffer{}
		},
	},
	data:  make(map[uint64]template.HTML),
	style: registerSeatbeltStyle(),
	formatter: html.New(
		html.Standalone(true),
		html.PreventSurroundingPre(true),
	),
}

// logger is logging HTTP middleware.
func logger(fn func(_ *seatbelt.Context) error) func(*seatbelt.Context) error {
	return func(c *seatbelt.Context) error {
		now := time.Now()
		defer func() {
			dur := time.Since(now)
			slog.Info("received request",
				slog.Duration("dur", dur),
				slog.String("path", c.Request().URL.Path),
			)
		}()

		return fn(c)
	}
}

func main() {
	log.Println("Starting server...")
	reload := os.Getenv("ENV") != "production"

	app := seatbelt.New(seatbelt.Option{
		TemplateDir: "templates",
		Reload:      reload,
		Funcs: func(w http.ResponseWriter, r *http.Request) template.FuncMap {
			return template.FuncMap{
				"CurrentPageClasses": func(path, active, inactive string) string {
					if strings.Contains(r.URL.Path, path) {
						return active
					}
					return inactive
				},
				"highlight":       hl.highlight,
				"highlightinline": hl.highlightInline,
			}
		},
	})

	app.Use(logger)

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

	log.Printf("Started the server on http://localhost:3000 with reload=%v", reload)
	srv := &http.Server{
		Addr:           ":3000",
		Handler:        app,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
	}
	log.Fatalln(srv.ListenAndServe())
}

type highlighter struct {
	mu        sync.Mutex
	pool      *sync.Pool
	data      map[uint64]template.HTML
	style     *chroma.Style
	formatter chroma.Formatter
}

func (h *highlighter) read(n uint64) (template.HTML, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	result, ok := h.data[n]
	return result, ok
}

func (h *highlighter) write(n uint64, data template.HTML) {
	h.mu.Lock()
	h.data[n] = data
	h.mu.Unlock()
}

// hash computes the hash of the given string.
func (h *highlighter) hash(s string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(s))
	return hash.Sum64()
}

// highlight adds syntax highlighting to the given string.
func (h *highlighter) highlight(lang, s string) template.HTML {
	// Check the cache first to see if there's a stored result.
	if result, ok := h.read(h.hash(s)); ok {
		return result
	}

	lexer := lexers.Get(lang)

	iterator, err := lexer.Tokenise(nil, s)
	if err != nil {
		return template.HTML(err.Error())
	}

	buf := h.pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer h.pool.Put(buf)

	if err := h.formatter.Format(buf, h.style, iterator); err != nil {
		return template.HTML(err.Error())
	}

	highlighted := buf.String()
	// Trim leading and trailing html and body tags that cause rendering
	// errors.
	highlighted = highlighted[23 : len(highlighted)-16]
	highlighted = strings.TrimSpace(highlighted)

	// Save the computed highlighted code in the cache.
	h.write(h.hash(s), template.HTML(highlighted))
	return template.HTML(highlighted)
}

func (h *highlighter) highlightInline(lang, s string) template.HTML {
	const open = `<code class="inline-block bg-slate-100 px-1 rounded-md text-sm sm:text-base">`
	const close = `</code>`
	return open + h.highlight(lang, s) + close
}

var samples = map[string]interface{}{
	"QuickstartGo": hl.highlight("go", `package main

import "github.com/go-seatbelt/seatbelt"

func main() {
    app := seatbelt.New()
    app.Get("/", func(c *seatbelt.Context) error {
    	return c.String(200, "Hello, world!")
	})
    app.Start(":3000")
}`),

	"RenderGo": hl.highlight("go", `package main

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

	"RenderHTML": hl.highlight("go html template", `<!DOCTYPE html>
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

	"SessionGo": hl.highlight("go", `package main

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

	"SessionHTML": hl.highlight("go html template", `<!DOCTYPE html>
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
