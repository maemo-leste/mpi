package main

import (
	"bufio"
	"compress/gzip"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/template"
)

type Package struct {
	Architecture string
	Breaks       []string
	Bugtracker   string
	// Build-Ids string
	Conflicts     []string
	Depends       []string
	Description   string
	Enhances      string
	Filename      string
	Homepage      string
	InstalledSize string
	Maintainer    string
	MD5sum        string
	Package       string
	PreDepends    []string
	Priority      string
	Provides      []string
	Recommends    []string
	Replaces      []string
	Section       string
	SHA1          string
	SHA256        string
	Size          string
	Source        string
	Suggests      []string
	Version       string
	Extras        bool
}

func populatePackage(s map[string]string) Package {
	p := Package{}
	if v, ok := s["Architecture"]; ok {
		p.Architecture = v
	}
	if v, ok := s["Breaks"]; ok {
		p.Breaks = strings.Split(v, ", ")
	}
	if v, ok := s["Bugtracker"]; ok {
		p.Bugtracker = v
	}
	if v, ok := s["Conflicts"]; ok {
		p.Conflicts = strings.Split(v, ", ")
	}
	if v, ok := s["Depends"]; ok {
		p.Depends = strings.Split(v, ", ")
	}
	if v, ok := s["Description"]; ok {
		p.Description = v
	}
	if v, ok := s["Enhances"]; ok {
		p.Enhances = v
	}
	if v, ok := s["Filename"]; ok {
		p.Filename = v
	}
	if v, ok := s["Homepage"]; ok {
		p.Homepage = v
	}
	if v, ok := s["Installed-Size"]; ok {
		p.InstalledSize = v
	}
	if v, ok := s["Maintainer"]; ok {
		p.Maintainer = v
	}
	if v, ok := s["MD5sum"]; ok {
		p.MD5sum = v
	}
	if v, ok := s["Package"]; ok {
		p.Package = v
	}
	if v, ok := s["Pre-Depends"]; ok {
		p.PreDepends = strings.Split(v, ", ")
	}
	if v, ok := s["Priority"]; ok {
		p.Priority = v
	}
	if v, ok := s["Provides"]; ok {
		p.Provides = strings.Split(v, ", ")
	}
	if v, ok := s["Recommends"]; ok {
		p.Recommends = strings.Split(v, ", ")
	}
	if v, ok := s["Replaces"]; ok {
		p.Replaces = strings.Split(v, ", ")
	}
	if v, ok := s["Section"]; ok {
		if strings.Contains(v, "user/") {
			p.Extras = true
		} else {
			p.Extras = false
		}
		p.Section = v
	}
	if v, ok := s["SHA1"]; ok {
		p.SHA1 = v
	}
	if v, ok := s["SHA256"]; ok {
		p.SHA256 = v
	}
	if v, ok := s["Size"]; ok {
		p.Size = v
	}
	if v, ok := s["Source"]; ok {
		p.Source = v
	} else {
		p.Source = p.Package
	}
	if v, ok := s["Suggests"]; ok {
		p.Suggests = strings.Split(v, ", ")
	}
	if v, ok := s["Version"]; ok {
		p.Version = v
	}

	return p
}

func genTemplate(s map[string]string, path string) {
	out := path + s["Package"] + ".html"
	data := populatePackage(s)

	tpl, err := template.ParseFiles("page.tmpl")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

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

func parsePackages(url, suite, comp, arch string) error {
	//log.Println("Fetching and parsing", pkgs)
	pkgs := strings.Join([]string{url, suite, comp, arch, "Packages.gz"}, "/")

	f, err := http.Head(pkgs)
	if err != nil {
		return err
	}

	if f.StatusCode != 200 {
		return nil
	}

	last_modified, err := http.ParseTime(f.Header["Last-Modified"][0])
	if err != nil {
		return err
	}

	f, err = http.Get(pkgs)
	if err != nil {
		return err
	}

	checkpath := strings.Join([]string{outputPath, suite, comp, arch}, "/")
	if fi, err := os.Stat(checkpath); err == nil {
		modtime := fi.ModTime()

		if last_modified.Before(modtime) {
			return nil
		}
	}

	os.RemoveAll(strings.Join([]string{outputPath, suite, comp, arch}, "/"))

	uncompressed, err := gzip.NewReader(f.Body)
	if err != nil {
		//log.Printf("Warning: %s: %s", pkgs, err)
		return nil
	}

	defer f.Body.Close()

	r := bufio.NewScanner(uncompressed)

	path := strings.Replace(pkgs, mainurl, "", 1)
	path = strings.Replace(path, extrasurl, "", 1)
	path = strings.Replace(path, "Packages.gz", "", 1)
	path = outputPath + path
	os.MkdirAll(path, 0775)

	for s, err := scanStanza(r); s["Package"] != ""; s, err = scanStanza(r) {
		if err == nil {
			genTemplate(s, path)
		} else {
			log.Println("error:", err)
		}
	}

	return nil
}

func main() {
	var wg sync.WaitGroup
	maxChan := make(chan string, 100)

	for _, suite := range suites {
		for _, comp := range components {
			for _, arch := range architectures {
				for _, url := range []string{mainurl, extrasurl} {
					pkgs := strings.Join([]string{url, suite, comp, arch, "Packages.gz"}, "/")
					maxChan <- pkgs
					wg.Add(1)
					go func(maxChan chan string) {
						defer wg.Done()
						defer func(maxChan chan string) { <-maxChan }(maxChan)
						err := parsePackages(url, suite, comp, arch)
						if err != nil {
							log.Fatal(err)
						}
					}(maxChan)
				}
			}
		}
	}

	wg.Wait()
}
