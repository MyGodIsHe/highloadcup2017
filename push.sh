#!/bin/bash

go build &&
docker build -t traveler . &&
docker tag traveler stor.highloadcup.ru/travels/dark_albatross &&
docker push stor.highloadcup.ru/travels/dark_albatross
#docker tag traveler stor.highloadcup.ru/travels/white_oyster &&
#docker push stor.highloadcup.ru/travels/white_oyster
