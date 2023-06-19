package pkg

import (
	"context"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	service "k8s.io/client-go/informers/core/v1"
	ingress "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	servicelister "k8s.io/client-go/listers/core/v1"
	ingresslister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const workNum = 5
const maxRetry = 10

type controller struct {
	client        kubernetes.Interface //客户端
	ingressLister ingresslister.IngressLister
	serviceLister servicelister.ServiceLister
	queue         workqueue.RateLimitingInterface //限速队列
}

func (c *controller) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
	}
	c.queue.Add(key)
}

func (c *controller) updateService(oldObj, newObj interface{}) {
	//todo 比较annotation
	if reflect.DeepEqual(oldObj, newObj) { //如果新旧对象一直，不处理
		return
	}

	//操作资源对象让其保持一致
	c.enqueue(newObj)
}

func (c *controller) addService(obj interface{}) {
	c.enqueue(obj)
}
func (c *controller) deleteService(obj interface{}) {
	service := obj.(*corev1.Service)
	ingress, err := c.ingressLister.Ingresses(service.Namespace).Get(service.Name)
	if err != nil || ingress == nil { //如果获取ingress出错，或者根本没有对应的ingress
		return
	}
	c.queue.Add(ingress.Namespace + "/" + ingress.Name)
}

/*
这段代码的目的是在 Ingress 资源删除时，找到关联的控制器，
并将 Ingress 添加到控制器的队列中，以便进行后续的处理逻辑。
具体的处理逻辑可能在控制器的其他部分实现，这里只是负责将删除事件添加到队列中。
*/
func (c *controller) deleteIngress(obj interface{}) {
	ingress := obj.(*network.Ingress)
	ownref := metav1.GetControllerOf(ingress) //获取与ingress关联的控制器
	if ownref == nil {
		return
	}
	if ownref.Kind != "Service" {
		return
	}

	c.queue.Add(ingress.Namespace + "/" + ingress.Name)
}
func (c *controller) Run(stopch chan struct{}) {
	for i := 0; i < workNum; i++ {
		go wait.Until(c.worker, time.Minute, stopch) //wait.Until会在定时的时间间隔不断的执行传入的函数，直到接收到stopch信号
	}

	<-stopch
}

func (c *controller) worker() {
	for c.processNextItem() {

	}
}

func (c *controller) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)

	key := item.(string)
	if err := c.syncService(key); err != nil {
		c.HandleError(key, err)
	}
	return true
}
func (c *controller) HandleError(key string, err error) {
	if c.queue.NumRequeues(key) <= maxRetry { //重试次数在限定次数内
		c.queue.AddRateLimited(key)
	}
	runtime.HandleError(err)
	c.queue.Forget(key)
}
func (c *controller) syncService(key string) error {
	namespaceKey, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	// 删除
	service, err := c.serviceLister.Services(namespaceKey).Get(name)
	if errors.IsNotFound(err) {
		c.client.NetworkingV1().Ingresses(namespaceKey).Delete(context.TODO(), name, metav1.DeleteOptions{})
		return nil
	}
	if err != nil {
		return err
	}
	//新增和删除
	_, ok := service.GetAnnotations()["ingress/http"] //这行代码用于检查 service 对象的 Annotations 字段中是否存在键为 "ingress/http" 的注解。
	ingress, err := c.ingressLister.Ingresses(namespaceKey).Get(name)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if ok && errors.IsNotFound(err) { //如果service存在而ingress不存在
		//创建ingress
		ig := c.constructIngress(service)
		_, err := c.client.NetworkingV1().Ingresses(namespaceKey).Create(context.TODO(), ig, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if !ok && ingress != nil {
		//删除ingress
		err := c.client.NetworkingV1().Ingresses(namespaceKey).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) constructIngress(service *corev1.Service) *network.Ingress {
	ingress := network.Ingress{}
	//这样设置 OwnerReferences 字段的作用是，在删除 service 对象时，会同时删除与之关联的 ingress 对象，以确保资源之间的一致性和关联性。
	ingress.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(service, metav1.SchemeGroupVersion.WithKind("Service")),
	}

	ingress.Name = service.Name
	ingress.Namespace = service.Namespace
	pathType := network.PathTypePrefix
	icn := "nginx"
	ingress.Spec = network.IngressSpec{
		IngressClassName: &icn,
		Rules: []network.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: network.IngressRuleValue{
					HTTP: &network.HTTPIngressRuleValue{
						Paths: []network.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: network.IngressBackend{
									Service: &network.IngressServiceBackend{
										Name: service.Name,
										Port: network.ServiceBackendPort{
											Number: 80,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return &ingress
}
func NewController(client kubernetes.Interface, serviceinformer service.ServiceInformer, ingressinformer ingress.IngressInformer) controller {
	c := controller{
		client:        client,
		serviceLister: serviceinformer.Lister(),
		ingressLister: ingressinformer.Lister(),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingressManager"),
	}

	serviceinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addService,
		UpdateFunc: c.updateService,
		DeleteFunc: c.deleteService,
	})

	ingressinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.deleteIngress,
	})
	return c
}
