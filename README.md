# configmap-restart main logical

- opsappv1.Configrestart{} define the configMap to watch and deployemnts to restart
- watch configMap data update, then get the same namespace opsappv1.ConfigrestartList{}
  and restart deployment

```yaml
# example cr
apiVersion: opsapp.someapp.cn/v1
kind: Configrestart
metadata:
  name: configrestart-sample
spec:
  // configMap name to watch
  configName: my-config
  // 指定deployments to restart, if not set, it will restart all deployment which have mount the configMap
  deployments:
    - nginx-tmp
  // 此配置是否生效的开关
  suspend: false
```

## Description

### 思路梳理:

1. ctrl.Option 下的 Cache: cache.Option{} 针对的是要 cache 资源，添加过滤条件，可以为
   namepspace/labelSelecter/FieldSelector/transform, 使用 transform 裁掉了 managerdFields, 减少缓存
2. SutupWithManager 中的 For() 用来选择 controller 要监听的对象，这里要监听两个对象, configmap 和 configrestart(此 crd 定义的对象)
   所以创建了两个 controller
3. SetupWithManager()中的 watch() 是在 ctrl.NewManager() 后的配置的更进一步配置，
   其中 watch() 方法等同于 WatchRawSource(),

```go
Watch(
            &source.Kind{Type: &corev1.ConfigMap{}},
            &handler.EnqueueRequestForObject{},
            configMapUpdatePredicate(), // 应用自定义谓词
        ).
```

- 关于第二个参数,事件处理函数:
- EnqueueRequestForObject：最简单直接的处理器，适用于大多数情况。
- EnqueueRequestsFromMapFunc：允许自定义映射函数，适用于需要生成多个请求的情况。
- EnqueueRequestForOwner：适用于需要根据子资源的变化来更新父资源的情况。

4. WithEventFilter(p Predicate.Predicate) 方法用来过滤事件（事件是否触发此 reconcile），比如只监听 configMap 的内容更新，标签及 annotations 变化不触发 reconcile

## Getting Started

### Prerequisites

- go version v1.22.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/configmap-restart:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/configmap-restart:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
> privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

> **NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall

**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following are the steps to build the installer and distribute this project to users.

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/configmap-restart:tag
```

NOTE: The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without
its dependencies.

2. Using the installer

Users can just run kubectl apply -f <URL for YAML BUNDLE> to install the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/configmap-restart/<tag or branch>/dist/install.yaml
```

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
