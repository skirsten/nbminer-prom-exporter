## Super simple prometheus exporter for NBMiner

Currently only the hashrate is exposed. I might add the rest later

### Use in k8s sidecar container

```yaml
metadata:
  ...
  annotations:
    prometheus.io/port: "2121"
    prometheus.io/scrape: "true"
spec:
  containers:
    - name: prom-exporter
      image: ghcr.io/skirsten/nbminer-prom-exporter@...
      ports:
        - name: metrics
          containerPort: 2121

    - ... nbminer with --api "0.0.0.0:8080"
```

### Publish the image

```sh
KO_DOCKER_REPO=ghcr.io/skirsten ko build  -B --platform=all github.com/skirsten/nbminer-prom-exporter
```
