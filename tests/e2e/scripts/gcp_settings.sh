#!/bin/bash

cat > $TMP_DIR/gcp_settings.json <<EOF
{
  "metadata": {
    "namespace": "$NAMESPACE"
  },
  "spec": {
      "configuration":{
        "velero":{
          "defaultPlugins": [
            "openshift", "$PROVIDER"
          ]
        }
      },
      "backupLocations": [
        {
          "velero": {
            "default": true,
            "provider": "$PROVIDER",
            
            "objectStorage":{
              "bucket": "$BUCKET"
            }
          }
        }
      ],
#     ,"credential":{
#       "name": "$SECRET",
#       "key": "cloud"
#     },
      "snapshotLocations": [
        {
          "velero": {
            "provider": "$PROVIDER",
            "config": {
              "snapshotLocation" : "$REGION",
              "project": "openshift-qe"
            }
          }
        }
      ]
  }
}
EOF

x=$(cat $TMP_DIR/gcp_settings.json); echo "$x" | grep -o '^[^#]*'  > $TMP_DIR/gcp_settings.json
