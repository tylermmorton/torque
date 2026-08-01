---
name: technical-writer
description: This skill should be used when the user asks to "write a README", "create API documentation", "draft release notes", "write a tutorial", "structure documentation", or "review docs for clarity". Automatically activates when working with .md files in docs/ directories, README files, or when discussing documentation structure, style guides, or content organization. Not for creative/marketing writing, academic papers, code comments/docstrings, or internal chat.
---

# Technical Documentation Expert

## Overview

Expert guidance for creating clear, comprehensive, and user-friendly technical documentation following industry best practices and structured content models.

**Core principle:** Write for the audience with clarity, accessibility, and actionable content using standardized documentation patterns.

## When to Use

Automatically activates when:

- Working with `.md` files in `docs/` directories
- Creating or editing README files
- Editing documentation files or markup
- Discussing documentation structure, information architecture, or style guides
- Creating API documentation with examples and parameter tables
- Writing user guides, tutorials, or quickstarts
- Drafting release notes or change logs
- Structuring specifications or technical proposals

**Style reference:** For detailed formatting rules, consult `./references/style-guide.md`.

## When NOT to Use

Do not apply this skill for:

- Creative writing or marketing copy
- Code implementation (documentation only)
- Project management documentation
- Internal team chat or informal notes
- Academic papers or research documentation

## Core Expertise Areas

### Content Types

1. **Conceptual** -- Explains what something is and why it matters ("About..." articles)
2. **Referential** -- Detailed reference information (API docs, syntax guides, parameter tables)
3. **Procedural** -- Step-by-step task completion with numbered lists and gerund titles
4. **Troubleshooting** -- Error resolution, known issues, and debugging guidance
5. **Quickstart** -- Essential setup in 5 minutes / 600 words maximum
6. **Tutorial** -- End-to-end workflow with real-world examples and conversational tone
7. **Release Notes** -- Version changes categorized by type (Features, Fixes, Breaking Changes)

### Documentation Structure

**Standard Article Elements:**

- Titles: Sentence case, gerund for procedures, character limits by level
- Intros: 1-2 sentences explaining content
- Prerequisites and permissions when applicable
- Clear next steps

**Information Architecture:**

- Hierarchical structure with maximum 4 navigation levels
- Content ordering: Conceptual -> Referential -> Procedural -> Troubleshooting

### Style Guide Principles

Apply formatting and style rules from the comprehensive style guide:

- **Language:** Clear, simple, active voice, sentence case. Avoid jargon without definition. Prefer short sentences (under 25 words). Eliminate filler words ("just", "simply", "basically").
- **Technical formatting:** Code in backticks, UI elements in bold, placeholders in ALL-CAPS with descriptive names (e.g., `YOUR_API_KEY`).
- **Structure:** Numbered lists for procedures (sequential steps), bullets for non-sequential information. Limit list items to 7 or fewer when possible.
- **Links:** Descriptive text that tells the reader where the link leads. Never use "click here" or bare URLs in prose.
- **Alerts:** Use Note, Tip, Important, Warning, Caution sparingly. Reserve Warning and Caution for data loss or security risks.
- **Code examples:** Include runnable examples with expected output. Annotate non-obvious lines with inline comments. Verify all examples compile and run against the documented version.

### Procedural Content Ordering

Follow standard procedural sequence: Enabling -> Using -> Managing -> Disabling -> Destructive

Within individual steps: Optional info -> Reason -> Location -> Action

### Writing Procedures

When writing step-by-step procedures:

1. Start each step with an action verb (imperative form)
2. Include only one action per step
3. Provide expected results after each significant step
4. Add screenshots or output examples for complex steps
5. Keep the total number of steps under 10 when possible; break longer procedures into sub-procedures
6. Place optional steps clearly marked with "Optionally" at the start
7. Include a "Before starting" or "Prerequisites" section listing required tools, permissions, and knowledge

## Development Workflow

### 1. Understand the Audience

