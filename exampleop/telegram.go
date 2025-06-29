package exampleop

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/zitadel/oidc/v3/pkg/op"

	"github.com/AlexeyKupershtokh/telegram-oidc-server/internal/pkg/telegram"
)

type tgAuthenticate interface {
	CheckTelegramData(id string, data telegram.UserData) error
}

type telegramLogin struct {
	tgAuthenticate tgAuthenticate
	router         chi.Router
	callback       func(context.Context, string) string
}

func NewTelegramLogin(tgAuthenticate tgAuthenticate, callback func(context.Context, string) string, issuerInterceptor *op.IssuerInterceptor) *telegramLogin {
	l := &telegramLogin{
		tgAuthenticate: tgAuthenticate,
		callback:       callback,
	}
	l.createRouter(issuerInterceptor)
	return l
}

func (l *telegramLogin) createRouter(issuerInterceptor *op.IssuerInterceptor) {
	l.router = chi.NewRouter()
	l.router.Get("/", l.loginHandler)
	l.router.Get("/check", l.checkLoginHandler)
}

func (l *telegramLogin) loginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot parse form:%s", err), http.StatusInternalServerError)
		return
	}
	// the oidc package will pass the id of the auth request as query parameter
	// we will use this id through the login process and therefore pass it to the login page
	renderTelegramLogin(w, r.FormValue(queryAuthRequestID), nil)
}

func renderTelegramLogin(w http.ResponseWriter, id string, err error) {
	data := &struct {
		ID    string
		Error string
	}{
		ID:    id,
		Error: errMsg(err),
	}
	err = templates.ExecuteTemplate(w, "telegram_login", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (l *telegramLogin) checkLoginHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("flow_id")

	t := telegram.DefaultVerifier{}
	userData, err := t.ParseData(r.URL.Query())
	if err != nil {
		redirectBackToTelegramLogin(w, r, err.Error())
		return
	}

	fmt.Printf("id: %s, userData: %v\n", id, userData)

	err = l.tgAuthenticate.CheckTelegramData(id, userData)
	if err != nil {
		redirectBackToTelegramLogin(w, r, err.Error())
		return
	}

	http.Redirect(w, r, l.callback(r.Context(), id), http.StatusFound)
}

func redirectBackToTelegramLogin(w http.ResponseWriter, r *http.Request, prompt string) {
	values := make(url.Values)
	values.Set("prompt", url.QueryEscape(prompt))

	url := url.URL{
		Path:     "/telegram",
		RawQuery: values.Encode(),
	}
	http.Redirect(w, r, url.String(), http.StatusSeeOther)
}
