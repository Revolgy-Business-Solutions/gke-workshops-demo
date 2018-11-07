# kube-lego
```
helm upgrade --install  kube-lego stable/kube-lego --set rbac.create=true,config.LEGO_EMAIL=user@example.com,config.LEGO_URL=https://acme-v01.api.letsencrypt.org/directory
```
# cloudflare-dns
```
helm upgrade --install cloudflare-dns --set rbac.create=true,sources={ingress},logLevel=debug,provider=cloudflare,cloudflare.apiKey=SECRET_KEY,cloudflare.email=user@example.com stable/external-dns
```
