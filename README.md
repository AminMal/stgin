
# STGIN

stgin is a functional RESTful API server framework that provides easy APIs in order to maintain your application RESTful API server.

# Installation
You can use the following command to add stgin into your application.
```
go get github.com/AminMal/stgin
```

# Quick Start
stgin concentrates a lot about making a well-structured application.
Let's take a look at the following application structure, this is just a part of a simple application (health checking part).
```
  /project
    ... other packages you might need (i.e., users, authentication, inventory, ...)
    /health
      init.go
      models.go
      apis.go
      ... other things you might need
    init.go
    main.go
```
Let's start defining health APIs:
```go
// /health/apis.go file

package health

import "github.com/AminMal/stgin"
// #recommended
func getHealthAPI(request stgin.RequestContext) stgin.Status {
    healthResponse := getHealth() // some method implemented in package
    // model is defined earlier in models.go
    if healthResponse.ok {
        return stgin.Ok(stgin.Json(&healthResponse))
    } else {
        return stgin.InternalServerError(stgin.Json(&healthResponse))
    }
}
// you can also define APIs as variable:
// var someAPI stgin.API = func(request stgin.RequestContext) stgin.Status {...}

// but using functions is more convenient and is recommended
// or inline the implementation in route definition #not recommended

```
Fine for now, let's create the controller in `/health/init.go`. It's better to use init to initialize controllers and apis.
```go
// health/init.go file

package health

import "github.com/AminMal/stgin"

var Controller *stgin.Controller // exposed to main package

func init() {
    Controller = stgin.NewController("HealthController", "/health")
    Controller.AddRoutes(
      stgin.GET("/status", getHealthAPI), // defined in apis.go
      // this route will be interpreted to /health/status
    )
}
```
So the health APIs are implemented, let's integrate them into the server.
Now in the root directory of the project, let's open `init.go` file.
```go
package main

import "github.com/AminMal/stgin"
import "fmt"

var server *stgin.Server

func init() {
        portNo := 9000
        server = stgin.DefaultServer(fmt.Sprintf(":%d", portNo))
        // default server included default logger and error handler
        server.Register(health.Controller)
}
```
Almost done, let's just run the server (main.go).
```go
package main

import "log"

func main() {
        log.Fatal(server.Start()) // all done
}
```
Your application will be up and running.

Another approach to define routes is route builder. You might want to use some method which is not defined in stgin default routing methods.
You can use:
```go
stgin.OnPattern("/your/path/$path_param").WithMethod("something").Do(
        func(req stgin.RequestContext) stgin.Status{...},
)
```

A comparison between stgin and gin-gonic, writing a simple API:
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
***stgin implementation:***
```go
health := stgin.POST("/health", func(request stgin.RequestContext) stgin.Status {
    var reqBody HealthCheckRequest
    request.Body().JSONInto(&reqBody)
    // do something with reqBody
    var response HealthCheckResponse = GetHealth()
    if response.DBConnection {
        return stgin.Ok(stgin.Json(&response)) 
    } else {
        return stgin.InternalServerError(stgin.Json(&response))
    }
})
```
***gin implementation:***
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
stgin.Ok(...).
      WithHeaders(...).
      WithCookies(...)
```

## Structure

The structure of stgin types and interfaces is pretty simple, a `Server` may have several `Controller`s, and each controller may have serveral `Route`s.
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

**RequestContext**: Holds the information about the requests, such as uri, body, headers, ... . Can parse request entity into the desired variable, using helper functions like `request.JSONInto`, `request.XMLInto`.

**Status:** Is a wrapper around an actual http response, holds status code, response headers, response body, ... (i.e., Ok, BadRequest, ...)
* ResponseEntity: A response could be of different content types (i.e., JSON, XML, Text, file, ...). A response entity is an interface which defines the content type of the response, and entity bytes. There are some helper functions provided in stgin to ease the use of these entities, like `stgin.Json`, `stgin.Text`, `stgin.Xml`, `stgin.File`.

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
    username, exists := request.PathParams.Get("username")
    // or if you're sure about the existence, 
    username := request.PathParams.MustGet("username")
    // or if you have specified the type in path pattern
    purchaseId := request.PathParams.MustGetInt("purchase_id")
    // or
    purchaseId, err := request.PathParams.GetInt("purchase_id")
    ```
  
# Query Parameters
* When to define?

  Define query parameter specifications only for mandatory and non-empty query parameter expectations.


* How to define?

  When defining a route, you can specify which query params of what type the route should expect. If a request could not satisfy the expected queries, it will be rejected by the route and will be passed into the next route and so on.
  Specifying query parameters does not mean that the route will not accept other query parameters which are not specified.
  By specifying the query parameters, you just make sure that when a request is accepted by the route, it always contains those query parameters with the right type.
  After defining the path pattern, use a question mark `?` to start defining query parameters, write the name of the parameter (if it has a certain type, use `:` and put the type name, i.e., int, string, float),
  and when done, put `&` to define the next query parameter. The order of queries does not matter.
  ```go
    stgin.GET("/users?username&id:int")
    // Just like path parameters, if you do not specify the type of query, it's assumed to be string, 
    // so "username" here is interpreted into "username:string" 
   ```
  As mentioned earlier, this pattern will match urls like `/users?id=7&username=JohnDoe&otherquery=whatever&anotherone=true`. 
  And you can access those easily in the request, so no worries about not specifying all the query parameters.
  
