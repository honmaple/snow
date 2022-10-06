package server

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/honmaple/snow/config"
)

func Serve(conf *config.Config) error {
	baseURL := conf.GetString("baseURL")
	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(conf.GetString("output_dir"))))
	return http.ListenAndServe(fmt.Sprintf("%s:%d", u.Host, u.Port()), mux)
}
