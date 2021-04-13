Exporter Test
=============

This is a Prometheus Exporter testing application using [Open
Telemetry](https://opentelemetry.io/) (OTEL) instrumentation for the metrics.


Usage
-----

Add Helm chart repo:

```shell
helm repo add exporter-test https://jtyr.github.io/exporter-test
```

Install Helm chart:

```shell
helm upgrade --create-namespace --namespace exporter-test --install exporter-test/exporter-test
```

Test the `metrics` endpoint:

```shell
kubectl run curl \
    --image curlimages/curl \
    --restart=Never \
    --rm \
    --tty \
    --stdin \
    --command -- \
    curl http://exporter-test.exporter-test/metrics
```


License
-------

MIT


Author
------

Jiri Tyr
