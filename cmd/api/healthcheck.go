package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "status: available")
	_, _ = fmt.Fprintf(w, "environment: %s\n\n", app.config.env)
	_, _ = fmt.Fprintf(w, "version: %s\n", version)
}
