## About

This is a simple tool that reads the Docker API and generates Prometheus scrape configs for all Docker configs that have the `io.prometheus.scrape_config` label set to `true`.

## Example

```yaml
# docker-stack.yml

configs:
  example-scrape-config:
    file: ./example-scrape-config.yaml
    labels:
      - "io.prometheus.scrape_config=true"
```
