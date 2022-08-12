package middleware

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func changeStringToIntArray(s string) []int {
	var result []int
	if s != "" {
		for _, v := range strings.Split(s, ",") {
			i, err := strconv.Atoi(v)
			if err != nil {
				log.Fatal(err)
			}
			result = append(result, i)
		}
	}
	return result
}

type waftypes map[string]string

type wafRule struct {
	ID   int    `json:"id"`
	Reg  string `json:"reg"`
	Type int    `json:"type"`
	Cmt  string `json:"cmt"`
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func wafDetect(rules []wafRule, match string, exclude []int) []int {
	var result []int
	for _, v := range rules {
		if contains(exclude, v.ID) {
			continue
		}

		// regexp
		regexp := regexp.MustCompile(v.Reg)
		if regexp.MatchString(match) {
			result = append(result, v.ID)
		}
	}
	return result
}

func Waf(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("WAF") != "on" {
			next(w, r)
			return
		}

		waf_url_exclude_rules := os.Getenv("WAF_URL_EXCLUDE_RULES") // comma separated list of integers
		waf_header_exclude_rules := os.Getenv("WAF_HEADER_EXCLUDE_RULES")
		waf_body_exclude_rules := os.Getenv("WAF_BODY_EXCLUDE_RULES")

		exclude_url := changeStringToIntArray(waf_url_exclude_rules)
		exclude_header := changeStringToIntArray(waf_header_exclude_rules)
		exclude_body := changeStringToIntArray(waf_body_exclude_rules)

		// open wafTypes.json
		file, err := os.Open("wafTypes.json")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		var waftypes waftypes
		err = json.NewDecoder(file).Decode(&waftypes)
		if err != nil {
			log.Fatal(err)
		}

		// open file wafRules.json and parse it to wafRules
		file, err = os.Open("wafRules.json")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		var wafRules []wafRule
		err = json.NewDecoder(file).Decode(&wafRules)
		if err != nil {
			log.Fatal(err)
		}

		// wafdetect for url
		url := r.URL.String()
		url_result := wafDetect(wafRules, url, exclude_url)
		if len(url_result) > 0 {
			http.Redirect(w, r, "/pow?status=waf", http.StatusFound)
			return
		}

		// wafdetect for header
		for k, v := range r.Header {
			header := k + ": " + strings.Join(v, ",")
			header_result := wafDetect(wafRules, header, exclude_header)
			if len(header_result) > 0 {
				http.Redirect(w, r, "/pow?status=waf", http.StatusFound)
				return
			}
		}

		// wafdetect for body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}
		r.Body.Close()
		body_result := wafDetect(wafRules, string(body), exclude_body)
		if len(body_result) > 0 {
			http.Redirect(w, r, "/pow?status=waf", http.StatusFound)
			return
		}

		next(w, r)
	}
}
