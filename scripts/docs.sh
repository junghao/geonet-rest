#!/bin/bash

mkdir 'api-docs/endpoints/geojsonV1'
packages=("geojsonV1/quakeV1" "geojsonV1/regionV1")

for package in "${packages[@]}"
do 
	`grep '^//' ${package}_test.go  | awk -F '^//' '{print $2}' > api-docs/endpoints/${package}.md`
done