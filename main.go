package main

import (
	"html/template"
	"net/http"

	"github.com/MatTwix/Web-Scraper/services"
)

var tpl = template.Must(template.New("reault").Parse(`
<html>
<head>
	<title>Dead Link Checker</title>
</head>
<body>
	<h1>Dead Link Checker</h1>
	<form method="GET">
		URL: <input type="text" name="url" value="{{.URL}}" required>
		<input type="submit" value="Check Links">
	</form>
	{{if .Checked}}
		<h2>Results for URL: {{.URL}}</h2>
		{{if .AliveLinks}}
			<h3>Alive Links:</h3>
		<ul>
			{{range .AliveLinks}}
			<li>{{.}}</li>
			{{end}}
		</ul>
		{{else}}
			<p>No alive links found.</p>
		{{end}}
		{{if .DeadLinks}}
			<h3>Dead Links:</h3>
		<ul>
			{{range .DeadLinks}}
			<li>{{.}}</li>
			{{end}}
		</ul>
		{{else}}
			<p>No dead links found.</p>
		{{end}}
	{{else}}
		<p>Please enter a URL to check for dead links.</p>
	{{end}}
</body>
</html>
`))

func main() {
	// var url string
	// println("Enter the URL to check for dead links: ")
	// fmt.Scanln(&url)
	// if url == "" {
	// 	fmt.Println("No URL provided. Exiting.")
	// 	return
	// }

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		inputURL := r.URL.Query().Get("url")
		var aliveLinks, deadLinks []string
		checked := false

		if inputURL != "" {
			checked = true
			aliveLinks, deadLinks, _ = services.ScrapeLinks(inputURL)
		}

		tpl.Execute(w, map[string]interface{}{
			"URL":        inputURL,
			"AliveLinks": aliveLinks,
			"DeadLinks":  deadLinks,
			"Checked":    checked,
		})
	})

	http.ListenAndServe(":8080", nil)
}
