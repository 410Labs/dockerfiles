# Example nginx-graphite-proxy image

This is simple example of how to configure nginx to proxy to graphite-web.

 * based on official nginx image
 * copy in the static assets from the graphite-web image.
 * basic nginx server directive, serving static files, proxy_pass to graphite.

To use it, `graphite-web` must be available at `graphite`.  Use docker
networking DNS with services, or `--add-host`, or `--link` to accomplish this.

If you want e.g. HTTP Basic Auth to prevent unauthenticated access to graphite,
you can easily add that to the config file.
