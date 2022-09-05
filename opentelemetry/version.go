package opentelemetry

// Version is the current release version of the tracer instrumentation.
func Version() string {
	return "0.0.1" //git tag version
}

// SemVersion is the semantic version to be supplied to tracer/meter creation.
func SemVersion() string {
	return "semver:" + Version()
}
