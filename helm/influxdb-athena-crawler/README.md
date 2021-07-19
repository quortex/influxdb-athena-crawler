# influxdb-athena-crawler

![Version: 0.2.0](https://img.shields.io/badge/Version-0.2.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.2.0](https://img.shields.io/badge/AppVersion-0.2.0-informational?style=flat-square)

A cronjob that get athena reports on s3 and writes to influxdb periodically.

## Overview
This project is a utility designed to get AWS Athena results (CSV objects stored in AWS S3), parse them and write InfluxDB points.

## Prerequisites

### <a id="Prerequisites_AWS"></a>AWS
To be used with AWS and interact with the s3 bucket, an AWS account with the following permissions on s3 is required (note that `s3:DeleteObject` is only required if clean-objects is set):
```json
{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect" : "Allow",
        "Action" : [
          "s3:ListBucket"
        ],
        "Resource" : "<BUCKET_NAME>"
      },
      {
        "Effect" : "Allow",
        "Action" : [
          "s3:ListObjects",
          "s3:GetObject",
          "s3:DeleteObject"
        ],
        "Resource" : "<BUCKET_NAME>/*"
      }
    ]
  }
```

## Installation

1. Add influxdb-athena-crawler helm repository

```sh
helm repo add influxdb-athena-crawler https://quortex.github.io/influxdb-athena-crawler
```

2. Deploy the appropriate release in desired namespace.

```sh
helm install influxdb-athena-crawler influxdb-athena-crawler/influxdb-athena-crawler -n <NAMESPACE>>
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| region | string | `""` | The AWS region. |
| bucket | string | `""` | The AWS bucket to watch. |
| prefix | string | `""` | The bucket prefix. |
| cleanObjects | bool | `false` | Whether to delete S3 objects after processing them. |
| timeout | string | `"10m"` | The global timeout. |
| influxServers | list | `[]` | The InfluxDB servers addresses. |
| influxToken | string | `""` | The InfluxDB token. |
| influxOrg | string | `""` | The InfluxDB org to write to. |
| influxBucket | string | `""` | The InfluxDB bucket write to. |
| timestampRow | string | `"timestamp"` | The timestamp row in CSV. |
| timestampLayout | string | `"2006-01-02T15:04:05.000Z"` | The layout to parse timestamp. |
| tags | list | `[]` |  |
| fields | list | `[]` |  |
| awsCredsSecret | string | `"aws-creds"` | A reference to a secret wit AWS credentials (must contain awsKeyId / awsSecretKey). |
| schedule | string | `"0 0 * * *"` | The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron. |
| concurrencyPolicy | string | `"Forbid"` | Specifies how to treat concurrent executions of a Job (Allow / Forbid / Replace). |
| backoffLimit | int | `6` | Specifies the number of retries before marking a job as failed. |
| successfulJobsHistoryLimit | int | `3` | The number of successful finished jobs to retain. |
| failedJobsHistoryLimit | int | `1` | The number of failed finished jobs to retain. |
| suspend | bool | `false` | This flag tells the controller to suspend subsequent executions, it does not apply to already started executions. |
| image.repository | string | `"eu.gcr.io/quortex-registry-public/influxdb-athena-crawler"` | influxdb-athena-crawler image repository. |
| image.tag | string | `"0.2.0"` | influxdb-athena-crawler image tag. |
| image.pullPolicy | string | `"IfNotPresent"` | influxdb-athena-crawler image pull policy. |
| rbac.create | bool | `true` | Specifies whether rbac resources should be created. |
| restartPolicy | string | `"OnFailure"` | influxdb-athena-crawler restartPolicy (supported values: "OnFailure", "Never"). |
| imagePullSecrets | list | `[]` | A list of secrets used to pull containers images. |
| nameOverride | string | `""` | Helm's name computing override. |
| fullnameOverride | string | `""` | Helm's fullname computing override. |
| resources | object | `{}` | influxdb-athena-crawler container required resources. |
| podAnnotations | object | `{}` | Annotations to be added to pods. |
| nodeSelector | object | `{}` | Node labels for influxdb-athena-crawler pod assignment. |
| tolerations | list | `[]` | Node tolerations for influxdb-athena-crawler scheduling to nodes with taints. |
| affinity | object | `{}` | Affinity for influxdb-athena-crawler pod assignment. |
