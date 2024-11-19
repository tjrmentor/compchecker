package main

import (
	"context"
	"flag"
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

	// compsJSON, err := json.MarshalIndent(comps, "", "\t")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// os.WriteFile("comps.json", compsJSON, 0444)

	resourcesModeReports, pipelineModeReports := auditCompositions(comps)

	err = report(*resourcesModeReports, *pipelineModeReports)

	if err != nil {
		log.Fatal(err)
	}
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

func auditCompositions(comps crossplanev1.CompositionList) (
	resourcesModeReports *[]ResourcesModeCompositionReport,
	pipelineModeReports *[]PipelineModeCompositionReport) {
	resourcesModeReports = new([]ResourcesModeCompositionReport)
	pipelineModeReports = new([]PipelineModeCompositionReport)
	for _, comp := range comps.Items {
		switch getCompositionMode(comp) {
		case "Pipeline":
			// auditPipelineComposition(pipelineModeReports, comp)
		case "Resources":
			auditResourcesComposition(resourcesModeReports, comp)
		}
	}
	return
}

func getCompositionMode(comp crossplanev1.Composition) crossplanev1.CompositionMode {
	return *comp.Spec.Mode
}

// creates a function checking whether a Patch maps a given field path to a Composite
func isMatchingToCompositeFieldPathPatchFactory(fromFieldPath string) func(crossplanev1.Patch) bool {
	return func(p crossplanev1.Patch) bool {
		return p.Type == ToCompositeFieldPath &&
			*p.FromFieldPath == fromFieldPath
	}
}

func isMatchingPatchSetPatchFactory(name string) func(crossplanev1.Patch) bool {
	return func(p crossplanev1.Patch) bool {
		return p.Type == PatchSet &&
			*p.PatchSetName == name
	}
}
