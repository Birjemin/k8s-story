
#
## dac-demo
参见[官网](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)了解更多

### 生成证书cert
k8s内部访问时采用的是https协议，所以需要自签证书（或者使用k8s自带的证书）
```
ls ~/Developer/Go/src/k8s-story/k8s-dac

sh generate_key.sh  cert
kubectl -n devops create secret tls webhook-server-tls --cert "cert/webhook-server-tls.crt" --key "cert/webhook-server-tls.key"
```
### 打包镜像docker build
打包镜像，并上传到minikube中的docker里
```sh
ls ~/Developer/Go/src/k8s-story/

# build image
docker build -t webhook-server:v0.0.1 -f k8s-dac/Dockerfile .

# loading image
minikube image load webhook-server:v0.0.1

#verify registry(optional)
curl http://127.0.0.1:5000/v2/_catalog
curl http://172.20.10.4:5000/v2/_catalog

#minikube start(optional)
minikube start --cpus 2 --memory 3072 --registry-mirror=https://765qw7sx.mirror.aliyuncs.com --insecure-registry=192.168.31.203:5000 --image-mirror-country=cn --image-repository=registry.cn-hangzhou.aliyuncs.com/google_containers
```

### deploy
部署webhook-server服务，并且注册动态准入控制配置，caBundle字段为ca.crt的base64格式
```sh
# caBundle
openssl base64 -A < "cert/ca.crt"

#init mutating config
kubectl -n devops apply -f mutating-webhook-configuration.yaml

#deploy webhook-server
kubectl -n devops apply -f k8s-dac/webhook-server.yaml
kubectl -n devops get all
```
### test1
在test空间部署helloword.yaml和helloword-with-label.yaml，可以发现前一个部署失败（ 没有名为k8s-dac的label），后一个部署成功（由于dac缘故，所以replica为3），main.go中35行代码
```sh
#apply pod
kubectl -n test apply -f k8s-dac/helloword-with-label.yaml
kubectl -n test apply -f k8s-dac/helloword.yaml
```

结果
![image](./cert/demo.png)


### test2
在webhook-server中添加环境变量OP（值如下）
```sh
base64(`[{"op":"add","path":"/spec/template/spec/containers/1","value":{"name":"curl-container","image":"curlimages/curl:latest","imagePullPolicy":"IfNotPresent","restartPolicy":"Always","command":["sleep","infinity"]}}]`)
=>
W3sib3AiOiJhZGQiLCJwYXRoIjoiL3NwZWMvdGVtcGxhdGUvc3BlYy9jb250YWluZXJzLzEiLCJ2YWx1ZSI6eyJuYW1lIjoiY3VybC1jb250YWluZXIiLCJpbWFnZSI6ImN1cmxpbWFnZXMvY3VybDpsYXRlc3QiLCJpbWFnZVB1bGxQb2xpY3kiOiJJZk5vdFByZXNlbnQiLCJyZXN0YXJ0UG9saWN5IjoiQWx3YXlzIiwiY29tbWFuZCI6WyJzbGVlcCIsImluZmluaXR5Il19fV0=
```
重新部署在test空间部署helloword.yaml和helloword-with-label.yaml，可以发现前一个部署失败（ 没有名为k8s-dac的label），后一个部署成功（由于dac缘故，所以container为2），OP变量新增了一个container

结果
![image](./cert/demo2.png)

## reference
1. [registry](https://researchlab.github.io/2019/08/24/minikube-pull-image-from-docker-registry/)
2. [caBundle](https://cuisongliu.github.io/2020/07/kubernetes/admission-webhook/)
3. [jianshu](https://www.jianshu.com/p/00c69b992e3f)
4. [how to get jsonpatch](https://json-patch-builder-online.github.io/)

## how

## Question
```
W0718 11:24:23.417416       1 dispatcher.go:170] Failed calling webhook, failing open webhook-server-svc.devops.svc: failed calling webhook "webhook-server-svc.devops.svc": Post "https://webhook-server-svc.devops.svc:8080/mutating?timeout=10s": x509: certificate signed by unknown authority

# get caBundle
openssl base64 -A < "cert/ca.crt"
```
