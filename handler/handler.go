package handler

import (
	"context"
	"fmt"
	"github.com/autom8ter/api"
	"github.com/autom8ter/api/common"
	"github.com/autom8ter/auth0/endpoints"
	"github.com/autom8ter/objectify"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var util = objectify.Default()

type Handler struct {
	*api.OAuth2
	Port         int
	APIAddr      string
	Domain       string
	HomePath     string
	LoggedInPath string
	LoginPath    string
	LogoutPath   string
	CallbackPath string
}

func NewHandler(domain, clientID, clientSecret, redirect, homePath, loggedInPath, loginPath, logoutPath, callbackPath string) *Handler {

	return &Handler{
		OAuth2: &api.OAuth2{
			ClientId:     common.ToString(clientID),
			ClientSecret: common.ToString(clientSecret),
			TokenUrl:     common.ToString(endpoints.TokenURL(domain)),
			AuthUrl:      common.ToString(endpoints.AuthURL(domain)),
			Scopes:       common.ToStringArray([]string{"openid", "profile", "email"}),
			Redirect:     common.ToString(redirect),
		},
		HomePath:     homePath,
		LoggedInPath: loggedInPath,
		LoginPath:    loginPath,
		LogoutPath:   logoutPath,
		CallbackPath: callbackPath,
	}
}

func (h *Handler) callbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.Code = common.ToString(r.URL.Query().Get("code"))
		t, err := h.Token()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r.AddCookie(&http.Cookie{
			Name:  "access_token",
			Value: t.AccessToken,
		})
		r.AddCookie(&http.Cookie{
			Name:  "id_token",
			Value: string(util.MarshalJSON(t.Extra("id_token"))),
		})
		set, err := h.NewAPIClientSet(context.TODO(), h.APIAddr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		usr, err := set.Auth.GetUser(r.Context(), &common.AuthToken{Token: common.ToString(t.AccessToken)})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if usr == nil {
			http.Error(w, "user not found", http.StatusInternalServerError)
			return
		}
		r.AddCookie(&http.Cookie{
			Name:  "user",
			Value: string(util.MarshalJSON(usr)),
		})
		http.Redirect(w, r, h.LoggedInPath, http.StatusSeeOther)
	}

}

func (a *Handler) ListenAndServe(home, loggedIn http.HandlerFunc) error {
	return http.ListenAndServe(fmt.Sprintf(":%v", a.Port), a.Router(home, loggedIn))
}

func (c *Handler) logoutURL(returnTo string) (string, error) {
	var Url *url.URL
	Url, err := url.Parse("https://" + c.Domain)
	if err != nil {
		return "", err
	}

	Url.Path += "/v2/logout"
	parameters := url.Values{}
	parameters.Add("returnTo", returnTo)
	parameters.Add("client_id", c.ClientId.Text)
	Url.RawQuery = parameters.Encode()

	return Url.String(), nil
}

func (a *Handler) Router(home, loggedIn http.HandlerFunc) *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc(a.LogoutPath, a.Logout(fmt.Sprintf("http://localhost:%v", a.Port)))
	m.HandleFunc(a.LoginPath, a.Login(""))
	m.HandleFunc(a.CallbackPath, a.callbackHandler())
	m.HandleFunc(a.HomePath, home)
	m.Handle(a.LoggedInPath, a.RequireLogin(loggedIn))

	return m
}

func (a *Handler) RequireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usr, err := r.Cookie("user")
		if usr == nil {
			http.Redirect(w, r, a.LoginPath, http.StatusSeeOther)
			return
		}

		u := &api.User{}
		err = u.UnmarshalJSONFrom(util.MarshalJSON(usr.Value))
		if err != nil {
			http.Redirect(w, r, a.LoginPath, http.StatusSeeOther)
			return
		}
		if u == nil {
			http.Redirect(w, r, a.LoginPath, http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func (a *Handler) Logout(returnTo string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := a.logoutURL(returnTo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	}
}

func (a *Handler) Login(aud string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if aud == "" {
			aud = "https://" + a.Domain + "/userinfo"
		}
		state := util.RandomString(32)
		r.AddCookie(&http.Cookie{
			Name:  "state",
			Value: state,
		})
		http.Redirect(w, r, a.AuthCodeURL(state, aud), http.StatusTemporaryRedirect)
	}
}

func RenderFileFunc(name string, data []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bits, err := ioutil.ReadFile(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bitstring := string(bits)
		if strings.Contains(bitstring, "{{") {
			templ, err := template.New("").Funcs(common.FuncMap()).Parse(string(bits))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = templ.Execute(w, data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}
		_, err = io.WriteString(w, bitstring)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

}

func WriteFileFunc(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bits, err := ioutil.ReadFile(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		bitstring := string(bits)
		_, err = io.WriteString(w, bitstring)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

}
