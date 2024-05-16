package main

import (
	"flag"
	"fmt"
	htmlt "html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	textt "text/template"

	"github.com/npmania/scut"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: sscut file.json\n")
	os.Exit(2)
}

var (
	postTmpl *htmlt.Template
	osTmpl   *textt.Template
	scuts    map[string]scut.Scut
	prefix   = "."
	defsc    = "DEFAULT"
	hostname = "http://localhost:8080"
)

func main() {
	log.SetPrefix("sscut: ")
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scuts, err = scut.Load(f)
	if err != nil {
		log.Fatal(err)
	}
	postTmpl, err = htmlt.New("post.html").ParseFiles("post.html")
	if err != nil {
		log.Fatal(err)
	}
	osTmpl, err = textt.New("opensearch.xml").ParseFiles("opensearch.xml")
	if err != nil {
		log.Fatal(err)
	}
	s := os.Getenv("SSCUT_PREFIX")
	if s != "" {
		prefix = s
	}
	s = os.Getenv("SSCUT_DEFAULT")
	if s != "" {
		defsc = s
	}
	s = os.Getenv("SSCUT_HOSTNAME")
	if s != "" {
		hostname = s
	}
	_, ok := scuts[defsc]
	if !ok {
		log.Fatalf("default scut \"%s\" doesn't exist in %s\n", defsc, flag.Arg(0))
	}

	http.HandleFunc("/s", shortcut)
	http.HandleFunc("/opensearch.xml", func(w http.ResponseWriter, req *http.Request) {
		err := osTmpl.Execute(w, hostname)
		if err != nil {
			log.Print(err)
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, "./main.html")
	})
	http.ListenAndServe(":8080", nil)
}

func shortcut(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Referrer-Policy", "no-referrer")

	var scstr string
	q := req.FormValue("q")
	qfields := strings.Fields(q)
	if strings.HasPrefix(q, prefix) && len(qfields[0]) != len(prefix) {
		scstr = qfields[0][len(prefix):]
	}

	var sc scut.Scut
	if scstr == "" {
		sc, _ = scuts[defsc]
	} else {
		var ok bool
		sc, ok = scuts[scstr]
		if !ok {
			sc, _ = scuts[defsc]
		} else {
			q = q[len(qfields[0]):]
		}
	}

	query := strings.TrimSpace(q)
	if query == "" {
		http.Redirect(w, req, sc.MainUrl, http.StatusMovedPermanently)
		return
	}
	if sc.UsePOST {
		err := postTmpl.Execute(w, map[string]interface{}{
			"url":   sc.SearchUrl,
			"key":   sc.Key,
			"value": query,
		})
		if err != nil {
			log.Print(err)
		}
		return
	}
	params := &url.Values{}
	params.Set(sc.Key, query)
	u, err := url.Parse(sc.SearchUrl)
	if err != nil {
		log.Print(err)
		return
	}
	u.RawQuery = strings.ReplaceAll(params.Encode(), "+", "%20")
	http.Redirect(w, req, u.String(), http.StatusMovedPermanently)
}
