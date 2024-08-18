
#
## crd-demo
参见[官网](https://kubernetes.io/zh-cn/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)了解更多

### 创建crd
```sh
ls ~/Developer/Go/src/k8s-story/

kubectl -n test apply -f resourcedefinition.yaml
```

### 创建crontab

```sh
kubectl -n test apply -f my-crontab.yaml

kubectl -n test delete -f resourcedefinition.yaml
kubectl -n test get crontabs
```

## 进阶
[kubebuilder](https://book.kubebuilder.io/)  vs [Operator SDK](https://sdk.operatorframework.io/)


