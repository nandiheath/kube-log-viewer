package ws

import (
	"fmt"
	"nandiheath/kube-log-viewer/internal/kube"

	"github.com/gorilla/websocket"
	"k8s.io/apimachinery/pkg/util/uuid"
)

type Hub struct {
	// key: id
	clients    map[string]*Client
	send       chan kube.LogMessage
	kubeClient *kube.Client
}

func NewHub(kubeClient *kube.Client) *Hub {
	hub := &Hub{
		clients:    map[string]*Client{},
		send:       make(chan kube.LogMessage, 100),
		kubeClient: kubeClient,
	}
	return hub
}

func (h *Hub) Start() {
	for {
		select {
		case msg := <-h.send:
			for _, client := range h.clients {
				if client.id == msg.ClientId {
					go func() {
						client.send <- []byte(msg.Message)
					}()
				}
			}
		}
	}
}

func (h *Hub) WatchLog(conn *websocket.Conn, namespace, pod string) {
	client := Client{
		namespace: namespace,
		id:        string(uuid.NewUUID()),
		pod:       pod,
		conn:      conn,
		send:      make(chan []byte),
		stop:      make(chan struct{}),
	}
	h.clients[client.id] = &client
	fmt.Printf("connection established. client_id: %s\n", client.id)
	defer func() {
		fmt.Printf("connection closed. client_id: %s\n", client.id)
		delete(h.clients, client.id)
	}()

	go client.StartListening()
	go client.StartWriting()

	h.kubeClient.WatchLog(namespace, pod, client.id, h.send, client.stop)
}
