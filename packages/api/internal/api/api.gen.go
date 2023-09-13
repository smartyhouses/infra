// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.15.0 DO NOT EDIT.
package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (POST /envs)
	PostEnvs(c *gin.Context)

	// (GET /health)
	GetHealth(c *gin.Context)

	// (POST /instances)
	PostInstances(c *gin.Context)

	// (POST /instances/{instanceID}/refreshes)
	PostInstancesInstanceIDRefreshes(c *gin.Context, instanceID InstanceID)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandler       func(*gin.Context, error, int)
}

type MiddlewareFunc func(c *gin.Context)

// PostEnvs operation middleware
func (siw *ServerInterfaceWrapper) PostEnvs(c *gin.Context) {

	c.Set(ApiKeyAuthScopes, []string{})

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.PostEnvs(c)
}

// GetHealth operation middleware
func (siw *ServerInterfaceWrapper) GetHealth(c *gin.Context) {

	c.Set(ApiKeyAuthScopes, []string{})

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.GetHealth(c)
}

// PostInstances operation middleware
func (siw *ServerInterfaceWrapper) PostInstances(c *gin.Context) {

	c.Set(ApiKeyAuthScopes, []string{})

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.PostInstances(c)
}

// PostInstancesInstanceIDRefreshes operation middleware
func (siw *ServerInterfaceWrapper) PostInstancesInstanceIDRefreshes(c *gin.Context) {

	var err error

	// ------------- Path parameter "instanceID" -------------
	var instanceID InstanceID

	err = runtime.BindStyledParameter("simple", false, "instanceID", c.Param("instanceID"), &instanceID)
	if err != nil {
		siw.ErrorHandler(c, fmt.Errorf("Invalid format for parameter instanceID: %w", err), http.StatusBadRequest)
		return
	}

	c.Set(ApiKeyAuthScopes, []string{})

	for _, middleware := range siw.HandlerMiddlewares {
		middleware(c)
		if c.IsAborted() {
			return
		}
	}

	siw.Handler.PostInstancesInstanceIDRefreshes(c, instanceID)
}

// GinServerOptions provides options for the Gin server.
type GinServerOptions struct {
	BaseURL      string
	Middlewares  []MiddlewareFunc
	ErrorHandler func(*gin.Context, error, int)
}

// RegisterHandlers creates http.Handler with routing matching OpenAPI spec.
func RegisterHandlers(router gin.IRouter, si ServerInterface) {
	RegisterHandlersWithOptions(router, si, GinServerOptions{})
}

// RegisterHandlersWithOptions creates http.Handler with additional options
func RegisterHandlersWithOptions(router gin.IRouter, si ServerInterface, options GinServerOptions) {
	errorHandler := options.ErrorHandler
	if errorHandler == nil {
		errorHandler = func(c *gin.Context, err error, statusCode int) {
			c.JSON(statusCode, gin.H{"msg": err.Error()})
		}
	}

	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandler:       errorHandler,
	}

	router.POST(options.BaseURL+"/envs", wrapper.PostEnvs)
	router.GET(options.BaseURL+"/health", wrapper.GetHealth)
	router.POST(options.BaseURL+"/instances", wrapper.PostInstances)
	router.POST(options.BaseURL+"/instances/:instanceID/refreshes", wrapper.PostInstancesInstanceIDRefreshes)
}
