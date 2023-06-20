# 安装controller-tools
教程：https://www.cnblogs.com/huiyichanmian/p/16261635.html
```shell
$ git clone https://github.com/kubernetes-sigs/controller-tools.git
$ cd controller-gen
$ go install ./cmd/{controller-gen,type-scaffold}
```

生成deepcopy文件
type.go的结构体需要包含 // +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
> controller-gen object paths=./pkg/apis/baiding.tech/v1

生成CRD文件
register.go需要包含 // +groupName=baiding.tech
> controller-gen crd paths=./... output:crd:dir=config/crd

