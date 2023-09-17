package backend

import (
	"fmt"
	"highload-arch/pkg/backend/endpoints"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/",
		Index,
	},

	Route{
		"LoginPost",
		strings.ToUpper("Post"),
		"/login",
		endpoints.LoginPost,
	},

	Route{
		"UserGetIdGet",
		strings.ToUpper("Get"),
		"/user/get/{id}",
		endpoints.UserGetIdGet,
	},

	Route{
		"UserRegisterPost",
		strings.ToUpper("Post"),
		"/user/register",
		endpoints.UserRegisterPost,
	},
}
