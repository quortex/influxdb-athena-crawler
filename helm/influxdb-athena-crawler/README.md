# influxdb-athena-crawler

![Version: 1.0.0](https://img.shields.io/badge/Version-1.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.2.1](https://img.shields.io/badge/AppVersion-0.2.1-informational?style=flat-square)

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
| defaults.region | string | `""` | The AWS region. |
| defaults.bucket | string | `""` | The AWS bucket to watch. |
| defaults.prefix | string | `""` | The bucket prefix. |
| defaults.cleanObjects | bool | `false` | Whether to delete S3 objects after processing them. |
| defaults.timeout | string | `"10m"` | The global timeout. |
| defaults.influxServers | list | `[]` | The InfluxDB servers addresses. |
| defaults.influxToken | string | `""` | The InfluxDB token. |
| defaults.influxOrg | string | `""` | The InfluxDB org to write to. |
| defaults.influxBucket | string | `""` | The InfluxDB bucket write to. |
| defaults.timestampRow | string | `"timestamp"` | The timestamp row in CSV. |
| defaults.timestampLayout | string | `"2006-01-02T15:04:05.000Z"` | The layout to parse timestamp. |
| defaults.tags | list | `[]` |  |
| defaults.fields | list | `[]` |  |
| defaults.awsCredsSecret | string | `"aws-creds"` | A reference to a secret wit AWS credentials (must contain awsKeyId / awsSecretKey). |
| defaults.schedule | string | `"0 0 * * *"` | The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron. |
| defaults.concurrencyPolicy | string | `"Forbid"` | Specifies how to treat concurrent executions of a Job (Allow / Forbid / Replace). |
| defaults.backoffLimit | int | `6` | Specifies the number of retries before marking a job as failed. |
| defaults.successfulJobsHistoryLimit | int | `3` | The number of successful finished jobs to retain. |
| defaults.failedJobsHistoryLimit | int | `1` | The number of failed finished jobs to retain. |
| defaults.suspend | bool | `false` | This flag tells the controller to suspend subsequent executions, it does not apply to already started executions. |
| defaults.image.repository | string | `"eu.gcr.io/quortex-registry-public/influxdb-athena-crawler"` | influxdb-athena-crawler image repository. |
| defaults.image.tag | string | `"0.2.1"` | influxdb-athena-crawler image tag. |
| defaults.image.pullPolicy | string | `"IfNotPresent"` | influxdb-athena-crawler image pull policy. |
| defaults.restartPolicy | string | `"OnFailure"` | influxdb-athena-crawler restartPolicy (supported values: "OnFailure", "Never"). |
| defaults.imagePullSecrets | list | `[]` | A list of secrets used to pull containers images. |
| defaults.nameOverride | string | `""` | Helm's name computing override. |
| defaults.fullnameOverride | string | `""` | Helm's fullname computing override. |
| defaults.resources | object | `{}` | influxdb-athena-crawler container required resources. |
| defaults.podAnnotations | object | `{}` | Annotations to be added to pods. |
| defaults.nodeSelector | object | `{}` | Node labels for influxdb-athena-crawler pod assignment. |
| defaults.tolerations | list | `[]` | Node tolerations for influxdb-athena-crawler scheduling to nodes with taints. |
| defaults.affinity | object | `{}` | Affinity for influxdb-athena-crawler pod assignment. |
| crawlers | object | `{}` | Crawlers map. Each of the elements of this map defines a crawler, merged with the default values |
| rbac.create | bool | `true` | Specifies whether rbac resources should be created. |

