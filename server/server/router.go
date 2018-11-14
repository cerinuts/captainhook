package server

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// VersionPath is the base path for most API calls. This has to be changed if there are major changes in the API that break compatibility
const VersionPath = "/v1"

// ClientPath is the REST-path to manage clients
const ClientPath = VersionPath + "/clients"

// HookByUUIDPath is the REST-path to manage Hooks given by a UUID
const HookByUUIDPath = VersionPath + "/hookByUUID"

// HookPath is the REST-path to manage hooks
const HookPath = VersionPath + "/hooks"

// ConnectPath is the REST-path to where clients can connect to a websocket to receive webhooks
const ConnectPath = VersionPath + "/connect"

// ExternalHookPath is the path external applications will call to notify a webhook
const ExternalHookPath = "/h"

// ApplicationVersionPath will return the version of the CaptainHook server
const ApplicationVersionPath = "/version"

// SetupAPI will set up the HTTP-REST-API server without SSL. extPort is the Port the public interfaces will listen to, intPort will be used for private connection, e.g. the CLI
func SetupAPI(hostname string, extPort, intPort int, server *Server) {
	setupExternalRouter(hostname, extPort, 0, server, "", "")
	setupInternalRouter(intPort, server)
}

// SetupSSLAPI will set up the HTTP-REST-API server with SSL. extPort will redirect everything to HTTPS,
// extSSLPort is the Port the public interfaces will listen to, intPort will be used for private connection, e.g. the CLI
func SetupSSLAPI(hostname string, extPort, extSSLPort, intPort int, server *Server, sslCertFile, sslKeyFile string) {
	setupExternalRouter(hostname, extPort, extSSLPort, server, sslCertFile, sslKeyFile)
	setupInternalRouter(intPort, server)
}

func setupInternalRouter(internalPort int, server *Server) {
	intRouter := gin.Default()

	// get all clients
	intRouter.GET(ClientPath, func(c *gin.Context) {
		_, v := clientMapToSlice(server.Clients)
		c.JSON(http.StatusOK, v)
	})

	// create a new client or fail if name exists
	intRouter.POST(ClientPath+"/:name", func(c *gin.Context) {
		clientname := c.Param("name")
		secret, err := server.AddClient(clientname)
		if err != nil {
			switch err.(type) {
			case *ErrClientExists:
				{
					c.JSON(http.StatusBadRequest, errorToStruct(err))
					return
				}
			default:
				{
					c.JSON(http.StatusInternalServerError, errorToStruct(err))
					return
				}
			}
		}

		c.String(http.StatusOK, secret)
	})

	// generate a new secret in case the old one is lost
	intRouter.PATCH(ClientPath+"/:name", func(c *gin.Context) {
		clientname := c.Param("name")
		if server.Clients[clientname] == nil {
			c.Status(http.StatusNotFound)
			return
		}

		secret, err := server.Clients[clientname].generateSecret()
		if err != nil{
			c.JSON(http.StatusInternalServerError, errorToStruct(err))
		}
		server.DB.Store(server.Clients[clientname])
		c.String(http.StatusOK, secret)
	})

	// delete a client
	intRouter.DELETE(ClientPath+"/:name", func(c *gin.Context) {
		clientname := c.Param("name")
		err := server.RemoveClient(clientname)
		if err != nil {
			switch err.(type) {
			case *ErrClientNotExists:
				{
					c.JSON(http.StatusNotFound, errorToStruct(err))
					return
				}
			default:
				{
					c.JSON(http.StatusInternalServerError, errorToStruct(err))
					return
				}
			}
		}

		c.Status(http.StatusOK)
	})

	//create a new hook
	intRouter.PUT(HookPath+"/:client/:identifier", func(c *gin.Context) {

		identifier := c.Param("identifier")
		hook, err := server.AddHook(c.Param("client"), identifier)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorToStruct(err))
			return
		}

		c.JSON(http.StatusCreated, hook)
	})

	// delete any hook by uuid
	intRouter.DELETE(HookByUUIDPath+"/:uuid", func(c *gin.Context) {
		uuid, err := url.QueryUnescape(c.Param("uuid"))
		if err != nil {
			c.JSON(http.StatusBadRequest, errorToStruct(err))
			return
		}

		err = server.DeleteHookByUUID(uuid)
		if err != nil {
			switch err.(type) {
			case *ErrHookNotExists:
				{
					c.JSON(http.StatusNotFound, errorToStruct(err))
					return
				}
			default:
				{
					c.JSON(http.StatusInternalServerError, errorToStruct(err))
					return
				}
			}
		}

		c.Status(http.StatusOK)
	})

	intRouter.GET(ApplicationVersionPath, func(c *gin.Context) {
		c.String(http.StatusOK, ApplicationName+" "+FullVersion)
	})

	go func() {
		err := intRouter.Run("127.0.0.1:" + strconv.Itoa(internalPort))
		if err != nil {
			log.Print(err)
		}
	}()

}

