import sys

import requests
import yaml

import fnio

ANNOTATION_KEY_AUTHOR = "quotable.io/author"
ANNOTATION_KEY_QUOTE = "quotable.io/quote"


def get_quote() -> tuple[str, str]:
    """Get a quote from quotable.io"""
    rsp = requests.get("https://api.quotable.io/random")
    rsp.raise_for_status()
    j = rsp.json()
    return (j["author"], j["content"])


def read_functionio() -> fnio.FunctionIO:
    """Read the FunctionIO from stdin."""
    return yaml.load(sys.stdin.read(), yaml.Loader)


def write_functionio(f: fnio.FunctionIO):
    """Write the FunctionIO to stdout and exit."""
    sys.stdout.write(yaml.dump(f))
    sys.exit(0)


def result_warning(f: fnio.FunctionIO, message: str):
    """Add a warning result."""
    if "results" not in f:
        f["results"] = []

    f["results"].append(fnio.Result(severity="Warning", message=message))


def main():
    """Annotate all desired composed resources with a quote from quotable.io"""

    try:
        functionio = read_functionio()
    except yaml.parser.ParserError as err:
        sys.stdout.write("cannot parse FunctionIO: {}\n".format(err))
        sys.exit(1)

    # Return early if there are no desired resources to annotate.
    if "desired" not in functionio or "resources" not in functionio["desired"]:
        write_functionio(functionio)

    # If we can't get our quote, add a warning and return early.
    try:
        quote, author = get_quote()
    except requests.exceptions.RequestException as err:
        result_warning(functionio, "Cannot get quote: {}".format(err))
        write_functionio(functionio)

    # Annotate all desired resources with our quote.
    for r in functionio["desired"]["resources"]:
        if "resource" not in r:
            # This shouldn't happen - add a warning and continue.
            result_warning(
                functionio,
                "Desired resource {name} missing resource body".format(
                    name=r.get("name", "unknown")
                ),
            )
            continue

        if "metadata" not in r["resource"]:
            r["resource"]["metadata"] = {}

        if "annotations" not in r["resource"]["metadata"]:
            r["resource"]["metadata"]["annotations"] = {}

        r["resource"]["metadata"]["annotations"][ANNOTATION_KEY_AUTHOR] = author
        r["resource"]["metadata"]["annotations"][ANNOTATION_KEY_QUOTE] = quote

    write_functionio(functionio)


if __name__ == "__main__":
    main()
