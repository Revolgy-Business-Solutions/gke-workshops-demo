# PHP FPM & nginx demo application - tcp & unix connection

## Description

Example php application with 3-container deployment (php-fpm, nginx, cloudsqlproxy). Each folder contain the same application files. Difference is in Dockerfile between the tcp and unix version.

For testing purposes you can build and deploy application and test on endpoints (database credentials are, for testing purposes just hardcoded in db.php file):

```
_dns/index.html_  for getting file dirrectly from nginx container
_dns/hello.php_   for getting file from php-fpm container
_dns/db.php_      for testing cloudsql connection (correct database, credentials on database/db.php side and proper cloudsql proxy configuration)
```

## File distribution and sharing

There are multiple solution for files distribution. In this case, we are copying all files into shared emptyDir volume after the container starts. That allows both nginx and php-fpm access same files depending on if its php file or not.

Another solution is to have same files prepared in both containers or have php files only in php-fpm container with turned off try-files check. 

Each solution has its advatages and disadvatages. 

## Nginx & PHP config files

each example include yaml manifest for nginx and php-fpm configuration which are needed for proper connection between php-fpm and nginx. Configurations are either for tcp connection between containers or unix socket connection. You will deploy these configurations with 

```
kubectl apply -f nginx.conf.yaml
kubectl apply -f www.conf.yaml
```

Configuration are then mounted into containers to proper path with:

```
        volumeMounts:
        - name: php-fpm-config
          mountPath: /usr/local/etc/php-fpm.d/www.conf
          subPath: www.conf
```
with corresponding volume specification:

```
      volumes:
      - name: php-fpm-config
        configMap:
          name: php-fpm-config
```

## CloudSQL configuration

```
NOTE: At this point we don't have to use cloudSQL proxy inside GKE provided that we turned on "VPC-native" on cluster creation and "Private IP" on CloudSQL creation. With this settings we can use directly private ip of cloudsql from inside GKE
```

In both examples we use tcp localhost connection between php and cloudSQL proxy. For cloudsql-proxy is used official image from google. For authorization into CloudSQL it is mounting a secret with json credentials that have appropriate permission (CloudSQL client)

```
        volumeMounts:
          - name: cloudsql-instance-credentials
            mountPath: /secrets/cloudsql
            readOnly: true
```

with corresponding volume specification:

```
      volumes:
      - name: cloudsql-instance-credentials
        secret:
          secretName: cloudsql-instance-credentials
```

for more information about setting cloudsql proxy on kubernetes see [Connecting from Google Kubernetes Engine](https://cloud.google.com/sql/docs/mysql/connect-kubernetes-engine)

This demo application has database credentials to sql database included in source files. This is of course not a correct solution. Ideally they should be injected as variables from kubernetes secrets [Kubernetes variables from secret](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables)
