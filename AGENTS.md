# AGENTS.md

## 1. Project Overview

This project is an **AI-assisted novel-to-screenplay tool**.

The goal is to help novel authors convert their works into an editable screenplay draft with lower effort and higher efficiency.

Core requirement:

* Accept novel text containing **3 or more chapters**.
* Automatically convert the novel into a structured screenplay draft.
* Output the screenplay in **YAML format**.
* Provide an additional document defining the screenplay YAML Schema.
* The Schema document must explain the design rationale behind the fields.

This is a hackathon / competition project. The priority is to build a stable, demonstrable MVP first, then improve product experience.

---

## 2. Core Product Goal

The product should support this flow:

```text
Novel text with 3+ chapters
-> chapter parsing
-> story understanding
-> character extraction
-> scene splitting
-> screenplay YAML generation
-> YAML validation
-> readable screenplay preview
-> YAML export/download
```

The generated YAML should be:

* Structured
* Editable
* Validatable
* Traceable back to source chapters
* Useful as a screenplay draft for further human polishing

The project should not be described as a fully production-ready screenplay system. It is an MVP / prototype focused on structured conversion and editing support.

---

## 3. Hard Requirements

The following requirements are mandatory:

1. The user must be able to input or upload novel text containing at least 3 chapters.
2. The system must convert the novel text into a structured screenplay draft.
3. The screenplay draft must be output in YAML format.
4. The YAML output must contain scenes, characters, actions, dialogue, and source chapter references.
5. The system must provide a YAML Schema design document.
6. The Schema document must explain why the fields are designed this way.
7. The output should be easy for authors to edit and further polish.

If a task conflicts with these hard requirements, prioritize the hard requirements.

---

## 4. MVP Priorities

### P0 - Must Have

These features are required for the project to be considered valid:

* Novel input supporting 3 or more chapters
* Chapter parsing or chapter segmentation
* AI generation of screenplay YAML
* YAML Schema design document
* Basic YAML display
* Basic screenplay preview
* YAML copy or download

### P1 - Core Quality

These features make the MVP reliable:

* Chapter count validation
* Character extraction
* Key event extraction
* Scene splitting
* Action and dialogue generation
* YAML parsing and validation
* Error handling for invalid AI output
* Example YAML file for demo fallback

### P2 - Product Polish

These features are valuable but should not block P0/P1:

* Visual scene editor
* Style selection, such as faithful adaptation, web drama, movie script, short drama
* Multi-version generation
* Source traceability to paragraph ranges
* Character consistency check
* Markdown / DOCX export
* Better UI animations and loading states

### P3 - Do Not Prioritize Early

Avoid these unless all P0/P1 tasks are stable:

* User login system
* Payment system
* Multi-user collaboration
* Complex drag-and-drop editor
* Large database system
* Voiceover generation
* Poster generation
* Full storyboard generation
* Large-scale agent architecture

---

## 5. Development Rules

Every AI coding agent must follow these rules:

1. Every task must be developed in a separate branch.
2. Every feature must be submitted as a separate pull request.
3. Do not push directly to `main`.
4. Do not merge pull requests.
5. Do not combine multiple features into one PR.
6. Do not make large unrelated refactors.
7. Do not modify files outside the task scope unless explicitly required.
8. Do not introduce complex dependencies unless the task requires them.
9. Do not add authentication, payment, database, or user system unless explicitly assigned.
10. Do not commit API keys, tokens, `.env` files, credentials, or secrets.
11. Do not claim incomplete features are finished.
12. Do not remove existing tests unless the task explicitly asks for it.
13. Do not silently change public API contracts.
14. If a required command cannot run, explain why in the final report and PR description.
15. If the task is unclear, implement the smallest safe version and document assumptions.
16. Keep commits as small as possible. Each commit should represent one focused step, such as one test group, one endpoint, one validation rule, or one small refactor.
17. Avoid commits with hundreds of changed lines. If a task naturally grows large, split it into multiple small commits before pushing.
18. Prefer TDD-sized commits when coding: failing test first, minimal implementation second, cleanup or comments third.

