package pathways

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	pathTransform = regexp.MustCompile(`{((\w+)(\.\.\.)?)}`)
)

// Allows for matching of requests.
type StageAcceptor interface {
	fmt.Stringer
	Accept(cx *Context) bool
}

type matchMethods []string

func realMatchMethods(methods ...string) matchMethods {
	return (matchMethods)(methods)
}

func MatchMethods(methods ...string) StageAcceptor {
	return realMatchMethods(methods...)
}

func (m matchMethods) Accept(cx *Context) bool {
	for _, method := range m {
		if method == cx.Request.Method {
			return true
		}
	}
	return false
}

func (m matchMethods) String() string {
	return fmt.Sprintf("Methods(%v)", ([]string)(m))
}

type matchPath struct {
	pattern *regexp.Regexp
	params  []string
}

func realMatchPath(path string) *matchPath {
	routePattern := "^" + path + "$"
	for _, match := range pathTransform.FindAllStringSubmatch(routePattern, 16) {
		pattern := `([^/]+)`
		if match[3] == "..." {
			pattern = `(.+)`
		}
		routePattern = strings.Replace(routePattern, match[0], pattern, 1)
	}

	pattern := regexp.MustCompile(routePattern)
	params := []string{}
	for _, arg := range pathTransform.FindAllString(path, 16) {
		params = append(params, arg[1:len(arg)-1])
	}

	return &matchPath{
		pattern: pattern,
		params:  params,
	}
}

func MatchPath(path string) StageAcceptor {
	return realMatchPath(path)
}

func (m *matchPath) Accept(cx *Context) bool {
	args := m.pattern.FindStringSubmatch(cx.Request.RequestURI)
	if args != nil {
		cx.PathVars = make(map[string]string)
		for i, name := range m.params {
			cx.PathVars[name] = args[i+1]
		}
		return true
	}
	return false
}

func (m *matchPath) String() string {
	return fmt.Sprintf("Path(%#v)", m.pattern.String())
}

type matchHeader struct {
	pattern *regexp.Regexp
	name    string
}

func MatchHeader(name, pattern string) StageAcceptor {
	re := regexp.MustCompile(pattern)
	return &matchHeader{re, name}
}

func (m *matchHeader) Accept(cx *Context) bool {
	return m.pattern.MatchString(cx.Request.Header.Get(m.name))
}

func (m *matchHeader) String() string {
	return fmt.Sprintf("Header(%#v, %#v)", m.name, m.pattern.String())
}

type matchQuery struct {
	pattern *regexp.Regexp
	name    string
}

func MatchQuery(name, pattern string) StageAcceptor {
	re := regexp.MustCompile(pattern)
	return &matchQuery{re, name}
}

func (m *matchQuery) Accept(cx *Context) bool {
	if values, ok := cx.Request.URL.Query()[m.name]; ok {
		for _, value := range values {
			if m.pattern.MatchString(value) {
				return true
			}
		}
	}
	return false
}

func (m *matchQuery) String() string {
	return fmt.Sprintf("Query(%#v, %#v)", m.name, m.pattern.String())
}
