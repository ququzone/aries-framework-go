/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hyperledger/aries-framework-go/pkg/doc/did"
	"github.com/hyperledger/aries-framework-go/pkg/doc/ldcontext"
	"github.com/hyperledger/aries-framework-go/pkg/internal/ldtestutil"
	"github.com/hyperledger/aries-framework-go/pkg/vdr"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/httpbinding"
	"github.com/hyperledger/aries-framework-go/pkg/vdr/key"
)

const (
	testDID    = "did:key:z6MkoTHsgNNrby8JzCNQ1iRLyW5QQ6R8Xuu6AA8igGrMVPUM"
	testDomain = "https://identity.foundation"

	contextV1 = "https://identity.foundation/.well-known/did-configuration/v1"
)

func TestNew(t *testing.T) {
	t.Run("success - default options", func(t *testing.T) {
		c := New()
		require.NotNil(t, c)
		require.Len(t, c.didConfigOpts, 0)
	})

	t.Run("success - did config options provided", func(t *testing.T) {
		loader, err := ldtestutil.DocumentLoader(ldcontext.Document{
			URL:     contextV1,
			Content: json.RawMessage(didCfgCtxV1),
		})
		require.NoError(t, err)

		c := New(WithJSONLDDocumentLoader(loader),
			WithVDRegistry(vdr.New(vdr.WithVDR(key.New()))),
			WithHTTPClient(&http.Client{}))
		require.NotNil(t, c)
		require.Len(t, c.didConfigOpts, 2)
	})
}

func TestVerifyDIDAndDomain(t *testing.T) {
	loader, err := ldtestutil.DocumentLoader(ldcontext.Document{
		URL:     contextV1,
		Content: json.RawMessage(didCfgCtxV1),
	})
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		httpClient := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(didCfg))),
				}, nil
			},
		}

		c := New(WithJSONLDDocumentLoader(loader), WithHTTPClient(httpClient))

		err := c.VerifyDIDAndDomain(testDID, testDomain)
		require.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		httpClient := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(didCfg))),
				}, nil
			},
		}

		c := New(WithJSONLDDocumentLoader(loader), WithHTTPClient(httpClient))

		err := c.VerifyDIDAndDomain(testDID, testDomain)
		require.NoError(t, err)
	})

	t.Run("error - http client error", func(t *testing.T) {
		c := New(WithJSONLDDocumentLoader(loader))

		err := c.VerifyDIDAndDomain(testDID, "https://non-existent-abc.com")
		require.Error(t, err)
		require.Contains(t, err.Error(),
			"Get \"https://non-existent-abc.com/.well-known/did-configuration.json\": dial tcp: "+
				"lookup non-existent-abc.com: no such host")
	})

	t.Run("error - http request error", func(t *testing.T) {
		c := New(WithJSONLDDocumentLoader(loader))

		err := c.VerifyDIDAndDomain(testDID, ":invalid.com")
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing protocol scheme")
	})

	t.Run("error - http status error", func(t *testing.T) {
		httpClient := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       io.NopCloser(bytes.NewReader([]byte("data not found"))),
				}, nil
			},
		}

		c := New(WithJSONLDDocumentLoader(loader), WithHTTPClient(httpClient))

		err := c.VerifyDIDAndDomain(testDID, testDomain)
		require.Error(t, err)
		require.Contains(t, err.Error(), "endpoint https://identity.foundation/.well-known/did-configuration.json "+
			"returned status '404' and message 'data not found'")
	})

	t.Run("error - did configuration missing linked DIDs", func(t *testing.T) {
		httpClient := &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(didCfgNoLinkedDIDs))),
				}, nil
			},
		}

		c := New(WithJSONLDDocumentLoader(loader),
			WithVDRegistry(vdr.New(vdr.WithVDR(key.New()))),
			WithHTTPClient(httpClient))

		err := c.VerifyDIDAndDomain(testDID, testDomain)
		require.Error(t, err)
		require.Contains(t, err.Error(), "did configuration: property 'linked_dids' is required ")
	})
}

func TestCloseResponseBody(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		closeResponseBody(&mockCloser{Err: fmt.Errorf("test error")})
	})
}

type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

type mockCloser struct {
	Err error
}

func (c *mockCloser) Close() error {
	return c.Err
}

// nolint: lll
const didCfg = `
{
  "@context": "https://identity.foundation/.well-known/did-configuration/v1",
  "linked_dids": [
    {
      "@context": [
        "https://www.w3.org/2018/credentials/v1",
        "https://identity.foundation/.well-known/did-configuration/v1"
      ],
      "issuer": "did:key:z6MkoTHsgNNrby8JzCNQ1iRLyW5QQ6R8Xuu6AA8igGrMVPUM",
      "issuanceDate": "2020-12-04T14:08:28-06:00",
      "expirationDate": "2025-12-04T14:08:28-06:00",
      "type": [
        "VerifiableCredential",
        "DomainLinkageCredential"
      ],
      "credentialSubject": {
        "id": "did:key:z6MkoTHsgNNrby8JzCNQ1iRLyW5QQ6R8Xuu6AA8igGrMVPUM",
        "origin": "https://identity.foundation"
      },
      "proof": {
        "type": "Ed25519Signature2018",
        "created": "2020-12-04T20:08:28.540Z",
        "jws": "eyJhbGciOiJFZERTQSIsImI2NCI6ZmFsc2UsImNyaXQiOlsiYjY0Il19..D0eDhglCMEjxDV9f_SNxsuU-r3ZB9GR4vaM9TYbyV7yzs1WfdUyYO8rFZdedHbwQafYy8YOpJ1iJlkSmB4JaDQ",
        "proofPurpose": "assertionMethod",
        "verificationMethod": "did:key:z6MkoTHsgNNrby8JzCNQ1iRLyW5QQ6R8Xuu6AA8igGrMVPUM#z6MkoTHsgNNrby8JzCNQ1iRLyW5QQ6R8Xuu6AA8igGrMVPUM"
      }
    }
  ]
}`

const didCfgNoLinkedDIDs = `
{
  "@context": "https://identity.foundation/.well-known/did-configuration/v1"
}`

