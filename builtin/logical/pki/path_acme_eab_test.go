// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pki

import (
	"crypto/x509"
	"encoding/base64"
	"testing"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/stretchr/testify/require"
)

// TestACME_EabVaultAPIs verify the various Vault auth'd APIs for EAB work as expected,
// with creation, listing and deletions.
func TestACME_EabVaultAPIs(t *testing.T) {
	b, s := CreateBackendWithStorage(t)

	var ids []string

	// Generate an EAB
	resp, err := CBWrite(b, s, "acme/eab", map[string]interface{}{})
	requireSuccessNonNilResponse(t, resp, err, "Failed generating eab")
	requireFieldsSetInResp(t, resp, "id", "key_type", "key_bits", "private_key", "created_on")
	require.Equal(t, "ec", resp.Data["key_type"])
	require.Equal(t, "256", resp.Data["key_bits"])
	ids = append(ids, resp.Data["id"].(string))
	_, err = uuid.ParseUUID(resp.Data["id"].(string))
	require.NoError(t, err, "failed parsing id as a uuid")

	key, err := base64.RawURLEncoding.DecodeString(resp.Data["private_key"].(string))
	require.NoError(t, err, "failed base64 decoding private key")
	_, err = x509.ParseECPrivateKey(key)
	require.NoError(t, err, "failed parsing private key")

	// Generate another EAB
	resp, err = CBWrite(b, s, "acme/eab", map[string]interface{}{})
	requireSuccessNonNilResponse(t, resp, err, "Failed generating eab")
	ids = append(ids, resp.Data["id"].(string))

	// List our EABs
	resp, err = CBList(b, s, "acme/eab")
	requireSuccessNonNilResponse(t, resp, err, "failed list")

	require.ElementsMatch(t, ids, resp.Data["keys"])
	keyInfo := resp.Data["key_info"].(map[string]interface{})
	id0Map := keyInfo[ids[0]].(map[string]interface{})
	require.Equal(t, "ec", id0Map["key_type"])
	require.Equal(t, "256", id0Map["key_bits"])
	require.NotEmpty(t, id0Map["created_on"])
	_, err = time.Parse(time.RFC3339, id0Map["created_on"].(string))
	require.NoError(t, err, "failed to parse created_on date: %s", id0Map["created_on"])

	id1Map := keyInfo[ids[1]].(map[string]interface{})

	require.Equal(t, "ec", id1Map["key_type"])
	require.Equal(t, "256", id1Map["key_bits"])
	require.NotEmpty(t, id1Map["created_on"])

	// Delete an EAB
	resp, err = CBDelete(b, s, "acme/eab/"+ids[0])
	requireSuccessNonNilResponse(t, resp, err, "failed deleting eab identifier")
	require.Len(t, resp.Warnings, 0, "no warnings should have been set on delete")

	// Make sure it's really gone
	resp, err = CBList(b, s, "acme/eab")
	requireSuccessNonNilResponse(t, resp, err, "failed list post delete")
	require.Len(t, resp.Data["keys"], 1)
	require.Contains(t, resp.Data["keys"], ids[1])

	// Delete the same EAB again, we should just get a warning but still success.
	resp, err = CBDelete(b, s, "acme/eab/"+ids[0])
	requireSuccessNonNilResponse(t, resp, err, "failed deleting eab identifier")
	require.Len(t, resp.Warnings, 1, "expected a warning to be set on repeated delete call")
}
