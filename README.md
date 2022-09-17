# boom

Practice application for managing a distributed log of software assets

## TODO

### Now

* streamline heartbeat and multicast
  * test
* inform failure
  * send missing member message to shortlist
  * these then first probe (heartbeat) the perceived failed member (node)
* Create OCI Image
  * Hadolint
  * Seccom Profiles
  * Carvel Package?
    * use dev version of `kctrl`
* setup minimal CI
  * TektonCD
  * Syft for SBOM
  * hadolint
* Deploy as DaemonSet or Deployment
  * deployment via FluxCD?

### Next

* consensus protocol
  * https://www.hashicorp.com/resources/raft-consul-consensus-protocol-explained
  * https://thesecretlivesofdata.com/raft/
  * discovery: multi modes (kubernetes, via client?)
  * internal clock for agreeing on message ordering
  * leadership election
  * distributed log
  * lampart clock
    * https://jakub-m.github.io/2022/07/17/laport-clocks-formal.html
    * https://sookocheff.com/post/time/lamport-clock/
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

### On Distributed Computing

* https://www.coursera.org/learn/cloud-computing/home/week/2
* https://www.serf.io/docs/internals/gossip.html#lifeguard
* https://www.hashicorp.com/blog/making-gossip-more-robust-with-lifeguard
* https://github.com/hashicorp/memberlist
* https://martinfowler.com/articles/patterns-of-distributed-systems/gossip-dissemination.html
* https://martinfowler.com/articles/patterns-of-distributed-systems/lamport-clock.html
* https://www.serf.io/docs/internals/simulator.html
* https://www.brianstorti.com/swim/

### On Go

* https://golang.ch/better-error-handling-with-golang/
* https://thenotexpert.com/simple-server-binary-protocol-golang/
* https://www.linode.com/docs/guides/developing-udp-and-tcp-clients-and-servers-in-go/
* https://fossa.com/blog/golang-best-practices-testing-go/
* https://go.dev/security/fuzz/
 
### On Telemetry

* https://www.robustperception.io/understanding-machine-cpu-usage/
* https://medium.com/@magstherdev/opentelemetry-on-kubernetes-c167f024b35f
* https://opentelemetry.io/docs/instrumentation/go/exporting_data/