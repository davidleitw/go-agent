package session

import "time"

// CreateOptions holds options for creating a session
type CreateOptions struct {
	ID       string
	TTL      time.Duration
	Metadata map[string]string
}

// CreateOption is a function that configures CreateOptions
type CreateOption func(*CreateOptions)

// WithID sets a specific ID for the session
func WithID(id string) CreateOption {
	return func(opts *CreateOptions) {
		opts.ID = id
	}
}

// WithTTL sets the time-to-live for the session
func WithTTL(ttl time.Duration) CreateOption {
	return func(opts *CreateOptions) {
		opts.TTL = ttl
	}
}

// WithMetadata adds metadata to the session
func WithMetadata(key, value string) CreateOption {
	return func(opts *CreateOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]string)
		}
		opts.Metadata[key] = value
	}
}

// ApplyOptions applies the given options to CreateOptions
func ApplyOptions(opts ...CreateOption) CreateOptions {
	options := CreateOptions{
		Metadata: make(map[string]string),
	}

	for _, opt := range opts {
		opt(&options)
	}

	return options
}
