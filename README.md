# Kahoot port in go
~~Quizaara (pron: Kwee-ZAH-rah) (mean: Quiz for Sarah)~~

This is a quick port from my nextjs version to go+htmx.

Demo link master: kahoot.mrdima98.dev/lobby

Demo link player: kahoot.mrdima98.dev/lobby

## How to run

Make sure you have a redis server running on the default port.

Using redis-cli copy paste this in:

```redis
SET 'questions' '[ {"question":"What is this?","answer1":"Car","answer2":"Ship","answer3":"Space ship","answer4":"Bike","correct":"Car","path":"/static/pictures/1.jpg"}, {"question":"How many fingers a human has?","answer1":"More than 2","answer2":"About 8","answer3":"10","answer4":"42","correct":"10","path":"/static/pictures/2.jpg"}, {"question":"Who\'s Ken?","answer1":"Barbie\'s boyfriend","answer2":"Marvel hero","answer3":"My hero","answer4":"A legend","correct":"Barbie\'s boyfriend","path":"/static/pictures/3.jpg"} ]'
```

This will load some question otherwise stuff will crash.

Using golang:

```shell
go run .
```

Using docker:

```shell
docker build --tag kahoot .
docker run kahoot
```

Downloading the image:

```shell
docker image pull mrdima98/kahoot
docker run mrdima98/kahoot
```

## How to reuse
You will have to manually add images in static and craft your question into a stringify json.
You can also swap the background images or the paws for the questions.

No further improvements will happen in this project because I'm tired of it.

