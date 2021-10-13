package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

type responsetype struct {
	Id       string
	Origin   string
	Headers  map[string]string
	Response interface{}
	Remote   string
	Uri      string
}

func wrap(fnc func(w http.ResponseWriter, r *http.Request) (interface{}, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		response, err := fnc(w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		headers := map[string]string{}
		for k, v := range r.Header {
			headers[k] = strings.Join(v, " ")
		}

		reqId := middleware.GetReqID(r.Context())

		resp := &responsetype{
			Id:       reqId,
			Response: response,
			Headers:  headers,
			Origin:   r.Host,
			Remote:   r.RemoteAddr,
			Uri:      r.RequestURI,
		}

		ret, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(ret)
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		res, _ := json.Marshal(struct {
			Message string
			Routes  []string
		}{
			Message: "welcome",
			Routes: []string{
				"GET / (here)",
				"GET /info",
				"GET /ip",
				"GET /sleep/:seconds",
				"GET /status/:statusCode",
			},
		})

		w.Header().Add("Content-Type", "application/json")
		w.Write(res)
	})

	r.Get("/ip", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")
		if rip := r.Header.Get(xRealIP); rip != "" {
			w.Write([]byte(rip))
			return
		}
		if rip := r.Header.Get(xForwardedFor); rip != "" {
			w.Write([]byte(rip))
			return
		}
		w.Write([]byte(r.RemoteAddr))
	})

	r.Get("/info", wrap(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		return nil, nil
	}))

	r.Get("/sleep/{seconds}", wrap(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		sec := chi.URLParam(r, "seconds")
		s, err := strconv.Atoi(sec)
		if err != nil {
			return nil, err
		}

		time.Sleep(time.Duration(s * 1000 * 1000 * 1000))

		return struct {
			Duration int
		}{
			Duration: s,
		}, nil
	}))

	r.Get("/status/{statusCode}", func(w http.ResponseWriter, r *http.Request) {
		sec := chi.URLParam(r, "statusCode")
		s, err := strconv.Atoi(sec)
		if err != nil {
			http.Error(w, http.StatusText(400), 400)
			return
		}
		if s < 100 || 599 < s {
			http.Error(w, http.StatusText(400), 400)
			return
		}

		http.Error(w, http.StatusText(s), s)
	})

	err := http.ListenAndServe(":3000", r)
	log.Println(err)
}
