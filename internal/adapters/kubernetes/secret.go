package kubernetes

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// GetSecret reads a Kubernetes Secret and returns its decoded data as a map
// of key to raw bytes. The Kubernetes API base64-encodes Secret values under
// the "data" field, so this method decodes each entry before returning it.
//
// This method never logs or serializes Secret values. It only propagates
// explicit errors when the Secret cannot be found or the payload is
// malformed.
func (c *Client) GetSecret(ctx context.Context, namespace string, name string) (map[string][]byte, error) {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return nil, errors.New("kubernetes GetSecret: namespace is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("kubernetes GetSecret: name is required")
	}
	path := fmt.Sprintf("/api/v1/namespaces/%s/secrets/%s", url.PathEscape(namespace), url.PathEscape(name))
	response, err := c.do(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("kubernetes GetSecret: empty response")
	}
	raw, ok := response["data"]
	if !ok {
		return map[string][]byte{}, nil
	}
	dataMap, ok := raw.(map[string]any)
	if !ok {
		return nil, errors.New("kubernetes GetSecret: unexpected data payload")
	}
	out := make(map[string][]byte, len(dataMap))
	for key, value := range dataMap {
		encoded, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("kubernetes GetSecret: unexpected value type for key %q", key)
		}
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("kubernetes GetSecret: base64 decode failed for key %q: %w", key, err)
		}
		out[key] = decoded
	}
	return out, nil
}
