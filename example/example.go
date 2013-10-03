package main

import (
	"github.com/alecthomas/pathways"
	"net/http"
)

type KeyValueService struct {
	service *pathways.Service
	kv      map[string]string
}

func KeyValueServiceMap(root string) *pathways.Service {
	s := pathways.NewService(root)
	str := ""
	s.Path("/").Name("List").Get().APIResponseType(map[string]string{})
	s.Path("/{key}").Name("Get").Get().APIResponseType(&str)
	s.Path("/{key}").Name("Create").Post().APIRequestType(&str)
	s.Path("/{key}").Name("Delete").Delete()
	return s
}

func NewClient(url string) *pathways.Client {
	s := KeyValueServiceMap(url)
	return pathways.NewClient(s, "application/json")
}

func NewKeyValueService(root string) *KeyValueService {
	k := &KeyValueService{
		service: KeyValueServiceMap(root),
		kv:      make(map[string]string),
	}
	k.service.Find("List").APIFunction(k.List)
	k.service.Find("Get").APIFunction(k.Get)
	k.service.Find("Create").APIFunction(k.Create)
	k.service.Find("Delete").APIFunction(k.Delete)
	return k
}

func (k *KeyValueService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	k.service.ServeHTTP(w, r)
}

func (k *KeyValueService) List(cx *pathways.Context) pathways.Response {
	return cx.APIResponse(http.StatusOK, k.kv)
}

func (k *KeyValueService) Get(cx *pathways.Context) pathways.Response {
	return cx.APIResponse(http.StatusOK, k.kv[cx.PathVars["key"]])
}

func (k *KeyValueService) Create(cx *pathways.Context, value *string) pathways.Response {
	k.kv[cx.PathVars["key"]] = *value
	return cx.APIResponse(http.StatusCreated, "ok")
}

func (k *KeyValueService) Delete(cx *pathways.Context) pathways.Response {
	delete(k.kv, cx.PathVars["key"])
	return cx.APIResponse(http.StatusOK, &struct{}{})
}

func main() {
	s := NewKeyValueService("/api/")
	http.ListenAndServe(":8080", s)
}
