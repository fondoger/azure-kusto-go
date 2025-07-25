package azkustodata

import (
	"context"
	"github.com/Azure/azure-kusto-go/azkustodata/errors"
	"github.com/Azure/azure-kusto-go/azkustodata/kql"
	"github.com/Azure/azure-kusto-go/azkustodata/value"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestHeaders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                              string
		kcsbApplication, kcsbUser         string
		propApplication, propUser         string
		expectedApplication, expectedUser string
	}{
		{
			name: "TestDefault",
		},
		{
			name:                "TestKcsb",
			kcsbApplication:     "kcsbApplication",
			kcsbUser:            "kcsbUser",
			expectedApplication: "kcsbApplication",
			expectedUser:        "kcsbUser",
		},
		{
			name:                "TestProp",
			propApplication:     "propApplication",
			propUser:            "propUser",
			expectedApplication: "propApplication",
			expectedUser:        "propUser",
		},
		{
			name:                "TestKcsbProp",
			kcsbApplication:     "kcsbApplication",
			kcsbUser:            "kcsbUser",
			propApplication:     "propApplication",
			propUser:            "propUser",
			expectedApplication: "propApplication",
			expectedUser:        "propUser",
		},
	}
	for _, tt := range tests {
		tt := tt // Capture
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			kcsb := NewConnectionStringBuilder("https://test.kusto.windows.net")

			if tt.kcsbApplication != "" {
				kcsb.ApplicationForTracing = tt.kcsbApplication
			}
			if tt.kcsbUser != "" {
				kcsb.UserForTracing = tt.kcsbUser
			}

			queryOptions := make([]QueryOption, 0)
			queryOptions = append(queryOptions, Application(tt.propApplication))
			queryOptions = append(queryOptions, User(tt.propUser))

			opts, err := setQueryOptions(context.Background(), errors.OpQuery, kql.New("test"), queryCall, queryOptions...)
			require.NoError(t, err)

			client, err := New(kcsb)
			require.NoError(t, err)

			headers := client.conn.(*Conn).getHeaders(*opts.requestProperties)

			if tt.expectedApplication != "" {
				assert.Equal(t, tt.expectedApplication, headers.Get("x-ms-app"))
			} else {
				assert.Greater(t, len(headers.Get("x-ms-app")), 0)
			}
			if tt.expectedUser != "" {
				assert.Equal(t, tt.expectedUser, headers.Get("x-ms-user"))
			} else {
				assert.Greater(t, len(headers.Get("x-ms-user")), 0)
			}
			assert.True(t, strings.HasPrefix(headers.Get("x-ms-client-version"), "Kusto.Go.Client:"))
		})
	}
}

func TestSchemas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		hasAuth bool
		err     error
	}{
		{
			name:    "TestNoAuthHttps",
			url:     "https://test.kusto.windows.net",
			hasAuth: false,
		},
		{
			name:    "TestNoAuthHttp",
			url:     "http://test.kusto.windows.net",
			hasAuth: false,
		},
		{
			name:    "TestAuthHttps",
			url:     "https://test.kusto.windows.net",
			hasAuth: true,
		},
		{
			name:    "TestAuthHttp",
			url:     "http://test.kusto.windows.net",
			hasAuth: true,
			err:     errors.ES(errors.OpServConn, errors.KClientArgs, "cannot use token provider with http endpoint, as it would send the token in clear text").SetNoRetry(),
		},
		{
			name: "TestEmptyUrl",
			url:  "",
			err:  errors.ES(errors.OpQuery, errors.KClientArgs, "endpoint cannot be empty"),
		},
	}
	for _, tt := range tests {
		tt := tt // Capture
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var kcsb = &ConnectionStringBuilder{}
			if tt.url != "" {
				kcsb = NewConnectionStringBuilder(tt.url)
			}
			if tt.hasAuth {
				kcsb = kcsb.WithApplicationToken("1", "1")
			}
			client, err := New(kcsb)
			if tt.err != nil {
				assert.Equal(t, tt.err, err)
				return
			}
			assert.NoError(t, err)

			_, err = client.Query(context.Background(), "test", kql.New("test"))
			assert.Contains(t, err.Error(), "no such host")
		})
	}
}

