This directory contains an example of a durable coroutine. The
coroutine will recursively scrape a website given an input URL
(Wikipedia in this case) and will yield URLs that have been
scraped. The durable coroutine can be restarted and it will
resume from the last yield point.

Build the durable coroutine:

```console
make
```

To run the example:

```console
./scrape
```
