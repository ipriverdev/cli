package app

var OutputFormat string

func JSON() bool {
	return OutputFormat == "json"
}
