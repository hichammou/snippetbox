package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"snippetbox.hichammou/internal/models"
	"snippetbox.hichammou/internal/validator"
)

// Define a snippetCreateForm struct to represent the form data and validation errors for the form fields.
type snippetCreateForm struct {
	Title   string
	Content string
	Expires int
	// Here we Embedded the Validator stuct, mean that our snippetCreateForm inherits all the fields and methods of the Validator stuct
	validator.Validator
}

type UserSignupForm struct {
	Name     string
	Email    string
	Password string
	validator.Validator
}

type UserLoginForm struct {
	Email    string
	Password string
	validator.Validator
}

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// Auth pages

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	data.Form = UserSignupForm{}

	app.render(w, r, http.StatusOK, "signup.html", data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := UserSignupForm{
		Name:     r.PostForm.Get("name"),
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NoBlank(form.Name), "name", "This field could not be empty")
	form.CheckField(validator.NoBlank(form.Email), "email", "This field could not be empty")
	form.CheckField(validator.Match(form.Email, validator.EmailRX), "email", "This field must be a valide email address")
	form.CheckField(validator.NoBlank(form.Password), "password", "This field could not be empty")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)

		data.Form = form

		app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "Email address is already in use")

			data := app.newTemplateData(r)
			data.Form = form

			app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			app.serverError(w, r, err)
		}

		return
	}

	app.sessionManager.Put(r.Context(), "flash", "You signup was successfull. Please log in.")

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = UserLoginForm{}

	app.render(w, r, http.StatusOK, "login.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := UserLoginForm{
		Email:    r.PostForm.Get("email"),
		Password: r.PostForm.Get("password"),
	}

	form.CheckField(validator.NoBlank(form.Email), "email", "This field could not be empty")
	form.CheckField(validator.Match(form.Email, validator.EmailRX), "email", "This field should be a valide email")
	form.CheckField(validator.NoBlank(form.Password), "password", "This field could not be empty")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	id, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalideCredentials) {
			form.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form

			app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	// RenewToken() to change the current session ID. it's a good practice to generate a new token when the auth state changes
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// add the ID of the current user to the session, so that they are now logged in
	app.sessionManager.Put(r.Context(), "authenticatedUserID", id)

	// Get where the user cam from before they were redirected to login page.
	path := app.sessionManager.PopString(r.Context(), "fromUri")

	to := "/snippet/create"

	if path != "" {
		to = path
	}

	// Redirect the user to the create snippet page.
	http.Redirect(w, r, to, http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	app.sessionManager.Put(r.Context(), "flash", "You've logged successfull!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// user
func (app *application) Account(w http.ResponseWriter, r *http.Request) {
	id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")

	user, err := app.users.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.User = user

	app.render(w, r, http.StatusOK, "account.html", data)
}

// Snippet pages

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)

	data.Snippets = snippets

	app.render(w, r, http.StatusOK, "home.html", data)
}

func (app *application) SnippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))

	if err != nil || id < 1 {
		app.serverError(w, r, err)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, r, http.StatusOK, "view.html", data)
}

func (app *application) SnippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// Initialize a new createSnippetForm instance and pass it to the template.
	data.Form = snippetCreateForm{
		Expires: 7,
	}
	app.render(w, r, http.StatusOK, "create.html", data)
}

func (app *application) SnippetCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	// doing this manualy is fine because our form has only three fields. But if the form is very large consider using a form decoder package
	// like go-playground/form to save you typing.

	form := snippetCreateForm{
		Title:   r.PostForm.Get("title"),
		Content: r.PostForm.Get("content"),
		Expires: expires,
	}

	form.CheckField(validator.NoBlank(form.Title), "title", "This field can't be empty")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field can't be more than 100 characters long")
	form.CheckField(validator.NoBlank(form.Content), "content", "This field can't be empty")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 356), "expires", "This field must equal to 1, 7 aor 365")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content, form.Expires)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}

// About pages
func (app *application) About(w http.ResponseWriter, r *http.Request) {
	form := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, "about.html", form)
}
