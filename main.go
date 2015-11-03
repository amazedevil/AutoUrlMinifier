package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/xilp/systray"
)

const (
	URLS_FILENAME  = "urls.txt"
	GOOGLE_API_KEY = "<GOOGLE_API_KEY>"
)

var urls []string
var lastClipboardValue string = ""

func main() {
	readConfig()
	tray := systray.New("", "", 6333)
	tray.OnClick(func() {
		os.Exit(0)
	})
	err := tray.Show("icon.ico", "Click to exit")
	if err != nil {
		println(err.Error())
	}

	go func() {
		for {
			time.Sleep(time.Duration(1) * time.Second)
			processClipboard()
		}
	}()

	err = tray.Run()
	if err != nil {
		println(err.Error())
	}
}

func minify(url string) string {
	var jsonStr = []byte(`{"longUrl":"` + url + `"}`)
	req, _ := http.NewRequest(
		"POST",
		"https://www.googleapis.com/urlshortener/v1/url?key="+GOOGLE_API_KEY,
		bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		type ShortUrl struct {
			Kind    string
			Id      string
			LongUrl string
		}
		var su ShortUrl
		decerr := json.Unmarshal(body, &su)
		if decerr == nil {
			return su.Id
		}
	}

	return ""
}

func processClipboard() {
	val, _ := clipboard.ReadAll()
	if lastClipboardValue != val {
		lastClipboardValue = val
		found := false
		for _, urlPattern := range urls {
			m, _ := regexp.MatchString(
				"\\A"+strings.Replace(urlPattern, "*", "(.*)", -1)+"\\z",
				val)
			if m {
				found = true
				break
			}
		}
		if found {
			go func() {
				min := minify(val)
				if len(min) > 0 {
					clipboard.WriteAll(min)
					lastClipboardValue = min
				}
			}()
		}
	}
}

func isFileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func readConfig() {
	if isFileExists(URLS_FILENAME) {
		bytes, _ := ioutil.ReadFile(URLS_FILENAME)
		urls = strings.Split(string(bytes), "\n")
	} else {
		urls = make([]string, 0, 5)
	}
}

func writeConfig() {
	ioutil.WriteFile(URLS_FILENAME, []byte(strings.Join(urls, "\n")), 0644)
}
