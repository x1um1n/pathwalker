# pathwalker
pathwalker is a restful webservice written in Golang that allows users to submit reports and images of the state of public footpaths.  

This should be easily integrated into the website of the council responsible for the paths.

## Endpoints
 - /v1/api/add
 - /v1/api/delete/{survey-id}
 - /v1/api/fetch/survey/{survey-id}
 - /v1/api/list/{path-id}
 - /v1/api/update/{survey-id}
 - /v1/api/upload

## Installation
The included docker-compose.yml & Dockerfile can be used to build a container image that can either be hosted on a cluster, such as DockerSwarm or Kubernetes; or run on a local Docker instance for testing.

To build & start a local instance, simply run these two commands from the project directory:
```
docker-compose build
docker-compose up
```

The docker-compose.yml includes a mysql container to run alongside the app container.  This can be used as a persistent database if desired, but it is primarily intended for testing locally.

#### Configuration
**To Do** KOANF_ENVIRONMENT_envname

## Usage
The service is designed to be embedded into an existing website, you will need to build a page to accept new surveys and handle authentication of users, if desired.

To add a survey you will need to POST to two endpoints: /v1/api/upload to upload any attached images and /v1/api/add to send the survey in JSON.  The upload endpoint will return a comma-separated-list of image IDs to be included with the survey.

```
{
    "path-id": "PATH-ID",
    "survey-date": "YYYY-MM-DD",
    "survey-submitted-by": "user@example.com",
    "detail": "The meat of the survey, this is being inserted into a mysql TEXT field, which has a limit of 65,535 characters"
    "image-ids": "a-series-of-guids,returned-by-image-upload"
}
```

A (very basic) example image upload form can be found in the examples directory.
