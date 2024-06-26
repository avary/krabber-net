package web

import (
	"fmt"
	"net/http"
	"os"
)

func (app *Application) crabmin(w http.ResponseWriter, r *http.Request) {
	// todo add in the check for allowed admins
	id := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")

	if id == "" {
		app.NotFound(w)
		return
	}

	ca := os.Getenv("CRABMIN")
	if id != ca {
		app.NotFound(w)
		return
	}

	data := app.NewTemplateData(r)
	app.Render(w, r, http.StatusOK, "crabmin.html", data)
}

func (app *Application) crabminCreateSea(w http.ResponseWriter, r *http.Request) {
	// todo add in the check for allowed admins
	id := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	if id == "" {
		app.NotFound(w)
		return
	}
	ca := os.Getenv("CRABMIN")
	if id != ca {
		app.NotFound(w)
		return
	}

	err := app.Molts.FillSea()
	if err != nil {
		app.serverError(w, r, err)
	}

	data := app.NewTemplateData(r)
	app.Render(w, r, http.StatusOK, "crabmin.html", data)
}

func (app *Application) allCrabs(w http.ResponseWriter, r *http.Request) {
	// for now show this logged in crabs molts
	id := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	if id == "" {
		app.NotFound(w)
		return
	}
	c, err := app.Crabs.ByID(id)
	if err != nil {
		fmt.Printf("ERROR getting crab %v", err)
		app.serverError(w, r, err)
		return
	}

	crabs, err := app.Crabs.Show() /// add get string here
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.NewTemplateData(r)
	data.Crabs = crabs
	data.Crab = c
	app.Render(w, r, http.StatusOK, "crabs.html", data)
}

func (app *Application) root(w http.ResponseWriter, r *http.Request) {
	// for now show this logged in crabs molts
	data := app.NewTemplateData(r)
	data.PageNumber = 1
	data.Page = "p"
	id := app.SessionManager.GetString(r.Context(), "authenticatedCrabID")
	// crab isn't logged in
	if id == "" {
		app.Render(w, r, http.StatusOK, "welcome.html", data)
	}
	// crab is logged in
	if id != "" {
		app.Render(w, r, http.StatusOK, "trench.html", data)
	}

}
