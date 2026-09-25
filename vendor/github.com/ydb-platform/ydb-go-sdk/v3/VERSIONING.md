# YDB-Go-SDK Versioning Policy

By adhering to these guidelines and exceptions, we aim to provide a stable and reliable development experience for our users (aka [LTS](https://en.wikipedia.org/wiki/Long-term_support)) while still allowing for innovation and improvement.

We endeavor to adhere to versioning guidelines as defined by [SemVer2.0.0](https://semver.org/).

We making the following exceptions to those guidelines:
## Experimental
   - We use the `// Experimental` comment for new features in the `ydb-go-sdk`. 
   - Early adopters of newest feature can report bugs and imperfections in functionality. 
   - For fix this issues we can make broken changes in experimental API. 
   - We reserve the right to remove or modify these experimental features at any time without increase of major part of version.
   - We want to make experimental API as stable in the future
## Deprecated
   - We use the `// Deprecated` comment for deprecated features in the `ydb-go-sdk`.
   - Usage of some entity marked with `// Deprecated` can be detected with linters such as [check-deprecated](https://github.com/black-06/check-deprecated), [staticcheck](https://github.com/dominikh/go-tools/tree/master/cmd/staticcheck) or [go-critic](https://github.com/go-critic/go-critic).
   - This helps to our users to soft decline to use the deprecated feature without any impact on their code.
   - Deprecated features will not be removed or changed for a minimum period of **six months** since the mark added.
   - We reserve the right to remove or modify these deprecated features without increase of major part of version.
## Internals
   - Some public API of `ydb-go-sdk` relate to the internals.
   - We use the `// Internals` comment for public internals in the `ydb-go-sdk`.
   - `ydb-go-sdk` internals can be changed at any time without increase of major part of version.
   - Internals will never marked as stable
   - `testutil` package can be changed at any time without increase of major part of version.
