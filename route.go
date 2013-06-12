package pathways

import (
	"fmt"
	"net/http"
	"strings"
)

type Service struct {
	root          string
	routes        []*Route
	defaultAction http.Handler
}

func NewService(root string) *Service {
	root = strings.TrimRight(root, "/") + "/"
	return &Service{
		root:          root,
		defaultAction: (http.HandlerFunc)(http.NotFound),
	}
}

func (s *Service) DefaultHandler(action http.Handler) *Service {
	s.defaultAction = action
	return s
}

// Default action to perform when no routes match.
func (s *Service) DefaultAction(action RouteAction) *Service {
	s.defaultAction = action
	return s
}

func (s *Service) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	for _, route := range s.routes {
		if route.apply(writer, request) {
			return
		}
	}
	s.defaultAction.ServeHTTP(writer, request)
}

func (s *Service) Path(path string) *Route {
	path = strings.TrimLeft(path, "/")
	route := NewRoute(s.root + path)
	s.routes = append(s.routes, route)
	return route
}

func (s *Service) Find(name string) *Route {
	for _, r := range s.routes {
		if r.name == name {
			return r
		}
	}
	return nil
}

type Route struct {
	name         string
	path         string
	filters      []StageAcceptor
	pathMatch    *matchPath
	methodMatch  matchMethods
	action       RouteAction
	requestType  interface{}
	responseType interface{}
}

func NewRoute(path string) *Route {
	return (&Route{}).Path(path)
}

func (r *Route) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if !r.apply(writer, request) {
		// TODO: Respond with a 404...
	}
}

func (r *Route) String() string {
	filters := []string{}
	for _, f := range r.filters {
		filters = append(filters, f.String())
	}
	s := "Route." + strings.Join(filters, ".")
	if r.name != "" {
		s += fmt.Sprintf(".Named(%#v)", r.name)
	}
	return s
}

func (r *Route) apply(writer http.ResponseWriter, request *http.Request) bool {
	cx := &Context{
		Request:  request,
		Response: writer,
		Vars:     make(map[string]interface{}),
	}
	for _, filter := range r.filters {
		if !filter.Accept(cx) {
			return false
		}
	}
	response := r.action(cx)
	response(writer)
	return true
}

func (r *Route) Filter(filter StageAcceptor) *Route {
	r.filters = append(r.filters, filter)
	return r
}

func (r *Route) Action(action RouteAction) *Route {
	if r.action != nil {
		panic("an action has already been applied to this route")
	}
	r.action = action
	return r
}

func (r *Route) Name(name string) *Route {
	r.name = name
	return r
}

// Specify the HTTP methods this endpoint accepts.
func (r *Route) Methods(methods ...string) *Route {
	r.methodMatch = realMatchMethods(methods...)
	return r.Filter(r.methodMatch)
}

func (r *Route) Get() *Route {
	return r.Methods("GET")
}

func (r *Route) Post() *Route {
	return r.Methods("POST")
}

func (r *Route) Delete() *Route {
	return r.Methods("DELETE")
}

func (r *Route) Put() *Route {
	return r.Methods("PUT")
}

// Request must contain the given header and match the regex for the route to match.
// eg. route.Header("Content-Type", "application/json|application/x-msgpack")...
func (r *Route) Header(name, pattern string) *Route {
	return r.Filter(MatchHeader(name, pattern))
}

// Request must have the given query parameter and match the regex for this route to match.
// eg. route.Header("id", `\d+`)
func (r *Route) Query(name, pattern string) *Route {
	return r.Filter(MatchQuery(name, pattern))
}

func (r *Route) Path(path string) *Route {
	r.path = path
	r.pathMatch = realMatchPath(path)
	return r.Filter(r.pathMatch)
}

func (r *Route) Handler(handler http.Handler) *Route {
	return r.Action(applyHandler(handler))
}

func (r *Route) HandlerFunc(handler http.HandlerFunc) *Route {
	return r.Action(applyHandler(handler))
}

func (r *Route) ApiRequestType(t interface{}) *Route {
	r.requestType = t
	return r
}

func (r *Route) ApiResponseType(t interface{}) *Route {
	r.responseType = t
	return r
}

// Handle this route with a function of the form func(*Context[, t]). If t is
// provided by ApiRequestType(), it must be a pointer to a structure. The
// request body will be decoded into a value of this type and passed to the
// callback as the second argument. If t is nil, the request body is not
// decoded, and no argument is passed.
func (r *Route) ApiFunction(f interface{}) *Route {
	return r.Action(applyFunction(f, r.requestType))
}

// The HTTP method associated with this route. Will panic if route does not
// have exactly one method.
func (r *Route) Method() string {
	if len(r.methodMatch) != 1 {
		panic("no methods available")
	}
	return r.methodMatch[0]
}

// Reverse the route path.
func (r *Route) Reverse(args map[string]string) string {
	path := r.path
	for arg, value := range args {
		// TODO: Handle remainder {arg...}
		path = strings.Replace(path, "{"+arg+"}", value, 1)
	}
	return path
}