func TestSetConnectorDetails(t *testing.T) {
	t.Parallel()
	tests := []struct {
		testName                          string
		name, version                     string
		sendUser                          bool
		overrideUser, appName, appVersion string
		additionalFields                  []StringPair
		expectedApp, expectedUser         string
		appPrefix                         bool
		expectAnyUser                     bool
	}{
		{
			testName: "TestNameAndVersion",
			name:     "testName", version: "testVersion",
			expectedApp:  "Kusto.testName:{testVersion}|App.",
			appPrefix:    true,
			expectedUser: "[none]",
		},
		{
			testName: "TestNameAndVersionAndUser",
			name:     "testName", version: "testVersion", sendUser: true,
			expectedApp:   "Kusto.testName:{testVersion}|App.",
			appPrefix:     true,
			expectAnyUser: true,
		},
		{
			testName: "TestAll",
			name:     "testName", version: "testVersion", sendUser: true, overrideUser: "testUser", appName: "testApp", appVersion: "testAppVersion", additionalFields: []StringPair{{"testKey", "testValue"}},
			expectedApp:  "Kusto.testName:{testVersion}|App.{testApp}:{testAppVersion}|testKey:{testValue}",
			expectedUser: "testUser",
		},
	}
	for _, tt := range tests {
		tt := tt // Capture
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel()
			kcsb := NewConnectionStringBuilder("https://test.kusto.windows.net")
			kcsb.SetConnectorDetails(tt.name, tt.version, tt.appName, tt.appVersion, tt.sendUser, tt.overrideUser, tt.additionalFields...)

			if tt.appPrefix {
				assert.True(t, strings.HasPrefix(kcsb.ApplicationForTracing, tt.expectedApp))
			} else {
				assert.Equal(t, tt.expectedApp, kcsb.ApplicationForTracing)
			}

			if tt.expectAnyUser {
				assert.Greater(t, len(kcsb.UserForTracing), 0)
			} else {
				assert.Equal(t, tt.expectedUser, kcsb.UserForTracing)
			}
		})
	}
}

func TestTimeout(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(1997, 3, 9, 6, 14, 6, 3, time.UTC)
	nower = func() time.Time {
		return fixedTime
	}

	newContextWithTimeout := func(duration time.Duration) context.Context {
		ctx, _ := context.WithDeadline(context.Background(), fixedTime.Add(duration))
		return ctx
	}

	tests := []struct {
		name                  string
		ctx                   context.Context
		callType              int
		queryOptions          []QueryOption
		expectedServerTimeout time.Duration
	}{
		{
			name:                  "TestDefaultQuery",
			ctx:                   context.Background(),
			callType:              queryCall,
			expectedServerTimeout: defaultQueryTimeout,
		},
		{
			name:                  "TestDefaultMgmt",
			ctx:                   context.Background(),
			callType:              mgmtCall,
			expectedServerTimeout: defaultMgmtTimeout,
		},
		{
			name:                  "OptionSet",
			ctx:                   context.Background(),
			callType:              queryCall,
			queryOptions:          []QueryOption{ServerTimeout(15 * time.Second)},
			expectedServerTimeout: 15 * time.Second,
		},
		{
			name:                  "Context",
			ctx:                   newContextWithTimeout(15 * time.Second),
			callType:              queryCall,
			expectedServerTimeout: 15 * time.Second,
		},
		{
			name:                  "ContextAndOption",
			ctx:                   newContextWithTimeout(15 * time.Second),
			queryOptions:          []QueryOption{ServerTimeout(250 * time.Second)},
			callType:              queryCall,
			expectedServerTimeout: 250 * time.Second,
		},
	}
	for _, tt := range tests {
		tt := tt // Capture
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opts, err := setQueryOptions(tt.ctx, errors.OpUnknown, kql.New("test"), tt.callType, tt.queryOptions...)
			require.NoError(t, err)

			require.Equal(t, value.TimespanString(tt.expectedServerTimeout), opts.requestProperties.Options[ServerTimeoutValue])
		})
	}
}
