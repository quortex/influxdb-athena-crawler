# influxdb-athena-crawler
An AWS Athena crawler for InfluxDB.

## Overview
This project is a utility designed to get AWS Athena results (CSV objects stored in AWS S3), parse them and write InfluxDB points.

## Configuration

### <a id="Configuration_Optional_args"></a>Optional args
influxdb-athena-crawler takes as argument the parameters below.
| Key              | Description                                                                                                                                                                     | Default                      |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------- |
| region           | The AWS region.                                                                                                                                                                 | `""`                         |
| bucket           | The AWS bucket to watch.                                                                                                                                                        | `""`                         |
| prefix           | The bucket prefix.                                                                                                                                                              | `""`                         |
| clean-objects    | Whether to delete S3 objects after processing them.                                                                                                                             | `false`                      |
| timeout          | The global timeout.                                                                                                                                                             | `""`                         |
| influx-server    | The InfluxDB server address.                                                                                                                                                    | `""`                         |
| influx-token     | The InfluxDB token.                                                                                                                                                             | `""`                         |
| influx-org       | The InfluxDB org to write to.                                                                                                                                                   | `""`                         |
| influx-bucket    | The InfluxDB bucket write to.                                                                                                                                                   | `""`                         |
| timestamp-row    | The timestamp row in CSV.                                                                                                                                                       | `"timestamp"`                |
| timestamp-layout | The layout to parse timestamp.                                                                                                                                                  | `"2006-01-02T15:04:05.000Z"` |
| tag              | Tags to add to InfluxDB point. Could be of the form `--tag=foo` if tag name matches CSV row or `--tag='foo={row:bar}'` to specify row.                                          | `""`                         |
| field            | Fields to add to InfluxDB point. Could be of the form `--field='foo={type:int,row:bar}'`, if not specified, CSV row matches field name. Type can be float, int, string or bool. | `""`                         |

## License
Distributed under the Apache 2.0 License. See `LICENSE` for more information.

## Versioning
We use [SemVer](http://semver.org/) for versioning.

## Help
Got a question?
File a GitHub [issue](https://github.com/quortex/influxdb-athena-crawler/issues).
