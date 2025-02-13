#!/usr/bin/env bash

set -o errexit -o nounset

check_prerequisites() {
    if [[ -n ${CI:-} && -z ${RD_LINT_SPELLING:-} ]]; then
        echo "Skipping spell checking in CI."
        exit
    fi

    case $(uname -s) in # BSD uname doesn't support long option `--kernel-name`
        Darwin) check_prerequisites_darwin;;
        Linux) check_prerequisites_linux;;
        CYGWIN*|MINGW*|MSYS*) check_prerequisites_windows;;
        *) printf "Prerequisites not checked on %s\n" "$(uname -s)" >&2 ;;
    esac
}

check_prerequisites_darwin() {
    if ! command -v cpanm &>/dev/null; then
        echo "Please install cpanminus first:" >&2
        if command -v brew &>/dev/null; then
            echo "brew install cpanminus" >&2
        fi
        exit 1
    fi
    # On macOS, the spell checker fails to skip expected long words.
    # Disable spell checking there until check-spelling releases v0.0.25.
    # https://github.com/check-spelling/check-spelling/issues/84
    echo "Skipping spell checking, macOS has false positives."
    exit
}

check_prerequisites_linux() {
    if command -v wslpath >&/dev/null; then
        check_prerequisites_windows
        return
    fi
    if [[ -z "${PERL5LIB:-}" ]]; then
        export PERL5LIB=$HOME/perl5/lib/perl5
    fi
    if command -v cpanm &>/dev/null; then
        return
    fi
    echo "Please install cpanminus first:" >&2
    if command -v zypper &>/dev/null; then
        echo "zypper install perl-App-cpanminus" >&2
    elif command -v apt &>/dev/null; then
        echo "apt install cpanminus" >&2
    fi
    exit 1
}

check_prerequisites_windows() {
    # cygwin, mingw, msys, or WSL2.
    echo "Skipping spell checking, Windows is not supported."
    exit
}

# Locate the spell checking script, cloning the GitHub repository if necessary.
find_script() {
    # Put the check-spelling files in `$PWD/resources/host/check-spelling`
    local checkout=$PWD/resources/host/check-spelling
    local script=$checkout/unknown-words.sh
    local repo=https://github.com/check-spelling/check-spelling
    local version
    version="v$(yq --exit-status .check-spelling pkg/rancher-desktop/assets/dependencies.yaml)"

    if [[ ! -d "$checkout" ]]; then
        git clone --branch "$version" --depth 1 "$repo" "$checkout" >&2
    else
        git -C "$checkout" fetch origin "$version" >&2
        git -C "$checkout" checkout "$version" >&2
    fi

    if [[ ! -x "$script" ]]; then
        printf "Failed to checkout check-spelling@%s: %s not found.\n" "$version" "$script" >&2
        exit 1
    fi

    echo "$script"
}

check_prerequisites
script=$(find_script)

INPUTS=$(yq --output-format=json <<EOF
    suppress_push_for_open_pull_request: 1
    checkout: true
    check_file_names: 1
    post_comment: 0
    use_magic_file: 1
    report-timing: 1
    warnings: bad-regex,binary-file,deprecated-feature,large-file,limited-references,no-newline-at-eof,noisy-file,non-alpha-in-dictionary,token-is-substring,unexpected-line-ending,whitespace-in-dictionary,minified-file,unsupported-configuration,no-files-to-check
    use_sarif: ${CI:-0}
    extra_dictionary_limit: 20
    extra_dictionaries:
        cspell:software-terms/dict/softwareTerms.txt
        cspell:k8s/dict/k8s.txt
        cspell:node/dict/node.txt
        cspell:aws/aws.txt
        cspell:golang/dict/go.txt
        cspell:php/dict/php.txt
        cspell:python/src/python/python-lib.txt
        cspell:typescript/dict/typescript.txt
        cspell:npm/dict/npm.txt
        cspell:shell/dict/shell-all-words.txt
        cspell:html/dict/html.txt
        cspell:filetypes/filetypes.txt
        cspell:fullstack/dict/fullstack.txt
        cspell:python/src/common/extra.txt
        cspell:java/src/java.txt
        cspell:dotnet/dict/dotnet.txt
        cspell:css/dict/css.txt
        cspell:django/dict/django.txt
        cspell:docker/src/docker-words.txt
        cspell:cpp/src/stdlib-cmath.txt
        cspell:python/src/python/python.txt
        cspell:powershell/dict/powershell.txt
EOF
)

export INPUTS

exec "$script"
