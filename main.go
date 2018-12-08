package main

import (
	"bytes"
	"fmt"
	"html/template"
    "io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/russross/blackfriday/v2"
)

const (
	// postsRoot is the path to markdown posts on disk.
	postsRoot = "posts"
	// draftsRoot is the path to markdown drafts on disk.
	draftsRoot = "drafs"
	// templateRoot is the path to page templates on disk.
	templateRoot = "templates"
	// wwwRoot is the path to output rendered pages on disk.
	wwwRoot = "www"

	// dateLayout is the expected format of post names, prefixed with a date.
	dateLayout = "2006-01-02-"
)

var (
	// templateFuncs are custom display functions for templates
	templateFuncs = map[string]interface{}{ "date": uiDate }

	// tmplPage is the standard page template. Contains the header and footer and
	// embeds results of content rendering.
	tmplPage = template.Must(loadTemplate("page.html"))
)

type Page struct {
	// Path of the page on the site
	Path string
	// Title of the page
	Title string
	// Content is the rendered HTML string
	Content template.HTML
}

// Post is a parsed markdown document for use on the site.
type Post struct {
	// Path of the post on the site
	Path string
	// Title of the post, extracted from the first heading.
	Title string
	// Date of the post, extracted from the path prefix.
	Date time.Time
	// Doc is the parsed markdown to be rendered.
	Doc *blackfriday.Node
}

// LoadPost reads the markdown file at path and generates a Post. The basename
// must be prefixed with a date followed by the title.
//
//  "path/to/file/2006-01-02-example-title.md"
func LoadPost(path string) (*Post, error) {
	fmt.Println("loading", path)

	// TODO: make this more robust to bad names, etc.
	base := filepath.Base(path)
	date, err := time.Parse(dateLayout, base[:len(dateLayout)])
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
    exts := blackfriday.CommonExtensions | blackfriday.Footnotes
	doc := blackfriday.New(blackfriday.WithExtensions(exts)).Parse(buf)

	// Calculate the target path
	name := strings.TrimPrefix(path, postsRoot)
	name = strings.TrimSuffix(name, filepath.Ext(name))

	// Pull title from the initial header
	heading := doc.FirstChild
	if heading == nil || heading.Type != blackfriday.Heading {
		return nil, fmt.Errorf("%s: invalid title: %q", path, heading)
	}
	title := heading.FirstChild
	if title == nil || len(title.Literal) == 0 {
		return nil, fmt.Errorf("%s: empty title", path)
	}

	return &Post{
		Path:  name,
		Title: string(title.Literal),
		Date:  date,
		Doc:   doc,
	}, nil
}

// loadPosts reads markdown files under root and converts them to posts. The
// returned slice is in reverse chronological order.
func loadPosts(root string) ([]*Post, error) {
	posts := make([]*Post, 0)
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		post, err := LoadPost(p)
		if err != nil {
			return err
		}
		posts = append(posts, post)
		return nil
	})
	sort.Slice(posts, func(i, j int) bool {
		return posts[j].Date.Before(posts[i].Date)
	})
	return posts, err
}

// loadTemplate parses a template with the given name on disk, relative to
// the templateRoot.
func loadTemplate(name string) (*template.Template, error) {
	path := filepath.Join(templateRoot, name)
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Have to install functions before parsing templates that use them, hence
	// can't use ParseFiles(...) directly here and install funcs after.
	t := template.New(name).Funcs(templateFuncs)
	return t.Parse(string(buf))
}

// writePage renders a new page on disk relative to wwwRoot.
func writePage(page *Page) error {
	target := filepath.Join(wwwRoot, page.Path)
	fmt.Println("writing", target)
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	return tmplPage.Execute(f, page)
}

// renderIndex generates the index.html for the site.
func renderIndex(posts []*Post) error {
	t, err := loadTemplate("index.html")
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, posts); err != nil {
		return err
	}
	return writePage(&Page{
		Path:    "index.html",
		Title:   "neilpa.me",
		Content: template.HTML(buf.String()),
	})
}

func renderTree(w io.Writer, renderer blackfriday.Renderer, node *blackfriday.Node) {
    node.Walk(func(node *blackfriday.Node, enterring bool) blackfriday.WalkStatus {
        return renderer.RenderNode(w, node, enterring)
    })
}

// renderPosts generates all the post pages for the site.
func renderPosts(posts []*Post) error {
	params := blackfriday.HTMLRendererParameters{
		HeadingLevelOffset: 1,
		Flags: blackfriday.CommonHTMLFlags | blackfriday.FootnoteReturnLinks,
	}
	renderer := blackfriday.NewHTMLRenderer(params)

	for _, p := range posts {
		var buf bytes.Buffer
        // HACK: simplest way to inject the date for now, h3 assumes level offest
        title := p.Doc.FirstChild
        renderTree(&buf, renderer, title)
        fmt.Fprintf(&buf, "<h3 class=\"published\"><time datetime=%q>%s</time></h3>\n\n",
            p.Date.Format(time.RFC3339), uiDate(p.Date))
        // remainder of the document
        title.Unlink()
        renderTree(&buf, renderer, p.Doc)

		err := writePage(&Page{
			Path:    p.Path,
			Title:   p.Title,
			Content: template.HTML(buf.String()),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	posts, err := loadPosts(postsRoot)
	if err != nil {
		log.Fatal(err)
	}
	if err := renderIndex(posts); err != nil {
		log.Fatal(err)
	}
	if err := renderPosts(posts); err != nil {
		log.Fatal(err)
	}
}

func uiDate(t time.Time) string {
	return t.Format("2006/01/02")
}

