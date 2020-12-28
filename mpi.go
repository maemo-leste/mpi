package main

import (
	"bufio"
	"compress/gzip"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Package struct {
	Name            string
	Version         string
	Description     string
	LongDescription string
	Provides        []string
	Depends         []string
	Recommends      []string
	Maintainer      string
	Filename        string
	Suite           string
	Component       string
	Section         string
	Priority        string
	Origin          string
}

func toGopher(s map[string]string, path string) {
	return
}

func toHTML(s map[string]string, path string) {
	out := path + s["Package"] + ".html"

	log.Println(s)

	data := Package{
		Name:            s["Package"],
		Version:         s["Version"],
		Description:     s["Description"],
		LongDescription: s["LongDescription"],
		Depends:         strings.Split(s["Depends"], ", "),
	}

	tpl, err := template.ParseFiles("page.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}

	if err = tpl.Execute(f, data); err != nil {
		log.Fatal(err)
	}
}

func scanStanza(stream *bufio.Scanner) (map[string]string, error) {
	s := make(map[string]string)

	curField := ""
	for stream.Scan() {
		if strings.TrimSpace(stream.Text()) == "" {
			return s, nil
		}

		field := strings.SplitN(stream.Text(), ":", 2)
		if len(field) > 1 {
			// This is a new field
			curField = field[0]
			s[curField] = strings.TrimSpace(field[1])
		} else {
			// This is the continuation of the last field
			s[curField] += "\n" + field[0]
		}
	}

	err := stream.Err()
	return s, err
}

func parsePackages(pkgs string) error {
	log.Println("Fetching and parsing", pkgs)

	f, err := http.Get(pkgs)
	if err != nil {
		return err
	}

	uncompressed, err := gzip.NewReader(f.Body)
	if err != nil {
		return err
	}

	defer f.Body.Close()

	r := bufio.NewScanner(uncompressed)

	path := strings.Replace(pkgs, url, "", 1)
	path = strings.Replace(path, "Packages.gz", "", 1)
	path = outputPath + path
	os.MkdirAll(path, 0775)

	for s, err := scanStanza(r); s["Package"] != ""; s, err = scanStanza(r) {
		if err == nil {
			toHTML(s, path)
			toGopher(s, path)
		} else {
			log.Println("error:", err)
		}
	}

	return nil
}

func main() {
	var indexes = []string{}

	for _, arch := range architectures {
		for _, comp := range components {
			for _, suite := range suites {
				indexes = append(indexes, strings.Join([]string{
					url, suite, comp, arch, "Packages.gz"}, "/"))
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(indexes))

	for _, i := range indexes {
		go func(i string) {
			defer wg.Done()
			err := parsePackages(i)
			if err != nil {
				log.Fatal(err)
			}
		}(i)
	}

	wg.Wait()
}