- Identify user expertise level (beginner, intermediate, advanced)
- Determine user goals and tasks
- Consider context where documentation will be consumed
- Plan appropriate content depth and technical level
- Match vocabulary to the audience (avoid over-simplifying for experts; avoid under-explaining for beginners)

### 2. Choose Content Type

Select the appropriate content type based on user intent:

| User need | Content type | Key characteristics |
|-----------|-------------|-------------------|
| Understand a concept | Conceptual | "About..." title, explains what and why |
| Look up API or syntax | Referential | Parameter tables, return types, examples |
| Complete a specific task | Procedural | Numbered steps, gerund title, prerequisites |
| Fix a problem | Troubleshooting | Symptom-cause-fix tables, error messages |
| Get started quickly | Quickstart | Under 5 minutes, under 600 words |
| Learn end-to-end | Tutorial | Real-world example, conversational tone |
| Review version changes | Release Notes | Categorized by type, links to details |

### 3. Structure Content

**Standard content sequence:**

1. Title (sentence case, descriptive, within character limits)
2. Brief intro (1-2 sentences summarizing what the reader will learn or accomplish)
3. Prerequisites (if applicable)
4. Permissions statement (if required)
5. Main content (ordered appropriately by content type)
6. Troubleshooting (embedded when helpful, or linked to dedicated troubleshooting doc)
7. Next steps / Further reading (link to related content)

### 4. Apply Style Guide

Follow the comprehensive style guide for:

- Formatting code, UI elements, and placeholders
- Writing clear procedures with proper structure
- Adding accessibility features (alt text for images, sufficient color contrast)
- Ensuring proper link formatting and context
- Using alerts appropriately and sparingly
- Verifying content accuracy: do not invent information not in source material

### 5. Content Accuracy

**Critical rule:** Do not invent or assume information not present in source material.

- If gaps exist, ask the user for missing information
- Do not create placeholder or speculative content
- Verify technical accuracy with authoritative sources
- Include working examples whenever possible
- Check that examples work as documented
- Validate accessibility (alt text, heading hierarchy, structure)

### 6. Review and Iterate

After drafting:

- Read through the document from the reader's perspective
- Verify all links resolve to valid targets
- Confirm code examples are complete and runnable
- Check heading hierarchy (no skipped levels)
- Ensure consistent terminology throughout
- Validate that prerequisites actually cover what the reader needs

## Communication Style

**Clear and Actionable:**

- Use simple, direct language in imperative form
- Provide specific examples and code snippets
- Break complex topics into digestible sections
- Include visual aids when they clarify concepts

**Serve Multiple Expertise Levels:**

- Layer content from simple to complex
- Provide quick reference sections for experienced users
- Link to deeper explanations for beginners
- Set expectations with prerequisites

**Focus on User Goals:**

- Organize by tasks users want to accomplish
- Use gerund titles for procedures ("Creating...", "Configuring...")
- Include "what this covers" or "what to expect" statements
- Provide clear next steps after each article

## Documentation Anti-patterns

Avoid these common mistakes:

| Anti-pattern | Fix |
|-------------|-----|
| Assuming reader context ("As you know...") | State prerequisites explicitly |
| Burying critical info in long paragraphs | Lead with the most important information |
| Writing procedures without numbered steps | Always use numbered lists for sequential tasks |
| Using jargon without definition | Define terms on first use or link to glossary |
| Missing prerequisites section | List what the reader needs before starting |
| "Click here" link text | Use descriptive text that tells where the link goes |
| Outdated code examples | Verify all examples work with current versions |
| Mixing content types in one document | Separate conceptual from procedural content |
| Walls of text without headings | Add headings every 2-4 paragraphs |

## Success Criteria

Documentation is successful when:

- Content is accessible to the target audience
- Structure follows the appropriate content type
- Examples clarify complex concepts and are verifiably correct
- Style guide rules are consistently applied
- Users can complete tasks using the documentation alone
- Information architecture supports easy navigation
- Content is accurate, up-to-date, and free of speculation
- Heading hierarchy is correct (no skipped levels)
- All links resolve to valid targets