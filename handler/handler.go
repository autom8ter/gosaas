package handler

import (
	"bytes"
	"github.com/autom8ter/api"
	"github.com/autom8ter/api/common"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	*api.Auth
	*api.ClientSet
	HomePath     string
	LoggedInPath string
	LoginPath    string
	LogoutPath   string
	CallbackPath string
	HomeURL      string
	BlogPath     string
}

func NewHandler(auth *api.Auth, clientSet *api.ClientSet, homePath string, loggedInPath string, loginPath string, logoutPath string, callbackPath string, homeURL string, blogPath string) *Handler {
	return &Handler{Auth: auth, ClientSet: clientSet, HomePath: homePath, LoggedInPath: loggedInPath, LoginPath: loginPath, LogoutPath: logoutPath, CallbackPath: callbackPath, HomeURL: homeURL, BlogPath: blogPath}
}

func (h *Handler) callbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.DefaultIfEmpty()
		if err := h.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		state := r.URL.Query().Get("state")
		session, err := common.GetStateSession(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if state != session.Values["state"] {
			http.Error(w, "Invalid state parameter", http.StatusInternalServerError)
			return
		}

		code := r.URL.Query().Get("code")
		t, err := h.Token(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.ToSession(session)
		var req = &api.ResourceRequest{}
		req.Token = t
		req.Domain = h.Domain
		req.Url = api.URL_USER_INFOURL
		req.Method = common.HTTPMethod_GET
		resp, err := h.Resource.GetResource(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var userinfo map[string]interface{}
		err = resp.UnmarshalJSON(bytes.NewBuffer(common.Util.MarshalJSON(userinfo)))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sesh, err := common.GetAuthSession(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		common.AuthSessionValues(sesh, "user", userinfo)
		common.SaveSession(w, r)
		// Redirect to logged in page
		http.Redirect(w, r, h.LoggedInPath, http.StatusSeeOther)
	}

}

func (a *Handler) ListenAndServe(addr string, blog, home, loggedIn http.HandlerFunc) error {
	return http.ListenAndServe(addr, a.Router(blog, home, loggedIn))
}

func (c *Handler) logoutURL() (string, error) {
	var Url *url.URL
	Url, err := url.Parse("https://" + c.Domain.Text)
	if err != nil {
		return "", err
	}

	Url.Path += "/v2/logout"
	parameters := url.Values{}
	parameters.Add("returnTo", c.HomeURL)
	parameters.Add("client_id", c.ClientId.Text)
	Url.RawQuery = parameters.Encode()

	return Url.String(), nil
}

func (a *Handler) Router(home, blog, loggedIn http.HandlerFunc) *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc(a.LogoutPath, a.logoutHandler())
	m.HandleFunc(a.LoginPath, a.loginHandler())
	m.HandleFunc(a.CallbackPath, a.callbackHandler())
	m.HandleFunc(a.HomePath, home)

	m.Handle(a.LoggedInPath, a.RequireLogin(loggedIn))
	m.HandleFunc(a.BlogPath, blog)

	return m
}

func (a *Handler) RequireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := common.GetStateSession(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, ok := session.Values["user"]; !ok {
			http.Redirect(w, r, a.LoginPath, http.StatusSeeOther)
		} else {
			next(w, r)
		}
	}
}

func (a *Handler) logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.DefaultIfEmpty()
		if err := a.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u, err := a.logoutURL()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	}
}

func (a *Handler) loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.DefaultIfEmpty()
		if err := a.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		state := common.RandomString()
		session, err := common.GetStateSession(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values["state"] = state
		common.SaveSession(w, r)
		http.Redirect(w, r, a.AuthCodeURL(state.Text, api.URL_USER_INFOURL), http.StatusTemporaryRedirect)
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
