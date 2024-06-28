## About

This is a simple tool that reads the Docker API and generates Prometheus scrape configs for all Docker configs that have the `io.prometheus.scrape_config` label set to `true`.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://github.com/swarmlibs/prometheus-configs-provider/assets/4363857/beff289a-ef7b-4c43-b026-5bb89625e77f">
  <source media="(prefers-color-scheme: light)" srcset="https://github.com/swarmlibs/prometheus-configs-provider/assets/4363857/9c697fc3-f876-461f-bfd5-e3d9486f4e77">
  <img src="https://github.com/swarmlibs/prometheus-configs-provider/assets/4363857/9c697fc3-f876-461f-bfd5-e3d9486f4e77">
</picture>

## Example

```yaml
# docker-stack.yml

configs:
  example-scrape-config:
    file: ./example-scrape-config.yaml
    labels:
      - "io.prometheus.scrape_config=true"
```
