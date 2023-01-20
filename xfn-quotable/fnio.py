import typing


class ExplicitConnectionDetail(typing.TypedDict):
    name: str
    value: str


class ObservedComposite(typing.TypedDict):
    resource: dict
    connectionDetails: list[ExplicitConnectionDetail]


class ObservedResource(typing.TypedDict):
    name: str
    resource: dict
    connectionDetails: list[ExplicitConnectionDetail]


class Observed(typing.TypedDict):
    composite: ObservedComposite
    resources: list[ObservedResource]


class DesiredComposite(typing.TypedDict):
    resource: dict
    connectionDetails: list[ExplicitConnectionDetail]


class DesiredResource(typing.TypedDict):
    name: str
    resource: dict

    # TODO(negz): Define these as TypedDicts.
    connectionDetails: dict
    readinessChecks: dict


class Desired(typing.TypedDict):
    composite: DesiredComposite
    resources: list[DesiredResource]


class Result(typing.TypedDict):
    severity: str
    message: str


class FunctionIO(typing.TypedDict):
    """The I/O of a Crossplane Composition Function."""

    apiVersion: str
    kind: str

    config: dict
    observed: Observed
    desired: Desired
    results: list[Result]
