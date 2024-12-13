package stinfluxdb

// DbParams provides various configuration options for influxDB.
type DbParams struct {
	URL    string
	Org    string
	Token  string
	Bucket string
}
