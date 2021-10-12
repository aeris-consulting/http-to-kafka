//    Copyright 2021 AERIS-Consulting e.U.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package handlers

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"github.com/spf13/cobra"
	"http-to-kafka/datapublisher"
	"log"
	"net/http"
)

const defaultUsername = "test"
const defaultPassword = "test"
const sessionName = "aeris-http-to-kafka-session"

// loginForm represents the details for a user to sign in.
type loginForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// sessionConfiguration is the full application configuration to manage the HTTP sessions.
type sessionConfiguration struct {
	secret string // Secret to create the HTTP session store.

	redisSession  bool   // Enables the storage of the HTTP sessions in Redis when enabled.
	redisUri      string // URI made of host:port to access to the single Redis instance to store the HTTP sessions.
	redisDatabase int    // Number of the Redis database to store the HTTP sessions.
	redisPassword string // Password to authenticate to Redis to store the HTTP sessions.
}

var (
	config struct {
		user    loginForm            // Default user credentials to sign into the application.
		session sessionConfiguration // Configuration of the application.
	}

	AnonymousHandlers     []gin.HandlerFunc // Handlers for calls to anonymous endpoints (login, logout...).
	AuthenticatedHandlers []gin.HandlerFunc // Handlers for calls to authenticated endpoints.

)

func init() {
	AnonymousHandlers = []gin.HandlerFunc{}
	AuthenticatedHandlers = []gin.HandlerFunc{verifySession}

	config.user = loginForm{}
	config.session = sessionConfiguration{}
}

// InitCommand configures the command-line options for the HTTP server.
func InitCommand(rootCommand *cobra.Command) {

	rootCommand.PersistentFlags().StringVar(&(config.user.Username), "username", defaultUsername, "username for the HTTP login")
	rootCommand.PersistentFlags().StringVar(&(config.user.Password), "password", defaultPassword, "password for the HTTP login")

	rootCommand.PersistentFlags().StringVar(&(config.session.secret), "session-secret", "", "secret for the for session store")

	rootCommand.PersistentFlags().BoolVar(&(config.session.redisSession), "session-redis", false, "enables the HTTP session persistence in Redis")
	rootCommand.PersistentFlags().StringVar(&(config.session.redisUri), "session-redis-uri", "localhost:6379", "URI to connect to Redis for the HTTP session persistence")
	rootCommand.PersistentFlags().IntVar(&(config.session.redisDatabase), "session-redis-database", 0, "index for the Redis database for the HTTP session persistence")
	rootCommand.PersistentFlags().StringVar(&(config.session.redisPassword), "session-redis-auth", "", "auth secret for the Redis database for the HTTP session persistence")
}

// ConfigureEngine prepares the gin.Engine to support to support requests and sessions.
func ConfigureEngine(router *gin.Engine) {
	var sessionSecret []byte
	if config.session.secret != "" {
		sessionSecret = []byte(config.session.secret)
	} else {
		sessionSecret = securecookie.GenerateRandomKey(32)
	}
	if config.session.redisSession {
		store, err := redis.NewStore(1000, "tcp", config.session.redisUri, config.session.redisPassword, sessionSecret)
		if err != nil {
			log.Fatalf("An error occured while connecting to Redis: %s", err.Error())
		}
		router.Use(sessions.Sessions(sessionName, store))
	} else {
		sessionStore := cookie.NewStore(sessionSecret)
		router.Use(sessions.Sessions(sessionName, sessionStore))
	}
}

// verifySession only lets request through, when a valid session is already open. Otherwise, a response with http.StatusUnauthorized is returned.
func verifySession(ctx *gin.Context) {
	session := sessions.Default(ctx)
	if session.Get("username") == nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	} else {
		ctx.Next()
	}
}

// Home returns the welcome message.
func Home(ctx *gin.Context) {
	ctx.String(200, "HTTP to Kafka, by AERIS-Consulting e.U.")
}

// Login proceeds with a login request, verifies the credentials and create a session.
func Login(ctx *gin.Context) {
	var loginData loginForm
	if err := ctx.Bind(&loginData); err == nil {
		if loginData == config.user {
			session := sessions.Default(ctx)
			session.Set("username", loginData.Username)
			session.Save()
			ctx.Redirect(http.StatusSeeOther, "/")
		} else {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login/password"})
		}
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// PushData publishes the received data.
func PushData(ctx *gin.Context) {
	message, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {

		messageKey := []byte(ctx.GetHeader("message-key"))
		for _, publisher := range datapublisher.RegisteredPublishers {
			publisher.Publish(messageKey, message)
		}
		ctx.JSON(http.StatusAccepted, gin.H{})
	}
}

// Logout clears the active session.
func Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	session.Clear()
	session.Save()
	ctx.Redirect(http.StatusSeeOther, "/")
}
