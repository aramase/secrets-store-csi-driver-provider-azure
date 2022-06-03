package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		desc   string
		config *Config
	}{
		{
			desc: "using pod identity",
			config: &Config{
				UsePodIdentity: true,
			},
		},
		{
			desc: "using system-assigned identity",
			config: &Config{
				UseVMManagedIdentity: true,
			},
		},
		{
			desc: "using user-assigned identity",
			config: &Config{
				UseVMManagedIdentity:   true,
				UserAssignedIdentityID: "test-id",
			},
		},
		{
			desc: "using workload identity",
			config: &Config{
				WorkloadIdentityClientID: "client-id",
				WorkloadIdentityToken:    "token",
			},
		},
		{
			desc: "using service principal",
			config: &Config{
				ServicePrincipalCreds: map[string]string{"clientid": "testclientid", "clientsecret": "testclientsecret"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if err := tc.config.Validate(); err != nil {
				t.Fatalf("Validate() = %v, want nil", err)
			}
		})
	}
}

func TestValidateError(t *testing.T) {
	cases := []struct {
		desc   string
		config *Config
	}{
		{
			desc: "pod identity and vm managed identity enabled",
			config: &Config{
				UsePodIdentity:       true,
				UseVMManagedIdentity: true,
			},
		},
		{
			desc: "service principal client id is empty",
			config: &Config{
				ServicePrincipalCreds: map[string]string{"clientid": "", "clientsecret": "testclientsecret"},
			},
		},
		{
			desc: "service principal client secret is empty",
			config: &Config{
				ServicePrincipalCreds: map[string]string{"clientid": "testclientid", "clientsecret": ""},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.config.Validate()
			if err == nil {
				t.Fatalf("Validate() = nil, want error")
			}
		})
	}
}

func TestGetAuthorizerForServicePrincipal(t *testing.T) {
	env := &azure.PublicCloud
	authorizer, err := getAuthorizerForServicePrincipal("AADClientID", "AADClientSecret", env.KeyVaultEndpoint, env.ActiveDirectoryEndpoint, "tenantID")
	assert.NoError(t, err)

	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, "tenantID")
	assert.NoError(t, err)

	spt, err := adal.NewServicePrincipalToken(*oauthConfig, "AADClientID", "AADClientSecret", env.KeyVaultEndpoint)
	assert.NoError(t, err)

	assert.Equal(t, authorizer, autorest.NewBearerAuthorizer(spt))
}

func TestGetAuthorizerForPodIdentity(t *testing.T) {
	env := &azure.PublicCloud

	cases := []struct {
		desc        string
		tokenResp   NMIResponse
		podName     string
		expectedErr error
	}{
		{
			desc:        "pod name is empty",
			tokenResp:   NMIResponse{},
			podName:     "",
			expectedErr: fmt.Errorf("pod information is not available. deploy a CSIDriver object to set podInfoOnMount: true"),
		},
		{
			desc:        "token response is empty",
			tokenResp:   NMIResponse{},
			podName:     "pod",
			expectedErr: fmt.Errorf("nmi did not return expected values in response: token and clientid"),
		},
		{
			desc: "valid token response",
			tokenResp: NMIResponse{
				Token: adal.Token{
					AccessToken: "accessToken",
					ExpiresIn:   "0",
					ExpiresOn:   "0",
					NotBefore:   "0",
				},
				ClientID: "clientID",
			},
			podName:     "pod",
			expectedErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			// mock NMI server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.String(), "/host/token/")
				tr, err := json.Marshal(tc.tokenResp)
				assert.NoError(t, err)

				w.Write(tr)
			}))
			defer ts.Close()

			splitURL := strings.Split(ts.URL, ":")
			mockNMIPort := splitURL[len(splitURL)-1]

			authorizer, err := getAuthorizerForPodIdentity(tc.podName, "default", env.KeyVaultEndpoint, env.ActiveDirectoryEndpoint, "tenantID", mockNMIPort)
			assert.Equal(t, tc.expectedErr, err)

			if tc.expectedErr == nil {
				oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, "tenantID")
				assert.NoError(t, err)

				spt, err := adal.NewServicePrincipalTokenFromManualToken(*oauthConfig, "clientID", env.KeyVaultEndpoint, tc.tokenResp.Token, nil)
				assert.NoError(t, err)
				assert.Equal(t, authorizer, autorest.NewBearerAuthorizer(spt))
			}
		})
	}
}

