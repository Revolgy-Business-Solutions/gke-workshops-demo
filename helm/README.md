# helm pipeline

```    - sed -i "s#IMAGE_TAG#$IMAGE_REPO:$IMAGE_TAG#" k8s/kuard-canary.yaml
    - cat k8s/kuard-canary.yaml
    - kubectl apply -f k8s/kuard-canary.yaml
    - set -x
    - |
       helm upgrade --install --force --debug $RELEASE_NAME \
         --set replicas=1 \
         --set image.repository="$IMAGE_REPO" \
         --set image.tag="$IMAGE_TAG" \
         ./helm/
    - time kubectl rollout status deployment/$RELEASE_NAME
```

# test helm
```
helm install --name wordpress stable/wordpress
```