---

## 6. PR Rules

Each PR must solve exactly one clear problem.

Each PR may contain multiple commits, but every commit should stay small and reviewable. Do not squash unrelated steps into one large commit. A reviewer should be able to understand each commit independently.

A PR must include:

```text
## What changed

## How to test

## Risk

## Notes / TODO
```

Commit message format:

```text
type(scope): summary
```

Examples:

```text
docs(schema): add YAML schema design document
chore(backend): add backend scaffold
feat(parser): add chapter parser
feat(generator): add screenplay YAML prompt
feat(validator): add YAML validation
feat(frontend): add script generation page
test(parser): add chapter parser tests
fix(generator): handle invalid YAML output
```

Suggested branch naming:

```text
docs/yaml-schema-design
docs/task-board
chore/backend-scaffold
feat/chapter-parser
feat/script-generation-prompt
feat/script-yaml-generator
feat/yaml-validator
feat/frontend-script-page
feat/script-preview
feat/download-yaml
test/poc-smoke-script
ci/basic-checks
```

---

## 7. Repository Structure

Recommended structure:

```text
.
├── AGENTS.md
├── TASK_BOARD.md
├── README.md
├── docs/
│   ├── yaml-schema-design.md
│   ├── prompt-design.md
│   └── examples/
│       ├── script-example.yaml
│       └── novel-example.md
├── backend/
│   ├── src/
│   │   ├── parser/
│   │   ├── ai/
│   │   │   └── prompts/
│   │   ├── generator/
│   │   ├── validator/
│   │   ├── routes/
│   │   └── tests/
│   └── package.json
├── frontend/
│   ├── src/
│   └── package.json
└── scripts/
    └── poc-smoke.sh
```

If the current repository structure is different, follow the existing structure and avoid unnecessary restructuring.

---

## 8. Core YAML Schema Direction

The screenplay YAML should roughly follow this structure:

```yaml
script:
  title: ""
  logline: ""
  genre: ""

  source_chapters:
    - index: 1
      title: ""
      summary: ""

  characters:
    - name: ""
      role: ""
      description: ""
      first_appearance: ""

  scenes:
    - scene_id: "S001"
      source_chapter: 1
      title: ""
      location: ""
      time: ""
      characters:
        - ""
      summary: ""
      action:
        - ""
      dialogue:
        - speaker: ""
          line: ""
      notes: ""
```

Required design principles:

1. `source_chapters` preserves the relationship between the screenplay and the original novel.
2. `characters` helps authors maintain character consistency.
3. `scenes` are the core unit of screenplay editing.
4. `action` converts narration into performable visual action.
5. `dialogue` converts conversations into script-friendly dialogue.
6. `notes` allows AI or human editors to explain adaptation choices.
7. `scene_id` makes each scene easy to reference and edit.
8. `source_chapter` allows users to trace a scene back to the original chapter.

Do not overcomplicate the schema in the MVP. Prefer a stable, readable schema over an excessively detailed one.

---

## 9. Backend Guidelines

The backend should stay simple and reliable.

Expected MVP endpoints:

```text
GET /api/health
POST /api/chapters/parse
POST /api/scripts/generate
POST /api/scripts/validate
```

### `GET /api/health`

Purpose:

* Confirm that the backend is running.

Expected response:

```json
{
  "ok": true
}
```

### `POST /api/chapters/parse`

Purpose:

* Parse raw novel text into chapter objects.

Expected input:

```json
{
  "text": "第1章 ... 第2章 ... 第3章 ..."
}
```

Expected output:

```json
{
  "chapters": [
    {
      "index": 1,
      "title": "第1章",
      "content": "..."
    }
  ]
}
```

Validation:

* Must reject input with fewer than 3 chapters.
* Must return a clear error message.

### `POST /api/scripts/generate`

