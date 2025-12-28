# Architecture

## Go

- Library code :-
    - `/` - The main library module (`github.com/gophero/guardian`), should expose needed `internal/**` as public APIs. For example - Function to create stores and services implented in `internal/**`.
    - `internal/**` - Contains implentation of services and stores, and other packages that should not be imported.
    - `core/**` - Mainly contain interfaces for all services and stores.

- App code :-
    - `cmd/guardian` - The main guardian binary.

- External packages :- These packages are currently kept inside this project for rapid development but they will be extracted later in there own module as (`github.com/gophero/{PACKAGE_NAME}`). They are kept in `pkg/*`
    - `pkg/bedrock` -
        - Application framework on top of which applications are built on.
        - It should not be imported by any library code.
        - Any packages inside with a config should use `kong` tags for configuration parsing and validation.
    - `pkg/migration` -
        - A factory interface for `github.com.golang-migrate/migrate.Migrate `
