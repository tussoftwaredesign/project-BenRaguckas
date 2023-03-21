from root directory:

    helm install deployment-name . -n msgb

or

    helm upgrade --install deployment-name . -n msgb


(Not needed, covered by services) expose using kubernetes loadbalancer:

    kubectl expose deployment rabbitmq --type=LoadBalancer --name=balancer-name -n msgb

example exec:

    kubectl exec --stdin --tty go-api-686bbfcbb7-wv2jg -n msgb  -- /bin/sh