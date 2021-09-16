package runtime

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/dogmatiq/protean/internal/proteanpb"
	"github.com/dogmatiq/protean/internal/protomime"
	"github.com/dogmatiq/protean/rpcerror"
	"google.golang.org/protobuf/proto"
)

// ClientOptions contains options for a client.
type ClientOptions struct {
	HTTPClient      *http.Client
	InputMediaType  string
	OutputMediaType string
}

// Client implements the common logic for generated clients.
type Client struct {
	m       sync.RWMutex
	baseURL *url.URL
	opts    ClientOptions
}

// NewClient returns a new client with the given options.
func NewClient(
	baseURL *url.URL,
	opts ClientOptions,
) *Client {
	if opts.HTTPClient == nil {
		opts.HTTPClient = http.DefaultClient
	}

	if opts.InputMediaType == "" {
		opts.InputMediaType = protomime.MediaTypes[0]
	}

	if opts.OutputMediaType == "" {
		opts.OutputMediaType = protomime.MediaTypes[0]
	}

	return &Client{
		baseURL: baseURL,
		opts:    opts,
	}
}

// CallUnary invokes a unary RPC method.
func (c *Client) CallUnary(
	ctx context.Context,
	methodPath string,
	in, out proto.Message,
) error {
	c.m.Lock()
	opts := c.opts
	c.m.Unlock()

	data, err := c.marshal(opts.InputMediaType, in)
	if err != nil {
		return fmt.Errorf("unable to marshal RPC input message: %w", err)
	}

	methodURL := *c.baseURL // clone
	methodURL.Path = path.Join(methodURL.Path, methodPath)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		methodURL.String(),
		io.NopCloser(
			bytes.NewReader(data),
		),
	)
	if err != nil {
		return fmt.Errorf("unable to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", opts.InputMediaType)
	req.Header.Set("Accept", acceptHeader(opts.OutputMediaType))

	res, err := c.opts.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to perform HTTP request: %w", err)
	}
	defer res.Body.Close()

	data, err = io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("unable to read HTTP response body: %w", err)
	}

	contentType := res.Header.Get("Content-Type")
	if contentType == "" {
		return fmt.Errorf("unable to unmarshal RPC output message: response has no Content-Type header")
	}

	if res.StatusCode == http.StatusOK {
		if err := c.unmarshal(contentType, data, out); err != nil {
			return fmt.Errorf("unable to unmarshal RPC output message: %w", err)
		}

		return nil
	}

	protoErr := &proteanpb.Error{}
	if err := c.unmarshal(contentType, data, protoErr); err != nil {
		return fmt.Errorf("unable to unmarshal RPC error: %w", err)
	}

	rpcErr, err := rpcerror.FromProto(protoErr)
	if err != nil {
		return fmt.Errorf("unable to unmarshal RPC error: %w", err)
	}

	return rpcErr
}

// marshal unmarshals a Protocol Buffers message based on the given media type.
func (c *Client) marshal(mediaType string, in proto.Message) ([]byte, error) {
	m, ok := protomime.MarshalerForMediaType(mediaType)
	if !ok {
		return nil, fmt.Errorf("unsupported media type (%s)", mediaType)
	}

	return m.Marshal(in)
}

// unmarshal unmarshals a Protocol Buffers message based on the given media type.
func (c *Client) unmarshal(mediaType string, data []byte, out proto.Message) error {
	u, ok := protomime.UnmarshalerForMediaType(mediaType)
	if !ok {
		return fmt.Errorf("unsupported media type (%s)", mediaType)
	}

	return u.Unmarshal(data, out)
}

// acceptHeader builds an HTTP Accept header value that allows all of the
// supported media-types, with preference given to a specific media type.
func acceptHeader(preferredMediaType string) string {
	// Reduce the q-value by even steps for each decreasingly preferable
	// media-type.
	q := 1.0
	step := q / float64(len(protomime.MediaTypes)+1)

	var header strings.Builder
	fmt.Fprintf(&header, "%s;q=%.02f", preferredMediaType, q)
	q -= step

	for _, mediaType := range protomime.MediaTypes {
		if !strings.EqualFold(mediaType, preferredMediaType) {
			fmt.Fprintf(&header, ", %s;q=%.02f", mediaType, q)
			q -= step
		}
	}

	return header.String()
}
