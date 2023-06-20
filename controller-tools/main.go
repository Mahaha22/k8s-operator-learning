package main

import (
	"context"
	"fmt"
	"log"

	v1 "controller-tools/pkg/apis/baiding.tech/v1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatal(err)
	}
	config.APIPath = "/apis/"
	config.NegotiatedSerializer = v1.Codes.WithoutConversion()
	config.GroupVersion = &v1.Groupversion
	// clienset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	client, err := rest.RESTClientFor(config)
	if err != nil {
		log.Fatal(err)
	}

	foo := v1.Foo{}
	err = client.Get().Namespace("default").Resource("foos").Name("test-crd").Do(context.TODO()).Into(&foo)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(foo.Spec)

	newObj := foo.DeepCopy()
	newObj.Spec.Name = "test"

	fmt.Println(newObj.Spec.Name)
}
