# Graphite docker image

Similar to the official graphite image, but:
 * squashes all of the `RUN` layers
 * removes nginx and statsd. run those seperately.
 * removes statsd
 * config files are not volumes (by default; you are free to override at runtime)
 * TODO: seperate graphite, carbon, and carbon-aggregator into seperate containers.
 * TODO: not based on phusion base image? (need to ensure logging is properly handled)
