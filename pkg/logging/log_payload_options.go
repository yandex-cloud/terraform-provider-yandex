package logging

type logPayloadOptions struct {
	marshaller JSONPBMarshaller
	header     HeaderLoggingDecider
}

type LogPayloadClientOption func(*logPayloadOptions)

func LogPayloadClientHeader(x HeaderLoggingDecider) LogPayloadClientOption {
	return func(o *logPayloadOptions) {
		o.header = x
	}
}

func LogPayloadClientMarshaller(x JSONPBMarshaller) LogPayloadClientOption {
	return func(o *logPayloadOptions) {
		o.marshaller = x
	}
}
