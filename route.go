package stgin

import "strings"

type Route struct {
	Path	 	string
	Method      string
	Action 		API
}

func GET(path string, api API) Route {
	return Route{
		Path: path,
		Method: "GET",
		Action: api,
	}
}

func PUT(path string, api API) Route {
	return Route{
		Path: path,
		Method: "PUT",
		Action: api,
	}
}

func POST(path string, api API) Route {
	return Route{
		Path: path,
		Method: "POST",
		Action: api,
	}
}

func DELETE(path string, api API) Route {
	return Route{
		Path: path,
		Method: "DELETE",
		Action: api,
	}
}

func PATCH(path string, api API) Route {
	return Route{
		Path: path,
		Method: "PATCH",
		Action: api,
	}
}

func OPTIONS(path string, api API) Route {
	return Route{
		Path: path,
		Method: "OPTIONS",
		Action: api,
	}
}

type RouteCreationStage struct {
	method 		string
	path 		string
}

func (stage RouteCreationStage) Do(api API) Route {
	switch strings.ToUpper(stage.method) {
	case "GET":
		return GET(stage.path, api)
	case "PUT":
		return PUT(stage.path, api)
	case "POST":
		return POST(stage.path, api)
	case "DELETE":
		return DELETE(stage.path, api)
	case "PATCH":
		return PATCH(stage.path, api)
	default:
		return GET(stage.path, api)
	}
}

func OnGET(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "GET",
		path: path,
	}
}

func OnPUT(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "PUT",
		path:   path,
	}
}

func OnPOST(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "POST",
		path: path,
	}
}

func OnDelete(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "DELETE",
		path:   path,
	}
}

func OnPatch(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "PATCH",
		path: path,
	}
}

func OnOptions(path string) RouteCreationStage {
	return RouteCreationStage{
		method: "OPTIONS",
		path: path,
	}
}

func OnPath(path string) RouteCreationStage {
	return RouteCreationStage{path: path}
}

func (stage RouteCreationStage) WithMethod(method string) RouteCreationStage {
	stage.method = method
	return stage
}