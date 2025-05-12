#!/bin/sh

echo "Loading the questions"

redis-cli -p 6969 <<EOF
SET question:1 '{"question":"What is this?","answer1":"Car","answer2":"Ship","answer3":"Space ship","answer4":"Bike","correct":"Car","path":"/static/pictures/1.jpg"}'
SET question:2 '{"question":"How many fingers a human has?","answer1":"More than 2","answer2":"About 8","answer3":"10","answer4":"42","correct":"10","path":"/static/pictures/2.jpg"}'
SET question:3 '{"question":"Who's Ken?","answer1":"Barbie'\''s boyfriend","answer2":"Marvel hero","answer3":"My hero","answer4":"A legend","correct":"Barbie'\''s boyfriend","path":"/static/pictures/3.jpg"}'
EOF
