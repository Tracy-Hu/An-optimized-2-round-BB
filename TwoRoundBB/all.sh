#!/bin/bash

n=`cat nodetable.csv|wc -l`
for ((i=1;i<$n;i=i+1))
do
  ./main "$i" &
done
