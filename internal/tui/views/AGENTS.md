## TUI View Design Sources (Progressive Disclosure)

When creating or updating a view in this directory, read `@design/` artifacts in this order and at the specified time:

1. Before planning the view scope:
`@design/UX-DESIGN-PLAN.md`, `@design/workflows.yaml`, `@design/flows.yaml`

2. Before wiring navigation/state entry points:
`@design/screens.yaml`

3. Before writing the view layout:
`@design/views.yaml`

4. Before composing reusable UI pieces and styling:
`@design/components.yaml`, `@design/tokens.yaml`, `@design/config.yaml`, `@design/paradigm.yaml`

5. Immediately before and during implementation of the target view:
`@design/mocks/<view>.blueprint.md`

6. During final visual pass and before marking complete:
`@design/mocks/<view>.mock.html`

`/design/mocks` blueprints are mandatory. The matching `*.blueprint.md` and `*.mock.html` pair is the final source of truth for:
- layout and composition
- spacing and visual hierarchy
- component structure and interaction intent
