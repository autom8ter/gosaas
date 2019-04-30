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
	"github.com/autom8ter/gosaas/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net"
)

var port int
var homeTemplatePath string
var blogTemplatePath string
var loggedInTemplatePath string
var apiaddr string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "start the GoSaaS server",
	/*
		Run: func(cmd *cobra.Command, args []string) {
			conn, err := grpc.DialContext(common.ClientContext, apiaddr, grpc.WithInsecure())
			if err != nil {
				util.Util.Entry().Fatal(err.Error())
			}

			h := handler.NewHandler(&api.Auth{
				Domain:       common.StringFromEnv("AUTH0_DOMAIN"),
				ClientId:     common.StringFromEnv("AUTH0_CLIENT_ID"),
				ClientSecret: common.StringFromEnv("AUTH0_CLIENT_SECRET"),
				Redirect:     api.DEFAULT_OAUTH_REDIRECT,
				Scopes:       api.DEFAULT_OAUTH_SCOPES,
			}, api.NewClientSet(conn), "/", "/dashboard", "/login", "/logout", "/callback", "http://localhost:8080", "/blog")

			log.Debugln(fmt.Sprintf("starting server: %v:", port))
			if err := h.ListenAndServe(
				":8080",
				func(w http.ResponseWriter, r *http.Request) {

					http.ServeFile(w, r, blogTemplatePath)
				},
				func(w http.ResponseWriter, r *http.Request) {

					http.ServeFile(w, r, homeTemplatePath)
				},
				func(w http.ResponseWriter, r *http.Request) {

					http.ServeFile(w, r, loggedInTemplatePath)
				},
			); err != nil {
				util.Util.Entry().Fatalln(err.Error())
			}
		},
	*/

}

func init() {
	viper.SetDefault("port", 8080)
	viper.SetDefault("home", "static/home.html")
	viper.SetDefault("loggedin", "static/loggedin.html")
	viper.SetDefault("blog", "static/blog.html")
	viper.SetDefault("api", "localhost:3000")
	viper.SetConfigFile("gosaas.yaml")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		util.Util.Warnf("error: %s", err.Error())
	}
	serveCmd.Flags().IntVar(&port, "port", viper.GetInt("port"), "port to serve on")
	serveCmd.Flags().StringVar(&homeTemplatePath, "home", viper.GetString("home"), "path to home template")
	serveCmd.Flags().StringVar(&loggedInTemplatePath, "loggedin", viper.GetString("loggedin"), "path to loggedin template")
	serveCmd.Flags().StringVar(&blogTemplatePath, "blog", viper.GetString("blog"), "path to blog template")
	serveCmd.Flags().StringVarP(&apiaddr, "api", "a", viper.GetString("api"), "api url")

	rootCmd.AddCommand(serveCmd)
	if _, err := net.Dial("tcp", apiaddr); err != nil {
		util.Util.Fatalf("api is unreachable: %s", err.Error())
	}
}