// nolint: lll
const didCfgCtxV1 = `
{
  "@context": [
    {
      "@version": 1.1,
      "@protected": true,
      "LinkedDomains": "https://identity.foundation/.well-known/resources/did-configuration/#LinkedDomains",
      "DomainLinkageCredential": "https://identity.foundation/.well-known/resources/did-configuration/#DomainLinkageCredential",
      "origin": "https://identity.foundation/.well-known/resources/did-configuration/#origin",
      "linked_dids": "https://identity.foundation/.well-known/resources/did-configuration/#linked_dids"
    }
  ]
}`

func TestInterop(t *testing.T) {
	const contextV0 = "https://identity.foundation/.well-known/contexts/did-configuration-v0.0.jsonld"

	const contextV1 = "https://identity.foundation/.well-known/did-configuration/v1"

	loader, err := ldtestutil.DocumentLoader(
		ldcontext.Document{
			URL:     contextV1,
			Content: json.RawMessage(didCfgCtxV1),
		},
		ldcontext.Document{
			URL:     contextV0,
			Content: json.RawMessage(didCfgCtxV0),
		},
	)
	require.NoError(t, err)

	httpClient := &mockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(msCfg))),
			}, nil
		},
	}

	t.Run("success - MS interop", func(t *testing.T) {
		testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			require.Equal(t, "/"+msDID, req.URL.String())
			res.Header().Add("Content-type", "application/did+ld+json")
			res.WriteHeader(http.StatusOK)
			_, err := res.Write([]byte(msResolutionResponse))
			require.NoError(t, err)
		}))

		defer func() { testServer.Close() }()

		resolver, err := httpbinding.New(testServer.URL)
		require.NoError(t, err)

		gotDocument, err := resolver.Read(msDID)
		require.NoError(t, err)
		didDoc, err := did.ParseDocument([]byte(msDoc))
		require.NoError(t, err)
		require.Equal(t, didDoc.ID, gotDocument.DIDDocument.ID)

		c := New(WithJSONLDDocumentLoader(loader), WithHTTPClient(httpClient), WithVDRegistry(vdr.New(vdr.WithVDR(resolver))))

		err = c.VerifyDIDAndDomain(msDID, msDomain)
		require.NoError(t, err)
	})
}

