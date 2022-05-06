#STGIN

STgin is a functional rest framework that provides easy APIs in order to maintain your application RESTful APIS.

It is currently built upon go-gin framework, but has the ability to easily switch between any other library/framework using the translator function in server file.

A comparison between STgin and go-gin, writing a simple API:
```go
// Given response type as
type HealthCheckResponse struct {
	DBConnection    bool    `json:"database_connection"`
	Message         string  `json:"message"`
}
// and request type as
type HealthCheckRequest struct {
	Whatever        string   `json:"whatever"`
}
```
***STgin implementation:***
```go
health := healthController.POST("/health", func(request stgin.RequestContext) stgin.Status {
    var reqBody HealthCheckRequest
    request.Body.WriteInto(&reqBody)
    // do something with reqBody
    var response HealthCheckResponse = GetHealth()
    if response.DBConnection {
        return stgin.Ok(&response) 
    } else {
        return stgin.InternalServerError(&response)
    }
})
```
***go-gin implementation:***
```go
r.POST("/health", func(c *gin.Context) {
    var reqBody HealthCheckRequest
    bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	// potential error handling
    err = json.Unmarshal(bodyBytes, &reqBody)
	// potential error handling
	// do something with reqBody
    var response HealthCheckResponse = GetHealth()
    jsonResponse, _ := json.Marshal(response)
    if response.DBConnection {
    	c.Status(200)
    	_, writeError := c.Writer.Write(jsonResponse)
    	// potential error handling
    } else {
    	c.Status(500)
    	_, writeError = c.Writer.Write(jsonResponse)
    	// potential error handling
    }
})
```
Or just easily add headers with a receiver function instead of manually writing headers:
```go
stgin.Ok(&body).WithHeaders(...)
```

## Structure

The structure of STgin types and interfaces is pretty simple, a `Server` may have several `Controller`s, and each controller may have serveral `Route`s.
```
    
    -Server =>
        -Controller 1 ->
            -Route 1
            -Route 2
        -Cotroller 2 ->
            -Route 1
```
**Server:** Is run on the specified port, contains the controllers.

**Controller:** Contains routes which are exposed to the server, has a name, and may have a route prefix (i.e., /home)

**Route:** Holds route specifications (i.e., method, path, API action)

**RequestContext**: Holds the information about the requests, such as uri, body, headers, ...

**Status:** Is a wrapper around an actual http response, holds status code, response headers, response body, ... (i.e., Ok, BadRequest, ...)

**API:** Is a type alias for a function which accepts a request context and returns a status.

## Custom Actions
STgin does not provide actions about stuff like Authentication, because simple authentication is not useful most of the time, and you may need customized authentications.

For instance:
```go
type AuthInfo struct {
	Username    string      `json:"username"`
	AccountId   int         `json:"account_id"`
	Roles       []string    `json:"roles"`
}

func authenticate(rc stgin.RequestContext) (AuthInfo, bool) {
    if name, found := rc.GetQuery("user"); !found {
    	...
    } else {
        ...
    }
}
// written once in your base package
func Authenticated(rc stgin.RequestContext, action func(AuthInfo) stgin.Status) stgin.Status {
    authInfo, isAuthenticated := authenticate(rc)
    if !isAuthenticated {
        return stgin.Unauthorized(...)
    } else {
        return action(authInfo)
    }
}

// In the apis section
myAPI := stgin.GET("/test", func(request stgin.RequestContext) stgin.Status {
    return Authenticated(request, func(authInfo AuthInfo) stgin.Status {
        return stgin.Ok(&Greet(authInfo.Username))
    })
})

```

#TODOs
* Add most common statuses as predefined functions
* Add support for cookies
