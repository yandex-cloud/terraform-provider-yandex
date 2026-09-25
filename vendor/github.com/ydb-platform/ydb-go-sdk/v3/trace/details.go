package trace

import (
	"regexp"
	"sort"
	"strings"
)

type Detailer interface {
	Details() Details
}

var _ Detailer = Details(0)

type Details uint64

func (d Details) Details() Details {
	return d
}

func (d Details) String() string {
	ss := make([]string, 0)
	for bit, name := range detailsMap {
		if d&bit != 0 {
			ss = append(ss, name)
		}
	}
	sort.Strings(ss)

	return strings.Join(ss, "|")
}

const (
	DriverNetEvents Details = 1 << iota // for bitmask: 1, 2, 4, 8, 16, 32, ...
	DriverConnEvents
	DriverConnStreamEvents
	DriverBalancerEvents
	DriverResolverEvents
	DriverRepeaterEvents
	DriverCredentialsEvents

	TableSessionLifeCycleEvents
	TableSessionQueryInvokeEvents
	TableSessionQueryStreamEvents
	TableSessionTransactionEvents
	TablePoolLifeCycleEvents
	TablePoolSessionLifeCycleEvents
	TablePoolAPIEvents

	QuerySessionEvents
	QueryResultEvents
	QueryTransactionEvents
	QueryPoolEvents

	TopicControlPlaneEvents

	TopicReaderCustomerEvents

	TopicReaderStreamLifeCycleEvents
	TopicReaderStreamEvents
	TopicReaderTransactionEvents
	TopicReaderMessageEvents
	TopicReaderPartitionEvents

	TopicListenerStreamEvents
	TopicListenerWorkerEvents
	TopicListenerPartitionEvents

	TopicWriterStreamLifeCycleEvents
	TopicWriterStreamEvents
	TopicWriterStreamGrpcMessageEvents

	DatabaseSQLConnectorEvents
	DatabaseSQLConnEvents
	DatabaseSQLTxEvents
	DatabaseSQLStmtEvents

	RetryEvents

	DiscoveryEvents

	SchemeEvents

	ScriptingEvents

	RatelimiterEvents

	CoordinationEvents

	DriverEvents = DriverConnEvents |
		DriverConnStreamEvents |
		DriverBalancerEvents |
		DriverResolverEvents |
		DriverRepeaterEvents |
		DriverCredentialsEvents

	TableEvents = TableSessionLifeCycleEvents |
		TableSessionQueryInvokeEvents |
		TableSessionQueryStreamEvents |
		TableSessionTransactionEvents |
		TablePoolLifeCycleEvents |
		TablePoolSessionLifeCycleEvents |
		TablePoolAPIEvents

	QueryEvents = QuerySessionEvents |
		QueryPoolEvents |
		QueryResultEvents |
		QueryTransactionEvents

	TablePoolEvents = TablePoolLifeCycleEvents |
		TablePoolSessionLifeCycleEvents |
		TablePoolAPIEvents

	TableSessionQueryEvents = TableSessionQueryInvokeEvents |
		TableSessionQueryStreamEvents
	TableSessionEvents = TableSessionLifeCycleEvents |
		TableSessionQueryEvents |
		TableSessionTransactionEvents

	TopicReaderEvents = TopicReaderCustomerEvents | TopicReaderStreamEvents | TopicReaderMessageEvents |
		TopicReaderPartitionEvents |
		TopicReaderStreamLifeCycleEvents

	TopicListenerEvents = TopicListenerStreamEvents | TopicListenerWorkerEvents | TopicListenerPartitionEvents

	TopicWriterEvents = TopicWriterStreamLifeCycleEvents | TopicWriterStreamEvents |
		TopicWriterStreamGrpcMessageEvents

	TopicEvents = TopicControlPlaneEvents | TopicReaderEvents | TopicListenerEvents | TopicWriterEvents

	DatabaseSQLEvents = DatabaseSQLConnectorEvents |
		DatabaseSQLConnEvents |
		DatabaseSQLTxEvents |
		DatabaseSQLStmtEvents

	DetailsAll = ^Details(0) // All bits enabled
)

