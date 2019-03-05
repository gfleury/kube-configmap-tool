package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const usage string = `
kube-prom-config-update configMapName configMapReference fileToImport
`

// Usage is default used by flag to print the default tool usage
var Usage = func() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	fmt.Print(usage)
	flag.PrintDefaults()
}

func main() {
	var context = flag.String("C", "", "Specify Kubernetes config context to use (same from kubectl config).")
	var kubeNamespace = flag.String("N", "default", "Specify Kubernetes namespace.")
	var newReference = flag.Bool("n", false, "Create reference inside configMap if it doesn't exists.")

	flag.Parse()

	var configMapName = flag.Arg(0)
	var configMapReference = flag.Arg(1)
	var file = flag.Arg(2)

	if configMapName == "" || configMapReference == "" || file == "" {
		Usage()
		os.Exit(2)
	}

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf(err.Error())
	}

	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	if *context != "" {
		overrides.CurrentContext = *context
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides).ClientConfig()

	if err != nil {
		// creates the in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	var configMap *v1.ConfigMap

	configMap, err = clientset.CoreV1().ConfigMaps(*kubeNamespace).Get(configMapName, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("Failed to get configMap %s on namespace %s: %s\n", configMapName, *kubeNamespace, err.Error())
	}

	if _, ok := configMap.Data[configMapReference]; !ok && !*newReference {
		log.Fatalf("Reference %s not found on configMap %s", configMapName, configMapReference)
	}

	configMap.Data[configMapReference] = string(fileContent)

	_, err = clientset.CoreV1().ConfigMaps(*kubeNamespace).Update(configMap)
	if err != nil {
		log.Fatalf("Failed to update configMap %s on namespace %s: %s\n", configMapName, *kubeNamespace, err.Error())
	}
	log.Printf("ConfigMap %s updated sucessfully.", configMapName)
}
