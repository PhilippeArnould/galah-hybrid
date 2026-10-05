# Galah Hybrid

**Galah Hybrid** is an experimental extension of **Galah 1.1.1**, the
[LLM-powered web honeypot](https://github.com/0x4D31/galah) created by
Adel "0x4D31" Karimi.

This project explores a hybrid response-generation architecture that combines
deterministic HTML templates with dynamically generated LLM content. Its goal is
to keep page structure and visual identity consistent while allowing the honeypot
to adapt its content to incoming HTTP requests.

> This project is an experimental research prototype. Its implementation and
> configuration format may change.

## Why Galah Hybrid?

Galah can ask an LLM to generate a complete HTTP response, including headers and
body content. For HTML pages, that may involve generating both presentation and
dynamic content for each uncached request.

The hybrid approach addresses potential limitations of full-page generation:

- first-response latency;
- repeated generation of HTML and CSS;
- inconsistent appearance between pages;
- long outputs and increased token consumption;
- truncated or invalid structured responses.

Galah Hybrid separates the **presentation layer** from the **dynamic content
layer**. Templates define the page structure; the LLM generates the compact data
needed to populate them.

## Hybrid Architecture

For a request handled by a scenario, the intended generation flow is:

```text
Incoming HTTP request
        |
        v
Galah request handling and cache lookup
        |
        +-- Cache hit ----------------------> Cached HTTP response
        |
        v
Route / scenario matching
        |
        +-- HTML template and expected fields
        |
        v
LLM generates compact JSON data
        |
        v
JSON parsing and field validation
        |
        v
Local rendering with Go html/template
        |
        v
Final HTTP response, caching and logging
```

A template may contain fields such as:

```html
<h1>{{.Title}}</h1>
<p>Status: {{.Status}}</p>
<p>Priority: {{.Priority}}</p>
<p>{{.Description}}</p>
```

The LLM only needs to generate the corresponding data:

```json
{
  "Title": "Ticket #TN-5821",
  "Status": "In Progress",
  "Priority": "High",
  "Description": "Intermittent VPN authentication failures."
}
```

The final page is rendered locally. Requests that do not match a scenario route
can use the original Galah response-generation path, where fallback is enabled
by the implementation.

## Expected Benefits

The architecture aims to provide:

- lower generation latency and reduced output token usage;
- shorter, easier-to-validate structured responses;
- consistent visual identity across pages;
- reusable organization-specific scenarios;
- deterministic navigation and page structure;
- dynamic content adapted to incoming requests;
- reuse of Galah's caching and logging mechanisms.

These are design goals, not quantified performance guarantees. Actual results
depend on the model, hardware, prompt, template and cache hit rate.

## Organization Scenarios

A scenario describes a fictional organization and its simulated web environment.
Possible scenarios include technology companies, research laboratories,
universities, industrial organizations and training centers.

An example directory layout is:

```text
scenarios/
└── technova/
    ├── profile.yaml
    └── templates/
        ├── home.html
        ├── ticket.html
        ├── projects.html
        ├── kb.html
        └── generic.html
```

A scenario can define:

- organization identity, mission and terminology;
- canonical URLs and navigation;
- route-to-template mappings;
- HTML templates and shared visual conventions;
- expected dynamic fields;
- runtime prompts and content constraints.

Use fictional names and synthetic data throughout the scenario.

## Example Route Configuration

The following YAML illustrates the configuration proposed for the prototype.
Check the configuration loader and examples in your checkout before using it:
the keys and route-matching behavior are experimental.

```yaml
scenario:
  enabled: true
  name: technova
  templates_dir: /galah/scenarios/technova/templates

  routes:
    - pattern: "^/it-support(?:[-/].*)?$"
      template: ticket.html
      fields:
        - Title
        - TicketID
        - Status
        - Priority
        - Description

    - pattern: "^/kb(?:/.*)?$"
      template: kb.html
      fields:
        - Title
        - Heading
        - Summary
```

The template directory must exist inside the runtime environment. When running
in a container, mount or copy the scenario files to the configured path.

Template placeholders must match the dynamic field names. Validate generated
JSON and handle missing fields, malformed responses and rendering failures
before returning a page. Keep generated strings subject to `html/template`
escaping rather than treating them as trusted HTML.

## Tested Prototype Architecture

The development setup described for this prototype is:

```text
macOS host
   |
   +-- Ollama
   |     └── Qwen model
   |
   +-- UTM Linux virtual machine
         |
         └── T-Pot
               |
               └── Galah Hybrid
```

The LLM runs locally on the macOS host, while Galah Hybrid and the T-Pot stack run
inside the Linux virtual machine. Galah connects to the host's Ollama endpoint
over the network available to the VM.

This documents the reported development topology, not a compatibility test
matrix or a reproducible benchmark. Exact model versions, hardware and timing
measurements are not specified here. The host endpoint must be reachable from
the Galah runtime; `localhost` inside a VM or container refers to that environment.

## Experimental LLM Parameters

The proposed experimental settings favor compact dynamic responses:

| Parameter | Example value | Purpose |
| --- | --- | --- |
| Temperature | `0.6` | Moderate content variation |
| Context window | `4096` tokens | Bound the request and scenario context |
| Maximum generation | `384` tokens | Keep generated JSON short |
| Thinking | Disabled, where supported | Avoid unnecessary reasoning output |

These values are tuning examples, not a portable configuration block. Parameter
names and support depend on the provider, model and integration. In particular,
context size, generation limits and thinking controls are not necessarily
exposed by Galah's command-line interface.

Ask the model for a single JSON object containing only the required fields, with
no Markdown fences or surrounding commentary. Increase the generation limit if
the field set cannot fit reliably within the output budget.

Previously generated responses can be served from Galah's cache. When comparing
models or prompts, distinguish cache hits from newly generated responses.

## Scenario Generation with a Meta-Prompt

One objective is to generate complete scenarios from a higher-level
**meta-prompt**, then review and store the resulting assets before deployment.
This preparation step keeps runtime generation focused on dynamic content.

Example meta-prompt:

```text
Create a fictional organization scenario for an experimental web honeypot.

Organization type: technology company
Organization name: TechNova
Mission: managed IT services and internal technical support
Internal sections: home, IT support, projects, knowledge base
Visual style: restrained corporate interface with consistent navigation

Produce:
1. An organization profile using only fictional identities and synthetic data.
2. A list of canonical URLs and a route-to-template mapping.
3. HTML templates using Go html/template placeholders for dynamic fields.
4. The required JSON field names and example values for each template.
5. Runtime prompts that request only those fields as a single JSON object.

Keep navigation and organization terminology consistent across all pages.
Do not include real credentials, personal information, external tracking,
or executable server-side functionality.
Clearly label each output file and its intended location.
```

Review generated templates and configuration manually. Verify route patterns,
links, placeholder names and JSON examples before enabling a scenario. Scenario
generation is a workflow goal; this README does not imply that an automated
scenario-generation command is available.

## Compatibility

Galah Hybrid uses Galah 1.1.1 as its stated baseline and aims to preserve as much
of the original architecture as possible. The template engine is designed as an
additional generation path, with the original LLM response generation available
as a fallback.

Upstream Galah documents support for multiple LLM providers, response caching
and event logging. Provider support in upstream does not guarantee that every
provider or model has been tested with the hybrid path. Likewise, the development
setup above does not establish compatibility with every T-Pot release.

For installation and upstream behavior, refer to the
[original Galah documentation](https://github.com/0x4D31/galah). Use the files and
command-line help in this fork as the authority for its actual configuration.

## Project Status

This repository is a research and experimental project. Configuration formats,
scenario layouts and generation behavior may evolve as the approach is evaluated.

Useful evaluation criteria include uncached response latency, output token count,
valid JSON rate, template-rendering success, visual consistency and fallback
behavior. No numerical benchmark or production-readiness claim is made here.

## Credits

- **Original project:** [Galah — an LLM-powered web honeypot](https://github.com/0x4D31/galah)
- **Original author:** Adel "0x4D31" Karimi
- **Experimental extension:** [PhilippeArnould/galah-hybrid](https://github.com/PhilippeArnould/galah-hybrid)

Thanks to the original Galah author and contributors. Galah Hybrid is an
independent experimental derivative; it does not imply endorsement by the
upstream project or the developers of Ollama, Qwen, UTM or T-Pot.

## License

This project is derived from Galah and retains the **Apache License 2.0**.
See [LICENSE](LICENSE) for the full terms. Preserve applicable upstream copyright,
license and attribution notices when redistributing or modifying the software.

## Disclaimer

This software is intended for cybersecurity research, education and authorized
honeypot experimentation. Deploy it only in environments where you have
permission to do so.

Generated organizations, identities and systems should remain fictional. Do not
include real credentials, personal information or sensitive infrastructure data.
Treat incoming requests and model output as untrusted, and isolate the honeypot
from production systems.

LLM responses can be inaccurate, inconsistent or identifiable as synthetic.
Requests sent to an external LLM provider may contain attacker-supplied data;
review what is transmitted and logged. Set resource and API spending limits to
control abuse and unexpected costs.

The software is provided on an "AS IS" basis, without warranties or conditions,
as described in the Apache License 2.0.
