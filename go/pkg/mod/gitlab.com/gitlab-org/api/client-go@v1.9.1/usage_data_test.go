package gitlab

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestUsageDataService_GetServicePing(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/usage_data/service_ping", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"recorded_at": "2024-01-15T23:33:50.387Z",
			"license": {},
			"counts": {
				"assignee_lists": 0,
				"ci_builds": 463,
				"ci_external_pipelines": 0
			}
		}`)
	})

	sp, _, err := client.UsageData.GetServicePing()
	require.NoError(t, err)

	want := &ServicePingData{
		RecordedAt: Ptr(time.Date(2024, time.January, 15, 23, 33, 50, 387000000, time.UTC)),
		License:    map[string]string{},
		Counts: map[string]int64{
			"assignee_lists":        int64(0),
			"ci_builds":             int64(463),
			"ci_external_pipelines": int64(0),
		},
	}
	require.Equal(t, want, sp)
}

func TestUsageDataService_GetMetricDefinitionsAsYAML(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	expectedYAML := `---
- key_path: redis_hll_counters.search.i_search_paid_monthly
  description: Calculated unique users to perform a search with a paid license enabled
    by month
  product_group: global_search
  value_type: number
  status: active
  time_frame: 28d
  data_source: redis_hll
  tier:
  - premium
  - ultimate
`

	mux.HandleFunc("/api/v4/usage_data/metric_definitions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "text/yaml")
		fmt.Fprint(w, expectedYAML)
	})

	yaml, _, err := client.UsageData.GetMetricDefinitionsAsYAML()
	require.NoError(t, err)

	var want bytes.Buffer
	want.Write([]byte(expectedYAML))

	require.Equal(t, &want, yaml)
}

func TestUsageDataService_GetQueries(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/usage_data/queries", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"recorded_at": "2021-03-23T06:31:21.267Z",
			"uuid": "123",
			"hostname": "localhost",
			"version": "13.11.0-pre",
			"installation_type": "gitlab-development-kit",
			"active_user_count": "SELECT COUNT(\"users\".\"id\") FROM \"users\" WHERE (\"users\".\"state\" IN ('active')) AND (\"users\".\"user_type\" IS NULL OR \"users\".\"user_type\" IN (NULL, 6, 4))",
			"edition": "EE",
			"license_md5": "c701acc03844c45366dd175ef7a4e19c",
			"license_sha256": "366dd175ef7a4e19cc701acc03844c45366dd175ef7a4e19cc701acc03844c45",
			"license_id": "abc",
			"historical_max_users": 0,
			"licensee": {
				"Name": "John Doe1"
			},
			"license_user_count": 5,
			"license_starts_at": "1970-01-01",
			"license_expires_at": "2022-02-23",
			"license_plan": "starter",
			"license_subscription_id": "0000",
			"license_add_ons": {
				"GitLab_FileLocks": 1,
				"GitLab_Auditor_User": 1
			},
			"license_trial": "license",
			"license": {},
			"settings": {
				"ldap_encrypted_secrets_enabled": "false",
				"operating_system": "mac_os_x-11.2.2"
			},
			"counts": {
				"assignee_lists": "SELECT COUNT(\"lists\".\"id\") FROM \"lists\" WHERE \"lists\".\"list_type\" = 3",
				"boards": "SELECT COUNT(\"boards\".\"id\") FROM \"boards\""
			}
		}`)
	})

	sq, _, err := client.UsageData.GetQueries()
	require.NoError(t, err)

	want := &ServicePingQueries{
		RecordedAt:         Ptr(time.Date(2021, time.March, 23, 6, 31, 21, 267000000, time.UTC)),
		UUID:               "123",
		Hostname:           "localhost",
		Version:            "13.11.0-pre",
		InstallationType:   "gitlab-development-kit",
		ActiveUserCount:    "SELECT COUNT(\"users\".\"id\") FROM \"users\" WHERE (\"users\".\"state\" IN ('active')) AND (\"users\".\"user_type\" IS NULL OR \"users\".\"user_type\" IN (NULL, 6, 4))",
		Edition:            "EE",
		LicenseMD5:         "c701acc03844c45366dd175ef7a4e19c",
		LicenseSHA256:      "366dd175ef7a4e19cc701acc03844c45366dd175ef7a4e19cc701acc03844c45",
		LicenseID:          "abc",
		HistoricalMaxUsers: 0,
		Licensee: map[string]string{
			"Name": "John Doe1",
		},
		LicenseUserCount:      5,
		LicenseStartsAt:       "1970-01-01",
		LicenseExpiresAt:      "2022-02-23",
		LicensePlan:           "starter",
		LicenseSubscriptionID: "0000",
		LicenseAddOns: map[string]int64{
			"GitLab_FileLocks":    1,
			"GitLab_Auditor_User": 1,
		},
		LicenseTrial: "license",
		License:      map[string]string{},
		Settings: map[string]string{
			"ldap_encrypted_secrets_enabled": "false",
			"operating_system":               "mac_os_x-11.2.2",
		},
		Counts: map[string]string{
			"assignee_lists": "SELECT COUNT(\"lists\".\"id\") FROM \"lists\" WHERE \"lists\".\"list_type\" = 3",
			"boards":         "SELECT COUNT(\"boards\".\"id\") FROM \"boards\"",
		},
	}
	require.Equal(t, want, sq)
}

