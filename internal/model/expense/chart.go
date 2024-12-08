package expense

import (
	"fmt"
	"time"
)

type ChartData struct {
	Labels   [][]string     `json:"labels"`
	Datasets []CategoryData `json:"datasets"`
}

type CategoryData struct {
	Label string    `json:"label"`
	Data  []float64 `json:"data"`
}

func getLastSixMonths() ([]string, []string) {
	months := []string{}
	monthKeys := []string{}
	currentTime := time.Now()

	for i := 0; i < 6; i++ {
		monthName := currentTime.Format("January")
		monthKey := currentTime.Format("2006-01")

		months = append([]string{monthName}, months...)
		monthKeys = append([]string{monthKey}, monthKeys...)

		currentTime = currentTime.AddDate(0, -1, 0)
	}

	return months, monthKeys
}

func TransformToChartData(data []MonthlySum) ChartData {
	months, monthKeys := getLastSixMonths()
	categoryMap := map[string]map[string]float64{}

	for _, record := range data {
		monthYear := record.SK[:7]

		if categoryMap[record.Category] == nil {
			categoryMap[record.Category] = make(map[string]float64)
		}

		categoryMap[record.Category][monthYear] = record.Sum
	}

	datasets := []CategoryData{}
	for category, monthData := range categoryMap {
		dataPoints := make([]float64, len(monthKeys))
		for i, monthKey := range monthKeys {
			dataPoints[i] = monthData[monthKey]
		}

		datasets = append(datasets, CategoryData{
			Label: category,
			Data:  dataPoints,
		})
	}

	monthAmounts := make([]float64, len(months))

	for _, dataset := range datasets {
		for k, v := range dataset.Data {
			monthAmounts[k] += v
		}
	}

	labels := [][]string{}

	for i, month := range months {
		labels = append(labels, []string{
			fmt.Sprintf("%d PLN", int(monthAmounts[i])),
			month,
		})
	}

	return ChartData{
		Labels:   labels,
		Datasets: datasets,
	}
}
