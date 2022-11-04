package app

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var app *App = &App{}

type App struct {
	Router   *mux.Router
	Config   Config
	sessions map[string]*Session
}

type Session struct {
	prevPage string
	expiry   time.Time
	loggedIn bool
}

func (s Session) isExpired() bool {
	return s.expiry.Before(time.Now())
}

type Config struct {
	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`
}

func GetApp() *App {
	return app
}

func (a *App) Init() {
	configJson, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(configJson, &a.Config)
	if err != nil {
		panic(err)
	}
	a.sessions = map[string]*Session{}
	a.Router = initRouter()
}

func initRouter() *mux.Router {
	r := mux.NewRouter()

	unprotectedRoutes := r.PathPrefix("/").Subrouter()
	protectedRoutes := r.PathPrefix("/").Subrouter()

	unprotectedRoutes.HandleFunc("/login", LoginHandler)
	protectedRoutes.HandleFunc("/", HomePageHandler)
	protectedRoutes.HandleFunc("/view/{fileName}", ViewHandler)
	protectedRoutes.HandleFunc("/{fileName}", EditHandler)
	protectedRoutes.Use(SessionMiddleware)
	unprotectedRoutes.Use(SessionMiddleware)
	protectedRoutes.Use(AuthorizationMiddleware)
	http.Handle("/", r)

	return r
}
