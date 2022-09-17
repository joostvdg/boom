# Logs To Fix

## Not Stopping Heartbeat Requests and Failure Propagation In Time

```shell
We are not able to reach &{Alan Boreas 172.30.169.33 7780 0.0.0.0 2022-09-13 23:25:38.841450861 +0200 CEST m=+22.667857276 0} for 5 times, initiating failure propagation
Received Multicast from Member: Emily @Boreas(172.30.169.33:7784 / 0.0.0.0)
Sending heartbeat request message to Alan @172.30.169.33:7780
We are not able to reach &{Alan Boreas 172.30.169.33 7780 0.0.0.0 2022-09-13 23:25:38.841450861 +0200 CEST m=+22.667857276 0} for 5 times, initiating failure propagation
Sending heartbeat request message to Alan @172.30.169.33:7780
We are not able to reach &{Alan Boreas 172.30.169.33 7780 0.0.0.0 2022-09-13 23:25:38.841450861 +0200 CEST m=+22.667857276 0} for 5 times, initiating failure propagation
Sending heartbeat request message to Alan @172.30.169.33:7780
We are not able to reach &{Alan Boreas 172.30.169.33 7780 0.0.0.0 2022-09-13 23:25:38.841450861 +0200 CEST m=+22.667857276 0} for 5 times, initiating failure propagation
```

Once we enter Failure propagation, stop sending hearbeat requests
if we receive a failure propagation from others for the same node, stop and remove the node