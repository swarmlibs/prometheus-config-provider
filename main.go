package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/oklog/run"
	"github.com/prometheus-operator/prometheus-operator/pkg/versionutil"
	"github.com/prometheus/common/version"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

var (
	defaultOutputDir                   = "/etc/prometheus/configs"
	defaultOuputExt                    = "yaml"
	defaultPrometheusScrapeConfigLabel = "io.prometheus.scrape_config"
	defaultEvaluationInterval          = 15 * time.Second
)

func main() {
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	app := kingpin.New("prometheus-configs-provider", "")

	outputDir := app.Flag("output-dir", "directory for the configs").Default(defaultOutputDir).String()
	outputExt := app.Flag("output-ext", "extension for the configs").Default(defaultOuputExt).String()
	keepExisting := app.Flag("keep-existing", "keep existing files in output directory").Bool()
	evaluationInterval := app.Flag("evaluation_interval", "How frequently to evaluate service configs").Default(defaultEvaluationInterval.String()).Duration()
	prometheusScrapeConfigLabel := app.Flag("prometheus-scrape-config-label", "label to identify prometheus scrape configs").Default(defaultPrometheusScrapeConfigLabel).String()

	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stdout)
	logger = level.NewFilter(logger, level.AllowAll())
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	versionutil.RegisterIntoKingpinFlags(app)

	if versionutil.ShouldPrintVersion() {
		versionutil.Print(os.Stdout, "prometheus-config-reloader")
		os.Exit(0)
	}

	if _, err := app.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(2)
	}

	level.Info(logger).Log("msg", "Starting prometheus-configs-provider", "version", version.Info())
	level.Info(logger).Log("build_context", version.BuildContext())

	var (
		g           run.Group
		ctx, cancel = context.WithCancel(context.Background())
	)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Check if outputDir exists, if not create it
	if _, err := os.Stat(*outputDir); os.IsNotExist(err) {
		level.Info(logger).Log("msg", "Creating output directory")
		if err := os.Mkdir(*outputDir, 0755); err != nil {
			level.Error(logger).Log("msg", "Failed to create output directory", "err", err)
			os.Exit(1)
		}
	} else {
		if !*keepExisting {
			level.Info(logger).Log("msg", "Cleaning up existing files in output directory")
			files, _ := os.ReadDir(*outputDir)
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				if err := os.Remove(fmt.Sprintf("%s/%s", *outputDir, file.Name())); err != nil {
					level.Error(logger).Log("msg", "Failed to remove file", "file", file.Name(), "err", err)
				}
			}
			time.Sleep(1 * time.Second)
		} else {
			level.Info(logger).Log("msg", "Keeping existing files in output directory")
		}
	}

	// Run the main loop
	mainCh := make(chan struct{})
	g.Add(func() error {
		<-mainCh
		cancel()
		cli.Close()
		return nil
	}, func(error) {})

	// Run the ticker to evaluate service configs
	serviceCh := make(chan struct{})
	g.Add(func() error {
		ticker := time.NewTicker(*evaluationInterval)
		for {
			select {
			case <-ticker.C:
				// Get all services
				services, err := cli.ServiceList(ctx, types.ServiceListOptions{})
				if err != nil {
					level.Error(logger).Log("msg", "Failed to list services", "err", err)
					os.Exit(1)
				}

				// Loop through all services and get the configs
				for _, service := range services {
					// Check if the service has a previous spec
					// If it does, we need to check the previous spec for configs
					// and remove them if they don't exist in the current spec
					if service.PreviousSpec != nil {
					prevConfigLoop:
						for _, prevConfig := range service.PreviousSpec.TaskTemplate.ContainerSpec.Configs {
							// Check if the config exists in the current spec
							for _, config := range service.Spec.TaskTemplate.ContainerSpec.Configs {
								if prevConfig.ConfigID == config.ConfigID {
									continue prevConfigLoop
								}
							}

							cfg, _, err := cli.ConfigInspectWithRaw(ctx, prevConfig.ConfigID)
							if err != nil {
								level.Error(logger).Log("msg", "Failed to read config", "type", "service", "id", prevConfig.ConfigID, "err", err)
								continue
							}

							if cfg.Spec.Labels[*prometheusScrapeConfigLabel] == "" {
								continue
							}

							configName := cfg.Spec.Name
							if cfg.Spec.Labels[*prometheusScrapeConfigLabel+".name"] != "" {
								configName = cfg.Spec.Labels[*prometheusScrapeConfigLabel+".name"]
							}

							// Prepare the output file name
							outFile := fmt.Sprintf("%s/%s.%s", *outputDir, configName, *outputExt)

							// Remove the config file if exists in the output directory
							if _, err := os.Stat(outFile); err == nil {
								os.Remove(outFile)
								level.Info(logger).Log("msg", "Removing config", "type", "service", "id", cfg.ID, "name", cfg.Spec.Name, "file", outFile)
							}
						}
					}

					// Get the service configs from the current task template
					for _, config := range service.Spec.TaskTemplate.ContainerSpec.Configs {
						cfg, _, err := cli.ConfigInspectWithRaw(ctx, config.ConfigID)
						if err != nil {
							level.Error(logger).Log("msg", "Failed to read config", "type", "service", "id", config.ConfigID, "err", err)
							continue
						}

						if cfg.Spec.Labels[*prometheusScrapeConfigLabel] == "" {
							continue
						}

						// Ability to override the config name with a label
						// e.g. io.prometheus.scrape_config.name
						configName := cfg.Spec.Name
						if cfg.Spec.Labels[*prometheusScrapeConfigLabel+".name"] != "" {
							configName = cfg.Spec.Labels[*prometheusScrapeConfigLabel+".name"]
						}

						// Prepare the output file name
						outFile := fmt.Sprintf("%s/%s.%s", *outputDir, configName, *outputExt)

						// Write the config to file if it doesn't exist
						if _, err := os.Stat(outFile); os.IsNotExist(err) {
							writeConfigToFile(outFile, cfg.Spec.Data)
							level.Info(logger).Log("msg", "Creating config", "type", "service", "id", cfg.ID, "name", cfg.Spec.Name, "file", outFile)
						}
					}
				}
			case <-serviceCh:
				ticker.Stop()
				return nil
			}
		}
	}, func(error) {})

	// Subscribe to Docker events for configs
	g.Add(func() error {
		filters := filters.NewArgs()
		filters.Add("type", "config")
		events, errCh := cli.Events(ctx, events.ListOptions{
			Filters: filters,
		})

		for {
			select {
			case event := <-events:
				switch event.Action {
				case "remove":
					configName := event.Actor.Attributes["name"]
					if event.Actor.Attributes[*prometheusScrapeConfigLabel+".name"] != "" {
						configName = event.Actor.Attributes[*prometheusScrapeConfigLabel+".name"]
					}

					outFile := fmt.Sprintf("%s/%s.%s", *outputDir, configName, *outputExt)

					// Remove the config file if exists in the output directory
					if _, err := os.Stat(outFile); err == nil {
						os.Remove(outFile)
						level.Info(logger).Log("msg", "Removing config", "type", "event", "id", event.Actor.ID, "name", event.Actor.Attributes["name"], "file", outFile)
					}
				}
			case err := <-errCh:
				level.Error(logger).Log("msg", "Failed to receive Docker events", "err", err)
				return err
			}
		}
	}, func(error) {})

	// Handle Interrupt & SIGTERM signals
	g.Add(func() error {
		select {
		case <-term:
			close(mainCh)
			close(serviceCh)
			level.Info(logger).Log("msg", "Received SIGTERM, exiting gracefully...")
		case <-ctx.Done():
		}

		return nil
	}, func(error) {})

	if err := g.Run(); err != nil {
		level.Error(logger).Log("msg", "Failed to run", "err", err)
		os.Exit(1)
	}
}

func writeConfigToFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	file.WriteString(string(data))
	file.Chmod(0777)
	return nil
}
