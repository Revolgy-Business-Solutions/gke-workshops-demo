# PHP FPM & nginx connection using unix socket

## Configuration files

For unix socket connection both nginx and php-fpm configuration needs to be setup correctly. see manifests **nginx.conf.yaml** and **www.conf.yaml**, especially upstream and php location settings on nginx side and "listen" line on php config side.

Furthemore, default configuration of "zz-docker.conf" is overwriting the setup of configs mentioned above so in this case we will overwrite it with our own config in Dockerfile.

## Unix socket connection

For unix socket connection between containers in kubernetes we are using mounted volume of type **emptydir**. first we need to mount the same volume into both containers

```        
        volumeMounts:
        - name: unix-sock
          mountPath: /sock
```

and then define corresponding volume

```
      volumes:
      - name: shared-files
        emptyDir: {}
      - name: unix-sock
``` 

