# How to contribute

You can contribute by using the library, opening issues, or opening pull requests.

These guidelines reduce unvetted AI code, false positives, drive-by contributions, uncertain authorship, and risks.

## Security vulnerabilities

Unvetted AI-generated security reports and reproducers increase operational risks and can delay fixes for real bugs.

Before you report a vulnerability, have a human (not AI) confirm it matters in real use and the reproducer code is valid.  Some AI-detected bugs are triggered by AI tools using anti-patterns or improper settings that real apps should not use.

To report security vulnerabilities, please email faye.github@gmail.com and allow time for the problem to be resolved before disclosing it to the public.  For more info, see [Security Policy](https://github.com/fxamacker/cbor/blob/master/SECURITY.md).

Please don't send data containing personally identifiable information.  That type of support requires payment and a contract where the maintainer is indemnified, held harmless, and defended for any data you send.

## Issues

Unvetted AI-generated issues and false positives increase operational risks and can delay real improvements.  Issues that do not follow this guide and the [issue templates](https://github.com/fxamacker/cbor/issues/new/choose) may be closed.

Before you open an issue, have a human (not AI) confirm it matters in real use.  Some AI-detected bugs are triggered by AI tools using anti-patterns or improper settings that real apps should not use.

Some issues must not be disclosed publicly (e.g., security vulnerabilities not yet known to the public).  See [Security Policy](https://github.com/fxamacker/cbor/blob/master/SECURITY.md).

Clearly describe the issue:
* If it's a bug, first reproduce it with the latest release, then provide: **version of this library** and **Go** (`go version`), **unmodified error message**, **encoding/decoding options used**, and **a minimal program that reproduces it**.  Also state **what you expected to happen** instead of the error.
* If you propose a change or addition, explain which project needs it and why existing options or a workaround (e.g., a user-defined type implementing `cbor.Marshaler` or `cbor.Unmarshaler`) do not meet that need, and give an example of how the improved code could look or how to use it.
* If you found a compilation error, please confirm you're using a supported version of Go. If you are, then provide the output of `go version` first, followed by the complete error message.

## Pull requests

Unvetted AI-generated code, drive-by pull requests, and uncertain authorship increase risks.  Pull requests that do not follow this guide and the [Pull Request Template](https://github.com/fxamacker/cbor/blob/master/.github/pull_request_template.md) may be closed.

Before opening a PR, comment on the related issue saying you plan to work on it, and wait for a maintainer to reply that you can proceed. If no related issue exists, [create one](https://github.com/fxamacker/cbor/issues/new/choose) first. Approval given to someone else, or approval of the idea in general, does not extend to you.  No reply means not approved. PRs opened without this approval may be closed.

Pull requests have signing requirements ([DCO 1.1](https://developercertificate.org/)) and must not be anonymous.  Exceptions are usually made for docs and CI scripts.

AI-generated contributions are typically considered anonymous by this project because AI training data includes content from unidentified authors.  Exceptions are decided case by case at the maintainer's discretion.

Additionally, a pull request has a greater chance of being approved if:
- it implements the approach approved in the issue.
- it has no negative impact on speed, memory use, security, etc. for users who do not use the new feature.
- it fixes a bug that affects a project used in production, and the bug can be reproduced without anti-patterns or improper settings.
- it provides a feature that an open source project used in production needs now, and there is no workaround (e.g., implementing `cbor.Marshaler` or `cbor.Unmarshaler`).
- it does not reduce code coverage (currently > 97%).

## Special thanks

- @lukseven for pointing out in 2021 that the contribution guidelines didn't mention signing requirements.
