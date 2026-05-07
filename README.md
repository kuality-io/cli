# Kuality CLI

Test any website for accessibility, performance, SEO, cross-browser, stress testing, and 30+ QA dimensions — from your terminal.

## Install

**Homebrew** (macOS and Linux):

```bash
brew install kuality-io/tap/kuality
```

**Go**:

```bash
go install github.com/kuality-io/cli@latest
```

**Binary**: download from [Releases](https://github.com/kuality-io/cli/releases).

## Quick start

```bash
# Authenticate with your API key (get one at https://kuality.io/settings/api-keys)
kuality auth login

# Run a test
kuality test example.com

# Run a specific test type
kuality test example.com --type a11y

# Fail CI if high-severity findings exist
kuality test example.com --type a11y --fail-on high

# JSON output for scripting
kuality test example.com --format json

# JUnit XML for CI test reporting
kuality test example.com --format junit > results.xml
```

## Commands

| Command | Description |
|---------|-------------|
| `kuality test <url>` | Run a quality test |
| `kuality status <test-id>` | Check test status |
| `kuality reports list` | List recent reports |
| `kuality reports show <id>` | Show report details |
| `kuality targets` | List configured targets |
| `kuality score` | Show Kuality Scores |
| `kuality auth login` | Store API key |
| `kuality auth status` | Check auth status |
| `kuality auth logout` | Remove stored API key |

## Test types

37 test types across 8 categories:

| Category | Types |
|----------|-------|
| Core quality | `a11y`, `webvitals`, `seo`, `formaudit`, `brokenlinks`, `cookie`, `headers`, `jsaudit`, `tech`, `cms`, `api` |
| Cross-browser | `firefox`, `webkit` |
| Advanced UX | `uxaudit`, `animation`, `colorblind`, `assets`, `screenreader` |
| Performance | `performancebudget`, `assetaudit`, `bundlesize`, `ttfb`, `throttle`, `memoryleak` |
| Mobile | `touchaudit`, `touchsize`, `orientation`, `pwa`, `mobilelighthouse` |
| API testing | `contract`, `graphql`, `openapi` |
| Web compliance | `privacyscan`, `csp`, `cors` |
| Monitoring | `synthetic`, `cdnaudit` |

## CI/CD integration

### GitHub Actions

```yaml
- name: Quality test
  run: |
    kuality test ${{ vars.SITE_URL }} --type a11y --fail-on high --format junit > kuality-results.xml

- name: Upload results
  uses: actions/upload-artifact@v4
  with:
    name: kuality-results
    path: kuality-results.xml
```

### GitLab CI

```yaml
quality_test:
  script:
    - kuality test $SITE_URL --type a11y --fail-on high --format junit > kuality-results.xml
  artifacts:
    reports:
      junit: kuality-results.xml
```

### Generic CI

```bash
export KUALITY_API_KEY="your-key"
kuality test example.com --type a11y --fail-on high --quiet
```

Exit codes: `0` = pass, `1` = findings exceed `--fail-on` threshold.

## Authentication

Three methods, in order of precedence:

1. `--api-key` flag
2. `KUALITY_API_KEY` environment variable
3. `~/.kuality/config.yaml` (set via `kuality auth login`)

API keys start with `ku_` and are stored with `0600` file permissions.

## Global flags

| Flag | Short | Description |
|------|-------|-------------|
| `--api-key` | | API key override |
| `--format` | `-f` | Output format: `table`, `json`, `junit` |
| `--quiet` | `-q` | Suppress progress output |
| `--version` | `-v` | Show version |
| `--help` | `-h` | Show help |

## License

MIT. See [LICENSE](LICENSE).
