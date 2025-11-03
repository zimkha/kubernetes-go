package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("failed to build kubeconfig: %v", err)
		//panic(err.Error())
	}
	configset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalln(fmt.Errorf("error building config: %s", err))
		//panic(err.Error())
	}

	configMapset := configset.CoreV1().ConfigMaps(corev1.NamespaceDefault)

	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "configmap",
			Labels: map[string]string{
				"app": "dev",
			},
		},
		Data: map[string]string{
			"foo": "bar",
			"baz": "qux",
		},
	}
	result, err := configMapset.Create(context.TODO(), configmap, metav1.CreateOptions{})
	if err != nil {
		log.Fatalln(fmt.Errorf("error creating configmap: %s", err))
	}
	fmt.Printf("Created configmap %s\n", result.GetObjectMeta().GetName())
}

func promptConfig() {
	fmt.Printf("Please press enter to procesed ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()

}
