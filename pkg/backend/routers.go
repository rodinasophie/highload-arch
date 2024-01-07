package backend

import (
	"fmt"
	"highload-arch/pkg/backend/endpoints"
	"highload-arch/pkg/common"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

const PREFIX_V1 = "/api/v1"
const PREFIX_V2 = "/api/v2"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	routes := append(routesV1, routesV2...)
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

var routesV1 = Routes{
	Route{
		"Index",
		"GET",
		PREFIX_V1,
		Index,
	},

	Route{
		"LoginPost",
		strings.ToUpper("Post"),
		PREFIX_V1 + "/login",
		endpoints.LoginPost,
	},

	Route{
		"CheckAuthGet",
		strings.ToUpper("Get"),
		PREFIX_V1 + "/checkAuth",
		endpoints.CheckAuthGet,
	},

	Route{
		"UserGetIdGet",
		strings.ToUpper("Get"),
		PREFIX_V1 + "/user/get/{id}",
		endpoints.UserGetIdGet,
	},

	Route{
		"UserRegisterPost",
		strings.ToUpper("Post"),
		PREFIX_V1 + "/user/register",
		endpoints.UserRegisterPost,
	},

	Route{
		"UserSearchGet",
		strings.ToUpper("Get"),
		PREFIX_V1 + "/user/search",
		endpoints.UserSearchGet,
	},

	Route{
		"FriendAddPut",
		strings.ToUpper("Put"),
		PREFIX_V1 + "/friend/add/{user_id}",
		endpoints.FriendAddPut,
	},

	Route{
		"FriendDeletePut",
		strings.ToUpper("Put"),
		PREFIX_V1 + "/friend/delete/{user_id}",
		endpoints.FriendDeletePut,
	},

	Route{
		"PostCreatePost",
		strings.ToUpper("Post"),
		PREFIX_V1 + "/post/create",
		endpoints.PostCreatePost,
	},

	Route{
		"PostDeletePut",
		strings.ToUpper("Put"),
		PREFIX_V1 + "/post/delete/{id}",
		endpoints.PostDeletePut,
	},

	Route{
		"PostGetGet",
		strings.ToUpper("Get"),
		PREFIX_V1 + "/post/get/{id}",
		endpoints.PostGetGet,
	},

	Route{
		"PostUpdatePut",
		strings.ToUpper("Put"),
		PREFIX_V1 + "/post/update",
		endpoints.PostUpdatePut,
	},

	Route{
		"PostFeedGet",
		strings.ToUpper("Get"),
		PREFIX_V1 + "/post/feed",
		endpoints.PostFeedGet,
	},

	Route{
		"DialogUserIdSendMessage",
		strings.ToUpper("Post"),
		PREFIX_V1 + "/dialog/{user_id}/send",
		endpoints.DialogUserIdSendMessage,
	},

	Route{
		"DialogUserIdListGet",
		strings.ToUpper("Get"),
		PREFIX_V1 + "/dialog/{user_id}/list",
		endpoints.DialogUserIdListGet,
	},
}

var routesV2 = Routes{
	Route{
		"Index",
		"GET",
		PREFIX_V2,
		Index,
	},

	Route{
		"LoginPost",
		strings.ToUpper("Post"),
		PREFIX_V2 + "/login",
		endpoints.LoginPost,
	},

	Route{
		"UserGetIdGet",
		strings.ToUpper("Get"),
		PREFIX_V2 + "/user/get/{id}",
		endpoints.UserGetIdGet,
	},

	Route{
		"UserRegisterPost",
		strings.ToUpper("Post"),
		PREFIX_V2 + "/user/register",
		endpoints.UserRegisterPost,
	},

	Route{
		"UserSearchGet",
		strings.ToUpper("Get"),
		PREFIX_V2 + "/user/search",
		endpoints.UserSearchGet,
	},

	Route{
		"FriendAddPut",
		strings.ToUpper("Put"),
		PREFIX_V2 + "/friend/add/{user_id}",
		endpoints.FriendAddPut,
	},

	Route{
		"FriendDeletePut",
		strings.ToUpper("Put"),
		PREFIX_V2 + "/friend/delete/{user_id}",
		endpoints.FriendDeletePut,
	},

	Route{
		"PostCreatePost",
		strings.ToUpper("Post"),
		PREFIX_V2 + "/post/create",
		endpoints.PostCreatePost,
	},

	Route{
		"PostDeletePut",
		strings.ToUpper("Put"),
		PREFIX_V2 + "/post/delete/{id}",
		endpoints.PostDeletePut,
	},

	Route{
		"PostGetGet",
		strings.ToUpper("Get"),
		PREFIX_V2 + "/post/get/{id}",
		endpoints.PostGetGet,
	},

	Route{
		"PostUpdatePut",
		strings.ToUpper("Put"),
		PREFIX_V2 + "/post/update",
		endpoints.PostUpdatePut,
	},

	Route{
		"PostFeedGet",
		strings.ToUpper("Get"),
		PREFIX_V2 + "/post/feed",
		endpoints.PostFeedGet,
	},

	Route{
		"DialogUserIdSendMessage",
		strings.ToUpper("Post"),
		PREFIX_V2 + "/dialog/{user_id}/send",
		endpoints.DialogUserIdSendMessage,
	},

	Route{
		"DialogUserIdListGet",
		strings.ToUpper("Get"),
		PREFIX_V2 + "/dialog/{user_id}/list",
		endpoints.DialogUserIdListGet,
	},

	Route{
		"CheckAuthGet",
		strings.ToUpper("Get"),
		PREFIX_V2 + "/checkAuth",
		endpoints.CheckAuthGet,
	},
}
