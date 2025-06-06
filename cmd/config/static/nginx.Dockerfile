FROM nginx:alpine

RUN apk add --no-cache \
  gcc \
  libc-dev \
  make \
  openssl-dev \
  pcre-dev \
  zlib-dev \
  linux-headers \
  curl \
  tar \
  wget

WORKDIR /tmp
RUN wget https://nginx.org/download/nginx-1.25.3.tar.gz && \
  wget https://github.com/aperezdc/ngx-fancyindex/archive/refs/tags/v0.5.2.tar.gz

RUN tar -xzf nginx-1.25.3.tar.gz && \
  tar -xzf v0.5.2.tar.gz

WORKDIR /tmp/nginx-1.25.3
RUN ./configure \
  --prefix=/etc/nginx \
  --sbin-path=/usr/sbin/nginx \
  --modules-path=/usr/lib/nginx/modules \
  --conf-path=/etc/nginx/nginx.conf \
  --error-log-path=/var/log/nginx/error.log \
  --http-log-path=/var/log/nginx/access.log \
  --pid-path=/var/run/nginx.pid \
  --lock-path=/var/run/nginx.lock \
  --http-client-body-temp-path=/var/cache/nginx/client_temp \
  --http-proxy-temp-path=/var/cache/nginx/proxy_temp \
  --http-fastcgi-temp-path=/var/cache/nginx/fastcgi_temp \
  --http-uwsgi-temp-path=/var/cache/nginx/uwsgi_temp \
  --http-scgi-temp-path=/var/cache/nginx/scgi_temp \
  --with-http_ssl_module \
  --with-http_realip_module \
  --with-http_addition_module \
  --with-http_sub_module \
  --with-http_dav_module \
  --with-http_flv_module \
  --with-http_mp4_module \
  --with-http_gunzip_module \
  --with-http_gzip_static_module \
  --with-http_random_index_module \
  --with-http_secure_link_module \
  --with-http_stub_status_module \
  --with-http_auth_request_module \
  --with-threads \
  --add-module=/tmp/ngx-fancyindex-0.5.2 \
  && make \
  && make install

RUN rm -rf /tmp/nginx-1.25.3 /tmp/ngx-fancyindex-0.5.2 \
  && rm /tmp/nginx-1.25.3.tar.gz /tmp/v0.5.2.tar.gz

COPY nginx.conf /etc/nginx/nginx.conf
COPY ./fancyindex /var/www/servers/cansu.dev/static/fancyindex

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]