# code-generator使用
项目地址 https://github.com/kubernetes/sample-controller
```shell
$ git clone https://github.com/kubernetes/code-generator.git
$ git checkout 0.23.3
# 进行安装
$ go install ./cmd/{client-gen,deepcopy-gen,informer-gen,lister-gen}
$ ls $GOPATH/bin
client-gen    defaulter-gen  go-callvis    go-outline  gopls    impl          lister-gen     protoc-gen-go-grpc 
deepcopy-gen  dlv            gomodifytags  goplay      gotests  informer-gen  protoc-gen-go  staticcheck

#创建项目
$ mkdir operator-test && cd operator-test
$ go mod init operator-test
$ mkdir -p pkg/apis/example.com/v1

➜  operator-test tree
.
├── go.mod
├── go.sum
└── pkg
    └── apis
        └── example.com
            └── v1
                ├── doc.go
                ├── register.go
                └── types.go

4 directories, 5 files
```

```go
// pkg/crd.example.com/v1/doc.go

// +k8s:deepcopy-gen=package
// +groupName=example.com

package v1
```

```go
// pkg/crd.example.com/v1/types.go

package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Bar is a specification for a Bar resource
type Bar struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec BarSpec `json:"spec"`
    // Status BarStatus `json:"status"`
}

// BarSpec is the spec for a Bar resource
type BarSpec struct {
    DeploymentName string `json:"deploymentName"`
    Image          string `json:"image"`
    Replicas       *int32 `json:"replicas"`
}

// BarStatus is the status for a Bar resource
type BarStatus struct {
    AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BarList is a list of Bar resources
type BarList struct {
    metav1.TypeMeta `json:",inline" :"metav1.TypeMeta"`
    metav1.ListMeta `json:"metadata" :"metav1.ListMeta"`

    Items []Bar `json:"items" :"items"`
}
```

```go
package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: "example.com", Version: "v1"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
    return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
    return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
    // SchemeBuilder initializes a scheme builder
    SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
    // AddToScheme is a global function that registers this API group & version to a scheme
    AddToScheme = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
    scheme.AddKnownTypes(SchemeGroupVersion,
        &Bar{},
        &BarList{},
    )
    metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
    return nil
}
```

```shell
# 运行 code-generator/generate-group.sh
./../../github/code-generator/generate-groups.sh all \
# 指定 group 和 version，生成deeplycopy以及client
operator-test/pkg/client operator-test/pkg/apis crd.example.com:v1 \
# 指定头文件
--go-header-file=./hack/boilerplate.go.txt \
# 指定输出位置，默认为GOPATH
--output-base ../

../../code-generator/generate-groups.sh all operator-crd/pkg/generator operator-crd/pkg/apis crd.example.com:v1 --go-header-file=/root/code-generator/hack/boilerplate.go.txt --output-base ../


Generating deepcopy funcs
Generating clientset for crd.example.com:v1 at operator-test/pkg/client/clientset
Generating listers for crd.example.com:v1 at operator-test/pkg/client/listers
Generating informers for crd.example.com:v1 at operator-test/pkg/client/informers
```

最终的项目结构
```shell
➜  operator-test tree
.
├── go.mod
├── go.sum
├── hack
│   └── boilerplate.go.txt
└── pkg
    ├── apis
    │   └── crd.example.com
    │       └── v1
    │           ├── doc.go
    │           ├── register.go
    │           ├── types.go
    │           └── zz_generated.deepcopy.go
    └── client
        ├── clientset
        │   └── versioned
        │       ├── clientset.go
        │       ├── doc.go
        │       ├── fake
        │       │   ├── clientset_generated.go
        │       │   ├── doc.go
        │       │   └── register.go
        │       ├── scheme
        │       │   ├── doc.go
        │       │   └── register.go
        │       └── typed
        │           └── crd.example.com
        │               └── v1
        │                   ├── bar.go
        │                   ├── crd.example.com_client.go
        │                   ├── doc.go
        │                   ├── fake
        │                   │   ├── doc.go
        │                   │   ├── fake_bar.go
        │                   │   └── fake_crd.example.com_client.go
        │                   └── generated_expansion.go
        ├── informers
        │   └── externalversions
        │       ├── crd.example.com
        │       │   ├── interface.go
        │       │   └── v1
        │       │       ├── bar.go
        │       │       └── interface.go
        │       ├── factory.go
        │       ├── generic.go
        │       └── internalinterfaces
        │           └── factory_interfaces.go
        └── listers
            └── crd.example.com
                └── v1
                    ├── bar.go
                    └── expansion_generated.go

22 directories, 29 files
```

至此，整个operator项目的框架就算是搭好了

# controller-tools
