> [!NOTE]
> The repository has been renamed from `prometheus-configs-provider` to `prometheus-config-provider`.

## About

This is a simple tool that reads the Docker API and generates Prometheus scrape configs for all Docker configs that have the `io.prometheus.scrape_config` label set to `true`.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://github.com/swarmlibs/prometheus-configs-provider/assets/4363857/5e790dd2-0d06-434a-98f7-a1e412388c96">
  <source media="(prefers-color-scheme: light)" srcset="https://github.com/swarmlibs/prometheus-configs-provider/assets/4363857/d439c204-fec4-492a-99f7-20df95ae1217">
  <img src="https://github.com/swarmlibs/prometheus-configs-provider/assets/4363857/d439c204-fec4-492a-99f7-20df95ae1217">
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

With the help of the [prometheus-operator/prometheus-operator/tree/main/cmd/prometheus-config-reloader](https://github.com/prometheus-operator/prometheus-operator/tree/main/cmd/prometheus-config-reloader) tool, we can automatically reload the Prometheus configuration when the Docker configs are create/update/remove.
