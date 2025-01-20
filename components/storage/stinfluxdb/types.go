package stinfluxdb

// DBParams provides various configuration options for influxDB.
type DBParams struct {
	URL    string
	Org    string
	Token  string
	Bucket string
}
