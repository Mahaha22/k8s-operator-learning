package main

import (
	"context"
	"fmt"
	"log"
	clientset "operator-crd/pkg/generator/clientset/versioned"
	informer "operator-crd/pkg/generator/informers/externalversions"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	//1.获取配置文件
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Fatal(err)
	}
	//2.创建客户端
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	//查询crd
	list, err := clientset.CrdV1().Bars("default").List(context.Background(), v1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, bar := range list.Items {
		fmt.Printf("name = %v\n", bar.Name)
	}

	//informer
	factory := informer.NewSharedInformerFactory(clientset, 0)
	factory.Crd().V1().Bars().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			//todo
		},
	})

}