// Vendored from https://github.com/Azure/go-autorest/blob/def88ef859fb980eff240c755a70597bc9b490d0/autorest/adal/token_test.go
func TestParseExpiresOn(t *testing.T) {
	// get current time, round to nearest second, and add one hour
	n := time.Now().UTC().Round(time.Second).Add(time.Hour)
	amPM := "AM"
	if n.Hour() >= 12 {
		amPM = "PM"
	}
	testcases := []struct {
		Name   string
		String string
		Value  int64
	}{
		{
			Name:   "integer",
			String: "3600",
			Value:  3600,
		},
		{
			Name:   "timestamp with AM/PM",
			String: fmt.Sprintf("%d/%d/%d %d:%02d:%02d %s +00:00", n.Month(), n.Day(), n.Year(), n.Hour(), n.Minute(), n.Second(), amPM),
			Value:  3600,
		},
		{
			Name:   "timestamp without AM/PM",
			String: fmt.Sprintf("%d/%d/%d %d:%02d:%02d +00:00", n.Month(), n.Day(), n.Year(), n.Hour(), n.Minute(), n.Second()),
			Value:  3600,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(subT *testing.T) {
			jn, err := parseExpiresOn(tc.String)
			if err != nil {
				subT.Error(err)
			}
			i, err := jn.Int64()
			if err != nil {
				subT.Error(err)
			}
			if i != tc.Value {
				subT.Logf("expected %d, got %d", tc.Value, i)
				subT.Fail()
			}
		})
	}
}

