FROM php:7.2-fpm

RUN mkdir /app

RUN mv "$PHP_INI_DIR/php.ini-production" "$PHP_INI_DIR/php.ini"

RUN docker-php-ext-install mysqli && docker-php-ext-enable mysqli


COPY hello.php /app
COPY index.html /app
COPY db.php /app
