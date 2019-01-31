# PHP-FPM & nginx connection using tcp socket

PHP-FPM default "www.conf" configuration is set for listening on tcp localhost:9000. For nginx configuration we need to set fastcgi pass on this port"

Both configuration are prepare in yaml manifest (nginx.conf.yaml, www.conf.yaml) and injected into deployment (app.yaml) as configmaps
