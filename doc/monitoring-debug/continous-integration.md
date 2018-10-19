# Continuous Intergration (CI)

We have [an instance of drone.io][1] to test the project each time a Pull Request is made on our Github repository.

Which tests are launched depends on the files changed (services modified).

## CI strategy

### Pull requests

When a PR is created or a new commit pushed to a PR, Drone captures the event `pull_request`. Depending on the files changed, Drone will launch a **lint check, unit tests and functional tests** in the case of the frontend (more info below), or unit tests for python. Go and Python tests are prepared to be launched based on the modifications made to the dependencies of each service, this means a service is only tested if any of its dependencies is modified. There are currently no GO tests and Python tests are global and not tied to a service, so the pipeline currently only tests python in case of a modification to the backend. Future improvements include individual service testing.

Those tests are made using the [merge and checkout strategy][2] so the tests will show the results as if the branch was be merged.

### Push

When a branch is merged or commit pushed to `develop`, the event is caught by drone. Every service **modified** is then built and published to the Docker registry. Each Docker image is published with 2 tags: develop and the commit sha. In the case of a merge, only services modified in the merged branch are built. In the case of a direct push, there is no way of checking which files are changed, so every service is built and published.

A push on the `master` branch does not trigger any test nor build.

### Tag

When a ``release-${version}`` tag is created on any branch, Drone will build and publish Docker images with 2 tags: latest and the version of the release. The number of the version is extracted from the tag, stripping the "release-" part.

## Frontend

Using a node container Drone launches the following tests and if one error is encountered it will stop the sequence and return a failure.

1. Lint JS
2. Lint SCSS
3. Unit tests JS
4. Functional tests

The functional tests release a bundle, start a mocked api (full JS using bouchon cf. devtools/api-mock) and open a socket on [Saucelabs](https://saucelabs.com) that will execute selenium commands on a real browser (Firefox 45/Linux for now). Connection to saucelabs is made through sauce-connect.
You can see the full video of the test by browsing open results on the project account: https://saucelabs.com/open_sauce/user/AllTheDey `> Automated tests > CaliOpen e2e - <name of the branch of the PR> > Watch or Metadata > Download video`.

[1]: https://drone.caliopen.org
[2]: https://github.com/drone-plugins/drone-git/blob/master/plugin.go#L49
