First let me say, Thank you very much for considering to contribute to this project! GoPlantUML can surelly benefit from the help of the community that is actually using the project.

**NOTICE**
This document is still a work in progress. Feel free to create issues regarding the modification of the process in general.

# Contributing
When contributing to this repository, please first discuss the change you wish to make via the issues feature in github before making any changes. Everyone is welcome to comment on the issues as well as to offer solutions to the problem that can help others pick it up and work on it.

Please note we have a [code of conduct](https://github.com/jfeliu007/goplantuml/blob/master/CODE_OF_CONDUCT.md "here"), please follow it in all your interactions with the project.

Any change to the project must be done through Pull Requests. Only the master branch will be present in this project unless it is necessary to introduce a new branch (e.g a new feature branch that needs multiple developers working on the same code base) . Tags will be cut from the master branch every time code is merged into master by using the following convention.
```v{major}.{minor}.{patch}```
- Major version numbers change whenever there is some significant change being introduced. For example, a large or potentially backward-incompatible change to a software package.
- Minor version numbers change when a new, minor feature is introduced or when a set of smaller features is rolled out.
- Patch numbers change when a new build of the software is released to customers. This is normally for small bug-fixes or the like.

# Development Process
Fork the project to start working on it. Always create a branch out of master to work on an issue. Since all work in this project is attached to issues, we suggest you name your branch after the issue you are working on, but this is not a requirement for development. Do your PRs always to the master branch.

## Pull Requests
- Every PR needs to have a title with information about what it is related to.
- PR description must include a brief description of the approach the developer followed to fix, or implement the feature.
- Part of the decription must include a reference to close the issue it relates to. See [Closing issues using keywords](https://help.github.com/en/articles/closing-issues-using-keywords "Closing issues using keywords") for more information. 

### Pull Request Acceptance Process
The following is a check list that is required for every Pull Request to be accepted.
- Make sure your code compiles for Golang version supported in the Readme file. If your code requires a new version of golang, please, specifie in the description the reason why it needed the new version and modify with your update the Readme file accordingly.
- Modify the Readme.md file accordingly for any changes in the command usage, or in changes on how the digrams generate (e.g changes in how classes are identified in the diagram)
- Provide 100% test coverage of the code for any new functionality or modification introduced in your PR.
- Maintain a go report card score of A+. Check  the Go Report Card [Here](https://goreportcard.com/report/github.com/jfeliu007/goplantuml "Go Report Card Here") for the specifics on how to maintain the score.
- Maintain the godoc if necessary.
- Make sure Travis builds still pass (This will happen automatically on the PR).
- PRs must be reviewed by at least on person before it can be accepted.
- Regenerate the ClassDiagram.puml by running ```./generate_diagram``` on the root of the project. This will keep our Diagram current.
- Make sure you have fun coding for the project ;-)

# Code of Conduct
Please, review the code of conduct [here](https://github.com/jfeliu007/goplantuml/blob/master/CODE_OF_CONDUCT.md "here").
