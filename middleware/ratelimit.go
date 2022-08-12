package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"pow-shield/utils"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Limiter struct {
	// limit is maximum number of events allowed in a period of time.
	Limit int
	// token bucket
	Token int
	// rate duration
	Rate time.Duration
	// last time token bucket is filled up with limit tokens
	Last time.Time
	// update every event happen
	Event time.Time
}

func NewLimiter(limit int, rate time.Duration) *Limiter {
	return &Limiter{
		Limit: limit,
		Rate:  rate,
	}
}

func (l *Limiter) Allow() bool {
	l.Event = time.Now()
	if l.Last.Add(l.Rate).Before(l.Event) {
		l.Token = l.Limit - 1
		l.Last = time.Now()
		return true
	} else {
		if l.Token <= 0 {
			return false
		} else {
			l.Token--
			return true
		}
	}
}

func Ratelimit(redisClient *redis.Client, next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("RATE_LIMIT") != "on" {
			next(w, r)
			return
		}

		var limit *Limiter

		// get ip address
		ip, _ := utils.GetIP(r)

		// get value form redis for ip address
		value, err := redisClient.Get(context.Background(), ip).Result()
		if err == redis.Nil {
			// if ip address is not in redis, create new limiter
			th, err := strconv.Atoi(os.Getenv("RATE_LIMIT_SESSION_THRESHOLD"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			n, err := strconv.Atoi(os.Getenv("RATE_LIMIT_SAMPLE_MINUTES"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			limit = NewLimiter(th, time.Minute*time.Duration(n))

			// marshal limiter to json
			b, _ := json.Marshal(limit)

			// set value to redis for ip address
			redisClient.Set(context.Background(), ip, string(b), 0)
		} else if err != nil {
			// if error, return error
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// unmarshal value
		json.Unmarshal([]byte(value), &limit)

		// if limiter is not allowed, return error
		if !limit.Allow() {
			http.Redirect(w, r, "/pow?status=banned", http.StatusFound)
			return
		}

		// save value to redis
		b, _ := json.Marshal(limit)
		redisClient.Set(context.Background(), ip, string(b), 0)

		next(w, r)
	}
}
