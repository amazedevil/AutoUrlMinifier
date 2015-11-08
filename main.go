package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/xilp/systray"
)

const (
	URLS_FILENAME            = "urls.txt"
	GOOGLE_API_KEY           = "<GOOGLE_API_KEY>"
	CLIPBOARD_CHECK_INTERVAL = time.Duration(500) * time.Millisecond
)

var urls []string
var lastClipboardValue string = ""

func main() {
	readConfig()
	tray := systray.New("icons", "systray", 6333)
	tray.OnClick(func() {
		tray.Stop()
		os.Exit(0)
	})
	err := tray.Show(map[string]string{
		"windows": "win_icon.ico",
		"darwin":  "mac_icon.png",
	}[runtime.GOOS], "Click to exit")
	if err != nil {
		println(err.Error())
	}

	go func() {
		for {
			time.Sleep(CLIPBOARD_CHECK_INTERVAL)
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
		rawUrls := strings.Split(string(bytes), "\n")
		urls = make([]string, 0, len(rawUrls))
		for _, v := range rawUrls {
			urls = append(urls, strings.TrimSpace(v))
		}
	}
}
