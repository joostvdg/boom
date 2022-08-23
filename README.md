# boom

Practice application for managing a distributed log of software assets

## TODO

### Now

* server shutdown
  * let server shutdown gracefully
  * let server say goodbye when _gracefully shutting down_
  * remove from membership list
* Member heartbeat
  * create a short list
  * send heartbeat to short list
* streamline heartbeat and multicast
  * test
* inform failure
  * send missing member message to shortlist


### Next

* consensus protocol
  * discovery: multi modes (kubernetes, via client?)
  * leadership election
  * distributed log
* store application identity
  * certificates
  * tags to use
* map application identity to (S)BOM
  * Syft (docker bom)
  * SPDX
  * should (S)BOM be Timeseries?
  * look at Grafana's Mimix?
* log where application is running/encountered?
  * maybe a Spring Native with a Database with JOOQ and so on?
* work with Backstage?

## Requirements

* can run as standalone process: Linux AMD64, Linux ARM64, Windows AMD64
* can run as container: Linux AMD64, Linux ARM64
* can run as Kubernetes StatefulSet
* GitHub actions pipeline(s) with linting, security scan
* Tekton pipeline, based on Continous Delivery with Kubernetes book

## Jaeger for Tracing

### Run In Kubernetes

* https://artifacthub.io/packages/helm/jaegertracing/jaeger

```shell
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
```

```shell
helm install jaeger jaegertracing/jaeger --namespace  jaeger --
```

### Run Local Jaeger

As described here: https://www.jaegertracing.io/docs/1.37/getting-started/

```shell
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 14250:14250 \
  -p 14268:14268 \
  -p 14269:14269 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.37
```

```shell
open http://localhost:16686
```

## Resources

* https://www.digitalocean.com/community/tutorials/how-to-use-contexts-in-go
* https://www.digitalocean.com/community/tutorials/understanding-maps-in-go
* https://medium.com/@leonardo5621_66451/how-to-shutdown-a-golang-application-in-a-cleaner-way-e9307b0ea505
* https://opentelemetry.io/docs/instrumentation/go/getting-started/
* https://github.com/open-telemetry/opentelemetry-go/blob/main/example/jaeger/main.go
* https://www.jaegertracing.io/docs/1.37/getting-started/