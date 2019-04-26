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
	"github.com/autom8ter/gosaas/handler"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net/http"

	"os"
)

var addr string
var homeTemplatePath string
var blogTemplatePath string
var loggedInTemplatePath string
var apiaddr string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start the GoSaaS server",
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := grpc.DialContext(api.Context, apiaddr, grpc.WithInsecure())
		if err != nil {
			log.Fatal(err.Error())
		}
		err = api.SecretFromEnv().InitSessions()
		if err != nil {
			log.Fatal(err.Error())
		}
		h := handler.NewHandler(&api.Auth{
			Domain:       os.Getenv("AUTH0_DOMAIN"),
			ClientId:     os.Getenv("AUTH0_CLIENT_ID"),
			ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		}, api.NewClientSet(conn), "/", "/dashboard", "/login", "/logout", "/callback", "http://localhost:8080", "/blog")

		log.Debugln("starting server: ", addr)
		if err := h.ListenAndServe(
			addr,
			func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, homeTemplatePath)
			},
			func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, homeTemplatePath)
			},
			func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, homeTemplatePath)
			},
		); err != nil {
			log.Fatalln(err.Error())
		}
	},
}

func init() {

	viper.SetDefault("addr", ":8080")
	viper.SetDefault("home", "static/home.html")
	viper.SetDefault("loggedin", "static/loggedin.html")
	viper.SetDefault("blog", "static/blog.html")

	serveCmd.Flags().StringVarP(&addr, "addr", "a", viper.GetString("addr"), "address to serve on")
	serveCmd.Flags().StringVar(&homeTemplatePath, "home", viper.GetString("home"), "path to home template")
	serveCmd.Flags().StringVar(&loggedInTemplatePath, "loggedin", viper.GetString("loggedin"), "path to loggedin template")
	serveCmd.Flags().StringVar(&blogTemplatePath, "blog", viper.GetString("blog"), "path to blog template")

	rootCmd.AddCommand(serveCmd)
}
