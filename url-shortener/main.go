package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/boltdb/bolt"
	"gopkg.in/yaml.v3"
)

func main() {
	yamlFilename := flag.String("yaml", "map.yaml", "supply map of urls in yaml file")
	jsonFilename := flag.String("json", "map.json", "supply map of urls in json file")
	flag.Parse()

	db, err := bolt.Open("bolt.db", 0600, nil)
	p(err)
	defer db.Close()
	err = createBucket(db)
	p(err)

	paths := []pathURL{
		{
			Path: "/urlshort-godoc",
			URL:  "https://godoc.org/github.com/gophercises/urlshort",
		},
		{
			Path: "/yaml-godoc",
			URL:  "https://godoc.org/gopkg.in/yaml.v2",
		},
	}

	err = putMany(db, paths)
	p(err)

	yamlData, err := ioutil.ReadFile(*yamlFilename)
	p(err)
	paths, err = unmarshal(yamlData, yaml.Unmarshal)
	p(err)
	err = putMany(db, paths)
	p(err)

	jsonData, err := ioutil.ReadFile(*jsonFilename)
	p(err)
	paths, err = unmarshal(jsonData, json.Unmarshal)
	p(err)
	err = putMany(db, paths)
	p(err)

	fmt.Println("Starting the serveron :8080")
	http.ListenAndServe(":8080", handler(db))
}

func p(e error) {
	if e != nil {
		panic(e)
	}
}

func handler(db *bolt.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		dest, err := get(db, path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if dest != "" {
			http.Redirect(w, r, dest, http.StatusFound)
			return
		}

		http.Redirect(w, r, "/", http.StatusNotFound)
	}
}

func createBucket(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("default"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})
}

func putMany(db *bolt.DB, p []pathURL) error {
	return db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("default"))

		for _, path := range p {
			err := b.Put([]byte(path.Path), []byte(path.URL))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func put(db *bolt.DB, p *pathURL) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("default"))
		err := b.Put([]byte(p.Path), []byte(p.URL))
		return err
	})
}

func get(db *bolt.DB, path string) (string, error) {
	var value string
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("default"))
		v := b.Get([]byte(path))
		value = string(v)
		return nil
	})

	return value, err
}

func unmarshal(data []byte, unmarshaller func([]byte, interface{}) error) ([]pathURL, error) {
	var pathURLs []pathURL
	err := unmarshaller(data, &pathURLs)
	if err != nil {
		return nil, err
	}

	return pathURLs, nil
}

type pathURL struct {
	Path string `yaml:"path" json:"path"`
	URL  string `yaml:"url"  json:"url"`
}
