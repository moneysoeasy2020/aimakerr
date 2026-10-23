package gitlab

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAlertManagement_UploadMetricImage(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/alert_management_alerts/2/metric_images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `
		{
			"id":17,
			"created_at":"2020-11-12T20:07:58.000Z",
			"filename":"sample_2054",
			"file_path":"/uploads/-/system/alert_metric_image/file/17/sample_2054.png",
			"url":"https://example.com/metric",
			"url_text":"An example metric"
		}
		`)
	})

	createdAt := time.Date(2020, time.November, 12, 20, 7, 58, 0, time.UTC)
	want := &MetricImage{
		ID:        17,
		CreatedAt: &createdAt,
		Filename:  "sample_2054",
		FilePath:  "/uploads/-/system/alert_metric_image/file/17/sample_2054.png",
		URL:       "https://example.com/metric",
		URLText:   "An example metric",
	}

	metricImage, resp, err := client.AlertManagement.UploadMetricImage(1, 2, strings.NewReader("image"), "sample_2054", &UploadMetricImageOptions{
		URL:     Ptr("https://example.com/metric"),
		URLText: Ptr("An example metric"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, metricImage)
}

func TestAlertManagement_ListMetricImages(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/alert_management_alerts/2/metric_images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `
		[
			{
				"id":17,
				"created_at":"2020-11-12T20:07:58.000Z",
				"filename":"sample_2054",
				"file_path":"/uploads/-/system/alert_metric_image/file/17/sample_2054.png",
				"url":"https://example.com/metric",
				"url_text":"An example metric"
			},
			{
				"id":18,
				"created_at":"2020-11-12T20:07:58.000Z",
				"filename":"sample_2054",
				"file_path":"/uploads/-/system/alert_metric_image/file/18/sample_2054.png",
				"url":"https://example.com/metric",
				"url_text":"An example metric"
			}
		]
		`)
	})

	createdAt := time.Date(2020, time.November, 12, 20, 7, 58, 0, time.UTC)
	want := []*MetricImage{
		{
			ID:        17,
			CreatedAt: &createdAt,
			Filename:  "sample_2054",
			FilePath:  "/uploads/-/system/alert_metric_image/file/17/sample_2054.png",
			URL:       "https://example.com/metric",
			URLText:   "An example metric",
		},
		{
			ID:        18,
			CreatedAt: &createdAt,
			Filename:  "sample_2054",
			FilePath:  "/uploads/-/system/alert_metric_image/file/18/sample_2054.png",
			URL:       "https://example.com/metric",
			URLText:   "An example metric",
		},
	}

	metricImages, resp, err := client.AlertManagement.ListMetricImages(1, 2, nil)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, metricImages)
}

func TestAlertManagement_UpdateMetricImage(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/alert_management_alerts/2/metric_images/17", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `
		{
			"id":17,
			"created_at":"2020-11-12T20:07:58.000Z",
			"filename":"sample_2054",
			"file_path":"/uploads/-/system/alert_metric_image/file/17/sample_2054.png",
			"url":"https://example.com/metric",
			"url_text":"An example metric"
		}
		`)
	})

	createdAt := time.Date(2020, time.November, 12, 20, 7, 58, 0, time.UTC)
	want := &MetricImage{
		ID:        17,
		CreatedAt: &createdAt,
		Filename:  "sample_2054",
		FilePath:  "/uploads/-/system/alert_metric_image/file/17/sample_2054.png",
		URL:       "https://example.com/metric",
		URLText:   "An example metric",
	}

	metricImage, resp, err := client.AlertManagement.UpdateMetricImage(1, 2, 17, &UpdateMetricImageOptions{
		URL:     Ptr("https://example.com/metric"),
		URLText: Ptr("An example metric"),
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, want, metricImage)
}

func TestAlertManagement_DeleteMetricImage(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/projects/1/alert_management_alerts/2/metric_images/17", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	resp, err := client.AlertManagement.DeleteMetricImage(1, 2, 17)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
