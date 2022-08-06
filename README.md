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

## Resources

* https://www.digitalocean.com/community/tutorials/how-to-use-contexts-in-go
* https://www.digitalocean.com/community/tutorials/understanding-maps-in-go
* https://medium.com/@leonardo5621_66451/how-to-shutdown-a-golang-application-in-a-cleaner-way-e9307b0ea505