> [!NOTE]
> The repository has been renamed from `prometheus-configs-provider` to `prometheus-config-provider`.

## About

This is a simple tool that reads the Docker API and generates Prometheus scrape configs for all Docker configs that have the `io.prometheus.scrape_config` label set to `true`.

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://github.com/swarmlibs/prometheus-configs-provider/assets/4363857/5e790dd2-0d06-434a-98f7-a1e412388c96">
  <source media="(prefers-color-scheme: light)" srcset="https://github.com/swarmlibs/prometheus-configs-provider/assets/4363857/d439c204-fec4-492a-99f7-20df95ae1217">
  <img src="https://github.com/swarmlibs/prometheus-configs-provider/assets/4363857/d439c204-fec4-492a-99f7-20df95ae1217">
</picture>

## Usage

```sh
usage: prometheus-configs-provider [<flags>]

Flags:
  --[no-]help           Show context-sensitive help (also try --help-long and --help-man).
  --output-dir="/etc/prometheus/configs"  
                        directory for the configs
  --output-ext="yaml"   extension for the configs
  --prometheus-scrape-config-label="io.prometheus.scrape_config"  
                        label to identify prometheus scrape configs
  --[no-]version        Prints current version.
  --[no-]short-version  Print just the version number.
```

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

## License

Licensed under [MIT](./LICENSE).
