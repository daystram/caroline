package config

var version string

func Version() string {
	if version == "" {
		return "v0.0.0-development"
	}

	return version
}
