package handler

import (
	"context"
	"github.com/autom8ter/api"
	"github.com/autom8ter/api/common"
	"github.com/autom8ter/auth0/endpoints"
	"github.com/autom8ter/gosaas/sessions"
	"github.com/autom8ter/gosaas/util"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	*api.OAuth2
	APIAddr string
	Domain  string
}

func NewHandler(domain, clientID, clientSecret, redirect string) *Handler {

	return &Handler{
		OAuth2: &api.OAuth2{
			ClientId:     common.ToString(clientID),
			ClientSecret: common.ToString(clientSecret),
			TokenUrl:     common.ToString(endpoints.TokenURL(domain)),
			AuthUrl:      common.ToString(endpoints.AuthURL(domain)),
			Scopes:       common.ToStringArray([]string{"openid", "profile", "email"}),
			Redirect:     common.ToString(redirect),
		},
	}
}

func (h *Handler) Callback(loggedInPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")

		if state != sessions.State(r).Values["state"] {
			http.Error(w, "Invalid state parameter", http.StatusInternalServerError)
			return
		}

		h.Code = common.ToString(r.URL.Query().Get("code"))
		t, err := h.Token()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

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
		s := sessions.Auth(r)
		s.Values["userinfo"] = util.Util.ToMap(usr)

		if err := s.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, loggedInPath, http.StatusSeeOther)
	}

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

func (a *Handler) RequireLogin(redirect string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usr := sessions.Auth(r).Values["userinfo"]

		if usr == nil {
			http.Redirect(w, r, redirect, http.StatusSeeOther)
			return
		}

		u := &api.User{}
		err := u.UnmarshalJSONFrom(util.Util.MarshalJSON(usr))
		if err != nil {
			http.Redirect(w, r, redirect, http.StatusSeeOther)
			return
		}
		if u == nil {
			http.Redirect(w, r, redirect, http.StatusSeeOther)
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
		state := util.Util.RandomString(32)
		s := sessions.State(r)
		s.Values["state"] = state
		_ = s.Save(r, w)
		http.Redirect(w, r, a.AuthCodeURL(state, endpoints.UserInfoURL(a.Domain)), http.StatusTemporaryRedirect)
	}
}

func (a *Handler) RenderFile(name string, data []byte) http.HandlerFunc {
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

func (a *Handler) WriteFile(name string) http.HandlerFunc {
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
