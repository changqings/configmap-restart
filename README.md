# config-restart
// Use crd to manage the configmap and restart deployments which mounted the configmap.

## Description

//

## tips:

1. mgr 是控制 controller 的，他的 cache.Option()针对的是要 cache 哪些资源，添加一些过滤解析，比如 ns/labels/annotations 等,其中的 cache.Options 可以使用 DefaultTransform:来剪切掉不需要缓存的资源
2. controller 中 SetupWithManager()中的 watch()和 predicate()是针对的已经 cached 的资源进行消费时的进一步过滤，其中 watch()是更细粒度的控制，一般使用 builder.WithPredicates()
   即可满足所需要的过滤条件。
3. 关于 watch(),这三个参数，第一个是资源类型，第二个是事件处理函数，第三个是过滤条件。
   重点讲一下第二个事件处理函数
   EnqueueRequestForObject：最简单直接的处理器，适用于大多数情况。
   EnqueueRequestsFromMapFunc：允许自定义映射函数，适用于需要生成多个请求的情况。
   EnqueueRequestForOwner：适用于需要根据子资源的变化来更新父资源的情况。

```go
Watch(
            &source.Kind{Type: &corev1.ConfigMap{}},
            &handler.EnqueueRequestForObject{},
            configMapUpdatePredicate(), // 应用自定义谓词
        ).
```
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
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

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

