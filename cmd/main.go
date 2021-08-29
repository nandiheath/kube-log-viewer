package main

import (
	"nandiheath/kube-log-viewer/internal/server"
)

func main() {
	s := server.New()
	s.Start()
}
