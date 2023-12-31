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

func EditProfile(ctx *gin.Context) {
    
}

func Profile(ctx *gin.Context) {
    db := database.DB
    var user models.User

    session := sessions.Default(ctx)
    profile := session.Get("profile")

    sub := profile.(map[string]interface{})["sub"].(string)

    // encontrar el userID
    if err := db.First(&user, "sub = ?", sub).Error; err != nil {
        ctx.JSON(404, gin.H{
            "error": "El usuario no existe",
        })
        return 
    }

    if err := db.Preload("Tasks").Find(&user).Error; err != nil {
        ctx.JSON(500, gin.H{
        "error": err,
    }) 
    return
    }

    fmt.Print(user)

    ctx.HTML(http.StatusOK, "profile.html", gin.H{
        "User": user,
        "Foo": "foo",
    })
}


func GetUsers(ctx *gin.Context) {
        db := database.DB 
	    var users []models.User
	    db.Find(&users)
        ctx.JSON(200, gin.H{
       "users": users,
    }) 
}

func DeleteTask(ctx *gin.Context) {
	time.Sleep(1 * time.Second)

    db := database.DB
    var user models.User
    id := ctx.Param("id") 

    session := sessions.Default(ctx)
    profile := session.Get("profile")

    sub := profile.(map[string]interface{})["sub"].(string) 

    // encontrar el userID
    if err := db.First(&user, "sub = ?", sub).Error; err != nil {
        ctx.JSON(404, gin.H{
            "error": "El usuario no existe",
        })
        return 
    }

	var task models.Task
    db.First(&task, id)

    if user.ID != task.UserID {
        ctx.Status(403)
        return 
    }

    db.Delete(&task)
    ctx.Status(204)
}

func GetTasks(ctx *gin.Context) {
        db := database.DB 
	    var tasks []models.Task
	    db.Find(&tasks)

        // Para que muestre el modelo user
        if err := db.Preload("User").Find(&tasks).Error; err != nil {
            ctx.JSON(500, gin.H{
            "error": err,
        }) 
        return
        }
        
        ctx.JSON(200, gin.H{
       "tasks": tasks,

    }) 
}

func CreateTask(ctx *gin.Context) {
	time.Sleep(1 * time.Second)

    db := database.DB
    var user models.User

    session := sessions.Default(ctx)
    profile := session.Get("profile")

    sub := profile.(map[string]interface{})["sub"].(string) 

    if err := db.First(&user, "sub = ?", sub).Error; err != nil {
        ctx.JSON(404, gin.H{
            "error": "El usuario no existe",
        })
        return
        // redirect
    }

	name := ctx.PostForm("name")

	var task models.Task
    task = models.Task{
        Name: name, 
        UserID: user.ID, 
    } 	
	db.Create(&task)
}


func User(ctx *gin.Context) {
    db := database.DB
    var tasks []models.Task
    
    if err := db.Preload("User").Find(&tasks).Error; err != nil {
        ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    session := sessions.Default(ctx)
    profile := session.Get("profile")

    ctx.HTML(http.StatusOK, "user.html", gin.H{
        "tasks":   tasks,
        "profile": profile,
    })
}

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

func Callback(auth *auth.Authenticator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		if ctx.Query("state") != session.Get("state") {
			ctx.String(http.StatusBadRequest, "Invalid state parameter.")
			return
		}

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
    // print profile
    // agarar sub y nickname
    // Si existe pass
    // Si no existe crealo

		if err := session.Save(); err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}

		ctx.Redirect(http.StatusTemporaryRedirect, "/user")
	}
}

// Esta logica en el callback 
func CheckIfExists(ctx *gin.Context) {
    db := database.DB
    var user models.User

    session := sessions.Default(ctx)
    profile := session.Get("profile")

    sub := profile.(map[string]interface{})["sub"].(string) 
    nickname := profile.(map[string]interface{})["nickname"].(string) 

    if err := db.First(&user, "sub = ?", sub).Error; err != nil {
        user = models.User{
            Sub:      sub,
            Nickname: nickname,
        }
        db.Create(&user)
        ctx.JSON(200, gin.H{
            "message": "Usuario creado con éxito :D",
        })
    } else {
        ctx.JSON(200, gin.H{
            "error": "El usuario ya existe y no se creó",
        })
    }
}
