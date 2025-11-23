package servicediscovery

import "time"

type ClientOption func(*clientOptions)

type clientOptions struct {
	timeout    time.Duration
	retries    int
	retryDelay time.Duration
}

func defaultOptions() *clientOptions {
	return &clientOptions{
		timeout:    10 * time.Second,
		retries:    3,
		retryDelay: 1 * time.Second,
	}
}

func WithTimeout(d time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = d
	}
}

func WithRetries(n int) ClientOption {
	return func(o *clientOptions) {
		o.retries = n
	}
}

func WithRetryDelay(d time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.retryDelay = d
	}
}

type RegisterOption func(*registerOptions)

type registerOptions struct {
	name        string
	host        string
	port        int
	protocol    string
	basePath    string
	routes      []Route
	healthCheck string
	tags        []string
	metadata    map[string]string
}

func WithName(name string) RegisterOption {
	return func(o *registerOptions) {
		o.name = name
	}
}

func WithHost(host string) RegisterOption {
	return func(o *registerOptions) {
		o.host = host
	}
}

func WithPort(port int) RegisterOption {
	return func(o *registerOptions) {
		o.port = port
	}
}

func WithProtocol(protocol string) RegisterOption {
	return func(o *registerOptions) {
		o.protocol = protocol
	}
}

func WithBasePath(basePath string) RegisterOption {
	return func(o *registerOptions) {
		o.basePath = basePath
	}
}

func WithRoutes(routes ...Route) RegisterOption {
	return func(o *registerOptions) {
		o.routes = routes
	}
}

func WithHealthCheck(healthCheck string) RegisterOption {
	return func(o *registerOptions) {
		o.healthCheck = healthCheck
	}
}

func WithTags(tags ...string) RegisterOption {
	return func(o *registerOptions) {
		o.tags = tags
	}
}

func WithMetadata(metadata map[string]string) RegisterOption {
	return func(o *registerOptions) {
		o.metadata = metadata
	}
}
