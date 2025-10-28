# Contributing to Cloudinary Go library

Contributions are welcome and greatly appreciated!

## Reporting a bug

- Make sure that the bug was not already reported by searching in GitHub under [Issues](https://github.com/cloudinary/cloudinary-go) and the Cloudinary [Support forms](https://support.cloudinary.com).
- If you're unable to find an open issue addressing the problem, [open a new one](https://github.com/cloudinary/cloudinary-go/issues/new/choose).
  Be sure to include a **title and clear description**, as relevant information as possible, and a **code sample**, or an **executable test case** demonstrating the expected behavior that is not occurring.
- If you require assistance in the implementation of cloudinary-go please [submit a request](https://support.cloudinary.com/hc/en-us/requests/new) on Cloudinary's site.

## Requesting a feature

We would love to receive your requests!
Please be aware that the library is used in a wide variety of environments and that some features may not be applicable to all users.

- Open a GitHub [issue](https://github.com/cloudinary/cloudinary-go) describing the benefits (and possible drawbacks) of the requested feature

## Fixing a bug / Implementing a new feature

- Follow the instructions detailed in [Code contribution](#code-contribution)
- Open a new GitHub pull request
- Ensure the PR description clearly describes the bug / feature. Include relevant issue number if applicable.
- Provide test code that covers the new code. See [TEST.md](TEST.md) for additional information about tests writing.

## Code contribution

When contributing code, either to fix a bug or to implement a new feature, please follow these guidelines:

#### Fork the Project

Fork [project on Github](https://github.com/cloudinary/cloudinary-go) and check your copy.

```
git clone https://github.com/contributor/cloudinary-go.git
cd cloudinary-go
git remote add upstream https://github.com/cloudinary/cloudinary-go.git
```

#### Create a Topic Branch

Make sure your fork is up-to-date and create a topic branch for your feature or bug fix.

```
git checkout main
git pull upstream main
git checkout -b feature/my-feature-branch
```
#### Rebase

If you've been working on a change for a while, rebase with upstream/main.

```
git fetch upstream
git rebase upstream/main
git push origin feature/my-feature-branch -f
```


#### Write Tests

Try to write a test that reproduces the problem you're trying to fix or describes a feature you would like to build.

We definitely appreciate pull requests that highlight or reproduce a problem, even without a fix.

See [TEST.md](TEST.md) for additional information about tests writing.

#### Write Code

Implement your feature or bug fix.
Follow the following Go coding standards, described in [Effective Go](https://golang.org/doc/effective_go) documentation.

Make sure that `go test ./...` completes without errors.

#### Write Documentation

Document any external behavior in the [README](README.md).

#### Commit Changes

Make sure git knows your name and email address:

```
git config --global user.name "Your Name"
git config --global user.email "contributor@example.com"
```

Writing good commit logs is important. A commit log should describe what changed and why.

```
git add ...
git commit
```


> Please squash your commits into a single commit when appropriate. This simplifies future cherry-picks and keeps the git log clean.

#### Push

```
git push origin feature/my-feature-branch
```

#### Make a Pull Request

Go to https://github.com/cloudinary/cloudinary-go/pulls. Click the 'New pull Request' button and fill out the form. Pull requests are normally reviewed within a few days.
Ensure the PR description clearly describes the problem and solution. Include relevant issue number if applicable.

#### Rebase

If you've been working on a change for a while, rebase with upstream/main.

```
git fetch upstream
git rebase upstream/main
git push origin feature/my-feature-branch -f
```

#### Check on Your Pull Request

Go back to your pull request after a few minutes and see whether it passed muster with Travis-CI. Everything should look green, otherwise - fix issues and amend your commit as described above.

#### Be Patient

It's likely that your change will not be merged and that the nitpicky maintainers will ask you to do more, or fix seemingly benign problems. Hang on there!

#### Thank You

Please do know that we really appreciate and value your time and work. We love you, really.
