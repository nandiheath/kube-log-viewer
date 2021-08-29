package server

import (
	"log"
	"nandiheath/kube-log-viewer/internal/kube"
	"nandiheath/kube-log-viewer/internal/ws"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"k8s.io/apimachinery/pkg/util/json"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Server struct {
	kubeClient *kube.Client
	hub        *ws.Hub
}

func New() *Server {
	m := Server{}
	m.kubeClient = kube.New()
	m.hub = ws.NewHub(m.kubeClient)
	go m.hub.Start()
	return &m
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func (s *Server) getPods(w http.ResponseWriter, r *http.Request) {
	pods, err := s.kubeClient.GetPods()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(pods)
	w.Write(data)
}

func (s *Server) watchLogs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	conn.SetReadLimit(maxMessageSize)

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	podname := vars["pod_name"]

	go s.severWs(conn, namespace, podname)
}

func (s *Server) severWs(conn *websocket.Conn, namespace, pod string) {

	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	conn.SetReadDeadline(time.Now().Add(pongWait))

	s.hub.WatchLog(conn, namespace, pod)

	//for {
	//	mt, message, err := conn.ReadMessage()
	//	if err != nil {
	//		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
	//			log.Printf("error: %v", err)
	//		}
	//		break
	//	}
	//	log.Printf("recv: %s", message)
	//	err = conn.WriteMessage(mt, []byte(message))
	//	if err != nil {
	//		log.Println("write:", err)
	//		break
	//	}
	//}
}

func (s *Server) Start() {

	r := mux.NewRouter()
	r.HandleFunc("/pods", s.getPods)
	r.HandleFunc("/namespace/{namespace}/pod/{pod_name}/logs", s.watchLogs)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