Purpose:

* Convert chapters into screenplay YAML.

Expected input:

```json
{
  "chapters": [
    {
      "index": 1,
      "title": "第1章",
      "content": "..."
    }
  ],
  "style": "faithful"
}
```

Expected output:

```json
{
  "yaml": "script:\n  title: ...",
  "validation": {
    "ok": true,
    "errors": []
  }
}
```

### `POST /api/scripts/validate`

Purpose:

* Validate generated or edited YAML.

Expected input:

```json
{
  "yaml": "script:\n  title: ..."
}
```

Expected output:

```json
{
  "ok": true,
  "errors": []
}
```

---

## 10. Chapter Parser Guidelines

The chapter parser should support common chapter headings:

```text
第1章
第一章
第 1 章
Chapter 1
CHAPTER 1
# 第一章
## 第一章
```

Each parsed chapter should include:

* `index`
* `title`
* `content`

Rules:

1. Preserve original chapter order.
2. Trim obvious leading/trailing whitespace.
3. Do not rewrite chapter content during parsing.
4. If fewer than 3 chapters are found, return a validation error.
5. Avoid complex NLP in the parser; keep it deterministic and testable.

---

## 11. AI Generation Guidelines

The AI generation step should not be a vague text generation call.

It should follow this pipeline when possible:

```text
chapters
-> identify characters and key events
-> split into scenes
-> convert narration into action
-> convert conversations into dialogue
-> output YAML
-> validate YAML
```

The prompt must instruct the model to:

1. Output YAML only.
2. Do not wrap output in Markdown code fences.
3. Do not add explanations before or after YAML.
4. Preserve source chapter references.
5. Generate scene-based screenplay structure.
6. Include characters, action, dialogue, summary, and notes.
7. Keep the adaptation faithful unless a style option says otherwise.
8. Avoid inventing major plot points not supported by the source text.
9. Keep dialogue natural and editable.
10. Prefer clarity and structure over literary flourish.

If the model output is invalid YAML:

1. Try a safe repair step if implemented.
2. Validate again.
3. If still invalid, return a clear error to the user.
4. Do not pretend invalid YAML is valid.

---

## 12. YAML Validation Guidelines

The validator should check at least:

* YAML can be parsed.
* Top-level `script` field exists.
* `script.title` exists.
* `script.source_chapters` is an array.
* `script.characters` is an array.
* `script.scenes` is an array.
* Each scene has `scene_id`.
* Each scene has `source_chapter`.
* Each scene has `location`.
* Each scene has `time`.
* Each scene has `characters`.
* Each scene has `summary`.
* Each scene has `action`.
* Each scene has `dialogue`.
* Each dialogue item has `speaker`.
* Each dialogue item has `line`.

Validation errors should be readable for users and developers.

Example error format:

```json
{
  "ok": false,
  "errors": [
    {
      "path": "script.scenes[0].dialogue[1].speaker",
      "message": "speaker is required"
    }
  ]
}
```

---

## 13. Frontend Guidelines

The frontend MVP should clearly show the conversion flow.

Required UI sections:

1. Novel input area
2. Chapter count / parse status
3. Generate button
4. Loading state
5. Error display
6. YAML output area
7. Readable screenplay preview
8. Copy YAML button
9. Download YAML button

Recommended layout:

```text
Left: novel input
Center/top: controls and status
Right: YAML output and screenplay preview
```

Do not build complex features too early.

Avoid in P0/P1:

* login page
* dashboard
* project management system
* complex drag-and-drop editor
* collaborative editing
* payment UI

---

## 14. Documentation Guidelines

The project must include a YAML Schema design document.

The document should explain:

1. What problem the schema solves.
2. Why YAML is used.
3. Why scenes are the core unit.
4. Why source chapter references are preserved.
5. Why characters are extracted globally.
6. Why action and dialogue are separated.
7. How the schema helps authors edit the result.
8. How the schema can support future features.
9. Validation rules.
10. A complete example.

