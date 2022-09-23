package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"pow-shield/middleware"
	"pow-shield/pow"
	"pow-shield/utils"
	"strconv"
	"text/template"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

// redis client
var redisClient *redis.Client

type PowAuth struct {
	Prefix string `json:"prefix"`
	Nonce  int    `json:"nonce"`
	Auth   bool   `json:"auth"`
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// path with request url path and queries
	path := r.URL.Path
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}

	// get cookie pow-shield
	cookie, err := r.Cookie("pow-shield")
	if err != nil {
		http.Redirect(w, r, "/pow?redirect="+url.QueryEscape(path), http.StatusFound)
		return
	}

	key := cookie.Value
	// get value from redis
	value, err := redisClient.Get(context.Background(), key).Result()
	if err != nil {
		http.Redirect(w, r, "/pow?redirect="+url.QueryEscape(path), http.StatusFound)
		return
	}

	// unmarshal value
	var powAuth PowAuth
	json.Unmarshal([]byte(value), &powAuth)
	if !powAuth.Auth {
		http.Redirect(w, r, "/pow?redirect="+url.QueryEscape(path), http.StatusFound)
		return
	}

	target, err := url.Parse(os.Getenv("BACKEND_URL") + path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	// set origin
	r.Host = target.Host
	// serve request
	proxy.ServeHTTP(w, r)

}

var powTmpl = template.Must(template.ParseFiles("templates/index.gohtml"))

func powHandler(w http.ResponseWriter, r *http.Request) {
	difficulty, _ := strconv.Atoi(os.Getenv("DIFFICULTY"))
	if r.Method == "GET" {
		prefix := utils.RandStrn(4)
		redirect, _ := url.PathUnescape(r.FormValue("redirect"))

		// key
		key := prefix + strconv.Itoa(rand.Intn(1000000))

		// set cookie
		http.SetCookie(w, &http.Cookie{
			Name:  "pow-shield",
			Value: key,
		})

		// set pow auth
		powAuth := PowAuth{
			Prefix: prefix,
		}

		b, _ := json.Marshal(powAuth)

		// set value to redis
		redisClient.Set(context.Background(), key, string(b), 0)

		powTmpl.Execute(w, map[string]interface{}{
			"prefix":     prefix,
			"difficulty": difficulty,
			"redirect":   redirect,
			"status":     r.FormValue("status"),
		})
	} else if r.Method == "POST" {
		// get cookie
		cookie, err := r.Cookie("pow-shield")
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		key := cookie.Value

		// get value from redis
		value, err := redisClient.Get(context.Background(), key).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// unmarshal value
		var powAuth PowAuth
		json.Unmarshal([]byte(value), &powAuth)

		// get post data
		var postData PowAuth
		err = json.NewDecoder(r.Body).Decode(&postData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if powAuth.Prefix != postData.Prefix && !pow.Verify(postData.Prefix, difficulty, postData.Nonce) {
			http.Error(w, "invalid nonce", http.StatusBadRequest)
			return
		}

		// update value to redis
		powAuth.Auth = true
		powAuth.Nonce = postData.Nonce
		b, _ := json.Marshal(powAuth)

		redisClient.Set(context.Background(), key, string(b), 0)

		// response with no content
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	// load env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	redisHost := os.Getenv("DATABASE_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("DATABASE_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisPassword := os.Getenv("DATABASE_PASSWORD")

	// get redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       0, // use default DB
	})

	// ping redis
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	// http server mux
	mux := http.NewServeMux()
	// static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
	// handlers
	mux.HandleFunc("/", middleware.Ratelimit(redisClient, middleware.Waf(rootHandler)))
	mux.HandleFunc("/pow", powHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s", port)

	if os.Getenv("SSL") == "on" {
		err = http.ListenAndServeTLS(":"+port, os.Getenv("SSL_CERT"), os.Getenv("SSL_KEY"), mux)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := http.ListenAndServe(":"+port, mux)
		if err != nil {
			log.Fatal(err)
		}
	}
}