func TestParseServiceAccountTokenError(t *testing.T) {
	cases := []struct {
		desc     string
		saTokens string
	}{
		{
			desc:     "empty serviceaccount tokens",
			saTokens: "",
		},
		{
			desc:     "invalid serviceaccount tokens",
			saTokens: "invalid",
		},
		{
			desc:     "token for audience not found",
			saTokens: `{"aud1":{"token":"eyJhbGciOiJSUzI1NiIsImtpZCI6InRhVDBxbzhQVEZ1ajB1S3BYUUxIclRsR01XakxjemJNOTlzWVMxSlNwbWcifQ.eyJhdWQiOlsiYXBpOi8vQXp1cmVBRGlUb2tlbkV4Y2hhbmdlIl0sImV4cCI6MTY0MzIzNDY0NywiaWF0IjoxNjQzMjMxMDQ3LCJpc3MiOiJodHRwczovL2t1YmVybmV0ZXMuZGVmYXVsdC5zdmMuY2x1c3Rlci5sb2NhbCIsImt1YmVybmV0ZXMuaW8iOnsibmFtZXNwYWNlIjoidGVzdC12MWFscGhhMSIsInBvZCI6eyJuYW1lIjoic2VjcmV0cy1zdG9yZS1pbmxpbmUtY3JkIiwidWlkIjoiYjBlYmZjMzUtZjEyNC00ZTEyLWI3N2UtYjM0MjM2N2IyMDNmIn0sInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJkZWZhdWx0IiwidWlkIjoiMjViNGY1NzgtM2U4MC00NTczLWJlOGQtZTdmNDA5ZDI0MmI2In19LCJuYmYiOjE2NDMyMzEwNDcsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDp0ZXN0LXYxYWxwaGExOmRlZmF1bHQifQ.ALE46aKmtTV7dsuFOwDZqvEjdHFUTNP-JVjMxexTemmPA78fmPTUZF0P6zANumA03fjX3L-MZNR3PxmEZgKA9qEGIDsljLsUWsVBEquowuBh8yoBYkGkMJmRfmbfS3y7_4Q7AU3D9Drw4iAHcn1GwedjOQC0i589y3dkNNqf8saqHfXkbSSLtSE0f2uzI-PjuTKvR1kuojEVNKlEcA4wsKfoiRpkua17sHkHU0q9zxCMDCr_1f8xbigRnRx0wscU3vy-8KhF3zQtpcWkk3r4C5YSXut9F3xjz5J9DUQn2vNMfZg4tOdcR-9Xv9fbY5iujiSlS58GEktSEa3SE9wrCw\",\"expirationTimestamp\":\"2022-01-26T22:04:07Z\"},\"gcp\":{\"token\":\"eyJhbGciOiJSUzI1NiIsImtpZCI6InRhVDBxbzhQVEZ1ajB1S3BYUUxIclRsR01XakxjemJNOTlzWVMxSlNwbWcifQ.eyJhdWQiOlsiZ2NwIl0sImV4cCI6MTY0MzIzNDY0NywiaWF0IjoxNjQzMjMxMDQ3LCJpc3MiOiJodHRwczovL2t1YmVybmV0ZXMuZGVmYXVsdC5zdmMuY2x1c3Rlci5sb2NhbCIsImt1YmVybmV0ZXMuaW8iOnsibmFtZXNwYWNlIjoidGVzdC12MWFscGhhMSIsInBvZCI6eyJuYW1lIjoic2VjcmV0cy1zdG9yZS1pbmxpbmUtY3JkIiwidWlkIjoiYjBlYmZjMzUtZjEyNC00ZTEyLWI3N2UtYjM0MjM2N2IyMDNmIn0sInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJkZWZhdWx0IiwidWlkIjoiMjViNGY1NzgtM2U4MC00NTczLWJlOGQtZTdmNDA5ZDI0MmI2In19LCJuYmYiOjE2NDMyMzEwNDcsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDp0ZXN0LXYxYWxwaGExOmRlZmF1bHQifQ.BT0YGI7bGdSNaIBqIEnVL0Ky5t-fynaemSGxjGdKOPl0E22UIVGDpAMUhaS19i20c-Dqs-Kn0N-R5QyDNpZg8vOL5KIFqu2kSYNbKxtQW7TPYIsV0d9wUZjLSr54DKrmyXNMGRoT2bwcF4yyfmO46eMmZSaXN8Y4lgapeabg6CBVVQYHD-GrgXf9jVLeJfCQkTuojK1iXOphyD6NqlGtVCaY1jWxbBMibN0q214vKvQboub8YMuvclGdzn_l_ZQSTjvhBj9I-W1t-JArVjqHoIb8_FlR9BSgzgL7V3Jki55vmiOdEYqMErJWrIZPP3s8qkU5hhO9rSVEd3LJHponvQ","expirationTimestamp":"2022-01-26T22:04:07Z"}}`, //nolint
		},
		{
			desc:     "token incorrect format",
			saTokens: `{"api://AzureADTokenExchange":{"tokens":"eyJhbGciOiJSUzI1NiIsImtpZCI6InRhVDBxbzhQVEZ1ajB1S3BYUUxIclRsR01XakxjemJNOTlzWVMxSlNwbWcifQ.eyJhdWQiOlsiYXBpOi8vQXp1cmVBRGlUb2tlbkV4Y2hhbmdlIl0sImV4cCI6MTY0MzIzNDY0NywiaWF0IjoxNjQzMjMxMDQ3LCJpc3MiOiJodHRwczovL2t1YmVybmV0ZXMuZGVmYXVsdC5zdmMuY2x1c3Rlci5sb2NhbCIsImt1YmVybmV0ZXMuaW8iOnsibmFtZXNwYWNlIjoidGVzdC12MWFscGhhMSIsInBvZCI6eyJuYW1lIjoic2VjcmV0cy1zdG9yZS1pbmxpbmUtY3JkIiwidWlkIjoiYjBlYmZjMzUtZjEyNC00ZTEyLWI3N2UtYjM0MjM2N2IyMDNmIn0sInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJkZWZhdWx0IiwidWlkIjoiMjViNGY1NzgtM2U4MC00NTczLWJlOGQtZTdmNDA5ZDI0MmI2In19LCJuYmYiOjE2NDMyMzEwNDcsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDp0ZXN0LXYxYWxwaGExOmRlZmF1bHQifQ.ALE46aKmtTV7dsuFOwDZqvEjdHFUTNP-JVjMxexTemmPA78fmPTUZF0P6zANumA03fjX3L-MZNR3PxmEZgKA9qEGIDsljLsUWsVBEquowuBh8yoBYkGkMJmRfmbfS3y7_4Q7AU3D9Drw4iAHcn1GwedjOQC0i589y3dkNNqf8saqHfXkbSSLtSE0f2uzI-PjuTKvR1kuojEVNKlEcA4wsKfoiRpkua17sHkHU0q9zxCMDCr_1f8xbigRnRx0wscU3vy-8KhF3zQtpcWkk3r4C5YSXut9F3xjz5J9DUQn2vNMfZg4tOdcR-9Xv9fbY5iujiSlS58GEktSEa3SE9wrCw\",\"expirationTimestamp\":\"2022-01-26T22:04:07Z\"},\"gcp\":{\"token\":\"eyJhbGciOiJSUzI1NiIsImtpZCI6InRhVDBxbzhQVEZ1ajB1S3BYUUxIclRsR01XakxjemJNOTlzWVMxSlNwbWcifQ.eyJhdWQiOlsiZ2NwIl0sImV4cCI6MTY0MzIzNDY0NywiaWF0IjoxNjQzMjMxMDQ3LCJpc3MiOiJodHRwczovL2t1YmVybmV0ZXMuZGVmYXVsdC5zdmMuY2x1c3Rlci5sb2NhbCIsImt1YmVybmV0ZXMuaW8iOnsibmFtZXNwYWNlIjoidGVzdC12MWFscGhhMSIsInBvZCI6eyJuYW1lIjoic2VjcmV0cy1zdG9yZS1pbmxpbmUtY3JkIiwidWlkIjoiYjBlYmZjMzUtZjEyNC00ZTEyLWI3N2UtYjM0MjM2N2IyMDNmIn0sInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJkZWZhdWx0IiwidWlkIjoiMjViNGY1NzgtM2U4MC00NTczLWJlOGQtZTdmNDA5ZDI0MmI2In19LCJuYmYiOjE2NDMyMzEwNDcsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDp0ZXN0LXYxYWxwaGExOmRlZmF1bHQifQ.BT0YGI7bGdSNaIBqIEnVL0Ky5t-fynaemSGxjGdKOPl0E22UIVGDpAMUhaS19i20c-Dqs-Kn0N-R5QyDNpZg8vOL5KIFqu2kSYNbKxtQW7TPYIsV0d9wUZjLSr54DKrmyXNMGRoT2bwcF4yyfmO46eMmZSaXN8Y4lgapeabg6CBVVQYHD-GrgXf9jVLeJfCQkTuojK1iXOphyD6NqlGtVCaY1jWxbBMibN0q214vKvQboub8YMuvclGdzn_l_ZQSTjvhBj9I-W1t-JArVjqHoIb8_FlR9BSgzgL7V3Jki55vmiOdEYqMErJWrIZPP3s8qkU5hhO9rSVEd3LJHponvQ","expirationTimestamp":"2022-01-26T22:04:07Z"}}`, //nolint

		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if _, err := ParseServiceAccountToken(tc.saTokens); err == nil {
				t.Errorf("ParseServiceAccountToken(%s) = nil, want error", tc.saTokens)
			}
		})
	}
}

