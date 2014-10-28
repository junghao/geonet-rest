#!/bin/bash

files=("quakeV1" "regionV1" "newsV1" "feltV1")

for file in "${files[@]}"
do 
	`grep '^//' ${file}_test.go  | awk -F '^//' '{print $2}' | awk '{sub("SERVER","http://ec2-54-253-219-100.ap-southeast-2.compute.amazonaws.com:8080", $0)}{print $0}' > api-docs/endpoints/${file}.md`
done