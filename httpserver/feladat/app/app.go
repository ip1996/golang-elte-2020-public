package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// App holds the common context for the handlers and stuff
// Preferably this should be in a different package and have a simple New() only instead of NewApp()
// You could also use a "config struct" instead of multiple parameters
type App struct {
	logger *zap.Logger
	db     database
}

type errorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type addJSON struct {
	Message string `json:"msg"`
}

type addResponse struct {
	ID int `json:"id"`
}

// TODO(feladat): Toltsd ki a hianyzo (<???>) struct taget a megfelelo ertekkel
type message struct {
	ID      int       `json:"id"`
	Name    string    `json:"user"`
	Message string    `json:"msg"`
	Created time.Time `json:"date"`
}
type listResponse struct {
	Messages []message `json:"messages"`
}

type database interface {
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func NewApp(logger *zap.Logger, db database) *App {
	return &App{
		logger: logger,
		db:     db,
	}
}

// Routes return the endpoint handled by the app
// It receives a router which could help you inject custom middleware or helpers for easier testing
func (a *App) Routes(r *chi.Mux) *chi.Mux {
	r.Use(a.authMiddleware)
	r.Get("/", a.handleMain)
	r.Post("/add", a.handleAdd)
	r.Get("/list", a.handleList)
	return r
}

func (a *App) handleMain(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "main app endpoint and stuff")
}

func (a *App) handleAdd(w http.ResponseWriter, r *http.Request) {
	var input addJSON

	name := getNameFromContext(r.Context())
	if name == "" {
		sendError(a.logger, w, errors.New("no name from context"), http.StatusInternalServerError)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		// Should we log all the invalid incoming format errors?
		sendError(a.logger, w, err, http.StatusBadRequest)
		return
	}

	// TODO(feladat): Adj hozza egy bemenet ellenorzest a kodhoz, ami ures input.Message eseten a megfelelo http hibaval ter vissza
	// Itt a `TestAddHandlerInvalidInput`-nak es a `TestAddHandlerInvalidInput` kell tudnia lefutnia megfeleloen.
	// <???>
	if input.Message == "" {
		sendError(a.logger, w, errors.New("message could not be empty"), http.StatusBadRequest)
		return
	}

	res, err := a.db.Exec("INSERT INTO message (`name`, `message`) VALUES (?,?)", name, input.Message)
	if err != nil {
		sendError(a.logger, w, err, http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	if err := json.NewEncoder(w).Encode(addResponse{ID: int(id)}); err != nil {
		a.logger.Error("add response send failed", zap.Error(err))
		return
	}
}

func (a *App) handleList(w http.ResponseWriter, r *http.Request) {
	var messages []message
	err := a.db.Select(&messages, "SELECT id, name, message, created FROM message ORDER BY id DESC")
	if err != nil {
		sendError(a.logger, w, err, http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(listResponse{Messages: messages}); err != nil {
		a.logger.Error("list response send failed", zap.Error(err))
		return
	}
}

type errorLogger interface {
	Error(msg string, fields ...zap.Field)
}

// sendError will send a JSON instead of a simple string and do some logging.
// Ideally this should have it's own package with other HTTP related helpers.
// And of course it could have it's own `logger` injected as a context at init etc...
func sendError(logger errorLogger, w http.ResponseWriter, err error, code int) {
	logger.Error("error on backend", zap.Error(err), zap.Int("code", code))

	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(errorResponse{
		Error: err.Error(),
		Code:  code,
	}); err != nil {
		logger.Error("json error response send failed", zap.Error(err))
	}
}
