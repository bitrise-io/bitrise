

SLA for Verified Step Authors
=========================================================


**Verified Steps** are marked differently from Community Steps to
communicate towards Bitrise users that they can expect: *secure*,
*maintained*, *consistent*, *high-quality* Steps, which *follows the
Step Development Guideline* and the *underlying tool/service changes* so
that our users\' expectations can be met.

These badges appear on all interfaces we display steps (workflow editor,
[integrations page, ...)

------------------------------------------------------------------------

SLA
---

**First response time:** The time it can take the reporter is notified
about the type (`critical-bug`, `bug`, `feature-request`,
`maintenance`) and status (`accepted`,
`rejected`) of the Contribution.

**Resolution time:** The time it can take the accepted Contribution gets
closed.

The **type of contribution** needs to be marked by adding one of the
following labels:

-   `critical-bug`: the current feature set has abnormal behavior, which
    blocks users in use of the step (no workaround exists for the
    issue) - must be fixed by the author

-   `bug`: the current feature set has abnormal behavior, which does not
    block users in use of the step (workaround exists for the issue) -
    must be fixed by the author

-   `feature-request`: request for not yet existing feature for step -
    the author can decide whether the feature is worth to implement or
    not

-   `maintenance`: improvement on the step source code, which does not
    add new feature/fixes issue - the author can decide whether the
    feature is worth to implement or not

If a contribution is `rejected`, it needs to be closed within the First
response time.

`accepted` contribution means that the given:
`critical-bug`, `bug`, `feature`, `maintenance` will be fixed/merged,
within the given resolution time.



| **Type**          | **First response time** | **Resolution time** |
|-------------------|-------------------------|---------------------|
| critical-bug      |    5 business days      |    10 business days |
| bug               |    5 business days      |    15 business days |
| feature-request   |    5 business days      |    20 business days |
| maintenance       |    5 business days      |    20 business days |

**Labeling Contributions**

-   add contribution type label to Issues and Pull Requests
    (`critical-bug`, `bug`, `feature`, `maintenance`)

-   rejected contribution means that the given:

    -   `critical-bug`, `bug` is not an abnormal behavior

    -   `feature`, `maintenance` does not worth to implement

    -   any rejection needs to explain to the contributor

    -   any rejected contribution needs to be closed at the first
        response