// ms constants.
const (
	// nolint: lll
	msDID    = "did:ion:EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA:eyJkZWx0YSI6eyJwYXRjaGVzIjpbeyJhY3Rpb24iOiJyZXBsYWNlIiwiZG9jdW1lbnQiOnsicHVibGljS2V5cyI6W3siaWQiOiI2NmRkNTFmZTBjYWM0ZjFhYWU4MTJkMGFhMTA5YmMyYXZjU2lnbmluZ0tleS0yZTk3NSIsInB1YmxpY0tleUp3ayI6eyJjcnYiOiJzZWNwMjU2azEiLCJrdHkiOiJFQyIsIngiOiJqNVQ4S1FfQ19IRGxSbXlFX1pwRjltbE1RZ3B4N19fMFJQRHhPVmM4dWt3IiwieSI6InpybDBWSllHWnhVLXFjZWt2SlY4NGs5U2x2STQxam53NG4yTS1WMnB4MGMifSwicHVycG9zZXMiOlsiYXV0aGVudGljYXRpb24iLCJhc3NlcnRpb25NZXRob2QiXSwidHlwZSI6IkVjZHNhU2VjcDI1NmsxVmVyaWZpY2F0aW9uS2V5MjAxOSJ9XSwic2VydmljZXMiOlt7ImlkIjoibGlua2VkZG9tYWlucyIsInNlcnZpY2VFbmRwb2ludCI6eyJvcmlnaW5zIjpbImh0dHBzOi8vZGlkLnJvaGl0Z3VsYXRpLmNvbS8iXX0sInR5cGUiOiJMaW5rZWREb21haW5zIn0seyJpZCI6Imh1YiIsInNlcnZpY2VFbmRwb2ludCI6eyJpbnN0YW5jZXMiOlsiaHR0cHM6Ly9iZXRhLmh1Yi5tc2lkZW50aXR5LmNvbS92MS4wL2E0OTJjZmYyLWQ3MzMtNDA1Ny05NWE1LWE3MWZjMzY5NWJjOCJdfSwidHlwZSI6IklkZW50aXR5SHViIn1dfX1dLCJ1cGRhdGVDb21taXRtZW50IjoiRWlDcXRpZnUwSHg4RUVkbGlrVnZIWGpYZzRLb0pZZUV0cDdZeGlvRzVYWmRKZyJ9LCJzdWZmaXhEYXRhIjp7ImRlbHRhSGFzaCI6IkVpQ1NVQklmYTBXZHBXNm5oVTdNaHlSczRucTFDeEg1V1ZyUjVkUFZYV09MYmciLCJyZWNvdmVyeUNvbW1pdG1lbnQiOiJFaUF1cGoxRWZsOHdjWlRQZTI3X0lGWEJ3MjlzOEN5SXBRX3UzVkRwUmswdkNRIn19"
	msDomain = "https://did.rohitgulati.com/"

	didCfgCtxV0 = `
{
  "@context": [
    {
      "@version": 1.1,
      "didcfg": "https://identity.foundation/.well-known/contexts/did-configuration-v0.0#",
      "domainLinkageAssertion": "didcfg:domainLinkageAssertion",
      "origin": "didcfg:origin",
      "linked_dids": "didcfg:linked_dids",
      "did": "didcfg:did",
      "vc": "didcfg:vc"
    }
  ]
}`

	// source: https://did.rohitgulati.com/.well-known/did-configuration.json
	// nolint: lll
	msCfg = `
{
  "@context": "https://identity.foundation/.well-known/contexts/did-configuration-v0.0.jsonld",
  "linked_dids": [
    "eyJhbGciOiJFUzI1NksiLCJraWQiOiJkaWQ6aW9uOkVpQ01kVkx0enFxVzVuNnpVQzNfc3JaeFdQQ3NlVnhLWHU5RnFROEx5UzFtVEE6ZXlKa1pXeDBZU0k2ZXlKd1lYUmphR1Z6SWpwYmV5SmhZM1JwYjI0aU9pSnlaWEJzWVdObElpd2laRzlqZFcxbGJuUWlPbnNpY0hWaWJHbGpTMlY1Y3lJNlczc2lhV1FpT2lJMk5tUmtOVEZtWlRCallXTTBaakZoWVdVNE1USmtNR0ZoTVRBNVltTXlZWFpqVTJsbmJtbHVaMHRsZVMweVpUazNOU0lzSW5CMVlteHBZMHRsZVVwM2F5STZleUpqY25ZaU9pSnpaV053TWpVMmF6RWlMQ0pyZEhraU9pSkZReUlzSW5naU9pSnFOVlE0UzFGZlExOUlSR3hTYlhsRlgxcHdSamx0YkUxUlozQjROMTlmTUZKUVJIaFBWbU00ZFd0M0lpd2llU0k2SW5weWJEQldTbGxIV25oVkxYRmpaV3QyU2xZNE5HczVVMngyU1RReGFtNTNORzR5VFMxV01uQjRNR01pZlN3aWNIVnljRzl6WlhNaU9sc2lZWFYwYUdWdWRHbGpZWFJwYjI0aUxDSmhjM05sY25ScGIyNU5aWFJvYjJRaVhTd2lkSGx3WlNJNklrVmpaSE5oVTJWamNESTFObXN4Vm1WeWFXWnBZMkYwYVc5dVMyVjVNakF4T1NKOVhTd2ljMlZ5ZG1salpYTWlPbHQ3SW1sa0lqb2liR2x1YTJWa1pHOXRZV2x1Y3lJc0luTmxjblpwWTJWRmJtUndiMmx1ZENJNmV5SnZjbWxuYVc1eklqcGJJbWgwZEhCek9pOHZaR2xrTG5KdmFHbDBaM1ZzWVhScExtTnZiUzhpWFgwc0luUjVjR1VpT2lKTWFXNXJaV1JFYjIxaGFXNXpJbjBzZXlKcFpDSTZJbWgxWWlJc0luTmxjblpwWTJWRmJtUndiMmx1ZENJNmV5SnBibk4wWVc1alpYTWlPbHNpYUhSMGNITTZMeTlpWlhSaExtaDFZaTV0YzJsa1pXNTBhWFI1TG1OdmJTOTJNUzR3TDJFME9USmpabVl5TFdRM016TXROREExTnkwNU5XRTFMV0UzTVdaak16WTVOV0pqT0NKZGZTd2lkSGx3WlNJNklrbGtaVzUwYVhSNVNIVmlJbjFkZlgxZExDSjFjR1JoZEdWRGIyMXRhWFJ0Wlc1MElqb2lSV2xEY1hScFpuVXdTSGc0UlVWa2JHbHJWblpJV0dwWVp6UkxiMHBaWlVWMGNEZFplR2x2UnpWWVdtUktaeUo5TENKemRXWm1hWGhFWVhSaElqcDdJbVJsYkhSaFNHRnphQ0k2SWtWcFExTlZRa2xtWVRCWFpIQlhObTVvVlRkTmFIbFNjelJ1Y1RGRGVFZzFWMVp5VWpWa1VGWllWMDlNWW1jaUxDSnlaV052ZG1WeWVVTnZiVzFwZEcxbGJuUWlPaUpGYVVGMWNHb3hSV1pzT0hkaldsUlFaVEkzWDBsR1dFSjNNamx6T0VONVNYQlJYM1V6VmtSd1Vtc3dka05SSW4xOSM2NmRkNTFmZTBjYWM0ZjFhYWU4MTJkMGFhMTA5YmMyYXZjU2lnbmluZ0tleS0yZTk3NSJ9.eyJzdWIiOiJkaWQ6aW9uOkVpQ01kVkx0enFxVzVuNnpVQzNfc3JaeFdQQ3NlVnhLWHU5RnFROEx5UzFtVEE6ZXlKa1pXeDBZU0k2ZXlKd1lYUmphR1Z6SWpwYmV5SmhZM1JwYjI0aU9pSnlaWEJzWVdObElpd2laRzlqZFcxbGJuUWlPbnNpY0hWaWJHbGpTMlY1Y3lJNlczc2lhV1FpT2lJMk5tUmtOVEZtWlRCallXTTBaakZoWVdVNE1USmtNR0ZoTVRBNVltTXlZWFpqVTJsbmJtbHVaMHRsZVMweVpUazNOU0lzSW5CMVlteHBZMHRsZVVwM2F5STZleUpqY25ZaU9pSnpaV053TWpVMmF6RWlMQ0pyZEhraU9pSkZReUlzSW5naU9pSnFOVlE0UzFGZlExOUlSR3hTYlhsRlgxcHdSamx0YkUxUlozQjROMTlmTUZKUVJIaFBWbU00ZFd0M0lpd2llU0k2SW5weWJEQldTbGxIV25oVkxYRmpaV3QyU2xZNE5HczVVMngyU1RReGFtNTNORzR5VFMxV01uQjRNR01pZlN3aWNIVnljRzl6WlhNaU9sc2lZWFYwYUdWdWRHbGpZWFJwYjI0aUxDSmhjM05sY25ScGIyNU5aWFJvYjJRaVhTd2lkSGx3WlNJNklrVmpaSE5oVTJWamNESTFObXN4Vm1WeWFXWnBZMkYwYVc5dVMyVjVNakF4T1NKOVhTd2ljMlZ5ZG1salpYTWlPbHQ3SW1sa0lqb2liR2x1YTJWa1pHOXRZV2x1Y3lJc0luTmxjblpwWTJWRmJtUndiMmx1ZENJNmV5SnZjbWxuYVc1eklqcGJJbWgwZEhCek9pOHZaR2xrTG5KdmFHbDBaM1ZzWVhScExtTnZiUzhpWFgwc0luUjVjR1VpT2lKTWFXNXJaV1JFYjIxaGFXNXpJbjBzZXlKcFpDSTZJbWgxWWlJc0luTmxjblpwWTJWRmJtUndiMmx1ZENJNmV5SnBibk4wWVc1alpYTWlPbHNpYUhSMGNITTZMeTlpWlhSaExtaDFZaTV0YzJsa1pXNTBhWFI1TG1OdmJTOTJNUzR3TDJFME9USmpabVl5TFdRM016TXROREExTnkwNU5XRTFMV0UzTVdaak16WTVOV0pqT0NKZGZTd2lkSGx3WlNJNklrbGtaVzUwYVhSNVNIVmlJbjFkZlgxZExDSjFjR1JoZEdWRGIyMXRhWFJ0Wlc1MElqb2lSV2xEY1hScFpuVXdTSGc0UlVWa2JHbHJWblpJV0dwWVp6UkxiMHBaWlVWMGNEZFplR2x2UnpWWVdtUktaeUo5TENKemRXWm1hWGhFWVhSaElqcDdJbVJsYkhSaFNHRnphQ0k2SWtWcFExTlZRa2xtWVRCWFpIQlhObTVvVlRkTmFIbFNjelJ1Y1RGRGVFZzFWMVp5VWpWa1VGWllWMDlNWW1jaUxDSnlaV052ZG1WeWVVTnZiVzFwZEcxbGJuUWlPaUpGYVVGMWNHb3hSV1pzT0hkaldsUlFaVEkzWDBsR1dFSjNNamx6T0VONVNYQlJYM1V6VmtSd1Vtc3dka05SSW4xOSIsImlzcyI6ImRpZDppb246RWlDTWRWTHR6cXFXNW42elVDM19zclp4V1BDc2VWeEtYdTlGcVE4THlTMW1UQTpleUprWld4MFlTSTZleUp3WVhSamFHVnpJanBiZXlKaFkzUnBiMjRpT2lKeVpYQnNZV05sSWl3aVpHOWpkVzFsYm5RaU9uc2ljSFZpYkdsalMyVjVjeUk2VzNzaWFXUWlPaUkyTm1Sa05URm1aVEJqWVdNMFpqRmhZV1U0TVRKa01HRmhNVEE1WW1NeVlYWmpVMmxuYm1sdVowdGxlUzB5WlRrM05TSXNJbkIxWW14cFkwdGxlVXAzYXlJNmV5SmpjbllpT2lKelpXTndNalUyYXpFaUxDSnJkSGtpT2lKRlF5SXNJbmdpT2lKcU5WUTRTMUZmUTE5SVJHeFNiWGxGWDFwd1JqbHRiRTFSWjNCNE4xOWZNRkpRUkhoUFZtTTRkV3QzSWl3aWVTSTZJbnB5YkRCV1NsbEhXbmhWTFhGalpXdDJTbFk0TkdzNVUyeDJTVFF4YW01M05HNHlUUzFXTW5CNE1HTWlmU3dpY0hWeWNHOXpaWE1pT2xzaVlYVjBhR1Z1ZEdsallYUnBiMjRpTENKaGMzTmxjblJwYjI1TlpYUm9iMlFpWFN3aWRIbHdaU0k2SWtWalpITmhVMlZqY0RJMU5tc3hWbVZ5YVdacFkyRjBhVzl1UzJWNU1qQXhPU0o5WFN3aWMyVnlkbWxqWlhNaU9sdDdJbWxrSWpvaWJHbHVhMlZrWkc5dFlXbHVjeUlzSW5ObGNuWnBZMlZGYm1Sd2IybHVkQ0k2ZXlKdmNtbG5hVzV6SWpwYkltaDBkSEJ6T2k4dlpHbGtMbkp2YUdsMFozVnNZWFJwTG1OdmJTOGlYWDBzSW5SNWNHVWlPaUpNYVc1clpXUkViMjFoYVc1ekluMHNleUpwWkNJNkltaDFZaUlzSW5ObGNuWnBZMlZGYm1Sd2IybHVkQ0k2ZXlKcGJuTjBZVzVqWlhNaU9sc2lhSFIwY0hNNkx5OWlaWFJoTG1oMVlpNXRjMmxrWlc1MGFYUjVMbU52YlM5Mk1TNHdMMkUwT1RKalptWXlMV1EzTXpNdE5EQTFOeTA1TldFMUxXRTNNV1pqTXpZNU5XSmpPQ0pkZlN3aWRIbHdaU0k2SWtsa1pXNTBhWFI1U0hWaUluMWRmWDFkTENKMWNHUmhkR1ZEYjIxdGFYUnRaVzUwSWpvaVJXbERjWFJwWm5Vd1NIZzRSVVZrYkdsclZuWklXR3BZWnpSTGIwcFpaVVYwY0RkWmVHbHZSelZZV21SS1p5SjlMQ0p6ZFdabWFYaEVZWFJoSWpwN0ltUmxiSFJoU0dGemFDSTZJa1ZwUTFOVlFrbG1ZVEJYWkhCWE5tNW9WVGROYUhsU2N6UnVjVEZEZUVnMVYxWnlValZrVUZaWVYwOU1ZbWNpTENKeVpXTnZkbVZ5ZVVOdmJXMXBkRzFsYm5RaU9pSkZhVUYxY0dveFJXWnNPSGRqV2xSUVpUSTNYMGxHV0VKM01qbHpPRU41U1hCUlgzVXpWa1J3VW1zd2RrTlJJbjE5IiwibmJmIjoxNjU0NzUxMjc3LCJleHAiOjI0NDM2Njk2NzcsInZjIjp7IkBjb250ZXh0IjpbImh0dHBzOi8vd3d3LnczLm9yZy8yMDE4L2NyZWRlbnRpYWxzL3YxIiwiaHR0cHM6Ly9pZGVudGl0eS5mb3VuZGF0aW9uLy53ZWxsLWtub3duL2NvbnRleHRzL2RpZC1jb25maWd1cmF0aW9uLXYwLjAuanNvbmxkIl0sImlzc3VlciI6ImRpZDppb246RWlDTWRWTHR6cXFXNW42elVDM19zclp4V1BDc2VWeEtYdTlGcVE4THlTMW1UQTpleUprWld4MFlTSTZleUp3WVhSamFHVnpJanBiZXlKaFkzUnBiMjRpT2lKeVpYQnNZV05sSWl3aVpHOWpkVzFsYm5RaU9uc2ljSFZpYkdsalMyVjVjeUk2VzNzaWFXUWlPaUkyTm1Sa05URm1aVEJqWVdNMFpqRmhZV1U0TVRKa01HRmhNVEE1WW1NeVlYWmpVMmxuYm1sdVowdGxlUzB5WlRrM05TSXNJbkIxWW14cFkwdGxlVXAzYXlJNmV5SmpjbllpT2lKelpXTndNalUyYXpFaUxDSnJkSGtpT2lKRlF5SXNJbmdpT2lKcU5WUTRTMUZmUTE5SVJHeFNiWGxGWDFwd1JqbHRiRTFSWjNCNE4xOWZNRkpRUkhoUFZtTTRkV3QzSWl3aWVTSTZJbnB5YkRCV1NsbEhXbmhWTFhGalpXdDJTbFk0TkdzNVUyeDJTVFF4YW01M05HNHlUUzFXTW5CNE1HTWlmU3dpY0hWeWNHOXpaWE1pT2xzaVlYVjBhR1Z1ZEdsallYUnBiMjRpTENKaGMzTmxjblJwYjI1TlpYUm9iMlFpWFN3aWRIbHdaU0k2SWtWalpITmhVMlZqY0RJMU5tc3hWbVZ5YVdacFkyRjBhVzl1UzJWNU1qQXhPU0o5WFN3aWMyVnlkbWxqWlhNaU9sdDdJbWxrSWpvaWJHbHVhMlZrWkc5dFlXbHVjeUlzSW5ObGNuWnBZMlZGYm1Sd2IybHVkQ0k2ZXlKdmNtbG5hVzV6SWpwYkltaDBkSEJ6T2k4dlpHbGtMbkp2YUdsMFozVnNZWFJwTG1OdmJTOGlYWDBzSW5SNWNHVWlPaUpNYVc1clpXUkViMjFoYVc1ekluMHNleUpwWkNJNkltaDFZaUlzSW5ObGNuWnBZMlZGYm1Sd2IybHVkQ0k2ZXlKcGJuTjBZVzVqWlhNaU9sc2lhSFIwY0hNNkx5OWlaWFJoTG1oMVlpNXRjMmxrWlc1MGFYUjVMbU52YlM5Mk1TNHdMMkUwT1RKalptWXlMV1EzTXpNdE5EQTFOeTA1TldFMUxXRTNNV1pqTXpZNU5XSmpPQ0pkZlN3aWRIbHdaU0k2SWtsa1pXNTBhWFI1U0hWaUluMWRmWDFkTENKMWNHUmhkR1ZEYjIxdGFYUnRaVzUwSWpvaVJXbERjWFJwWm5Vd1NIZzRSVVZrYkdsclZuWklXR3BZWnpSTGIwcFpaVVYwY0RkWmVHbHZSelZZV21SS1p5SjlMQ0p6ZFdabWFYaEVZWFJoSWpwN0ltUmxiSFJoU0dGemFDSTZJa1ZwUTFOVlFrbG1ZVEJYWkhCWE5tNW9WVGROYUhsU2N6UnVjVEZEZUVnMVYxWnlValZrVUZaWVYwOU1ZbWNpTENKeVpXTnZkbVZ5ZVVOdmJXMXBkRzFsYm5RaU9pSkZhVUYxY0dveFJXWnNPSGRqV2xSUVpUSTNYMGxHV0VKM01qbHpPRU41U1hCUlgzVXpWa1J3VW1zd2RrTlJJbjE5IiwiaXNzdWFuY2VEYXRlIjoiMjAyMi0wNi0wOVQwNTowNzo1Ny42NjRaIiwiZXhwaXJhdGlvbkRhdGUiOiIyMDQ3LTA2LTA5VDA1OjA3OjU3LjY2NFoiLCJ0eXBlIjpbIlZlcmlmaWFibGVDcmVkZW50aWFsIiwiRG9tYWluTGlua2FnZUNyZWRlbnRpYWwiXSwiY3JlZGVudGlhbFN1YmplY3QiOnsiaWQiOiJkaWQ6aW9uOkVpQ01kVkx0enFxVzVuNnpVQzNfc3JaeFdQQ3NlVnhLWHU5RnFROEx5UzFtVEE6ZXlKa1pXeDBZU0k2ZXlKd1lYUmphR1Z6SWpwYmV5SmhZM1JwYjI0aU9pSnlaWEJzWVdObElpd2laRzlqZFcxbGJuUWlPbnNpY0hWaWJHbGpTMlY1Y3lJNlczc2lhV1FpT2lJMk5tUmtOVEZtWlRCallXTTBaakZoWVdVNE1USmtNR0ZoTVRBNVltTXlZWFpqVTJsbmJtbHVaMHRsZVMweVpUazNOU0lzSW5CMVlteHBZMHRsZVVwM2F5STZleUpqY25ZaU9pSnpaV053TWpVMmF6RWlMQ0pyZEhraU9pSkZReUlzSW5naU9pSnFOVlE0UzFGZlExOUlSR3hTYlhsRlgxcHdSamx0YkUxUlozQjROMTlmTUZKUVJIaFBWbU00ZFd0M0lpd2llU0k2SW5weWJEQldTbGxIV25oVkxYRmpaV3QyU2xZNE5HczVVMngyU1RReGFtNTNORzR5VFMxV01uQjRNR01pZlN3aWNIVnljRzl6WlhNaU9sc2lZWFYwYUdWdWRHbGpZWFJwYjI0aUxDSmhjM05sY25ScGIyNU5aWFJvYjJRaVhTd2lkSGx3WlNJNklrVmpaSE5oVTJWamNESTFObXN4Vm1WeWFXWnBZMkYwYVc5dVMyVjVNakF4T1NKOVhTd2ljMlZ5ZG1salpYTWlPbHQ3SW1sa0lqb2liR2x1YTJWa1pHOXRZV2x1Y3lJc0luTmxjblpwWTJWRmJtUndiMmx1ZENJNmV5SnZjbWxuYVc1eklqcGJJbWgwZEhCek9pOHZaR2xrTG5KdmFHbDBaM1ZzWVhScExtTnZiUzhpWFgwc0luUjVjR1VpT2lKTWFXNXJaV1JFYjIxaGFXNXpJbjBzZXlKcFpDSTZJbWgxWWlJc0luTmxjblpwWTJWRmJtUndiMmx1ZENJNmV5SnBibk4wWVc1alpYTWlPbHNpYUhSMGNITTZMeTlpWlhSaExtaDFZaTV0YzJsa1pXNTBhWFI1TG1OdmJTOTJNUzR3TDJFME9USmpabVl5TFdRM016TXROREExTnkwNU5XRTFMV0UzTVdaak16WTVOV0pqT0NKZGZTd2lkSGx3WlNJNklrbGtaVzUwYVhSNVNIVmlJbjFkZlgxZExDSjFjR1JoZEdWRGIyMXRhWFJ0Wlc1MElqb2lSV2xEY1hScFpuVXdTSGc0UlVWa2JHbHJWblpJV0dwWVp6UkxiMHBaWlVWMGNEZFplR2x2UnpWWVdtUktaeUo5TENKemRXWm1hWGhFWVhSaElqcDdJbVJsYkhSaFNHRnphQ0k2SWtWcFExTlZRa2xtWVRCWFpIQlhObTVvVlRkTmFIbFNjelJ1Y1RGRGVFZzFWMVp5VWpWa1VGWllWMDlNWW1jaUxDSnlaV052ZG1WeWVVTnZiVzFwZEcxbGJuUWlPaUpGYVVGMWNHb3hSV1pzT0hkaldsUlFaVEkzWDBsR1dFSjNNamx6T0VONVNYQlJYM1V6VmtSd1Vtc3dka05SSW4xOSIsIm9yaWdpbiI6Imh0dHBzOi8vZGlkLnJvaGl0Z3VsYXRpLmNvbS8ifX19.Ek8mz8O9yw3ZT8ds3sfy0ELqhUJdgJM-DUpgQawubNyI2wfxM8nLeON_zzxBp1uafdsJujCb4KkFg-SKsRoD3A"
  ]
}`

	// nolint:lll
	msDoc = `
{
  "id": "did:ion:EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA:eyJkZWx0YSI6eyJwYXRjaGVzIjpbeyJhY3Rpb24iOiJyZXBsYWNlIiwiZG9jdW1lbnQiOnsicHVibGljS2V5cyI6W3siaWQiOiI2NmRkNTFmZTBjYWM0ZjFhYWU4MTJkMGFhMTA5YmMyYXZjU2lnbmluZ0tleS0yZTk3NSIsInB1YmxpY0tleUp3ayI6eyJjcnYiOiJzZWNwMjU2azEiLCJrdHkiOiJFQyIsIngiOiJqNVQ4S1FfQ19IRGxSbXlFX1pwRjltbE1RZ3B4N19fMFJQRHhPVmM4dWt3IiwieSI6InpybDBWSllHWnhVLXFjZWt2SlY4NGs5U2x2STQxam53NG4yTS1WMnB4MGMifSwicHVycG9zZXMiOlsiYXV0aGVudGljYXRpb24iLCJhc3NlcnRpb25NZXRob2QiXSwidHlwZSI6IkVjZHNhU2VjcDI1NmsxVmVyaWZpY2F0aW9uS2V5MjAxOSJ9XSwic2VydmljZXMiOlt7ImlkIjoibGlua2VkZG9tYWlucyIsInNlcnZpY2VFbmRwb2ludCI6eyJvcmlnaW5zIjpbImh0dHBzOi8vZGlkLnJvaGl0Z3VsYXRpLmNvbS8iXX0sInR5cGUiOiJMaW5rZWREb21haW5zIn0seyJpZCI6Imh1YiIsInNlcnZpY2VFbmRwb2ludCI6eyJpbnN0YW5jZXMiOlsiaHR0cHM6Ly9iZXRhLmh1Yi5tc2lkZW50aXR5LmNvbS92MS4wL2E0OTJjZmYyLWQ3MzMtNDA1Ny05NWE1LWE3MWZjMzY5NWJjOCJdfSwidHlwZSI6IklkZW50aXR5SHViIn1dfX1dLCJ1cGRhdGVDb21taXRtZW50IjoiRWlDcXRpZnUwSHg4RUVkbGlrVnZIWGpYZzRLb0pZZUV0cDdZeGlvRzVYWmRKZyJ9LCJzdWZmaXhEYXRhIjp7ImRlbHRhSGFzaCI6IkVpQ1NVQklmYTBXZHBXNm5oVTdNaHlSczRucTFDeEg1V1ZyUjVkUFZYV09MYmciLCJyZWNvdmVyeUNvbW1pdG1lbnQiOiJFaUF1cGoxRWZsOHdjWlRQZTI3X0lGWEJ3MjlzOEN5SXBRX3UzVkRwUmswdkNRIn19",
  "@context": [
    "https://www.w3.org/ns/did/v1",
    {
      "@base": "did:ion:EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA:eyJkZWx0YSI6eyJwYXRjaGVzIjpbeyJhY3Rpb24iOiJyZXBsYWNlIiwiZG9jdW1lbnQiOnsicHVibGljS2V5cyI6W3siaWQiOiI2NmRkNTFmZTBjYWM0ZjFhYWU4MTJkMGFhMTA5YmMyYXZjU2lnbmluZ0tleS0yZTk3NSIsInB1YmxpY0tleUp3ayI6eyJjcnYiOiJzZWNwMjU2azEiLCJrdHkiOiJFQyIsIngiOiJqNVQ4S1FfQ19IRGxSbXlFX1pwRjltbE1RZ3B4N19fMFJQRHhPVmM4dWt3IiwieSI6InpybDBWSllHWnhVLXFjZWt2SlY4NGs5U2x2STQxam53NG4yTS1WMnB4MGMifSwicHVycG9zZXMiOlsiYXV0aGVudGljYXRpb24iLCJhc3NlcnRpb25NZXRob2QiXSwidHlwZSI6IkVjZHNhU2VjcDI1NmsxVmVyaWZpY2F0aW9uS2V5MjAxOSJ9XSwic2VydmljZXMiOlt7ImlkIjoibGlua2VkZG9tYWlucyIsInNlcnZpY2VFbmRwb2ludCI6eyJvcmlnaW5zIjpbImh0dHBzOi8vZGlkLnJvaGl0Z3VsYXRpLmNvbS8iXX0sInR5cGUiOiJMaW5rZWREb21haW5zIn0seyJpZCI6Imh1YiIsInNlcnZpY2VFbmRwb2ludCI6eyJpbnN0YW5jZXMiOlsiaHR0cHM6Ly9iZXRhLmh1Yi5tc2lkZW50aXR5LmNvbS92MS4wL2E0OTJjZmYyLWQ3MzMtNDA1Ny05NWE1LWE3MWZjMzY5NWJjOCJdfSwidHlwZSI6IklkZW50aXR5SHViIn1dfX1dLCJ1cGRhdGVDb21taXRtZW50IjoiRWlDcXRpZnUwSHg4RUVkbGlrVnZIWGpYZzRLb0pZZUV0cDdZeGlvRzVYWmRKZyJ9LCJzdWZmaXhEYXRhIjp7ImRlbHRhSGFzaCI6IkVpQ1NVQklmYTBXZHBXNm5oVTdNaHlSczRucTFDeEg1V1ZyUjVkUFZYV09MYmciLCJyZWNvdmVyeUNvbW1pdG1lbnQiOiJFaUF1cGoxRWZsOHdjWlRQZTI3X0lGWEJ3MjlzOEN5SXBRX3UzVkRwUmswdkNRIn19"
    }
  ],
  "service": [
    {
      "id": "#linkeddomains",
      "type": "LinkedDomains",
      "serviceEndpoint": {
        "origins": [
          "https://did.rohitgulati.com/"
        ]
      }
    },
    {
      "id": "#hub",
      "type": "IdentityHub",
      "serviceEndpoint": {
        "instances": [
          "https://beta.hub.msidentity.com/v1.0/a492cff2-d733-4057-95a5-a71fc3695bc8"
        ],
        "origins": []
      }
    }
  ],
  "verificationMethod": [
    {
      "id": "#66dd51fe0cac4f1aae812d0aa109bc2avcSigningKey-2e975",
      "controller": "did:ion:EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA:eyJkZWx0YSI6eyJwYXRjaGVzIjpbeyJhY3Rpb24iOiJyZXBsYWNlIiwiZG9jdW1lbnQiOnsicHVibGljS2V5cyI6W3siaWQiOiI2NmRkNTFmZTBjYWM0ZjFhYWU4MTJkMGFhMTA5YmMyYXZjU2lnbmluZ0tleS0yZTk3NSIsInB1YmxpY0tleUp3ayI6eyJjcnYiOiJzZWNwMjU2azEiLCJrdHkiOiJFQyIsIngiOiJqNVQ4S1FfQ19IRGxSbXlFX1pwRjltbE1RZ3B4N19fMFJQRHhPVmM4dWt3IiwieSI6InpybDBWSllHWnhVLXFjZWt2SlY4NGs5U2x2STQxam53NG4yTS1WMnB4MGMifSwicHVycG9zZXMiOlsiYXV0aGVudGljYXRpb24iLCJhc3NlcnRpb25NZXRob2QiXSwidHlwZSI6IkVjZHNhU2VjcDI1NmsxVmVyaWZpY2F0aW9uS2V5MjAxOSJ9XSwic2VydmljZXMiOlt7ImlkIjoibGlua2VkZG9tYWlucyIsInNlcnZpY2VFbmRwb2ludCI6eyJvcmlnaW5zIjpbImh0dHBzOi8vZGlkLnJvaGl0Z3VsYXRpLmNvbS8iXX0sInR5cGUiOiJMaW5rZWREb21haW5zIn0seyJpZCI6Imh1YiIsInNlcnZpY2VFbmRwb2ludCI6eyJpbnN0YW5jZXMiOlsiaHR0cHM6Ly9iZXRhLmh1Yi5tc2lkZW50aXR5LmNvbS92MS4wL2E0OTJjZmYyLWQ3MzMtNDA1Ny05NWE1LWE3MWZjMzY5NWJjOCJdfSwidHlwZSI6IklkZW50aXR5SHViIn1dfX1dLCJ1cGRhdGVDb21taXRtZW50IjoiRWlDcXRpZnUwSHg4RUVkbGlrVnZIWGpYZzRLb0pZZUV0cDdZeGlvRzVYWmRKZyJ9LCJzdWZmaXhEYXRhIjp7ImRlbHRhSGFzaCI6IkVpQ1NVQklmYTBXZHBXNm5oVTdNaHlSczRucTFDeEg1V1ZyUjVkUFZYV09MYmciLCJyZWNvdmVyeUNvbW1pdG1lbnQiOiJFaUF1cGoxRWZsOHdjWlRQZTI3X0lGWEJ3MjlzOEN5SXBRX3UzVkRwUmswdkNRIn19",
      "type": "EcdsaSecp256k1VerificationKey2019",
      "publicKeyJwk": {
        "kty": "EC",
        "crv": "secp256k1",
        "x": "j5T8KQ_C_HDlRmyE_ZpF9mlMQgpx7__0RPDxOVc8ukw",
        "y": "zrl0VJYGZxU-qcekvJV84k9SlvI41jnw4n2M-V2px0c"
      }
    }
  ],
  "authentication": [
    "#66dd51fe0cac4f1aae812d0aa109bc2avcSigningKey-2e975"
  ],
  "assertionMethod": [
    "#66dd51fe0cac4f1aae812d0aa109bc2avcSigningKey-2e975"
  ]
}`

	msDocMetadata = `
{
  "method": {
    "published": true,
    "recoveryCommitment": "EiAupj1Efl8wcZTPe27_IFXBw29s8CyIpQ_u3VDpRk0vCQ",
    "updateCommitment": "EiCqtifu0Hx8EEdlikVvHXjXg4KoJYeEtp7YxioG5XZdJg"
  },
  "equivalentId": [
    "did:ion:EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA"
  ],
  "canonicalId": "did:ion:EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA"
}`

	// nolint:lll
	msResolutionMetadata = `
{
  "contentType": "application/did+ld+json",
  "pattern": "^(did:ion:(?!test).+)$",
  "driverUrl": "http://driver-did-ion:8080/1.0/identifiers/",
  "duration": 403,
  "did": {
    "didString": "did:ion:EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA:eyJkZWx0YSI6eyJwYXRjaGVzIjpbeyJhY3Rpb24iOiJyZXBsYWNlIiwiZG9jdW1lbnQiOnsicHVibGljS2V5cyI6W3siaWQiOiI2NmRkNTFmZTBjYWM0ZjFhYWU4MTJkMGFhMTA5YmMyYXZjU2lnbmluZ0tleS0yZTk3NSIsInB1YmxpY0tleUp3ayI6eyJjcnYiOiJzZWNwMjU2azEiLCJrdHkiOiJFQyIsIngiOiJqNVQ4S1FfQ19IRGxSbXlFX1pwRjltbE1RZ3B4N19fMFJQRHhPVmM4dWt3IiwieSI6InpybDBWSllHWnhVLXFjZWt2SlY4NGs5U2x2STQxam53NG4yTS1WMnB4MGMifSwicHVycG9zZXMiOlsiYXV0aGVudGljYXRpb24iLCJhc3NlcnRpb25NZXRob2QiXSwidHlwZSI6IkVjZHNhU2VjcDI1NmsxVmVyaWZpY2F0aW9uS2V5MjAxOSJ9XSwic2VydmljZXMiOlt7ImlkIjoibGlua2VkZG9tYWlucyIsInNlcnZpY2VFbmRwb2ludCI6eyJvcmlnaW5zIjpbImh0dHBzOi8vZGlkLnJvaGl0Z3VsYXRpLmNvbS8iXX0sInR5cGUiOiJMaW5rZWREb21haW5zIn0seyJpZCI6Imh1YiIsInNlcnZpY2VFbmRwb2ludCI6eyJpbnN0YW5jZXMiOlsiaHR0cHM6Ly9iZXRhLmh1Yi5tc2lkZW50aXR5LmNvbS92MS4wL2E0OTJjZmYyLWQ3MzMtNDA1Ny05NWE1LWE3MWZjMzY5NWJjOCJdfSwidHlwZSI6IklkZW50aXR5SHViIn1dfX1dLCJ1cGRhdGVDb21taXRtZW50IjoiRWlDcXRpZnUwSHg4RUVkbGlrVnZIWGpYZzRLb0pZZUV0cDdZeGlvRzVYWmRKZyJ9LCJzdWZmaXhEYXRhIjp7ImRlbHRhSGFzaCI6IkVpQ1NVQklmYTBXZHBXNm5oVTdNaHlSczRucTFDeEg1V1ZyUjVkUFZYV09MYmciLCJyZWNvdmVyeUNvbW1pdG1lbnQiOiJFaUF1cGoxRWZsOHdjWlRQZTI3X0lGWEJ3MjlzOEN5SXBRX3UzVkRwUmswdkNRIn19",
    "methodSpecificId": "EiCMdVLtzqqW5n6zUC3_srZxWPCseVxKXu9FqQ8LyS1mTA:eyJkZWx0YSI6eyJwYXRjaGVzIjpbeyJhY3Rpb24iOiJyZXBsYWNlIiwiZG9jdW1lbnQiOnsicHVibGljS2V5cyI6W3siaWQiOiI2NmRkNTFmZTBjYWM0ZjFhYWU4MTJkMGFhMTA5YmMyYXZjU2lnbmluZ0tleS0yZTk3NSIsInB1YmxpY0tleUp3ayI6eyJjcnYiOiJzZWNwMjU2azEiLCJrdHkiOiJFQyIsIngiOiJqNVQ4S1FfQ19IRGxSbXlFX1pwRjltbE1RZ3B4N19fMFJQRHhPVmM4dWt3IiwieSI6InpybDBWSllHWnhVLXFjZWt2SlY4NGs5U2x2STQxam53NG4yTS1WMnB4MGMifSwicHVycG9zZXMiOlsiYXV0aGVudGljYXRpb24iLCJhc3NlcnRpb25NZXRob2QiXSwidHlwZSI6IkVjZHNhU2VjcDI1NmsxVmVyaWZpY2F0aW9uS2V5MjAxOSJ9XSwic2VydmljZXMiOlt7ImlkIjoibGlua2VkZG9tYWlucyIsInNlcnZpY2VFbmRwb2ludCI6eyJvcmlnaW5zIjpbImh0dHBzOi8vZGlkLnJvaGl0Z3VsYXRpLmNvbS8iXX0sInR5cGUiOiJMaW5rZWREb21haW5zIn0seyJpZCI6Imh1YiIsInNlcnZpY2VFbmRwb2ludCI6eyJpbnN0YW5jZXMiOlsiaHR0cHM6Ly9iZXRhLmh1Yi5tc2lkZW50aXR5LmNvbS92MS4wL2E0OTJjZmYyLWQ3MzMtNDA1Ny05NWE1LWE3MWZjMzY5NWJjOCJdfSwidHlwZSI6IklkZW50aXR5SHViIn1dfX1dLCJ1cGRhdGVDb21taXRtZW50IjoiRWlDcXRpZnUwSHg4RUVkbGlrVnZIWGpYZzRLb0pZZUV0cDdZeGlvRzVYWmRKZyJ9LCJzdWZmaXhEYXRhIjp7ImRlbHRhSGFzaCI6IkVpQ1NVQklmYTBXZHBXNm5oVTdNaHlSczRucTFDeEg1V1ZyUjVkUFZYV09MYmciLCJyZWNvdmVyeUNvbW1pdG1lbnQiOiJFaUF1cGoxRWZsOHdjWlRQZTI3X0lGWEJ3MjlzOEN5SXBRX3UzVkRwUmswdkNRIn19",
    "method": "ion"
  }
}`

	msResolutionResponse = `{
  "@context": "https://w3id.org/did-resolution/v1",
  "didDocument": ` + msDoc + `,
  "didDocumentMetadata": ` + msDocMetadata + `,
  "didResolutionMetadata": ` + msResolutionMetadata + `
}`
)
