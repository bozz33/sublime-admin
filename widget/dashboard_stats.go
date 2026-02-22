package widget

// GetDashboardStats generates widgets for the dashboard page.
// By default, the dashboard is empty.
// Override this in your project by registering widgets via the Panel configuration.
func GetDashboardStats() []Widget {
	// Empty dashboard by default.
	// To add widgets, uncomment and customize the code below:

	/*
		var widgets []Widget

		userCount, err := client.User.Query().Count(ctx)
		if err != nil {
			userCount = 0
		}

		stats := NewStats(
			Stat{
				Label:       "Total Users",
				Value:       fmt.Sprintf("%d", userCount),
				Description: "+12% this month",
				Icon:        "users",
				Color:       "primary",
				Increase:    true,
				Chart:       []int{10, 25, 40, 30, 45, 50, 70, 85},
			},
		)
		widgets = append(widgets, stats)

		return widgets
	*/

	return []Widget{}
}