func setupExternalRouter(hostname string, extPort, extSSLPort int, server *Server, certFile, keyFile string) {
	extRouter := gin.Default()

	// get all hooks for client
	extRouter.GET(HookPath, func(c *gin.Context) {
		if client, authorized := auth(c, server); authorized {
			_, h := clientHookMapToSlice(client.Hooks)
			c.JSON(http.StatusOK, h)
		}
	})

	// create a new hook or return existing
	extRouter.PUT(HookPath+"/:identifier", func(c *gin.Context) {
		if client, authorized := auth(c, server); authorized {

			identifier := c.Param("identifier")
			hook, err := server.AddHook(client.Name, identifier)
			if err != nil {
				c.JSON(http.StatusInternalServerError, errorToStruct(err))
				return
			}

			c.JSON(http.StatusCreated, hook)
		}
	})

	// delete a hook by identifier
	extRouter.DELETE(HookPath+"/:identifier", func(c *gin.Context) {
		if client, authorized := auth(c, server); authorized {

			identifier := c.Param("identifier")
			err := server.DeleteHook(client.Name, identifier)
			if err != nil {
				c.JSON(http.StatusInternalServerError, errorToStruct(err))
				return
			}

			c.Status(http.StatusOK)
		}
	})

	// handle webhooks
	extRouter.POST(ExternalHookPath+"/:hook", func(c *gin.Context) {
		if err := server.HandleHook(c.Param("hook"), c.Request); err != nil {
			c.Status(http.StatusBadGateway)
			return
		}

		c.Status(http.StatusOK)
	})

	extRouter.GET(ConnectPath, func(c *gin.Context) {
		if client, authorized := auth(c, server); authorized {
			client.OpenWebsocket(c)
		}
	})

	start(extRouter, hostname, extPort, extSSLPort, certFile, keyFile)
}

func start(extRouter *gin.Engine, hostname string, extPort, extSSLPort int, certFile, keyFile string) {
	if certFile != "" && keyFile != "" {
		httpRouter := gin.Default()
		httpRouter.Any("*path", func(c *gin.Context) {
			u := c.Request.URL
			u.Host = strings.Split(c.Request.Host, ":")[0] + ":" + strconv.Itoa(extSSLPort)
			u.Scheme = "https"
			c.Redirect(307, u.String())
		})

		go func() {
			err := httpRouter.Run(hostname + ":" + strconv.Itoa(extPort))
			if err != nil {
				log.Fatal(err)
			}
		}()
		go func() {
			err := extRouter.RunTLS(hostname+":"+strconv.Itoa(extSSLPort), certFile, keyFile)
			if err != nil {
				log.Fatal(err)
			}
		}()
	} else {
		go func() {
			err := extRouter.Run(hostname + ":" + strconv.Itoa(extPort))
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

func auth(c *gin.Context, server *Server) (*Client, bool) {
	clientsecret := c.GetHeader("Authorization")

	if strings.HasPrefix(clientsecret, "Bearer ") {
		client := server.validateClient(strings.TrimPrefix(clientsecret, "Bearer "))
		if client == nil {
			c.Status(http.StatusForbidden)
			return nil, false
		}

		return client, true
	}

	c.Status(http.StatusForbidden)

	return nil, false
}

// Error is a simple error message struct used to represent an error in the api
type Error struct {
	Message string `json:"message"`
}

func errorToStruct(e error) Error {
	return Error{
		Message: e.Error(),
	}
}

func clientMapToSlice(inmap map[string]*Client) ([]string, []*Client) {
	keys := make([]string, 0)
	values := make([]*Client, 0)
	for k, v := range inmap {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}

func clientHookMapToSlice(inmap map[string]*Webhook) ([]string, []*Webhook) {
	keys := make([]string, 0)
	values := make([]*Webhook, 0)
	for k, v := range inmap {
		keys = append(keys, k)
		values = append(values, v)
	}
	return keys, values
}
