
## Todo List

### API
[x] Loader
[x] Templates
[x] Context (Provide, Inject, ContextProvider)
[x] Action
[x] RouterProvider
[x] Outlets
[x] Layouts

### V3 Features

[] Built-in HTML Layout API
    - Provide scripts, styles, assets
[] Mixins
    - Mixins provide layouts, scripts, assets
    - Encapsulates integrations like v8go, tailwind
[] v8go Experiment
    - Benchmarked flow 
    - React JSX?
    - esbuild pipeline
[] GitHub Pipeline
    - build/test
    - benchmarks comparisons
[] eslint-plugin-torque
    - eslint plugin for linting string literals in Go
[] Claude Code Plugin
    - Embedded documentation
    - Examples and Recipes
    - Go Templates
    - Evals

### Operations

[ ] Skills
  [ ] go-templates
  [ ] prototype
  [x] tdd
  [x] benchmark
  [x] technical-writer

[x] Docs Initial Draft
[ ] Agent.md
[ ] Torque Claude Plugin 
    [ ] torque skill
        - writing a template analyzer
        - writing handler tests
        - create different "design style" skills for claude's code gen
    [ ] recipes
    [ ] evals


## Ideas

- Benchmarking over time. Benchmark test results could be recorded on merge
  - Merge requests can compare/contrast the benchmark tests

- Routers can provide additional 'global' templates that are loaded. Think icon sheets