#!/bin/bash

files=("quakeV1" "regionV1" "newsV1" "feltV1")

for file in "${files[@]}"
do 
	`grep '^//' ${file}_test.go  | awk -F '^//' '{print $2}'  > api-docs/endpoints/${file}.md`
done