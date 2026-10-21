Third-party notices are maintained in `public/licenses/THIRD-PARTY-NOTICES.txt`.
The desktop licenses modal imports that file as text, and Vite also copies it
to `dist/licenses/THIRD-PARTY-NOTICES.txt`. Ordinary builds do not regenerate it.

After changing dependencies, the Go version in `go.mod`, bundled fonts, or licenses:

```sh
cd frontend
yarn install --frozen-lockfile --ignore-platform
yarn licenses:generate
yarn licenses:check
```

Commit the generated public license files with the dependency change. CI on Linux,
macOS, and Windows regenerates the expected content in memory and fails if the
committed files are stale. The check does not rewrite the notices.

The generator collects a union of production Go dependencies for macOS, Windows,
and Linux on amd64 and arm64, with CGO enabled. `go list` inspects dependencies;
it does not compile or require the other platforms' native SDKs. Go's toolchain
is selected from the exact `go` version in `go.mod`, regardless of the host's
default toolchain or build target. Generating notices may download that toolchain
and the required Go modules.

Frontend runtime dependencies and their transitive dependencies are read from
the installed packages. Use the committed Yarn lockfile and `--ignore-platform`
so platform-specific packages are present. Missing dependencies fail generation
instead of silently producing incomplete notices. The sibling web-adapters
package is explicitly excluded, as before; the Wails runtime is included.

Output uses LF line endings, removes trailing spaces, and sorts names and versions
without locale-dependent comparisons. The shared file intentionally includes
notices for other supported desktop platforms as well as the current platform.
