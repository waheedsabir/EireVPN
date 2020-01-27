package router

import (
	"eirevpn/api/config"
	"eirevpn/api/errors"
	"eirevpn/api/logger"
	"eirevpn/api/models"
	"eirevpn/api/server"
	"eirevpn/api/settings"
	"eirevpn/api/util/jwt"
	"io/ioutil"

	"eirevpn/api/plan"
	"eirevpn/api/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const secretkey = "verysecretkey1995"

func Init(logging bool) *gin.Engine {

	conf := config.GetConfig()

	var router *gin.Engine
	if logging {
		router = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = ioutil.Discard
		router = gin.New()
		router.Use(gin.Recovery())
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = conf.App.AllowedOrigins
	corsConfig.AllowCredentials = true
	corsConfig.ExposeHeaders = []string{"X-CSRF-Token", "X-Auth-Token"}
	corsConfig.AddAllowHeaders("Origin", "Content-Length", "Content-Type", "Authorization", "X-CSRF-Token", "X-Auth-Token")
	router.Use(cors.New(corsConfig))

	public := router.Group("/api")
	private := router.Group("/api/private")
	protected := router.Group("/api/protected")
	private.Use(auth(secretkey, false))
	protected.Use(auth(secretkey, true))

	public.POST("/user/signup", user.SignUpUser)
	public.POST("/user/login", user.LoginUser)
	protected.GET("/user/:id", user.User)
	protected.PUT("/user/update/:id", user.UpdateUser)
	protected.GET("/users", user.AllUsers)
	public.POST("/user/webhook", user.Webhook)
	private.GET("/user/updatepayment", user.StripeUpdatePaymentSession)
	private.GET("/user/session/:planid", user.StripeSession)
	private.GET("/user/cancel", user.CancelSubscription)

	protected.GET("/plans/:id", plan.Plan)
	protected.POST("/plans/create", plan.CreatePlan)
	protected.PUT("/plans/update/:id", plan.UpdatePlan)
	protected.DELETE("/plans/delete/:id", plan.DeletePlan)
	public.GET("/plans", plan.AllPlans)

	protected.GET("/servers/:id", server.Server)
	protected.POST("/servers/create", server.CreateServer)
	protected.PUT("/servers/update/:id", server.UpdateServer)
	protected.DELETE("/servers/delete/:id", server.DeleteServer)
	private.GET("/servers/connect/:id", server.Connect)
	private.GET("/servers", server.AllServers)

	protected.GET("/settings", settings.Settings)
	protected.PUT("/settings/update", settings.UpdateSettings)

	router.Static("/assets", "./assets")
	return router
}

func auth(secret string, protected bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		conf := config.GetConfig()
		if conf.App.EnableAuth {
			var usersession models.UserAppSession

			// Fetch authentification token
			authToken := c.GetHeader("X-Auth-Token")
			if authToken == "" {
				logger.Log(logger.Fields{
					Loc:  "router.go - auth()",
					Code: errors.AuthCookieMissing.Code,
				})
				c.AbortWithStatusJSON(errors.AuthCookieMissing.Status, errors.AuthCookieMissing)
				return
			}

			// Fetch refresh token
			refreshToken, err := c.Request.Cookie("refreshToken")
			if err != nil {
				logger.Log(logger.Fields{
					Loc:  "router.go - auth()",
					Code: errors.RefresCookieMissing.Code,
					Err:  err.Error(),
				})
				c.AbortWithStatusJSON(errors.RefresCookieMissing.Status, errors.RefresCookieMissing)
				return
			}
			// Check auth token is valid
			authClaims, err := jwt.ValidateToken(authToken)
			if err != nil {
				fmt.Println("auth invalid checking refresh")
				// If auth token is invalid check refresh token is valid
				refreshClaims, err := jwt.ValidateToken(refreshToken.Value)
				if err != nil {
					logger.Log(logger.Fields{
						Loc:  "router.go - auth()",
						Code: errors.TokenInvalid.Code,
						Err:  err.Error(),
					})
					c.AbortWithStatusJSON(errors.TokenInvalid.Status, errors.TokenInvalid)
					return
				}

				usersession = models.UserAppSession{
					UserID:     refreshClaims.UserID,
					Identifier: refreshClaims.SessionIdentifier,
				}

				if err := usersession.Find(); err != nil {
					logger.Log(logger.Fields{
						Loc:  "router.go - auth()",
						Code: errors.InvalidIdentifier.Code,
						Err:  err.Error(),
					})
					c.SetCookie("refreshToken", "", -1, "/", "localhost", true, false)
					c.AbortWithStatusJSON(errors.InvalidIdentifier.Status, errors.InvalidIdentifier)
					return
				}
			}

			// If auth token or refresh token is valid check if crsf token matches the one supplied
			// in the header
			if conf.App.EnableCSRF {
				if authClaims == nil || authClaims.CSRF != c.GetHeader("X-CSRF-Token") {
					var reason string
					if authClaims.CSRF == "" {
						reason = "CSRF token is missing from claims"
					}
					if c.GetHeader("X-CSRF-Token") == "" {
						reason = "CSRF token is missing from header"
					}
					if authClaims == nil {
						reason = "Auth Claims is nil"
					}
					logger.Log(logger.Fields{
						Loc:   "router.go - auth()",
						Code:  errors.CSRFTokenInvalid.Code,
						Extra: map[string]interface{}{"auth-CSRF": authClaims.CSRF, "head-CSRF": c.GetHeader("X-CSRF-Token")},
						Err:   reason,
					})
					c.AbortWithStatusJSON(errors.CSRFTokenInvalid.Status, errors.CSRFTokenInvalid)
					return
				}
			}

			if usersession == (models.UserAppSession{}) {
				usersession = models.UserAppSession{
					UserID: authClaims.UserID,
				}
			}

			if protected {
				var user models.User
				user.ID = usersession.UserID
				if err := user.Find(); err != nil {
					logger.Log(logger.Fields{
						Loc: "router.go - auth()",
						Extra: map[string]interface{}{
							"UserID": usersession.UserID,
							"Detail": "User Not found when checking user type",
						},
						Err: err.Error(),
					})
					c.AbortWithStatusJSON(errors.InternalServerError.Status, errors.InternalServerError)
					return
				}

				if user.Type != models.UserTypeAdmin {
					logger.Log(logger.Fields{
						Loc:  "router.go - auth()",
						Code: errors.ProtectedRouted.Code,
						Extra: map[string]interface{}{
							"UserID": usersession.UserID,
						},
						Err: "User does not have permission to access route",
					})
					c.AbortWithStatusJSON(errors.ProtectedRouted.Status, errors.ProtectedRouted)
					return
				}
			}

			// create a new user session
			var newUserSession models.UserAppSession
			if err := newUserSession.New(usersession.UserID); err != nil {
				logger.Log(logger.Fields{
					Loc:   "/login - LoginUser() - Create session",
					Code:  errors.InternalServerError.Code,
					Extra: map[string]interface{}{"UserID": usersession.UserID},
					Err:   err.Error(),
				})
				c.AbortWithStatusJSON(errors.InternalServerError.Status, errors.InternalServerError)
				return
			}

			// If all auth checks pass create fresh tokens
			newAuthToken, newRefreshToken, newCsrfToken, err := jwt.Tokens(newUserSession)
			if err != nil {
				logger.Log(logger.Fields{
					Loc:   "router.go - auth()",
					Code:  errors.InternalServerError.Code,
					Extra: map[string]interface{}{"UserID": newUserSession.UserID},
					Err:   err.Error(),
				})
				c.AbortWithStatusJSON(errors.InternalServerError.Status, errors.InternalServerError)
				return
			}

			// Add user id to the context for use within the routes
			c.Set("UserID", newUserSession.UserID)

			// TODO: Change the domain name and add correct maxAge time
			// authCookieMaxAge := 15 * 60 // 15 minutes in seconds
			// c.SetCookie("authToken", newAuthToken, authCookieMaxAge, "/", conf.App.Domain, false, false)
			c.Header("X-Auth-Token", newAuthToken)

			// TODO: Change the domain name and add correct maxAge time
			refreshCookieMaxAge := 24 * 60 * 60 // 72 hours in seconds
			c.SetCookie("refreshToken", newRefreshToken, refreshCookieMaxAge, "/", conf.App.Domain, false, false)
			c.Header("X-CSRF-Token", newCsrfToken)
		}
	}
}
