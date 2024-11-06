#!/bin/bash

echo -n ${GL_NEW_TOKEN} | gcloud secrets --project=abc-proj version add SECRET_ID --dafa-file=-