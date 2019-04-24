// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"github.com/autom8ter/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"

	"os"
)

var addr string
var homeTemplate string
var blogTemplate string
var dashboard string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start the GoSaaS server",
	Run: func(cmd *cobra.Command, args []string) {
		err := api.InitSessions("")
		if err != nil {
			log.Fatal(err.Error())
		}
		a := &api.Auth{
			Domain:       os.Getenv("AUTH0_DOMAIN"),
			ClientId:     os.Getenv("AUTH0_CLIENT_ID"),
			ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		}
		m := a.Mux("/dashboard")

		log.Debugln("loading templates: ", "static/home.html", "static/blog.html", "static/dashboard.html")

		m.HandleFunc("/dashboard", a.RequireLogin(api.RenderFileWithSessionValue("static/dashboard.html", "userinfo")))
		m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "static/home.html")
		})
		m.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "static/blog.html")
		})

		log.Debugln("registered handlers: ", "/", "/blog", "/callback", "/login", "/logout", "/dashboard")

		log.Debugln("starting server: ", addr)
		if err := http.ListenAndServe(addr, m); err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	serveCmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "address to serve on")

	rootCmd.AddCommand(serveCmd)
}
