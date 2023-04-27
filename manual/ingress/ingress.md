https://kubernetes.github.io/ingress-nginx/deploy/#docker-desktop

for enabling (installing) ingress for docker desktop:

helm upgrade --install ingress-nginx ingress-nginx --repo https://kubernetes.github.io/ingress-nginx --namespace ingress-nginx --create-namespace




##For Enabling TCP connection for RabbitMQ:
1.  Download values.yaml: https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/values.yaml
2.  Edit vales.yaml to include required TCP connection (line: 873)
```yaml
tcp:
    <SERVICE_TARGETPORT>: <SERVICE_NAMESPACE>/<ku>:<SERVICE_TARGETPORT>
```
```yaml
tcp:
    5672: msgb/rabbitmq-service:5672
```
3.  Change annotations to allow larger body size (NO CLUE YET)
4.  Install using values.yaml file:
```sh
helm upgrade --install ingress-nginx ingress-nginx --repo https://kubernetes.github.io/ingress-nginx --namespace ingress-nginx --create-namespace --values values.yaml --wait
```


For local code debuging port-forwarding helps:
```sh
kubectl port-forward svc/minio-service 9000:9000 -n msgb
kubectl port-forward svc/mongodb-service 27017:27017 -n msgb
kubectl port-forward svc/rabbitmq-service 5672:5672 -n msgb
```