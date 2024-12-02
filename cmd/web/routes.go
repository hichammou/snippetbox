package main

import (
	"net/http"

	"github.com/justinas/alice"
	"snippetbox.hichammou/ui"
)

func (app *application) routes() http.Handler {

	mux := http.NewServeMux()

	// ---  this was before implementing embed files
	// fileServer := http.FileServer(http.Dir("./ui/static/"))
	// mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	// Unprotected routes
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	mux.HandleFunc("GET /ping", ping)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.Home))
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.SnippetView))
	mux.Handle("GET /about", dynamic.ThenFunc(app.About))

	// Add the five new routes, all of which use our 'dynamic' middleware chain.
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	// Protected routes
	protected := dynamic.Append(app.requireAuthentification)

	mux.Handle("GET /snippet/create", protected.ThenFunc(app.SnippetCreate))
	mux.Handle("POST /snippet/create", protected.ThenFunc(app.SnippetCreatePost))
	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))
	mux.Handle("GET /user/account", protected.ThenFunc(app.Account))

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)

	return standard.Then(mux)
}
