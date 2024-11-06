#!/bin/bash

echo -n ${GL_NEW_TOKEN} | gcloud secrets --project=${GCP_PROJECT} version add "${SECRET_ID}" --dafa-file=-