* How to get?

  Just like path parameters, query parameters follow the same rules for receiver functions.
  ```go
    userId := request.QueryParams.MustGetInt("user_id")
    // so on, just like path parameters
  ```

* Query to object

  There is a special method in request context, which can convert queries into a struct object.
  There are some few notes to take before using this. When defining the expected struct that the queries will be converted into,
  if you need to use other naming in queries rather than the field name in struct, use `qp` (short for query parameter) tag to specify the name (just like json tag):
  ```go
    type UserSearchFilter struct {
        Username   string  `qp:"name"`
        Id         int     `qp:"id"`
        Joined     string  
    }
  ```
  * **Always pass pointers to the function**
  * **Non-exported fields will not be parsed from request query parameters**
  * **If you do not pass the name to qp tag, parser would look up for the actual field name in the queries:**
    Notice the `Joined` field in the struct, parser looks for `&Joined=...` in the url.
    
## Custom parameter patterns
stgin provides patterns for `int`, `string`, `float` and `uuid` already. but If you need to add some custom pattern to use in query or path parameter definition, you can add the name with a valid pattern(regex) that gets compiled. 
```go
startsWithJohnRegex := "^john.*"
err := AddMatchingPattern("john", startsWithJohnRegex)
```
the above function will fail with errors if:
* 1) Try to override default patterns (`int`, `string`, `float`, `uuid`)
* 2) The given pattern couldn't be compiled
-----
## Custom Actions
stgin does not provide actions about stuff like Authentication, because simple authentication is not useful most of the time, and you may need customized authentications.

For instance:
```go
type AuthInfo struct {
        Username    string      `json:"username"`
        AccountId   int         `json:"account_id"`
        Roles       []string    `json:"roles"`
}

func authenticate(rc stgin.RequestContext) (AuthInfo, bool) {
    if name, found := rc.QueryParams.Get("user"); !found {
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
        return stgin.Ok(stgin.Json(&Greet(authInfo.Username)))
    })
})

```

# Listeners
Listeners are functions, which can affect the request and response based on the defined behavior.
For instance, a `ResponseListener` is a function which receives a response, and returns a response, it can be used when you want to apply something to all the responses in server layer or controller layer (i.e., providing CORS headers).
There are 3 types of listeners:
* RequestListener: `RequestContext => RequestContext` [Can be used to return a new modified instance of the request]

  ```go
    func AddUserTrackingKey(request stgin.RequestContext) stgin.RequestContext {
      request.Headers["X-Tracking-Key"] = []string{"some-generated-header"}
      return request
    }
  ```
* ResponseListener: `Status => Status` [Can be used to add/remove additional information to a new instance of Status]

  ```go
    func AddUserKeyToResponse(response stgin.Status) stgin.Status {
      return response.WithHeaders(http.Header{"X-Tracking-Key": {"<Some random value>"}}
    }
  ```
* APIListener: `(RequestContext, Status) => void` [Can be used to do stuff like logging, ...]

  ```go
    func ApiLogger(request stgin.RequestContext, response stgin.Status) {
      fmt.Println("received", request, "returned", response)
    }
  ```
  There are some listeners provided inside the stgin package which can be used inside a server or a controller (API watcher/logger, recovery, they're used in `stgin.DefaultServer` as well).

  
# CORS Handling
Cors handling can be done using server APIs:
```go
server.CorsHandler(stgin.CorsHandler{
        AllowOrigin: []string{"first_host.com", "second_host.com"},
        AllowHeaders: []string{"*"},
        AllowMethods: []string{"GET", "PUT", "POST", "HEAD", "OPTIONS"},
        ...
})
```


# Timeout
You can adjust server request timeout using `server.SetTimeout`,
this way if the request processing takes longer than the specified timeout,
the server will automatically abort the request and complete with a `408 request timed out` response.

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

# Files And Directories
**Files:** 

Working with files and directories is pretty easy. 
They are dealt just as normal response entities. They have a content type depending on the file format, and file bytes.
So you can return them inside your APIs, just give stgin the file location. If the file could not be found, `404 not found` is returned to the client as the response, and if there was some problems reading the file, `500 internal server error` would be returned.

**Directories:**

Directories are a bit out of RESTful APIs concept, so It's not possible in stgin to return them as an http response.
Instead of defining routes for the file system, a special Routing function is available as `StaticDir`:
```go
SomeController.AddRoutes(...) // some routes
SomeController.AddRoutes(
    stgin.GET("/some/pattern", someAPI),
    stgin.StaticDir("/get_my_files", "/tmp"),
    // serves /tmp folder on "</controller_prefix_if_exists>/get_my_files"
)
```
# Http 2 Push
Http push is available if you're using go 1.18 above, and using http 2 as a communication protocol.
```go
// inside api definiction
if request.HttpPush.IsSupported {
        pusher := request.HttpPush.Pusher
        // do stuff with pusher
}
```

# Html Templates
STgin does not support "rendering" html templates (parsing and creating appropriate sources for images, etc. developers should take care of images),
but loading html templates using variables is supported.

* **Template variables:** To define template variables, wrap the variable name inside double curly braces. (i.e., {{ title }}, {{name}}, spaces are ignored).

```go
//    /
//    /templates
//      welcome.html
//      /images
stgin.GET("/welcome", func (request stgin.RequestContext) stgin.Status {
    return stgin.Ok(template.LoadTemplate("./templates/welcome.html", template.Variables{
        "title": "Welcome to the team",
        "name": "John Doe",
}))
})
```
WIP