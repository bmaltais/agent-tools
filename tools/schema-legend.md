# schema-legend

Dense key legend for tools/registry.json.

Global keys:
- sv: schema semver
- sm: schema major (compat gate)
- ts: registry timestamp (UTC date)
- t: tool records array

Tool record keys:
- id: stable tool id
- bin: binary name
- st: lifecycle status (ga | beta | plan)
- in: intent token (dense)
- cmd: canonical invocation form
- a: args spec list
- o: outputs list
- p: preconditions list
- s: side-effects list
- f: failure modes list
- x: examples list

Args spec item keys:
- k: arg key
- r: required bool
- t: type token
- d: dense description

Overlay keys (tools/overlays/*.json):
- rt: runtime id
- sm: supported schema major
- m: runtime mode fields
- t: per-tool runtime adapter map
