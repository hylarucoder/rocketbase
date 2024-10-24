package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hylarucoder/rocketbase/forms"
	"github.com/hylarucoder/rocketbase/tools/filesystem"
	"golang.org/x/net/html"

	"github.com/joho/godotenv"

	"github.com/gosimple/slug"
	rocketbase "github.com/hylarucoder/rocketbase"
	"github.com/hylarucoder/rocketbase/core"
	"github.com/hylarucoder/rocketbase/models"
	"github.com/hylarucoder/rocketbase/plugins/migratecmd"
)

type LinkPreview struct {
	Title       string
	Description string
	ImageURL    string
	FaviconURL  string
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func resolveURL(base, ref string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ref
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	resolvedURL := baseURL.ResolveReference(refURL)
	return resolvedURL.String()
}

func ExtractLinkPreview(urlStr string) (*LinkPreview, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	preview := &LinkPreview{}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil {
					preview.Title = n.FirstChild.Data
				}
			case "meta":
				name := getAttr(n, "name")
				property := getAttr(n, "property")
				content := getAttr(n, "content")

				switch {
				case name == "description":
					preview.Description = content
				case property == "og:description":
					preview.Description = content
				case property == "og:image":
					preview.ImageURL = resolveURL(urlStr, content)
				}
			case "link":
				rel := getAttr(n, "rel")
				href := getAttr(n, "href")

				if rel == "icon" || rel == "shortcut icon" {
					preview.FaviconURL = resolveURL(urlStr, href)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	// If no favicon found in HTML, try the default location
	if preview.FaviconURL == "" {
		defaultFaviconURL := resolveURL(urlStr, "/favicon.ico")
		resp, err := http.Head(defaultFaviconURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			preview.FaviconURL = defaultFaviconURL
		}
	}
	// if https://www.notion.sohttps://, then replace with https://
	if strings.HasPrefix(preview.ImageURL, "https://www.notion.sohttps://") {
		preview.ImageURL = strings.Replace(preview.ImageURL, "https://www.notion.sohttps://", "https://", 1)
	}

	return preview, nil
}

func slugify(name string) string {
	// remove all special characters
	return slug.Make(name)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	app := rocketbase.New()

	// ---------------------------------------------------------------
	// Optional plugin flags:
	// ---------------------------------------------------------------
	var automigrate bool
	app.RootCmd.PersistentFlags().BoolVar(
		&automigrate,
		"automigrate",
		false,
		"enable/disable auto migrations",
	)

	var queryTimeout int
	app.RootCmd.PersistentFlags().IntVar(
		&queryTimeout,
		"queryTimeout",
		30,
		"the default SELECT queries timeout in seconds",
	)

	app.RootCmd.ParseFlags(os.Args[1:])

	// ---------------------------------------------------------------
	// Plugins and hooks:
	// ---------------------------------------------------------------

	// migrate command (with js templates)
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  automigrate,
		// Dir:          migrationsDir,
	})

	// // GitHub selfupdate
	// ghupdate.MustRegister(app, app.RootCmd, ghupdate.Config{})

	app.OnAfterBootstrap().PreAdd(func(e *core.BootstrapEvent) error {
		app.Dao().ModelQueryTimeout = time.Duration(queryTimeout) * time.Second
		return nil
	})

	// app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
	// 	// serves static files from the provided public dir (if exists)
	// 	e.Router.GET("/*", apis.StaticDirectoryHandler(os.DirFS(publicDir), indexFallback))
	// 	return nil
	// })

	app.OnModelBeforeCreate("post_category", "post_tag").Add(func(e *core.ModelEvent) error {
		rec := e.Model.(*models.Record)

		if rec.GetString("slug") == "" {
			rec.Set("slug", slugify(rec.GetString("name")))
		}

		return nil
	})
	app.OnModelBeforeUpdate("post_category", "post_tag").Add(func(e *core.ModelEvent) error {
		rec := e.Model.(*models.Record)

		if rec.GetString("slug") == "" {
			rec.Set("slug", slugify(rec.GetString("name")))
		}

		return nil
	})

	app.OnModelBeforeCreate("post").Add(func(e *core.ModelEvent) error {
		rec := e.Model.(*models.Record)

		if rec.GetString("slug") == "" {
			rec.Set("slug", slugify(rec.GetString("title")))
		}

		return nil
	})
	app.OnModelBeforeUpdate("post").Add(func(e *core.ModelEvent) error {
		rec := e.Model.(*models.Record)

		if rec.GetString("slug") == "" {
			rec.Set("slug", slugify(rec.GetString("title")))
		}

		return nil
	})

	app.OnModelBeforeCreate("post_link_preview").Add(func(e *core.ModelEvent) error {
		rec := e.Model.(*models.Record)

		linkPreview, _ := ExtractLinkPreview(rec.GetString("url"))
		rec.Set("title", linkPreview.Title)
		rec.Set("description", linkPreview.Description)
		rec.Set("image_src", linkPreview.ImageURL)

		return nil
	})

	app.OnRecordAfterCreateRequest().Add(func(e *core.RecordCreateEvent) error {
		collection := e.Record.Collection()
		if collection.Name != "post_link_preview" {
			return nil
		}

		imageURL := e.Record.GetString("image_src")
		if imageURL == "" {
			return nil
		}

		// Create a new form for updating the record
		form := forms.NewRecordUpsert(app, e.Record)

		// Use PocketBase's built-in function to create a file from URL
		file, err := filesystem.NewFileFromUrl(e.HttpContext.Request().Context(), imageURL)
		if err != nil {
			app.Logger().Error("failed to create file from URL", "error", err)
			return nil
		}

		// Add the file to the form
		form.AddFiles("image", file)

		// Submit the form to update the record
		if err := form.Submit(); err != nil {
			app.Logger().Error("failed to save imageUrl -> image", "error", err)
			return nil
		}

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

}
