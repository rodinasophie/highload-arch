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

	router.HandleFunc("/post/feed/posted", endpoints.PostFeedGetWebsocket)
	return router
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to highload architecture homework!")
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

	Route{
		"UserSearchGet",
		strings.ToUpper("Get"),
		"/user/search",
		endpoints.UserSearchGet,
	},

	Route{
		"FriendAddPut",
		strings.ToUpper("Put"),
		"/friend/add/{user_id}",
		endpoints.FriendAddPut,
	},

	Route{
		"FriendDeletePut",
		strings.ToUpper("Put"),
		"/friend/delete/{user_id}",
		endpoints.FriendDeletePut,
	},

	Route{
		"PostCreatePost",
		strings.ToUpper("Post"),
		"/post/create",
		endpoints.PostCreatePost,
	},

	Route{
		"PostDeletePut",
		strings.ToUpper("Put"),
		"/post/delete/{id}",
		endpoints.PostDeletePut,
	},

	Route{
		"PostGetGet",
		strings.ToUpper("Get"),
		"/post/get/{id}",
		endpoints.PostGetGet,
	},

	Route{
		"PostUpdatePut",
		strings.ToUpper("Put"),
		"/post/update",
		endpoints.PostUpdatePut,
	},

	Route{
		"PostFeedGet",
		strings.ToUpper("Get"),
		"/post/feed",
		endpoints.PostFeedGet,
	},

	Route{
		"DialogUserIdSendMessage",
		strings.ToUpper("Post"),
		"/dialog/{user_id}/send",
		endpoints.DialogUserIdSendMessage,
	},

	Route{
		"DialogUserIdListGet",
		strings.ToUpper("Get"),
		"/dialog/{user_id}/list",
		endpoints.DialogUserIdListGet,
	},
}
