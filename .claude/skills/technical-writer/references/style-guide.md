# Comprehensive Technical Writing Style Guide

This reference provides detailed guidance for creating professional, accessible, and user-friendly technical documentation.

## Content Model

### Content Structure

- **Hierarchical organization**: Top-level doc sets → Categories → Map topics → Articles
- **Category creation threshold**: 10+ articles required for new categories
- **Navigation depth**: Maximum 4 levels deep
- **Consistent content types**: Each with specific purposes and formats
- **Standard elements**: Every article includes titles, intros, permissions, etc.

## Article Structure and Standard Elements

### Titles

**Character limits:**
- Categories: 67 character limit
- Map topics: 63 character limit
- Articles: 80 character limit

**Formatting:**
- Use sentence case (not title case)
- Gerund titles for procedural content ("Creating...", "Configuring...")
- Avoid "How to" prefix unless required by product convention

**Examples:**

✅ Do:
- "Set up OAuth for a single-page app"
- "Creating a service principal" (procedural/gerund)

❌ Don't:
- "Setting Up OAuth for a Single-Page App" (title case)
- "How to set up OAuth" (avoid "How to")

### Intros

**Length requirements:**
- Articles: 1-2 sentences
- Categories/Map topics: 1 sentence

**Content:**
- Must explain what the content covers
- Be specific and actionable
- Avoid vague language

**Examples:**

✅ Do:
"This guide shows you how to enable OAuth for single-page apps using PKCE. You'll configure an app, create secrets, and test the flow."

❌ Don't:
"In this document, we will be covering OAuth for single-page applications in depth." (vague, long-winded)

### Permissions Statements

**Required for:** All tasks requiring specific permissions

**Format options:**

1. **Frontmatter** (preferred):
```yaml
---
required-permissions:
  role: Owner
  scope: Subscription
---
```

2. **Inline**:
   "You need the Owner role at the subscription scope to complete these steps."

**Consistency:** Use the same language patterns across documentation.

### Required Sections

When applicable, include these sections:

**Prerequisites:**
- Use structured list format
- Be specific about versions, tools, access levels

Example:
- An Azure subscription
- The Azure CLI 2.60.0 or later
- Owner role at the subscription scope

**Product callouts:**
"This feature is available on the Enterprise plan only."

**Next steps and Further reading:**
- Provide clear next actions
- Link to related content
- Use descriptive link text

## Content Types Detailed

### 1. Conceptual Content

**Purpose:** Explains what something is and why it's useful

**Characteristics:**
- "About..." article pattern
- Defines concepts and terminology
- Explains when to use and tradeoffs
- Provides context without step-by-step instructions

**Example:**
"About service principals" — defines the concept, when to use it, and tradeoffs.

### 2. Referential Content

**Purpose:** Detailed information for active use

**Characteristics:**
- Syntax guides with complete parameter lists
- Tables for complex data relationships
- Alphabetical or logical organization
- Comprehensive but scannable format

**When to use tables vs. lists:**
- **Tables**: Multiple related properties per item
- **Lists**: Simple, single-dimension information

**Structure headers appropriately:**
- Long reference content benefits from clear section breaks
- Use consistent heading hierarchy
- Enable quick scanning and search

**Example:**
"CLI reference for `az ad sp`" — flags, options, examples.

### 3. Procedural Content

**Purpose:** Step-by-step task completion

**Characteristics:**
- Numbered lists (sequential steps)
- Gerund titles ("Creating...", "Configuring...")
- One action per step
- Clear prerequisites upfront

**Step structure:**
1. Optional information (if applicable)
2. Reason for the step (if not obvious)
3. Location where to perform action
4. Action to take

**Example:**
"Creating a service principal" — numbered steps start-to-finish.

### 4. Troubleshooting Content

**Purpose:** Error resolution and known issues

**Types:**

**Known issues** (separate articles):
- Complex problems requiring detailed explanation
- Multiple potential causes
- Ongoing or unresolved issues

**General troubleshooting** (embedded):
- Common errors for specific tasks
- Quick fixes and workarounds
- Embedded in main task articles

**Structure:**
- Symptom description
- Cause explanation
- Resolution steps
- Prevention guidance

**Example:**
"Fix: AADSTS700016 error" — symptoms, cause, resolution.

### 5. Quickstart Content

**Purpose:** Essential steps for quick setup

**Critical requirements:**
- **Time limit**: 5 minutes / 600 words maximum
- **Specific intro phrasing**: Must clarify audience and scope
- **Audience clarification**: Required for complex products

**Structure:**
1. Before you begin (2-3 prerequisites)
2. Create resources (1-2 commands or clicks)
3. Run and verify (1 command or screenshot)
4. Clean up resources

**Intro pattern:**
"In this quickstart, you [accomplish X]. This guide is for [audience] who want to [goal]."

**Example:**
"Create and deploy your first function app" — 3-5 steps, done in 5 minutes.

### 6. Tutorial Content

**Purpose:** Detailed workflow guidance with real-world examples

**Requirements:**
- **Conversational tone**: Required for engagement
- **Real examples**: Must use actual, working examples (no placeholder code)
- **Conclusion structure**: Specific requirements for wrap-up

