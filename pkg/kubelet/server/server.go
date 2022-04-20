package server

import (
	"net/http"

	restful "github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/util/sets"
	"minik8s.com/minik8s/pkg/api/v1"
)

// some kubelet interfaces for Restful service
type HostInterface interface {
	GetPods() []*v1.Pod
	GetPodByName(namespaces, name string) (*v1.Pod, bool)

}

type Server struct {
	host HostInterface
	metricsMethodBuckets sets.String
	restfulContainer restful.Container
}

func NewServer() Server  {
	server := Server{restfulContainer: *restful.NewContainer()}

	server.InstallDefaultHandlers()

	return server
}

// to encode/decode 
func (s *Server) getPods(request *restful.Request, response *restful.Response) {
	s.host.GetPods()
}

func (s *Server) InstallDefaultHandlers() {
	s.metricsMethodBuckets.Insert("pods")
	ws := new(restful.WebService)

	ws.Path("/pods")
	ws.Produces(restful.MIME_JSON)

	ws.Route(ws.GET("").To(s.getPods).Operation("getPods"))

	s.restfulContainer.Add(ws)

}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {

}