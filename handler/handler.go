package handler

import (
	"github.com/autom8ter/api"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
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
		session, err := api.GetStateSession(r)
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
		resp, err := h.Resource.GetResource(r.Context(), t.ResourceRequest(h.Domain, api.HTTPMethod_GET, api.URL_USER_INFOURL, nil, nil))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var userinfo map[string]interface{}
		err = resp.UnMarshalJSON(userinfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sesh, err := api.GetAuthSession(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		api.AuthSessionValues(sesh, "user", userinfo)
		api.SaveSession(w, r)
		// Redirect to logged in page
		http.Redirect(w, r, h.LoggedInPath, http.StatusSeeOther)
	}

}

func (a *Handler) ListenAndServe(addr string, blog, home, loggedIn http.HandlerFunc) error {
	return http.ListenAndServe(addr, a.Router(blog, home, loggedIn))
}

func (c *Handler) logoutURL() (string, error) {
	var Url *url.URL
	Url, err := url.Parse("https://" + c.Domain)
	if err != nil {
		return "", err
	}

	Url.Path += "/v2/logout"
	parameters := url.Values{}
	parameters.Add("returnTo", c.HomeURL)
	parameters.Add("client_id", c.ClientId)
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
		session, err := api.GetStateSession(r)
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
		state := api.CreateRandomState()
		session, err := api.GetStateSession(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values["state"] = state
		api.SaveSession(w, r)
		http.Redirect(w, r, a.AuthCodeURL(state, api.URL_USER_INFOURL), http.StatusTemporaryRedirect)
	}
}