**Characteristics:**
- End-to-end scenario
- Explains "why" not just "how"
- Includes best practices
- Shows complete, working solution

**Conclusion must include:**
- Summary of what was accomplished
- Link to related concepts
- Suggested next steps
- Resources for going deeper

**Example:**
"Build a news summarizer with Functions and Cosmos DB" — end-to-end scenario.

### 7. Release Notes

**Purpose:** Version-specific changes and updates

**Requirements:**
- **Comprehensive formatting**: Specific categorization rules
- **Change categorization**: Features, fixes, improvements, breaking changes, etc.

**Categories (in order):**
1. Breaking Changes (if any)
2. Features (new capabilities)
3. Improvements (enhancements to existing features)
4. Fixes (bug fixes)
5. Documentation (doc updates)
6. Dependencies (version updates)

**Format:**
```markdown
## Version X.Y.Z - YYYY-MM-DD

### Breaking Changes
- Change description with migration path

### Features
- New feature description with brief example

### Improvements
- Enhancement description

### Fixes
- Bug fix description (can reference issue numbers)
```

**Example:**
"v1.8.0" — Features, fixes, breaking changes.

### Combining Content Types

**Guidelines:**
- Longer articles can combine multiple content types
- Follow content ordering guidelines
- Use clear transitions between sections
- Include troubleshooting information frequently

**Transition examples:**
- "Now that you understand the key concepts, follow these steps to enable the feature."
- "With the basics covered, let's explore advanced configurations."

## Content Ordering Guidelines

### Standard Content Sequence

1. **Conceptual** - What it is and why it's useful
2. **Referential** - Detailed information for active use
3. **Procedural** (in this specific order):
    - **Enabling** - Initial setup/activation
    - **Using** - Regular operational tasks
    - **Managing** - Administrative/configuration tasks
    - **Disabling** - Deactivation procedures
    - **Destructive** - Deletion/removal tasks
4. **Troubleshooting** - Error resolution

### Within Procedural Steps

Order information within each step:

1. **Optional information** (if applicable)
    - "If you already have X, skip this step."
2. **Reason** for the step (if not obvious)
    - "This keeps latency low."
3. **Location** where to perform the action
    - "In the Azure portal, go to Resource groups..."
4. **Action** to take
    - "...and select Create."

**Example:**
"Create a resource group in the same region as your app. This keeps latency low. In the Azure portal, go to Resource groups and select Create. Name it `my-rg` and select a region."

## Style Guide Key Points

### Language and Tone

**Principles:**
- Clear, simple, approachable language
- Active voice preferred (passive acceptable when appropriate)
- Sentence case for titles and headers
- Avoid jargon, idioms, regional phrases
- Write for global, multilingual audience

**Examples:**

✅ Do:
"Run the command and wait for the confirmation message."

❌ Don't:
"One could conceivably run said command prior to proceeding." (jargon, passive, unclear)

### Technical Writing Conventions

**Code formatting:**
- Use `backticks` for code, file names, commands, parameters
- Keep code blocks to ~60 character line length
- Avoid command prompts in examples (just show commands)
- Use syntax highlighting with proper language tags

**UI elements:**
- Format UI text in **bold** (buttons, menu items, fields)
- Use exact text from the interface
- Use > for navigation paths: **File > Save As**

**Placeholders:**
- ALL-CAPS with dashes: YOUR-PROJECT, YOUR-RESOURCE-GROUP
- Must explain placeholders before or after use
- Be consistent with placeholder naming

**Images:**
- Include alt text for all images (40-150 characters)
- Alt text describes what's shown, not just "screenshot"
- Reference images in text: "as shown in the following image"

**Examples:**

✅ Good:
- "Select **Create** and run `az group create --name MY-RG --location westus`."
- "Replace YOUR-RG with the name of your resource group."
- Alt text: "Screenshot of the Azure portal showing the Create resource group form with fields completed"

### Structure and Format

**Lists:**

**Numbered lists** (procedures):
- Capitalize first word
- Use periods for complete sentences
- One action per step
- Sequential order matters

**Bullet lists** (non-sequential):
- Use asterisks (*) not dashes
- Capitalize first word
- Alphabetize when appropriate
- Parallel structure

**Tables:**
- Use consistent header formatting
- Align content properly (left for text, right for numbers)
- Include header row
- Keep cells concise

**Alert Types** (use sparingly):

- **Note**: Neutral, supplementary information
    - "The CLI caches credentials for 1 hour."
- **Tip**: Helpful shortcuts or best practices
    - "Use tags to organize resources across subscriptions."
- **Important**: Critical information that affects outcomes
    - "Do not share client secrets. Rotate them every 90 days."
- **Warning**: Potential risks or consequences
    - "Deleting the resource group removes all resources in it."
- **Caution**: Actions that could cause data loss or damage
    - "Export data before disabling the feature to avoid loss."

### Links and References

**Internal links:**
- Use appropriate formatting for your doc system
- Link to specific sections when helpful
- Keep link text descriptive

