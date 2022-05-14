# STGIN

STgin is a functional rest framework that provides easy APIs in order to maintain your application RESTful API server.

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
health := stgin.POST("/health", func(request stgin.RequestContext) stgin.Status {
    var reqBody HealthCheckRequest
    request.Body.JSONInto(&reqBody)
    // do something with reqBody
    var response HealthCheckResponse = GetHealth()
    if response.DBConnection {
        return stgin.Ok(stgin.Json(&response)) 
    } else {
        return stgin.InternalServerError(stgin.Json(&response))
    }
})
```
***common framework implementation:***
```go
r.POST("/health", func(c *framework.Context) {
    var reqBody HealthCheckRequest
    bodyBytes, err := ioutil.ReadAll(c.Request.Body)
    if err != nil {
        // potential error handling
    }
    err = json.Unmarshal(bodyBytes, &reqBody)
    if err != nil {
        // potential error handling
    }
	// do something with reqBody
    var response HealthCheckResponse = GetHealth()
    jsonResponse, _ := json.Marshal(response)
    if response.DBConnection {
    	_, writeError := c.Writer.Write(jsonResponse)
        c.Status(200)
        if writeError != nil {
            // potential error handling
        }
    } else {
    	c.Status(500)
    	_, writeError = c.Writer.Write(jsonResponse)
        if writeError != nil {
            // potential error handling
        }
    }
})
```
Or just easily add headers or cookies with a receiver function instead of manually writing:
```go
stgin.Ok(...).WithHeaders(...).WithCookies
```

## Structure

The structure of STgin types and interfaces is pretty simple, a `Server` may have several `Controller`s, and each controller may have serveral `Route`s.
```
    
    -Server =>
        -Controller 1 ->
            -Route 1 (path pattern, method, api)
            -Route 2 (path pattern, method, api)
        -Cotroller 2 ->
            -Route 1 (path pattern, method, api)
```
**Server:** Is run on the specified port, contains the controllers.

**Controller:** Contains routes which are exposed to the server, has a name, and may have a route prefix (i.e., /home)

**Route:** Holds route specifications (i.e., method, path, API action)

**RequestContext**: Holds the information about the requests, such as uri, body, headers, ...

**Status:** Is a wrapper around an actual http response, holds status code, response headers, response body, ... (i.e., Ok, BadRequest, ...)

**API:** Is a type alias for a function which accepts a request context and returns a status.

# Path Parameters
* How to define?

    When defining a route, there are 2 possible values between any 2 slashes, a literal string (like ".../home/..."), or a path parameter specification.
    Path parameters must have a name, and an optional type which the parameter might have (string is used as default, if no type or an invalid type is provided).
  
    ```
            //same as $username:string
    stgin.GET("/users/$username/purchases/$purchase_id:int". ...)
  
    // "/users/$username/purchases/$purchase_id" also accepts the correct route like "/users/John/purchases/12",
    // but in case an alphanumeric value is passed as purchase_id, conversion from string to int in the api is manual
    // and needs manual error checking
    ```
* How to get?
    ```
    username, exists := request.GetPathParam("username")
    // or if you're sure about the existence, 
    username := request.MustGetPathParam("username")
    ```
    
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

# Listeners
Listeners are functions, which can affect the request and response based on the defined behavior.
For instance, a `ResponseListener` is a function which receives a response, and returns a response, it can be used when you want to apply something to all the responses in server layer or controller layer (i.e., providing CORS headers).
There are 3 types of listeners:
* RequestListener: func(RequestContext) RequestContext [Can be used to mutate request before the controller receives it]
* ResponseListener: func(Status) Status [Can be used to add/remove additional information to a raw controller response]
* APIListener: func(RequestContext, Status) [Can be used to do stuff like logging, ...]

There are some listeners provided inside the STgin package which can be used inside a server or a controller [API watcher/logger, recovery].

# Custom Recovery
An `ErrorHandler` can be provided by the developer, to provide custom error handling behavior.
Definition of an `ErrorHandler` function is pretty straight forward, you just define a function which takes the request and the error, and decides what to return as the status.
```go
var myErrorHandler stgin.ErrorHandler = func(request RequestContext, err any) stgin.Status {
    if myCustomErr, isCustomErr := err.(CustomErr); isCustomErr {
        return stgin.BadRequest(...)
    } else {
        return stgin.InternalServerError(...)
    }
}
```

# Files
Working with files is easy in stgin. you can simply use `stgin.File`, and everything is all set.
It returns `404 not found` if the file does not exist, or `500 internal server error` in case of problems reading the file.

# Todos
* Add static file support
* Add Quick start, installation in readme
* Add html template integrations
* http 2 server push