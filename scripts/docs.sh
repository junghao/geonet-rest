#!/bin/bash

mkdir 'api-docs/endpoints/geojsonV1'
mkdir 'api-docs/endpoints/jsonV1'
packages=("geojsonV1/quakeV1" "geojsonV1/regionV1" "jsonV1/newsV1" "geojsonV1/feltV1")

for package in "${packages[@]}"
do 
	`grep '^//' ${package}_test.go  | awk -F '^//' '{print $2}' | awk '{sub("SERVER","http://ec2-54-253-219-100.ap-southeast-2.compute.amazonaws.com:8080", $0)}{print $0}' > api-docs/endpoints/${package}.md`
done