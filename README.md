cache-got-hit
=============

Let Nginx talk to Redis cache server with modified memcached protocol.

This needs to apply a patch to Nginx 1.9.x, for example:

```
tar xfvz nginx-1.9.15.tar.gz
patch nginx-1.9.15/src/http/modules/ngx_http_memcached_module.c ~/work/cache-got-hit/ngx_http_memcached_module.c.patch
```

Run cache-got-hit:

```
cache-got-hit
```

Sample nginx config:

```nginx
server {
	listen 9999;

	location / {
		set            $memcached_key $request_uri;
		memcached_pass 127.0.0.1:8080;
		error_page     404 502 504 = @fallback;
		gzip_types     "application/json";
		gzip           on;
	}

	location @fallback {
		proxy_pass     http://127.0.0.1:3000;
	}
}
```

Protocol:

```
nginx

  ---requests-->

    get /api/performances

      ---cache-got-hit-->

        VALUE /api/performances 0 261
        -----CACHE GOT HIT HEADERS INCLUDED-----
        Content-Type: application/json; charset=utf-8
        Cache-Got-Hit-Status: 200
        X-Frame-Options: SAMEORIGIN
        X-XSS-Protection: 1; mode=block
        X-Content-Type-Options: nosniff
        ETag: W/"35d0e0b801444bb1f4c2115720f8e4f6"

        []

        END

        ---nginx-->

          HTTP/1.1 200 OK
          Server: nginx/1.9.15
          Date: Sun, 24 Apr 2016 15:57:18 GMT
          Content-Type: application/json; charset=utf-8
          Content-Length: 4
          Connection: keep-alive
          X-Frame-Options: SAMEORIGIN
          X-XSS-Protection: 1; mode=block
          X-Content-Type-Options: nosniff
          ETag: W/"35d0e0b801444bb1f4c2115720f8e4f6"
          Accept-Ranges: bytes

          []

  ---requests-->

    get /api/performances W/"35d0e0b801444bb1f4c2115720f8e4f6"

      ---cache-got-hit-->

        VALUE /api/performances 0 257
        -----CACHE GOT HIT HEADERS INCLUDED-----
        Content-Type: application/json; charset=utf-8
        Cache-Got-Hit-Status: 304
        X-Frame-Options: SAMEORIGIN
        X-XSS-Protection: 1; mode=block
        X-Content-Type-Options: nosniff
        ETag: W/"35d0e0b801444bb1f4c2115720f8e4f6"


        END

        ---nginx-->

          HTTP/1.1 304 Not Modified
          Server: nginx/1.9.15
          Date: Sun, 24 Apr 2016 15:58:03 GMT
          Content-Type: application/json; charset=utf-8
          Content-Length: 0
          Connection: keep-alive
          X-Frame-Options: SAMEORIGIN
          X-XSS-Protection: 1; mode=block
          X-Content-Type-Options: nosniff
          ETag: W/"35d0e0b801444bb1f4c2115720f8e4f6"

```

Inspired by
[Enhanced Nginx Memcached Module](https://github.com/bpaquet/ngx_http_enhanced_memcached_module)
and
[mrproxy](https://github.com/zobo/mrproxy).