Do not leave this document until the end. It is part of the competition requirement.

---

## 15. Testing Guidelines

Before completing a task, run relevant tests if possible.

General:

```bash
git diff --check
```

Backend examples:

```bash
npm test
npm run lint
npm run build
```

Frontend examples:

```bash
npm run lint
npm run build
```

If using a different stack, use the equivalent commands.

For YAML-related tasks, test with:

```text
docs/examples/script-example.yaml
```

For chapter parsing, test with at least:

1. Chinese numeric chapters, such as `第1章`
2. Chinese character chapters, such as `第一章`
3. English chapters, such as `Chapter 1`
4. Markdown headings, such as `# 第一章`
5. Input with fewer than 3 chapters

If tests cannot be run, explain why.

---

## 16. Security Rules

Never commit:

* API keys
* tokens
* credentials
* `.env`
* private keys
* real user data
* private documents
* production secrets

Use placeholders such as:

```text
OPENAI_API_KEY=your_api_key_here
```

Do not read or modify:

```text
~/.ssh
.env
.env.local
.env.production
```

Do not add code that uploads user content to unknown external services.

---

## 17. Demo Stability Rules

This project is for a competition, so demo stability is important.

Always keep a fallback path:

1. Example novel input file
2. Example YAML output file
3. Mock generation mode
4. Clear error messages
5. Ability to copy/download YAML

If live AI generation fails during demo, the app should still be able to show the example YAML and screenplay preview.

Do not remove demo fallback files.

Recommended fallback files:

```text
docs/examples/novel-example.md
docs/examples/script-example.yaml
```

---

## 18. Risk Classification

### Low Risk

Can be done by AI agents at night:

* documentation
* README updates
* task board
* YAML examples
* mock frontend page
* backend scaffold
* simple parser tests
* simple validation utilities

### Medium Risk

Should be reviewed carefully:

* YAML validator
* LLM prompt template
* AI generation endpoint
* frontend-backend integration
* YAML preview renderer

### High Risk

Prefer daytime human supervision:

* changing schema after implementation
* changing core API contracts
* adding database
* adding authentication
* large refactors
* changing the generation pipeline
* changing deployment or CI in a breaking way

---

## 19. Night Development Rules

When working unattended at night:

1. Only work on low-risk tasks unless explicitly assigned.
2. Do not modify core generation logic unless the task is very specific.
3. Do not change schema format unless the task is about schema.
4. Do not make destructive changes.
5. Do not delete files.
6. Do not run dangerous commands.
7. Do not install global packages unless explicitly required.
8. Do not use `sudo`.
9. Do not access files outside the repository.
10. Stop after completing the assigned task.
11. Submit a PR, but do not merge it.

Good night tasks:

```text
docs/yaml-schema-design
docs/task-board
docs/prompt-design
chore/backend-scaffold
feat/frontend-script-page
test/poc-smoke-script
```

Bad night tasks:

```text
feat/full-ai-agent-system
feat/database-user-system
feat/core-generation-refactor
feat/auth-system
feat/large-editor-rewrite
```

---

## 20. Final Response Format for AI Agents

After completing any task, report:

```text
Task:
Branch:
PR:
Files changed:

What changed:
- 

How to test:
- 

Risk:
- 

Notes / TODO:
- 
```

If the task could not be completed, report:

```text
Task:
Status: incomplete

What was done:
- 

Blocked by:
- 

Recommended next step:
- 
```

---

## 21. Important Product Reminder

The product is not just “AI writes a script”.

The product is:

```text
AI converts novel chapters into a structured, editable, validatable screenplay YAML draft.
```

The most important evaluation points are:

1. Can it process 3+ chapters?
2. Can it output valid YAML?
3. Is the YAML schema reasonable?
4. Is the schema design document clear?
5. Can authors edit and continue polishing the result?
6. Is the demo stable?

Prioritize these points over flashy but unstable features.
