package main

func PreciseCodeIntelWorker() *Container {
	return &Container{
		Name:        "precise-code-intel-worker",
		Title:       "Precise Code Intel Worker",
		Description: "Handles conversion of uploaded precise code intelligence bundles.",
		Groups: []Group{
			{
				Title: "General",
				Rows: []Row{
					{
						{
							Name:              "upload_queue_size",
							Description:       "upload queue size",
							Query:             `max(src_precise_code_intel_worker_queue_size)`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 100},
							Critical:          Alert{GreaterOrEqual: 200},
							PanelOptions:      PanelOptions().LegendFormat("uploads queued for processing"),
							PossibleSolutions: "none",
						},
						{
							Name:              "upload_processed",
							Description:       "total uploads processed every 5m",
							Query:             `sum(increase(src_precise_code_intel_worker_jobs_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{LessOrEqual: -1},
							PanelOptions:      PanelOptions().LegendFormat("uploads processed"),
							PossibleSolutions: "none",
						},
						{
							Name:        "upload_process_errors",
							Description: "upload process errors every 5m",
							// TODO(efritz) - ensure these differentiate malformed dumps and system errors
							Query:             `sum(increase(src_precise_code_intel_worker_jobs_errors_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("errors"),
							PossibleSolutions: "none",
						},
					},
					// TODO(efritz) - add bundle manager request meter
					// TODO(efritz) - add gitserver request meter
					{
						{
							Name:        "99th_percentile_db_duration",
							Description: "99th percentile successful db query duration over 5m",
							// TODO(efritz) - ensure these exclude error durations
							Query:             `histogram_quantile(0.99, sum by (le,op)(rate(src_precise_code_intel_worker_db_duration_seconds_bucket[5m])))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{op}}").Unit(Seconds),
							PossibleSolutions: "none",
						},
						{
							Name:              "db_errors",
							Description:       "db errors every 5m",
							Query:             `sum by (op)(increase(src_precise_code_intel_worker_db_errors_total[5m]))`,
							DataMayNotExist:   true,
							Warning:           Alert{GreaterOrEqual: 5},
							Critical:          Alert{GreaterOrEqual: 20},
							PanelOptions:      PanelOptions().LegendFormat("{{op}}"),
							PossibleSolutions: "none",
						},
					},
					{
						sharedFrontendInternalAPIErrorResponses("precise-code-intel-worker"),
					},
				},
			},
			{
				Title:  "Container monitoring (not available on k8s or server)",
				Hidden: true,
				Rows: []Row{
					{
						sharedContainerRestarts("precise-code-intel-worker"),
						sharedContainerMemoryUsage("precise-code-intel-worker"),
						sharedContainerCPUUsage("precise-code-intel-worker"),
					},
				},
			},
		},
	}
}