func TestUsageDataService_GetServicePingNonSqlMetrics(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/usage_data/non_sql_metrics", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"recorded_at": "2021-03-26T07:04:03.724Z",
			"uuid": null,
			"hostname": "localhost",
			"version": "13.11.0-pre",
			"installation_type": "gitlab-development-kit",
			"active_user_count": 3,
			"edition": "EE",
			"license_md5": "bb8cd0d8a6d9569ff3f70b8927a1f949",
			"license_sha256": "366dd175ef7a4e19cc701acc03844c45366dd175ef7a4e19cc701acc03844c45",
			"license_id": null,
			"historical_max_users": 0,
			"licensee": {
				"Name": "John Doe1"
			},
			"license_user_count": 0,
			"license_starts_at": "1970-01-01",
			"license_expires_at": "2022-02-26",
			"license_plan": "starter",
			"license_add_ons": {
				"GitLab_FileLocks": 1,
				"GitLab_Auditor_User": 1
			},
			"license_trial": null,
			"license_subscription_id": "0000",
			"license": {},
			"settings": {
				"ldap_encrypted_secrets_enabled": "false",
				"operating_system": "mac_os_x-11.2.2"
			}
		}`)
	})

	nsm, _, err := client.UsageData.GetNonSQLMetrics()
	require.NoError(t, err)

	want := &ServicePingNonSQLMetrics{
		RecordedAt:         "2021-03-26T07:04:03.724Z",
		UUID:               "",
		Hostname:           "localhost",
		Version:            "13.11.0-pre",
		InstallationType:   "gitlab-development-kit",
		ActiveUserCount:    3,
		Edition:            "EE",
		LicenseMD5:         "bb8cd0d8a6d9569ff3f70b8927a1f949",
		LicenseSHA256:      "366dd175ef7a4e19cc701acc03844c45366dd175ef7a4e19cc701acc03844c45",
		LicenseID:          "",
		HistoricalMaxUsers: 0,
		Licensee: map[string]string{
			"Name": "John Doe1",
		},
		LicenseUserCount: 0,
		LicenseStartsAt:  "1970-01-01",
		LicenseExpiresAt: "2022-02-26",
		LicensePlan:      "starter",
		LicenseAddOns: map[string]int64{
			"GitLab_FileLocks":    1,
			"GitLab_Auditor_User": 1,
		},
		LicenseTrial:          "",
		LicenseSubscriptionID: "0000",
		License:               map[string]string{},
		Settings: map[string]string{
			"ldap_encrypted_secrets_enabled": "false",
			"operating_system":               "mac_os_x-11.2.2",
		},
	}
	require.Equal(t, want, nsm)
}

func TestUsageDataService_TrackEvent(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/usage_data/track_event", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testJSONBody(t, r,
			`{"event":"mr_name_changed","send_to_snowplow":true,"namespace_id":1,"project_id":1,"additional_properties":{"lang":"eng"}}`)
		w.WriteHeader(http.StatusNoContent)
	})

	sendToSnowplow := true
	namespaceID := int64(1)
	projectID := int64(1)
	additionalProperties := map[string]string{
		"lang": "eng",
	}

	opt := &TrackEventOptions{
		Event:                "mr_name_changed",
		SendToSnowplow:       &sendToSnowplow,
		NamespaceID:          &namespaceID,
		ProjectID:            &projectID,
		AdditionalProperties: additionalProperties,
	}

	resp, err := client.UsageData.TrackEvent(opt)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestUsageDataService_TrackEvents(t *testing.T) {
	t.Parallel()
	mux, client := setup(t)

	mux.HandleFunc("/api/v4/usage_data/track_events", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testJSONBody(t, r, `{"events":[{"event":"mr_name_changed","namespace_id":1,"project_id":1,"additional_properties":{"lang":"eng"}},{"event":"mr_name_changed","namespace_id":2,"project_id":2,"additional_properties":{"lang":"eng"}}]}`)
		w.WriteHeader(http.StatusNoContent)
	})

	namespaceID1 := int64(1)
	projectID1 := int64(1)
	namespaceID2 := int64(2)
	projectID2 := int64(2)

	additionalProperties := map[string]string{
		"lang": "eng",
	}

	event1 := TrackEventOptions{
		Event:                "mr_name_changed",
		NamespaceID:          &namespaceID1,
		ProjectID:            &projectID1,
		AdditionalProperties: additionalProperties,
	}

	event2 := TrackEventOptions{
		Event:                "mr_name_changed",
		NamespaceID:          &namespaceID2,
		ProjectID:            &projectID2,
		AdditionalProperties: additionalProperties,
	}

	opt := &TrackEventsOptions{
		Events: []TrackEventOptions{event1, event2},
	}

	resp, err := client.UsageData.TrackEvents(opt)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func testJSONBody(t *testing.T, r *http.Request, want string) {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)

	require.JSONEq(t, want, string(body))
}