**Link text:**
- Be descriptive (tell what user will find)
- Never use "click here" or "read more"
- Include context when needed

**External links:**
- Provide full context
- Explain why user should follow the link
- Consider link rot (prefer stable URLs)

**Examples:**

✅ Do:
- "See the authentication overview for background on OAuth 2.0."
- "For the OAuth 2.1 draft, see the IETF working group page."

❌ Don't:
- "Click here for more info."
- "Read more." (no context)

## Process and Workflow Elements

### Topics (Metadata)

**Character limits:**
- Keep topics concise
- Follow selection criteria for your platform

**Forbidden patterns:**
- Avoid duplicating title text
- Don't use as keyword stuffing

### Tool Switchers

**When to use:**
- Multiple interfaces for same task (Portal, CLI, PowerShell, Terraform)
- All methods achieve same outcome
- Each deserves equal treatment

**Format:**
- Use tabbed interface when supported
- Consistent ordering across docs
- Complete examples in each tab

**Example:**
"Portal | CLI | PowerShell | Terraform"

### Call to Action (CTA)

**Requirements:**
- Consistent formatting
- Clear, action-oriented language
- Specific next step

**Examples:**
- Next steps: "Next, secure your API keys with Key Vault."
- CTA button text: "Save", "Create", "Deploy" (not "Click to save")

### Reusables and Variables

**When to use reusables:**
- Content repeated verbatim across multiple articles
- Maintenance efficiency matters
- Content unlikely to need article-specific variations

**When to inline content:**
- Article-specific variations likely
- Better readability in source
- Single-use content

**Variable usage:**
- Define at document start or first use
- Be consistent with variable names
- Match placeholder conventions

**Examples:**
- "Set `YOUR-RG`, `YOUR-LOCATION`, and `YOUR-APP-NAME`."
- Reusable snippet: Common troubleshooting block for network timeouts

## Accessibility Requirements

### Alt Text

**Requirements:**
- 40-150 characters
- Describe what's shown, not "screenshot of..."
- Include relevant text from image
- Convey information, not aesthetics

**Examples:**
✅ "Azure portal showing the Create resource group form with fields completed"
❌ "Screenshot" (too vague)

### Structure

**Heading hierarchy:**
- Use proper heading levels (H1 → H2 → H3)
- Don't skip levels
- One H1 per page (the title)

**Reading order:**
- Content makes sense when read linearly
- Screen reader compatible

**Color and contrast:**
- Don't rely on color alone
- Ensure sufficient contrast
- Use patterns or labels in addition to color

### Links

**Descriptive link text:**
- Tells where link goes
- Makes sense out of context
- Avoid "click here"

### ARIA (when applicable)

- Use ARIA labels for complex interactions
- Ensure keyboard navigation works
- Test with screen readers

## Localization Considerations

**Write for translation:**
- Avoid idioms and colloquialisms
- Use simple sentence structure
- Be explicit (avoid implied information)
- Watch for cultural references

**Date and number formatting:**
- Use international formats
- Provide context for ambiguous formats
- Let localization system handle conversion

**Right-to-left (RTL) languages:**
- Ensure UI accommodates RTL
- Test with RTL languages
- Avoid directional language ("left sidebar")

**String externalization:**
- Use localization keys
- Avoid concatenating strings
- Plan for text expansion (some languages are longer)

## Quality Checklist

Use this checklist before publishing:

**Content:**
- [ ] Accurate and technically correct
- [ ] Appropriate content type for task
- [ ] Complete (no missing steps or information)
- [ ] Examples work as documented
- [ ] Tone appropriate for audience

**Structure:**
- [ ] Clear title (sentence case, within character limits)
- [ ] Brief intro (1-2 sentences)
- [ ] Prerequisites listed when needed
- [ ] Logical flow and organization
- [ ] Clear next steps

**Style:**
- [ ] Sentence case for headers
- [ ] Active voice (mostly)
- [ ] Code formatted correctly
- [ ] Placeholders explained
- [ ] Links descriptive and working

**Accessibility:**
- [ ] Alt text for all images
- [ ] Proper heading hierarchy
- [ ] Sufficient color contrast
- [ ] Keyboard accessible

**Polish:**
- [ ] Spell check passed
- [ ] Grammar correct
- [ ] Consistent terminology
- [ ] No jargon or explained when necessary

## Summary: Quick Decision Tree

**What type of content do I need?**

- Explaining a concept? → **Conceptual**
- Documenting API/syntax? → **Referential**
- How to do a task? → **Procedural**
- Fixing errors? → **Troubleshooting**
- Quick start (< 5 min)? → **Quickstart**
- Learning journey? → **Tutorial**
- Version changes? → **Release Notes**

**How should I structure it?**

1. Title (sentence case, descriptive)
2. Intro (1-2 sentences, specific)
3. Prerequisites (if needed)
4. Main content (type-appropriate)
5. Troubleshooting (when helpful)
6. Next steps

**What voice should I use?**

- Active, direct, clear
- Simple language
- Appropriate for global audience
- Professional but approachable

**Did I make it accessible?**

- Alt text on images
- Proper headings
- Descriptive links
- Sufficient contrast