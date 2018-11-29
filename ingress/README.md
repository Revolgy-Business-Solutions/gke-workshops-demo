# kube-lego
```
helm upgrade --install  kube-lego stable/kube-lego --set rbac.create=true,config.LEGO_EMAIL=user@example.com,config.LEGO_URL=https://acme-v01.api.letsencrypt.org/directory
```
# cloudflare-dns
```
helm upgrade --install cloudflare-dns --set rbac.create=true,sources={ingress},logLevel=debug,provider=cloudflare,cloudflare.apiKey=SECRET_KEY,cloudflare.email=user@example.com stable/external-dns
```

if you don't have your public dns domain, you can use http://nip.io/ to resolve it for you

`whatever.1.2.3.4.nip.io` will resolve to `1.2.3.4`, you can use this in your ingress resource to provisiong a le cert for that domain
