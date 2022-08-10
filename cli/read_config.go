package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/anycable/anycable-go/config"
	"github.com/anycable/anycable-go/version"
	"github.com/urfave/cli/v2"
)

// Flags ordering issue: https://github.com/urfave/cli/pull/1430

const (
	serverCategoryDescription      = "ANYCABLE-GO SERVER:"
	sslCategoryDescription         = "SSL:"
	adapterCategoryDescription     = "ADAPTER:"
	redisCategoryDescription       = "REDIS:"
	httpCategoryDescription        = "HTTP:"
	natsCategoryDescription        = "NATS:"
	rpcCategoryDescription         = "RPC:"
	disconnectCategoryDescription  = "DISCONNECT OPTIONS:"
	logCategoryDescription         = "LOG:"
	metricsCategoryDescription     = "METRICS:"
	wsCategoryDescription          = "WEBSOCKET:"
	pingCategoryDescription        = "PING:"
	jwtCategoryDescription         = "JWT:"
	miscCategoryDescription        = "MISCELLANEOUS:"
	natsServiceCategoryDescription = "EMBEDDED NATS SERVICE:"
)

// NewConfigFromCLI reads config from os.Args. It returns config, error (if any) and a bool value
// indicating that the usage message or version was shown, no further action required.
func NewConfigFromCLI() (*config.Config, error, bool) {
	c := config.NewConfig()

	var path, headers, routes string
	var helpOrVersionWereShown bool = true

	// Print raw version without prefix
	cli.VersionPrinter = func(cCtx *cli.Context) {
		_, _ = fmt.Fprintf(cCtx.App.Writer, "%v\n", cCtx.App.Version)
	}

	app := &cli.App{
		Name:            "anycable-go",
		Version:         version.Version(),
		Usage:           "AnyCable-Go, The WebSocket server for https://anycable.io",
		HideHelpCommand: true,
		Flags: []cli.Flag{
			// Server options
			&cli.StringFlag{
				Name:        "host",
				Category:    serverCategoryDescription,
				Value:       c.Host,
				Usage:       "Server host",
				EnvVars:     []string{"ANYCABLE_HOST"},
				Destination: &c.Host,
			},

			&cli.IntFlag{
				Name:        "port",
				Category:    serverCategoryDescription,
				Value:       c.Port,
				Usage:       "Server port",
				EnvVars:     []string{"PORT", "ANYCABLE_PORT"},
				Destination: &c.Port,
			},

			&cli.IntFlag{
				Name:        "max-conn",
				Category:    serverCategoryDescription,
				Usage:       "Limit simultaneous server connections (0 – without limit)",
				EnvVars:     []string{"ANYCABLE_MAX_CONN"},
				Destination: &c.MaxConn,
			},

			&cli.StringFlag{
				Name:        "path",
				Category:    serverCategoryDescription,
				Value:       strings.Join(c.Path, ","),
				Usage:       "WebSocket endpoint path (you can specify multiple paths using comma as separator)",
				EnvVars:     []string{"ANYCABLE_PATH"},
				Destination: &path,
			},

			&cli.StringFlag{
				Name:        "health-path",
				Category:    serverCategoryDescription,
				Value:       c.HealthPath,
				Usage:       "HTTP health endpoint path",
				EnvVars:     []string{"ANYCABLE_HEALTH_PATH"},
				Destination: &c.HealthPath,
			},

			// SSL
			&cli.PathFlag{
				Name:        "ssl_cert",
				Category:    sslCategoryDescription,
				Usage:       "SSL certificate path",
				EnvVars:     []string{"ANYCABLE_SSL_CERT"},
				Destination: &c.SSL.CertPath,
			},

			&cli.PathFlag{
				Name:        "ssl_key",
				Category:    sslCategoryDescription,
				Usage:       "SSL private key path",
				EnvVars:     []string{"ANYCABLE_SSL_KEY"},
				Destination: &c.SSL.KeyPath,
			},

			// Broadcast
			&cli.StringFlag{
				Name:        "broadcast_adapter",
				Category:    adapterCategoryDescription,
				Usage:       "Broadcasting adapter to use (redis, http or nats)",
				Value:       c.BroadcastAdapter,
				EnvVars:     []string{"ANYCABLE_BROADCAST_ADAPTER"},
				Destination: &c.BroadcastAdapter,
			},

			// Redis
			&cli.StringFlag{
				Name:        "redis_url",
				Category:    redisCategoryDescription,
				Usage:       "Redis url",
				Value:       c.Redis.URL,
				EnvVars:     []string{"ANYCABLE_REDIS_URL", "REDIS_URL"},
				Destination: &c.Redis.URL,
			},

			&cli.StringFlag{
				Name:        "redis_channel",
				Category:    redisCategoryDescription,
				Usage:       "Redis channel for broadcasts",
				Value:       c.Redis.Channel,
				EnvVars:     []string{"ANYCABLE_REDIS_CHANNEL"},
				Destination: &c.Redis.Channel,
			},

			&cli.StringFlag{
				Name:        "redis_sentinels",
				Category:    redisCategoryDescription,
				Usage:       "Comma separated list of sentinel hosts, format: 'hostname:port,..'",
				EnvVars:     []string{"ANYCABLE_REDIS_SENTINELS"},
				Destination: &c.Redis.Sentinels,
			},

			&cli.IntFlag{
				Name:        "redis_sentinel_discovery_interval",
				Category:    redisCategoryDescription,
				Usage:       "Interval to rediscover sentinels in seconds",
				Value:       c.Redis.SentinelDiscoveryInterval,
				EnvVars:     []string{"ANYCABLE_REDIS_SENTINEL_DISCOVERY_INTERVAL"},
				Destination: &c.Redis.SentinelDiscoveryInterval,
			},

			&cli.IntFlag{
				Name:        "redis_keeepalive_interval",
				Category:    redisCategoryDescription,
				Usage:       "Interval to periodically ping Redis to make sure it's alive",
				Value:       c.Redis.KeepalivePingInterval,
				EnvVars:     []string{"ANYCABLE_REDIS_KEEPALIVE_INTERVAL"},
				Destination: &c.Redis.KeepalivePingInterval,
			},

			// HTTP
			&cli.IntFlag{
				Name:        "http_broadcast_port",
				Category:    httpCategoryDescription,
				Usage:       "HTTP pub/sub server port",
				Value:       c.HTTPPubSub.Port,
				EnvVars:     []string{"ANYCABLE_HTTP_BROADCAST_PORT"},
				Destination: &c.HTTPPubSub.Port,
			},

			&cli.StringFlag{
				Name:        "http_broadcast_path",
				Category:    httpCategoryDescription,
				Usage:       "HTTP pub/sub endpoint path",
				Value:       c.HTTPPubSub.Path,
				EnvVars:     []string{"ANYCABLE_HTTP_BROADCAST_PATH"},
				Destination: &c.HTTPPubSub.Path,
			},

			&cli.StringFlag{
				Name:        "http_broadcast_secret",
				Category:    httpCategoryDescription,
				Usage:       "HTTP pub/sub authorization secret",
				EnvVars:     []string{"ANYCABLE_HTTP_BROADCAST_SECRET"},
				Destination: &c.HTTPPubSub.Secret,
			},

			// NATS
			&cli.StringFlag{
				Name:        "nats_servers",
				Category:    natsCategoryDescription,
				Usage:       "Comma separated list of NATS cluster servers",
				Value:       c.NATSPubSub.Servers,
				EnvVars:     []string{"ANYCABLE_NATS_SERVERS"},
				Destination: &c.NATSPubSub.Servers,
			},

			&cli.StringFlag{
				Name:        "nats_channel",
				Category:    natsCategoryDescription,
				Usage:       "NATS channel for broadcasts",
				Value:       c.NATSPubSub.Channel,
				EnvVars:     []string{"ANYCABLE_NATS_CHANNEL"},
				Destination: &c.NATSPubSub.Channel,
			},

			&cli.BoolFlag{
				Name:        "nats_dont_randomize_servers",
				Category:    natsCategoryDescription,
				Usage:       "Pass this option to disable NATS servers randomization during (re-)connect",
				EnvVars:     []string{"ANYCABLE_NATS_DONT_RANDOMIZE_SERVERS"},
				Destination: &c.NATSPubSub.DontRandomizeServers,
			},

			// RPC
			&cli.StringFlag{
				Name:        "rpc_host",
				Category:    rpcCategoryDescription,
				Usage:       "RPC service address",
				Value:       c.RPC.Host,
				EnvVars:     []string{"ANYCABLE_RPC_HOST"},
				Destination: &c.RPC.Host,
			},

			&cli.IntFlag{
				Name:        "rpc_concurrency",
				Category:    rpcCategoryDescription,
				Usage:       "Max number of concurrent RPC request; should be slightly less than the RPC server concurrency",
				Value:       c.RPC.Concurrency,
				EnvVars:     []string{"ANYCABLE_RPC_CONCURRENCY"},
				Destination: &c.RPC.Concurrency,
			},

			&cli.BoolFlag{
				Name:        "rpc_enable_tls",
				Category:    rpcCategoryDescription,
				Usage:       "Enable client-side TLS with the RPC server",
				EnvVars:     []string{"ANYCABLE_RPC_ENABLE_TLS"},
				Destination: &c.RPC.EnableTLS,
			},

			&cli.IntFlag{
				Name:        "rpc_max_call_recv_size",
				Category:    rpcCategoryDescription,
				Usage:       "Override default MaxCallRecvMsgSize for RPC client (bytes)",
				Value:       c.RPC.MaxRecvSize,
				EnvVars:     []string{"ANYCABLE_RPC_MAX_CALL_RECV_SIZE"},
				Destination: &c.RPC.MaxRecvSize,
			},

			&cli.IntFlag{
				Name:        "rpc_max_call_send_size",
				Category:    rpcCategoryDescription,
				Usage:       "Override default MaxCallSendMsgSize for RPC client (bytes)",
				Value:       c.RPC.MaxSendSize,
				EnvVars:     []string{"ANYCABLE_RPC_MAX_CALL_SEND_SIZE"},
				Destination: &c.RPC.MaxSendSize,
			},

			&cli.StringFlag{
				Name:        "headers",
				Category:    rpcCategoryDescription,
				Usage:       "List of headers to proxy to RPC",
				Value:       strings.Join(c.Headers, ","),
				EnvVars:     []string{"ANYCABLE_HEADERS"},
				Destination: &headers,
			},

			// Disconnect
			&cli.IntFlag{
				Name:        "disconnect_rate",
				Category:    disconnectCategoryDescription,
				Usage:       "Max number of Disconnect calls per second",
				Value:       c.DisconnectQueue.Rate,
				EnvVars:     []string{"ANYCABLE_DISCONNECT_RATE"},
				Destination: &c.DisconnectQueue.Rate,
			},

			&cli.IntFlag{
				Name:        "disconnect_timeout",
				Category:    disconnectCategoryDescription,
				Usage:       "Graceful shutdown timeouts (in seconds)",
				Value:       c.DisconnectQueue.ShutdownTimeout,
				EnvVars:     []string{"ANYCABLE_DISCONNECT_TIMEOUT"},
				Destination: &c.DisconnectQueue.ShutdownTimeout,
			},

			&cli.BoolFlag{
				Name:        "disable_disconnect",
				Category:    disconnectCategoryDescription,
				Usage:       "Disable calling Disconnect callback",
				EnvVars:     []string{"ANYCABLE_DISABLE_DISCONNECT"},
				Destination: &c.DisconnectorDisabled,
			},

			// Log
			&cli.StringFlag{
				Name:        "log_level",
				Category:    logCategoryDescription,
				Usage:       "Set logging level (debug/info/warn/error/fatal)",
				Value:       c.LogLevel,
				EnvVars:     []string{"ANYCABLE_LOG_LEVEL"},
				Destination: &c.LogLevel,
			},

			&cli.StringFlag{
				Name:        "log_format",
				Category:    logCategoryDescription,
				Usage:       "Set logging format (text/json)",
				Value:       c.LogFormat,
				EnvVars:     []string{"ANYCABLE_LOG_FORMAT"},
				Destination: &c.LogFormat,
			},

			&cli.BoolFlag{
				Name:        "debug",
				Category:    logCategoryDescription,
				Usage:       "Enable debug mode (more verbose logging)",
				EnvVars:     []string{"ANYCABLE_DEBUG"},
				Destination: &c.Debug,
			},

			// Metrics
			&cli.BoolFlag{
				Name:        "metrics_log",
				Category:    metricsCategoryDescription,
				Usage:       "Enable metrics logging (with info level)",
				EnvVars:     []string{"ANYCABLE_METRICS_LOG"},
				Destination: &c.Metrics.Log,
			},

			&cli.IntFlag{
				Name:        "metrics_rotate_interval",
				Category:    metricsCategoryDescription,
				Usage:       "Specify how often flush metrics to writers (logs, statsd) (in seconds)",
				Value:       c.Metrics.RotateInterval,
				EnvVars:     []string{"ANYCABLE_METRICS_ROTATE_INTERVAL"},
				Destination: &c.Metrics.RotateInterval,
			},

			&cli.IntFlag{
				Name:        "metrics_log_interval",
				Category:    metricsCategoryDescription,
				Usage:       "DEPRECATED. Specify how often flush metrics logs (in seconds)",
				Value:       c.Metrics.LogInterval,
				EnvVars:     []string{"ANYCABLE_METRICS_LOG_INTERVAL"},
				Destination: &c.Metrics.LogInterval,
			},

			&cli.StringFlag{
				Name:        "metrics_log_formatter",
				Category:    metricsCategoryDescription,
				Usage:       "Specify the path to custom Ruby formatter script (only supported on MacOS and Linux)",
				EnvVars:     []string{"ANYCABLE_METRICS_LOG_FORMATTER"},
				Destination: &c.Metrics.LogFormatter,
			},

			&cli.StringFlag{
				Name:        "metrics_http",
				Category:    metricsCategoryDescription,
				Usage:       "Enable HTTP metrics endpoint at the specified path",
				EnvVars:     []string{"ANYCABLE_METRICS_HTTP"},
				Destination: &c.Metrics.HTTP,
			},

			&cli.StringFlag{
				Name:        "metrics_host",
				Category:    metricsCategoryDescription,
				Usage:       "Server host for metrics endpoint",
				EnvVars:     []string{"ANYCABLE_METRICS_HOST"},
				Destination: &c.Metrics.Host,
			},

			&cli.IntFlag{
				Name:        "metrics_port",
				Category:    metricsCategoryDescription,
				Usage:       "Server port for metrics endpoint, the same as for main server by default",
				EnvVars:     []string{"ANYCABLE_METRICS_PORT"},
				Destination: &c.Metrics.Port,
			},

			// WebSocket
			&cli.IntFlag{
				Name:        "read_buffer_size",
				Category:    wsCategoryDescription,
				Usage:       "WebSocket connection read buffer size",
				Value:       c.WS.ReadBufferSize,
				EnvVars:     []string{"ANYCABLE_READ_BUFFER_SIZE"},
				Destination: &c.WS.ReadBufferSize,
			},

			&cli.IntFlag{
				Name:        "write_buffer_size",
				Category:    wsCategoryDescription,
				Usage:       "WebSocket connection write buffer size",
				Value:       c.WS.WriteBufferSize,
				EnvVars:     []string{"ANYCABLE_WRITE_BUFFER_SIZE"},
				Destination: &c.WS.WriteBufferSize,
			},

			&cli.Int64Flag{
				Name:        "max_message_size",
				Category:    wsCategoryDescription,
				Usage:       "Maximum size of a message in bytes",
				Value:       c.WS.MaxMessageSize,
				EnvVars:     []string{"ANYCABLE_MAX_MESSAGE_SIZE"},
				Destination: &c.WS.MaxMessageSize,
			},

			&cli.BoolFlag{
				Name:        "enable_ws_compression",
				Category:    wsCategoryDescription,
				Usage:       "Enable experimental WebSocket per message compression",
				EnvVars:     []string{"ANYCABLE_ENABLE_WS_COMPRESSION"},
				Destination: &c.WS.EnableCompression,
			},

			&cli.IntFlag{
				Name:        "hub_gopool_size",
				Category:    wsCategoryDescription,
				Usage:       "The size of the goroutines pool to broadcast messages",
				EnvVars:     []string{"ANYCABLE_HUB_GOPOOL_SIZE"},
				Value:       c.App.HubGopoolSize,
				Destination: &c.App.HubGopoolSize,
			},

			&cli.StringFlag{
				Name:        "allowed_origins",
				Category:    wsCategoryDescription,
				Usage:       `Accept requests only from specified origins, e.g., "www.example.com,*example.io". No check is performed if empty`,
				EnvVars:     []string{"ANYCABLE_ALLOWED_ORIGINS"},
				Destination: &c.WS.AllowedOrigins,
			},

			// Ping
			&cli.IntFlag{
				Name:        "ping_interval",
				Category:    pingCategoryDescription,
				Usage:       "Action Cable ping interval (in seconds)",
				Value:       c.App.PingInterval,
				EnvVars:     []string{"ANYCABLE_PING_INTERVAL"},
				Destination: &c.App.PingInterval,
			},

			&cli.StringFlag{
				Name:        "ping_timestamp_precision",
				Category:    pingCategoryDescription,
				Usage:       "Precision for timestamps in ping messages (s, ms, ns)",
				Value:       c.App.PingTimestampPrecision,
				EnvVars:     []string{"ANYCABLE_PING_TIMESTAMP_PRECISION"},
				Destination: &c.App.PingTimestampPrecision,
			},

			&cli.IntFlag{
				Name:        "stats_refresh_interval",
				Category:    pingCategoryDescription,
				Usage:       "How often to refresh the server stats (in seconds)",
				Value:       c.App.StatsRefreshInterval,
				EnvVars:     []string{"ANYCABLE_STATS_REFRESH_INTERVAL"},
				Destination: &c.App.StatsRefreshInterval,
			},

			// JWT
			&cli.StringFlag{
				Name:        "jwt_id_key",
				Category:    jwtCategoryDescription,
				Usage:       "The encryption key used to verify JWT tokens",
				EnvVars:     []string{"ANYCABLE_JWT_ID_KEY"},
				Destination: &c.JWT.Secret,
			},

			&cli.StringFlag{
				Name:        "jwt_id_param",
				Category:    jwtCategoryDescription,
				Usage:       "The name of a query string param or an HTTP header carrying a token",
				Value:       c.JWT.Param,
				EnvVars:     []string{"ANYCABLE_JWT_ID_PARAM"},
				Destination: &c.JWT.Param,
			},

			&cli.BoolFlag{
				Name:        "jwt_id_enforce",
				Category:    jwtCategoryDescription,
				Usage:       "Whether to enforce token presence for all connections",
				EnvVars:     []string{"ANYCABLE_JWT_ID_ENFORCE"},
				Destination: &c.JWT.Force,
			},

			// Misc
			&cli.StringFlag{
				Name:        "turbo_rails_key",
				Category:    miscCategoryDescription,
				Usage:       "Enable Turbo Streams fastlane with the specified signing key",
				EnvVars:     []string{"ANYCABLE_TURBO_RAILS_KEY"},
				Destination: &c.Rails.TurboRailsKey,
			},

			&cli.StringFlag{
				Name:        "cable_ready_key",
				Category:    miscCategoryDescription,
				Usage:       "Enable CableReady fastlane with the specified signing key",
				EnvVars:     []string{"ANYCABLE_CABLE_READY_KEY"},
				Destination: &c.Rails.CableReadyKey,
			},

			// NATS Service
			&cli.BoolFlag{
				Name:        "mnats_enable",
				Category:    natsServiceCategoryDescription,
				Usage:       "Pass this option to enable embedded NATS server",
				EnvVars:     []string{"ANYCABLE_MNATS_ENABLE"},
				Destination: &c.NATSService.Enable,
			},

			&cli.StringFlag{
				Name:        "mnats_service_addr",
				Category:    natsServiceCategoryDescription,
				Usage:       "NATS server bind address",
				Value:       c.NATSService.ServiceAddr,
				EnvVars:     []string{"ANYCABLE_MNATS_SERVICE_ADDR"},
				Destination: &c.NATSService.ServiceAddr,
			},

			&cli.StringFlag{
				Name:        "mnats_cluster_addr",
				Category:    natsServiceCategoryDescription,
				Usage:       "NATS cluster service bind address",
				Value:       c.NATSService.ClusterAddr,
				EnvVars:     []string{"ANYCABLE_MNATS_CLUSTER_ADDR"},
				Destination: &c.NATSService.ClusterAddr,
			},

			&cli.StringFlag{
				Name:        "mnats_cluster_name",
				Category:    natsServiceCategoryDescription,
				Usage:       "NATS cluster name",
				Value:       c.NATSService.ClusterName,
				EnvVars:     []string{"ANYCABLE_MNATS_CLUSTER_NAME"},
				Destination: &c.NATSService.ClusterName,
			},

			&cli.StringFlag{
				Name:        "mnats_routes",
				Category:    natsServiceCategoryDescription,
				Usage:       "Comma separated list of known other cluster service addresses.",
				Value:       strings.Join(c.NATSService.Routes, ","),
				EnvVars:     []string{"ANYCABLE_MNATS_ROUTES"},
				Destination: &routes,
			},

			&cli.BoolFlag{
				Name:        "mnats_log",
				Category:    natsServiceCategoryDescription,
				Usage:       "Enable NATS server logs",
				EnvVars:     []string{"ANYCABLE_MNATS_LOG"},
				Destination: &c.NATSService.EnableLogging,
			},

			&cli.BoolFlag{
				Name:        "mnats_debug",
				Category:    natsServiceCategoryDescription,
				Usage:       "Enable NATS server debug logs. Automatically set to true if --debug is passed.",
				EnvVars:     []string{"ANYCABLE_MNATS_DEBUG"},
				Destination: &c.NATSService.Debug,
			},

			&cli.BoolFlag{
				Name:        "mnats_trace",
				Category:    natsServiceCategoryDescription,
				Usage:       "Enable NATS server protocol trace logs.",
				EnvVars:     []string{"ANYCABLE_MNATS_TRACE"},
				Destination: &c.NATSService.Debug,
			},
		},

		Action: func(nc *cli.Context) error {
			helpOrVersionWereShown = false
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		return &config.Config{}, err, false
	}

	// helpOrVersionWereShown = false indicates that the default action has been run.
	// true means that help/version message was displayed.
	//
	// Unfortunately, cli module does not support another way of detecting if or which
	// command was run.
	if helpOrVersionWereShown {
		return &config.Config{}, nil, true
	}

	if path != "" {
		c.Path = strings.Split(path, " ")
	}

	if routes != "" {
		c.NATSService.Routes = strings.Split(routes, ",")
	}

	c.Headers = strings.Split(strings.ToLower(headers), ",")

	if c.Debug {
		c.LogLevel = "debug"
		c.LogFormat = "text"
		c.NATSService.Debug = true
	}

	if c.Metrics.Port == 0 {
		c.Metrics.Port = c.Port
	}

	if c.Metrics.LogInterval > 0 {
		fmt.Println(`DEPRECATION WARNING: metrics_log_interval option is deprecated
and will be deleted in the next major release of anycable-go.
Use metrics_rotate_interval instead.`)

		if c.Metrics.RotateInterval == 0 {
			c.Metrics.RotateInterval = c.Metrics.LogInterval
		}
	}

	return &c, nil, false
}
