#!/bin/bash

for ((i=1;i<10000;i++));
do
	curl 10.1.1.1:50080;
done
