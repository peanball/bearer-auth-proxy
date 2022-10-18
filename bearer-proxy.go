package main

import (
	"encoding/base64"
	"fmt"
	"github.com/labstack/echo"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func main() {
	e := echo.New()

	// create the reverse proxy
	fmt.Printf("ENV %v", os.Environ())

	proxyUrl, exists := os.LookupEnv("PROXY_URL")
	if !exists {
		panic("URL not set! Please set PROXY_URL env variable.")
	}
	proxiedUrl, _ := url.Parse(proxyUrl)
	proxy := httputil.NewSingleHostReverseProxy(proxiedUrl)

	e.Use(func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			req := context.Request()
			res := context.Response().Writer

			authHeader := req.Header.Get("Authorization")

			if authHeader == "" {
				return fmt.Errorf("no auth provided")
			}
			encodedAuth := strings.Replace(authHeader, "Basic ", "", -1)
			authPair, err := base64.StdEncoding.DecodeString(encodedAuth)
			if err != nil {
				return fmt.Errorf("error decoding string %s: %v", authHeader, err)
			}

			auth := strings.Split(string(authPair), ":")
			if len(auth) != 2 {
				return fmt.Errorf("no token provided")
			}

			token := auth[1]

			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

			// Update the headers to allow for SSL redirection
			req.Host = proxiedUrl.Host
			req.URL.Host = proxiedUrl.Host
			req.URL.Scheme = proxiedUrl.Scheme

			// ServeHttp is non blocking and uses a go routine under the hood
			proxy.ServeHTTP(res, req)
			return nil
		}
	})

	err := e.Start("127.0.0.1:60060")
	if err != nil {
		return
	}
}
