package kube

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"nandiheath/kube-log-viewer/internal/config"
	"time"

	"github.com/pkg/errors"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type PodInfo struct {
	Name      string `json:"pod"`
	Namespace string `json:"namespace"`
}

type Client struct {
	clientset *kubernetes.Clientset
}

type LogMessage struct {
	Namespace string
	ClientId  string
	Pod       string
	Message   string
}

type MessageHandler func(message string) error

func New() *Client {
	m := Client{}
	var kubeconfig *string
	kubeconfig = flag.String("kubeconfig", config.KubeConfigPath, "absolute path to the kubeconfig file")
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	m.clientset = clientset
	return &m
}

// TODO: handle mulitple client listening to same pod
func (m *Client) WatchLog(namespace, pod, clientId string, send chan<- LogMessage, stop <-chan struct{}) {

	req := m.clientset.CoreV1().Pods(namespace).GetLogs(
		pod,
		&v12.PodLogOptions{Follow: true, Timestamps: true, SinceTime: &v1.Time{
			Time: time.Now().Add(-60 * 60 * 24 * 31 * time.Second),
		}},
	)
	stream, err := req.Stream(context.TODO())
	if err != nil {
		fmt.Printf("cannot listen to pod: %s", err.Error())
		return
	}

	br := bufio.NewReaderSize(stream, 2000)
	defer stream.Close()
	write := make(chan string)
	go func() {
		for {

			buf, _, err := br.ReadLine()
			if len(buf) == 0 {
				continue
			}
			if err == io.EOF {
				return
			}
			if err != nil {
				fmt.Printf("error when reading logs: %s", err.Error())
				return
			}
			write <- string(buf)
			fmt.Println(string(buf))
		}

	}()

	for {
		select {
		case <-stop:
			{
				fmt.Printf("signal term. stop listening to logs\n")
				return
			}
		case msg := <-write:
			{
				send <- LogMessage{
					Pod:       pod,
					ClientId:  clientId,
					Namespace: namespace,
					Message:   msg,
				}
			}
		}
	}

}

func (m *Client) GetPods() ([]PodInfo, error) {
	ctx := context.Background()

	pods, err := m.clientset.CoreV1().Pods("").List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get pods")
	}
	var podInfos []PodInfo
	for _, pod := range pods.Items {
		podInfos = append(podInfos, PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		})
	}
	return podInfos, nil
}
