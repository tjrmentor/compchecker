package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	crossplanev1 "github.com/crossplane/crossplane/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	dynamicClient, err := createClient()
	if err != nil {
		log.Fatal(err)
	}

	unstructuredComps, err := getUnstructuredCompositions(&dynamicClient)
	if err != nil {
		log.Fatal(err)
	}

	comps, err := deserializeToCompositions(unstructuredComps)
	if err != nil {
		log.Fatal(err)
	}

	report(comps)
}

// createClient creates a dynamic client using the kubeconfig found at ~/.kube/config,
// or the absolute path passed in with the --kubeconfig flag.
func createClient() (ret dynamic.DynamicClient, err error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return
	}

	return *client, nil
}

// getUnstructuredCompositions queries the API for a list of all Crossplane Compositions,
// returning it as an unstructured list.
func getUnstructuredCompositions(client *dynamic.DynamicClient) (unstructuredComps unstructured.UnstructuredList, err error) {

	compositionRes := schema.GroupVersionResource{Group: "apiextensions.crossplane.io", Version: "v1", Resource: "compositions"}
	unstructuredCompsPtr, err := client.Resource(compositionRes).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}

	return *unstructuredCompsPtr, nil
}

// deserializeToCompositions deserializes the UnstructuredList to a CompositionList.
func deserializeToCompositions(unstructuredComps unstructured.UnstructuredList) (compositions crossplanev1.CompositionList, err error) {
	for _, item := range unstructuredComps.Items {
		var comp crossplanev1.Composition
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &comp)
		if err != nil {
			return
		}
		compositions.Items = append(compositions.Items, comp)
	}

	return
}

func report(comps crossplanev1.CompositionList) {
	for _, comp := range comps.Items {
		fmt.Println(comp.Name)
	}
	fmt.Println(len(comps.Items))
}
