package main

import (
	"informer/pkg"
	"log"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	//1.获取config对象
	//2.创建client
	//3.informer
	//4.添加事件
	//5.启动informer
	config, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
	if err != nil {
		//集群外部获取不到从集群内部获取
		inclusterconfig, err := rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
		config = inclusterconfig

	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	factory := informers.NewSharedInformerFactory(clientset, 0)

	serviceinformer := factory.Core().V1().Services()
	ingressinformer := factory.Networking().V1().Ingresses()

	controller := pkg.NewController(clientset, serviceinformer, ingressinformer)
	stopch := make(chan struct{})
	factory.Start(stopch)
	factory.WaitForCacheSync(stopch)

	controller.Run(stopch)
}