func TestParseServiceAccountToken(t *testing.T) {
	saTokens := `{"api://AzureADTokenExchange":{"token":"eyJhbGciOiJSUzI1NiIsImtpZCI6InRhVDBxbzhQVEZ1ajB1S3BYUUxIclRsR01XakxjemJNOTlzWVMxSlNwbWcifQ.eyJhdWQiOlsiYXBpOi8vQXp1cmVBRGlUb2tlbkV4Y2hhbmdlIl0sImV4cCI6MTY0MzIzNDY0NywiaWF0IjoxNjQzMjMxMDQ3LCJpc3MiOiJodHRwczovL2t1YmVybmV0ZXMuZGVmYXVsdC5zdmMuY2x1c3Rlci5sb2NhbCIsImt1YmVybmV0ZXMuaW8iOnsibmFtZXNwYWNlIjoidGVzdC12MWFscGhhMSIsInBvZCI6eyJuYW1lIjoic2VjcmV0cy1zdG9yZS1pbmxpbmUtY3JkIiwidWlkIjoiYjBlYmZjMzUtZjEyNC00ZTEyLWI3N2UtYjM0MjM2N2IyMDNmIn0sInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJkZWZhdWx0IiwidWlkIjoiMjViNGY1NzgtM2U4MC00NTczLWJlOGQtZTdmNDA5ZDI0MmI2In19LCJuYmYiOjE2NDMyMzEwNDcsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDp0ZXN0LXYxYWxwaGExOmRlZmF1bHQifQ.ALE46aKmtTV7dsuFOwDZqvEjdHFUTNP-JVjMxexTemmPA78fmPTUZF0P6zANumA03fjX3L-MZNR3PxmEZgKA9qEGIDsljLsUWsVBEquowuBh8yoBYkGkMJmRfmbfS3y7_4Q7AU3D9Drw4iAHcn1GwedjOQC0i589y3dkNNqf8saqHfXkbSSLtSE0f2uzI-PjuTKvR1kuojEVNKlEcA4wsKfoiRpkua17sHkHU0q9zxCMDCr_1f8xbigRnRx0wscU3vy-8KhF3zQtpcWkk3r4C5YSXut9F3xjz5J9DUQn2vNMfZg4tOdcR-9Xv9fbY5iujiSlS58GEktSEa3SE9wrCw","expirationTimestamp":"2022-01-26T22:04:07Z"},"aud2":{"token":"eyJhbGciOiJSUzI1NiIsImtpZCI6InRhVDBxbzhQVEZ1ajB1S3BYUUxIclRsR01XakxjemJNOTlzWVMxSlNwbWcifQ.eyJhdWQiOlsiZ2NwIl0sImV4cCI6MTY0MzIzNDY0NywiaWF0IjoxNjQzMjMxMDQ3LCJpc3MiOiJodHRwczovL2t1YmVybmV0ZXMuZGVmYXVsdC5zdmMuY2x1c3Rlci5sb2NhbCIsImt1YmVybmV0ZXMuaW8iOnsibmFtZXNwYWNlIjoidGVzdC12MWFscGhhMSIsInBvZCI6eyJuYW1lIjoic2VjcmV0cy1zdG9yZS1pbmxpbmUtY3JkIiwidWlkIjoiYjBlYmZjMzUtZjEyNC00ZTEyLWI3N2UtYjM0MjM2N2IyMDNmIn0sInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJkZWZhdWx0IiwidWlkIjoiMjViNGY1NzgtM2U4MC00NTczLWJlOGQtZTdmNDA5ZDI0MmI2In19LCJuYmYiOjE2NDMyMzEwNDcsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDp0ZXN0LXYxYWxwaGExOmRlZmF1bHQifQ.BT0YGI7bGdSNaIBqIEnVL0Ky5t-fynaemSGxjGdKOPl0E22UIVGDpAMUhaS19i20c-Dqs-Kn0N-R5QyDNpZg8vOL5KIFqu2kSYNbKxtQW7TPYIsV0d9wUZjLSr54DKrmyXNMGRoT2bwcF4yyfmO46eMmZSaXN8Y4lgapeabg6CBVVQYHD-GrgXf9jVLeJfCQkTuojK1iXOphyD6NqlGtVCaY1jWxbBMibN0q214vKvQboub8YMuvclGdzn_l_ZQSTjvhBj9I-W1t-JArVjqHoIb8_FlR9BSgzgL7V3Jki55vmiOdEYqMErJWrIZPP3s8qkU5hhO9rSVEd3LJHponvQ","expirationTimestamp":"2022-01-26T22:04:07Z"}}` //nolint
	expectedToken := `eyJhbGciOiJSUzI1NiIsImtpZCI6InRhVDBxbzhQVEZ1ajB1S3BYUUxIclRsR01XakxjemJNOTlzWVMxSlNwbWcifQ.eyJhdWQiOlsiYXBpOi8vQXp1cmVBRGlUb2tlbkV4Y2hhbmdlIl0sImV4cCI6MTY0MzIzNDY0NywiaWF0IjoxNjQzMjMxMDQ3LCJpc3MiOiJodHRwczovL2t1YmVybmV0ZXMuZGVmYXVsdC5zdmMuY2x1c3Rlci5sb2NhbCIsImt1YmVybmV0ZXMuaW8iOnsibmFtZXNwYWNlIjoidGVzdC12MWFscGhhMSIsInBvZCI6eyJuYW1lIjoic2VjcmV0cy1zdG9yZS1pbmxpbmUtY3JkIiwidWlkIjoiYjBlYmZjMzUtZjEyNC00ZTEyLWI3N2UtYjM0MjM2N2IyMDNmIn0sInNlcnZpY2VhY2NvdW50Ijp7Im5hbWUiOiJkZWZhdWx0IiwidWlkIjoiMjViNGY1NzgtM2U4MC00NTczLWJlOGQtZTdmNDA5ZDI0MmI2In19LCJuYmYiOjE2NDMyMzEwNDcsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDp0ZXN0LXYxYWxwaGExOmRlZmF1bHQifQ.ALE46aKmtTV7dsuFOwDZqvEjdHFUTNP-JVjMxexTemmPA78fmPTUZF0P6zANumA03fjX3L-MZNR3PxmEZgKA9qEGIDsljLsUWsVBEquowuBh8yoBYkGkMJmRfmbfS3y7_4Q7AU3D9Drw4iAHcn1GwedjOQC0i589y3dkNNqf8saqHfXkbSSLtSE0f2uzI-PjuTKvR1kuojEVNKlEcA4wsKfoiRpkua17sHkHU0q9zxCMDCr_1f8xbigRnRx0wscU3vy-8KhF3zQtpcWkk3r4C5YSXut9F3xjz5J9DUQn2vNMfZg4tOdcR-9Xv9fbY5iujiSlS58GEktSEa3SE9wrCw`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         //nolint

	token, err := ParseServiceAccountToken(saTokens)
	if err != nil {
		t.Fatalf("ParseServiceAccountToken(%s) = %v, want nil", saTokens, err)
	}
	if token != expectedToken {
		t.Errorf("ParseServiceAccountToken(%s) = %s, want %s", saTokens, token, expectedToken)
	}
}

func TestGetServicePrincipalClientID(t *testing.T) {
	cases := []struct {
		desc     string
		secrets  map[string]string
		expected string
	}{
		{
			desc:     "empty",
			secrets:  map[string]string{},
			expected: "",
		},
		{
			desc: "not empty",
			secrets: map[string]string{
				"clientid": "client-id",
			},
			expected: "client-id",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := GetServicePrincipalClientID(tc.secrets)
			if actual != tc.expected {
				t.Errorf("GetServicePrincipalClientID(%v) = %v, want %v", tc.secrets, actual, tc.expected)
			}
		})
	}
}

func TestGetServicePrincipalClientSecret(t *testing.T) {
	cases := []struct {
		desc     string
		secrets  map[string]string
		expected string
	}{
		{
			desc:     "empty",
			secrets:  map[string]string{},
			expected: "",
		},
		{
			desc: "not empty",
			secrets: map[string]string{
				"clientsecret": "client-secret",
			},
			expected: "client-secret",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := GetServicePrincipalClientSecret(tc.secrets)
			if actual != tc.expected {
				t.Errorf("GetServicePrincipalClientSecret(%v) = %v, want %v", tc.secrets, actual, tc.expected)
			}
		})
	}
}
