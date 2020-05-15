package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	defaultNamespace string          = "default"
	patchType        types.PatchType = types.MergePatchType
)

var (
	nodeName     string
	podName      string
	namespace    string
	removeLabels []string
	appendLabels []string
)

type metadata struct {
	Labels map[string]string `json:"labels"`
}

type patch struct {
	metadata `json:"metadata"`
}

func main() {
	// define cmd flags
	pflag.StringVar(&nodeName, "node-name", "", "node name")
	pflag.StringVar(&podName, "pod-name", "", "pod name")
	pflag.StringVar(&namespace, "namespace", "", "namespace")
	pflag.StringArrayVar(&removeLabels, "exclude-label", []string{}, "exclude this labels")
	pflag.StringArrayVar(&appendLabels, "include-label", []string{}, "include only this labels")
	pflag.Parse()

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// get node meta
	node, err := clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	log.Printf("available node lables: %s\n", node.Labels)

	// get pod meta
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	log.Printf("available pod lables: %s\n", pod.Labels)

	// generate patch body
	patchLabels := make(map[string]string)

	if len(appendLabels) == 0 {
		for key, value := range node.Labels {
			for _, item := range removeLabels {
				if key == item {
					continue
				}

				patchLabels[key] = value
			}
		}
	} else {
		for key, value := range node.Labels {
			for _, item := range appendLabels {
				if key == item {
					patchLabels[key] = value
				}
			}
		}
	}

	data, err := json.Marshal(patch{metadata: metadata{Labels: patchLabels}})
	if err != nil {
		panic(err.Error())
	}

	log.Printf("patch request body: %s\n", data)

	// patch pod
	_, err = clientset.CoreV1().Pods(namespace).Patch(
		context.TODO(),
		podName,
		patchType,
		data,
		metav1.PatchOptions{},
	)
	if err != nil {
		panic(err.Error())
	}

	log.Printf("succefully patched\n")
}
