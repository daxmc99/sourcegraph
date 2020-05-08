package main

func PreciseCodeIntelBundleManager() *Container {
	return &Container{
		Name:        "precise-code-intel-bundle-manager",
		Title:       "Precise Code Intel Bundle Manager",
		Description: "Stores and manages precise code intelligence bundles.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:        "99th_percentile_database_duration",
							Description: "99th percentile successful database query duration over 5m",
							// TODO - by op?
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_precise_code_intel_bundle_manager_database_duration_seconds_bucket[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("duration").Unit(Seconds),
							PossibleSolutions: "none",
						},
						{
							Name:              "database_errors",
							Description:       "database errors every 5m",
							Query:             `sum(sum by (op)(increase(src_precise_code_intel_bundle_manager_database_errors_total[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:        "99th_percentile_reader_duration",
							Description: "99th percentile successful reader query duration over 5m",
							// TODO - by op?
							Query:             `histogram_quantile(0.99, sum by (le)(rate(src_precise_code_intel_bundle_manager_reader_duration_seconds_bucket[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("duration").Unit(Seconds),
							PossibleSolutions: "none",
						},
						{
							Name:              "reader_errors",
							Description:       "reader errors every 5m",
							Query:             `sum(sum by (op)(increase(src_precise_code_intel_bundle_manager_reader_errors_total[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:            "disk_space_remaining",
							Description:     "disk space remaining by instance",
							Query:           `(src_disk_space_available_bytes / src_disk_space_total_bytes) * 100`,
							DataMayNotExist: true,
							Warning:         Alert{LessOrEqual: 25},
							Critical:        Alert{LessOrEqual: 15},
							PanelOptions:    PanelOptions().LegendFormat("{{instance}}").Unit(Percentage),
							PossibleSolutions: `
								TODO - update this
								- **Provision more disk space:** Sourcegraph will begin deleting least-used repository clones at 10% disk space remaining which may result in decreased performance, users having to wait for repositories to clone, etc.
							`,
						},
					},
					{
						{
							Name:              "cache_cost",
							Description:       "cache cost",
							Query:             `src_cache_cost`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{cache}}"),
							PossibleSolutions: "none",
						},
						{
							Name:              "cache_hits",
							Description:       "cache hits every 5m",
							Query:             `increase(src_cache_hits[5m])`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{cache}}"),
							PossibleSolutions: "none",
						},
						{
							Name:              "cache_misses",
							Description:       "cache misses every 5m",
							Query:             `increase(src_cache_misses[5m])`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{cache}}"),
							PossibleSolutions: "none",
						},
					},
					{
						{
							Name:              "janitor_errors",
							Description:       "janitor errors every 5m",
							Query:             `sum(sum by (op)(increase(src_precise_code_intel_bundle_manager_janitor_errors[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							PossibleSolutions: "none",
						},
						{
							Name:              "janitor_old_dumps",
							Description:       "janitor old dumps every 5m",
							Query:             `sum(sum by (op)(increase(src_precise_code_intel_bundle_manager_janitor_old_dumps[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							PossibleSolutions: "none",
						},
						{
							Name:              "janitor_old_uploads",
							Description:       "janitor old uploads every 5m",
							Query:             `sum(sum by (op)(increase(src_precise_code_intel_bundle_manager_janitor_old_uploads[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							PossibleSolutions: "none",
						},
						{
							Name:              "janitor_orphaned_dumps",
							Description:       "janitor orphaned dumps every 5m",
							Query:             `sum(sum by (op)(increase(src_precise_code_intel_bundle_manager_janitor_orphaned_dumps[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							PossibleSolutions: "none",
						},
					},
					{
						sharedFrontendInternalAPIErrorResponses("precise-code-intel-bundle-manager"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("precise-code-intel-bundle-manager"),
						sharedContainerMemoryUsage("precise-code-intel-bundle-manager"),
						sharedContainerCPUUsage("precise-code-intel-bundle-manager"),
					},
				},
			},
		},
	}
}