var (
	detailsMap = map[Details]string{
		DriverEvents:            "ydb.driver",
		DriverBalancerEvents:    "ydb.driver.balancer",
		DriverResolverEvents:    "ydb.driver.resolver",
		DriverRepeaterEvents:    "ydb.driver.repeater",
		DriverConnEvents:        "ydb.driver.conn",
		DriverConnStreamEvents:  "ydb.driver.conn.stream",
		DriverCredentialsEvents: "ydb.driver.credentials",

		DiscoveryEvents: "ydb.discovery",

		RetryEvents: "ydb.retry",

		SchemeEvents: "ydb.scheme",

		ScriptingEvents: "ydb.scripting",

		CoordinationEvents: "ydb.coordination",

		RatelimiterEvents: "ydb.ratelimiter",

		TableEvents:                     "ydb.table",
		TableSessionLifeCycleEvents:     "ydb.table.session",
		TableSessionQueryInvokeEvents:   "ydb.table.session.query.invoke",
		TableSessionQueryStreamEvents:   "ydb.table.session.query.stream",
		TableSessionTransactionEvents:   "ydb.table.session.tx",
		TablePoolLifeCycleEvents:        "ydb.table.pool",
		TablePoolSessionLifeCycleEvents: "ydb.table.pool.session",
		TablePoolAPIEvents:              "ydb.table.pool.api",

		QueryEvents:            "ydb.query",
		QueryPoolEvents:        "ydb.query.pool",
		QuerySessionEvents:     "ydb.query.session",
		QueryResultEvents:      "ydb.query.result",
		QueryTransactionEvents: "ydb.query.tx",

		DatabaseSQLEvents:          "ydb.database.sql",
		DatabaseSQLConnectorEvents: "ydb.database.sql.connector",
		DatabaseSQLConnEvents:      "ydb.database.sql.conn",
		DatabaseSQLTxEvents:        "ydb.database.sql.tx",
		DatabaseSQLStmtEvents:      "ydb.database.sql.stmt",

		TopicEvents:                        "ydb.topic",
		TopicControlPlaneEvents:            "ydb.topic.controlplane",
		TopicReaderEvents:                  "ydb.topic.reader",
		TopicReaderStreamEvents:            "ydb.topic.reader.stream",
		TopicReaderMessageEvents:           "ydb.topic.reader.message",
		TopicReaderPartitionEvents:         "ydb.topic.reader.partition",
		TopicReaderStreamLifeCycleEvents:   "ydb.topic.reader.lifecycle",
		TopicListenerEvents:                "ydb.topic.listener",
		TopicWriterStreamLifeCycleEvents:   "ydb.topic.writer.lifecycle",
		TopicWriterStreamEvents:            "ydb.topic.writer.stream",
		TopicWriterStreamGrpcMessageEvents: "ydb.topic.writer.grpc",

		TopicListenerStreamEvents:    "ydb.topic.listener.stream",
		TopicListenerWorkerEvents:    "ydb.topic.listener.worker",
		TopicListenerPartitionEvents: "ydb.topic.listener.partition",
	}
	defaultDetails = DetailsAll
)

type matchDetailsOptionsHolder struct {
	defaultDetails Details
	posixMatch     bool
}

type matchDetailsOption func(h *matchDetailsOptionsHolder)

func WithDefaultDetails(defaultDetails Details) matchDetailsOption {
	return func(h *matchDetailsOptionsHolder) {
		h.defaultDetails = defaultDetails
	}
}

func WithPOSIXMatch() matchDetailsOption {
	return func(h *matchDetailsOptionsHolder) {
		h.posixMatch = true
	}
}

func MatchDetails(pattern string, opts ...matchDetailsOption) (d Details) {
	var (
		h = &matchDetailsOptionsHolder{
			defaultDetails: defaultDetails,
		}
		re  *regexp.Regexp
		err error
	)

	for _, opt := range opts {
		if opt != nil {
			opt(h)
		}
	}
	if h.posixMatch {
		re, err = regexp.CompilePOSIX(pattern)
	} else {
		re, err = regexp.Compile(pattern)
	}
	if err != nil {
		return h.defaultDetails
	}
	for k, v := range detailsMap {
		if re.MatchString(v) {
			d |= k
		}
	}
	if d == 0 {
		return h.defaultDetails
	}

	return d
}
