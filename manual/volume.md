to create volume use:
    kubectl apply -f .\persistant-volume.yaml

Ensure that storageClass name is availabel using (in this case "hostpath"):
    kubectl get storageclasses --all-namespaces