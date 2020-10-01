# Specter 
[![CircleCI](https://circleci.com/gh/shipt/specter/tree/master.svg?style=svg&circle-token=00862a45c139d2a1a28c93ca104e38a7a6e3e83e)](https://circleci.com/gh/shipt/specter/tree/master)
![Specter Screenshot](/readmeFiles/SpecterScreenShot.gif)

Specter is an attack map style visualization that parses NGINX access logs. Specter then displays the source ip's location, the NGINX's ip's location, and the http status on a world map. Additionally allows filtering the traffic based on services. 

To read more about, how to run, and how to develop Specter, check out our [wiki](https://github.com/shipt/specter/wiki).

## Table of Contents
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
      - [Third Party](#third-party)
      - [Golang](#golang)
      - [Frontend](#frontend-web/public)
    - [Installing](#installing)
      - [Running Specter in Your Development Environment](#running-specter-in-your-development-environment)
    - [Running the Tests](#running-the-tests)
      - [What We Test](#what-we-test)
      - [Coding Style](#coding-style)
  - [Deployment](#deployment)
  - [Attributions](#attributions)
  - [Contributing](#contributing)
  - [Versioning](#versioning)
  - [Authors](#authors)
  - [License](#license)

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See [deployment](#deployment) for notes on how to deploy the project on a live system.

### Prerequisites
- #### Third Party
   - This project uses a MaxMind database to get geo data from IP addresses. You must download the [GeoLite2 City](https://dev.maxmind.com/geoip/geoip2/geolite2/) database and save it somewhere Specter can reach it.


- #### Golang
   - [Go version 1.11+](https://golang.org/)  
   - For go dependencies this project uses [dep](https://github.com/golang/dep), so ```dep ensure``` will install all needed dependencies. 

- #### Frontend (web/public)
   - [npm](https://www.npmjs.com) must be installed

```bash
# Move into web/public
cd web/public
# Just once 
npm install
# Watch for changes during development
npm run watch
# When ready to build/deploy
npm run build
```

### Installing

1. Clone this repo to your GOPATH.
2. From your local repository, run ```dep ensure```.
3. From your local repository, ```cd web/public``` and run ```npm install```.

#### Running Specter in Your Development Environment

1. From your local repository, start the Specter webserver:
```bash
   go run ./cmd/specter/main.go -db={{Where your GeoLite2-City database is}} -mbat={{Your MapBox Access Token}}
```
2. In a new console from your local repository, start the start the Specter data server:
```bash
   go run cmd/specter-data/main.go -log ./scripts/access.log
```
3. (Optional) From your local repository, start the start the load test:
```bash
   scripts/load.py
```

### Running the Tests

Your code editor should be setup to run tests on save. However, to run the tests manually, you can run ```go test ./...``` from the local repository directory to run all tests. To run just one test, run ```go test ./dir/package```.

- #### What We Test

   - The unit tests in this repo test our code, not code brought in through packages. 

      - For [example](internal/dataServer/dataServer.go#L86) in the dataServer package, we do not test the `tailFile` function since it only implements imported code.    
      - However, in the [same package](internal/dataServer/dataServer.go#L92), we do test the `processLog` function since it contains code that is untested elsewhere.

- #### Coding Style

   - Just use the default [revive](https://github.com/mgechev/revive) configuration.

## Building and Running Your Own Images

1. Ensure you have docker installed and working as expected.
2. Build Specter.Dockerfile:
   ```
   docker build -f Specter.Dockerfile .
   ```
3. Build the Specter-Data.Dockerfile:
   ```
   docker build -f Specter-Data.Dockerfile .
   ```
4. Start the Specter docker image:
   ```
   docker run -e DB=./db/GeoLite2-City.mmdb -e MBAT=<<YOUR_MAPBOX_API_TOKEN>> -v <<FOLDER_WHERE_YOUR_GEOLITE2_MMDB_EXISTS>>:/go/src/github.com/shipt/specter/db -p 1323:1323 <<YOUR_IMAGE_FROM_STEP_2>>
   ```
5. Start the Specter-Data docker image:
   ```
   docker run <<YOUR_IMAGE_FROM_STEP_3>>
   ```

Note: You will probably want to set some other ENV vars in order to get Specter working for you. The ENV vars can be found [here](https://github.com/shipt/specter/wiki/Running-Specter).

## Deployment

1. Download the appropriate version from the [Releases](https://github.com/shipt/specter/releases) page.
2. Deploy the Specter package to where you plan on running the Specter Webserver.  
    - Start Specter with the `-db={{Where your GeoLite2-City database is}}` `-mbat={{Your MapBox Access Token}}` flags.
3. Deploy the Specter-Data package to all NGINX servers you wish to monitor.
    - Start Specter-Data with the `-log ./scripts/access.log` flag and any other applicable flags.


## Attributions

Read the [Attributions](ATTRIBUTIONS.md) here. 

## Contributing

Please read our [CONTRIBUTING.md](./CONTRIBUTING.md) for details on our community guidelines and the process for submitting pull requests to us.

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/shipt/specter/tags).

To update versions, run the [provided python script](scripts/version.py) and follow the prompts.

## Maintainers

To find out who our Maintainers are, check out: [MAINTAINERS.md](MAINTAINERS.md)

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
