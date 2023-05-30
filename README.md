### About
Locust Operator for Kubernetes

### Features
* Write your own Locust File and deploy as a ConfigMap in your cluster
* Deploy a LocustTest Custom Resource and watch your Locust Master / Workers spin up
* Port forward to Service in cluster to access the Locust UI
* Supports Secret Mounting
* Supports additional Python package required on top of Locust

## Installation of Operator using Helm
```
make deploy-helm
```

### Running Locust on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
helm uninstall locust-operator
```


### How it works
This was built using Kubebuilder, following the Operator pattern.
Helm charts are stored under /chart.
Feel free to make a PR for additional features.
More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

