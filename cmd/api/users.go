package main

import (
	"context"
	"errors"
	"github.com/manunio/greenlight/internal/data"
	"github.com/manunio/greenlight/internal/validator"
	"net/http"
	"time"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	tx, err := app.models.Users.DB.BeginTx(context.Background(), nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.Insert(tx, user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// After the user record has been created in the database, generate a new activation
	// token for the user.
	token, err := app.models.Tokens.New(tx, user.ID, 3*24*time.Hour, data.ScopActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Use the background helper to execute an anonymous function that sends the welcome
	// email.
	app.background(func() {

		// create a map to act as a 'holding structure' for the data. This
		// contains the plaintext version of the activation token for the
		// user, along with their ID.
		mailData := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", mailData)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	// TODO: better handling of transaction
	defer func() {
		if err != nil {
			// rollbacks transaction
			if rbErr := tx.Rollback(); rbErr != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
			return
		}

		// commits transaction
		if err = tx.Commit(); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}()

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the plaintext activation token from the request body.
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the plaintext token provided by the client.
	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// create transaction
	tx, err := app.models.Users.DB.BeginTx(context.Background(), nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	defer func() {
		if err != nil {
			// rollbacks transaction
			if rbErr := tx.Rollback(); rbErr != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
			return
		}

		// commits transaction
		if err = tx.Commit(); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}()

	// Retrieve the details of the user associated with the token using the
	// GetForToken() method. If no matching record
	// is found, then we let the client know that the token they provided is not valid.
	user, err := app.models.Users.GetForToken(tx, data.ScopActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Update the user activation status
	user.Activated = true

	// Save the updated user record in our database, checking for any edit conflicts in
	// the same way that we did for our movie records.
	err = app.models.Users.Update(tx, user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// If everything went successfully, then we delete all activation tokens for the
	// user.
	err = app.models.Tokens.DeleteAllForUser(tx, data.ScopActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send the updated user details to the client JSON request.
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
