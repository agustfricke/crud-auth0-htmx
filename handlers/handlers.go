package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/agustfricke/crud-auth0-htmx/auth"
	"github.com/agustfricke/crud-auth0-htmx/database"
	"github.com/agustfricke/crud-auth0-htmx/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func CreateTask(ctx *gin.Context) {
	time.Sleep(1 * time.Second)

    session := sessions.Default(ctx)
    profile := session.Get("profile")
    var userId string

    if profileMap, ok := profile.(map[string]interface{}); ok {
    userId = profileMap["sid"].(string)
    } else {
        fmt.Println("Perfil no válido")
    }

	name := ctx.PostForm("name")

	var task models.Task
    db := database.DB
    task = models.Task{Name: name, UserId: userId}
	db.Create(&task)
}

func User(ctx *gin.Context) {
    db := database.DB
    var tasks []models.Task
    db.Find(&tasks)

	session := sessions.Default(ctx)
	profile := session.Get("profile")

    // Renderiza el archivo HTML y pasa las tareas como datos
    ctx.HTML(http.StatusOK, "user.html", gin.H{
        "tasks": tasks,
        "profile": profile,
    })
}

// Handler for our logout.
func Logout(ctx *gin.Context) {
	logoutUrl, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/v2/logout")
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}

	returnTo, err := url.Parse(scheme + "://" + ctx.Request.Host)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo.String())
	parameters.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	logoutUrl.RawQuery = parameters.Encode()

	ctx.Redirect(http.StatusTemporaryRedirect, logoutUrl.String())
}


func Login(auth *auth.Authenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		state, err := generateRandomState()
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Save the state inside the session.
		session := sessions.Default(ctx)
		session.Set("state", state)
		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.Redirect(http.StatusTemporaryRedirect, auth.AuthCodeURL(state))
	}
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)

	return state, nil
}

func Home(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "home.html", nil)
}

// Handler for our callback.
func Callback(auth *auth.Authenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		if ctx.Query("state") != session.Get("state") {
			ctx.String(http.StatusBadRequest, "Invalid state parameter.")
			return
		}

		// Exchange an authorization code for a token.
		token, err := auth.Exchange(ctx.Request.Context(), ctx.Query("code"))
		if err != nil {
			ctx.String(http.StatusUnauthorized, "Failed to convert an authorization code into a token.")
			return
		}

		idToken, err := auth.VerifyIDToken(ctx.Request.Context(), token)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Failed to verify ID Token.")
			return
		}

		var profile map[string]interface{}
		if err := idToken.Claims(&profile); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		session.Set("access_token", token.AccessToken)
		session.Set("profile", profile)
		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Redirect to logged in page.
		ctx.Redirect(http.StatusTemporaryRedirect, "/user")
	}
}
