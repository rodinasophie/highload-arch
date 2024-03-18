package routes

import (
	"fmt"
	"highload-arch/pkg/common"
	"highload-arch/pkg/counters_service/endpoints"
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

const PREFIX_V2 = "/api/v2"

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		common.Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to highload architecture homework!")
}

var routes = Routes{
	Route{
		"Index",
		"GET",
		PREFIX_V2,
		Index,
	},

	Route{
		"CountersGetUnreadMessages",
		strings.ToUpper("Get"),
		PREFIX_V2 + "/counters/{user_id}/unreadMessages",
		endpoints.CountersGetUnreadMessages,
	},
}
